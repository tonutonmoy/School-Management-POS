package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/school-management/pos/internal/config"
	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/payment"
	"github.com/school-management/pos/internal/repository"
)

type PaymentService struct {
	repos    *repository.Repositories
	audit    *AuditService
	cfg      *config.Config
	registry *payment.ProviderRegistry
	fees     *FeeService
	admission *AdmissionService
	notify   *NotificationService
	logger   *slog.Logger
}

func NewPaymentService(repos *repository.Repositories, audit *AuditService, cfg *config.Config, logger *slog.Logger) *PaymentService {
	return &PaymentService{
		repos: repos, audit: audit, cfg: cfg,
		registry: payment.NewRegistry(logger),
		logger:   logger,
	}
}

func (s *PaymentService) SetFees(f *FeeService)         { s.fees = f }
func (s *PaymentService) SetAdmission(a *AdmissionService) { s.admission = a }
func (s *PaymentService) SetNotifier(n *NotificationService) { s.notify = n }

func (s *PaymentService) ListGateways(ctx context.Context) ([]dto.PaymentGatewayResponse, error) {
	recs, err := s.repos.Payments.ListGateways(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.PaymentGatewayResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapGateway(&r))
	}
	return items, nil
}

func (s *PaymentService) UpdateGateway(ctx context.Context, id uuid.UUID, req dto.PaymentGatewayRequest, actorID uuid.UUID, ip string) (*dto.PaymentGatewayResponse, error) {
	rec, err := s.repos.Payments.UpdateGateway(ctx, id, repository.UpdateGatewayParams{
		Name: req.Name, IsActive: req.IsActive, IsSandbox: req.IsSandbox,
		APIKey: req.APIKey, APISecret: req.APISecret, MerchantID: req.MerchantID, StoreID: req.StoreID,
		CallbackURL: req.CallbackURL, SuccessURL: req.SuccessURL, FailURL: req.FailURL,
	})
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp := mapGateway(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityPaymentGateway, &id, ip, map[string]any{"slug": rec.Slug})
	return &resp, nil
}

func (s *PaymentService) InitiateStudentFeePayment(ctx context.Context, studentID uuid.UUID, parentID *uuid.UUID, req dto.InitiatePaymentRequest, ip string) (*dto.InitiatePaymentResponse, error) {
	if s.fees == nil {
		return nil, fmt.Errorf("fee service not configured")
	}
	bills, err := s.fees.PendingBillsForStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}
	if len(bills) == 0 {
		return nil, fmt.Errorf("%w: no pending bills", ErrValidation)
	}
	_, amount, billIDs, err := s.buildAllocations(bills, req.Amount, req.BillIDs)
	if err != nil {
		return nil, err
	}
	return s.initiate(ctx, req, model.GwPaymentStudentFee, studentID, studentID, parentID, nil, billIDs, amount, ip)
}

func (s *PaymentService) InitiateAdmissionPayment(ctx context.Context, appID uuid.UUID, req dto.InitiatePaymentRequest, ip string) (*dto.InitiatePaymentResponse, error) {
	rec, err := s.repos.Admissions.GetByID(ctx, appID)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	if rec.PaymentStatus == model.AdmissionPaymentPaid {
		return nil, fmt.Errorf("%w: admission fee already paid", ErrValidation)
	}
	amount := rec.AdmissionFeeAmount
	if req.Amount > 0 {
		amount = req.Amount
	}
	if amount <= 0 {
		return nil, fmt.Errorf("%w: invalid admission fee amount", ErrValidation)
	}
	return s.initiate(ctx, req, model.GwPaymentAdmission, appID, uuid.Nil, nil, &appID, nil, amount, ip)
}

func (s *PaymentService) initiate(ctx context.Context, req dto.InitiatePaymentRequest, paymentType string, referenceID, studentID uuid.UUID, parentID, appID *uuid.UUID, billIDs []uuid.UUID, amount float64, ip string) (*dto.InitiatePaymentResponse, error) {
	gw, err := s.repos.Payments.GetGatewayBySlug(ctx, req.GatewaySlug)
	if err != nil || gw == nil {
		return nil, fmt.Errorf("%w: gateway not found", ErrNotFound)
	}
	if !gw.IsActive {
		return nil, fmt.Errorf("%w: gateway is inactive", ErrValidation)
	}
	idemKey := req.IdempotencyKey
	if idemKey == "" {
		idemKey, _ = randomKey()
	}
	if existing, _ := s.repos.Payments.GetTransactionByIdempotency(ctx, idemKey); existing != nil {
		return &dto.InitiatePaymentResponse{
			TransactionID: existing.ID, TransactionRef: existing.TransactionRef,
			Amount: existing.Amount, Gateway: gw.Slug, Status: existing.Status,
		}, nil
	}
	ref, err := s.repos.Payments.NextTransactionRef(ctx)
	if err != nil {
		return nil, err
	}
	provider := s.registry.Get(gw.Slug)
	if provider == nil {
		return nil, fmt.Errorf("%w: unsupported gateway", ErrValidation)
	}
	baseURL := s.cfg.App.URL
	cfg := payment.GatewayConfig{
		IsSandbox: gw.IsSandbox, APIKey: gw.APIKey, APISecret: gw.APISecret,
		MerchantID: gw.MerchantID, StoreID: gw.StoreID, BaseURL: baseURL,
		CallbackURL: baseURL + gw.CallbackURL, SuccessURL: baseURL + gw.SuccessURL, FailURL: baseURL + gw.FailURL,
	}
	result, err := provider.Initiate(ctx, cfg, payment.InitiateRequest{
		TransactionRef: ref, Amount: amount, Currency: "BDT", Description: paymentType,
	})
	if err != nil {
		return nil, err
	}
	var txRec *repository.GatewayTransactionRecord
	err = s.repos.Payments.WithTx(ctx, func(tx pgx.Tx) error {
		var sid *uuid.UUID
		if studentID != uuid.Nil {
			sid = &studentID
		}
		p := repository.CreateGatewayTxParams{
			GatewayID: gw.ID, TransactionRef: ref, IdempotencyKey: idemKey,
			GatewayRef: result.GatewayRef, PaymentType: paymentType, ReferenceID: referenceID,
			StudentID: sid, ParentID: parentID, AdmissionApplicationID: appID,
			BillIDs: billIDs, Amount: amount, Currency: "BDT", Status: model.GwStatusPending,
			GatewayResponse: result.RawResponse, IPAddress: ip,
		}
		rec, err := s.repos.Payments.CreateTransaction(ctx, tx, p)
		if err != nil {
			return err
		}
		txRec = rec
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.audit.Log(ctx, nil, model.ActionCreate, model.EntityGatewayTransaction, &txRec.ID, ip, map[string]any{
		"ref": ref, "amount": amount, "gateway": gw.Slug, "type": paymentType,
	})
	return &dto.InitiatePaymentResponse{
		TransactionID: txRec.ID, TransactionRef: ref, RedirectURL: result.RedirectURL,
		Amount: amount, Gateway: gw.Slug, Status: model.GwStatusPending,
	}, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, slug string, payload dto.WebhookPayload, ip string) error {
	gw, err := s.repos.Payments.GetGatewayBySlug(ctx, slug)
	if err != nil || gw == nil {
		return ErrNotFound
	}
	txRec, err := s.repos.Payments.GetTransactionByRef(ctx, payload.TransactionRef)
	if err != nil || txRec == nil {
		return ErrNotFound
	}
	if txRec.Status == model.GwStatusCompleted {
		return nil
	}
	provider := s.registry.Get(slug)
	cfg := payment.GatewayConfig{IsSandbox: gw.IsSandbox, APISecret: gw.APISecret}
	verified, err := provider.VerifyWebhook(ctx, cfg, payment.WebhookData{
		TransactionRef: payload.TransactionRef, GatewayRef: payload.GatewayRef,
		Amount: payload.Amount, Status: payload.Status, Signature: payload.Signature, Raw: payload.RawBody,
	})
	if err != nil {
		return err
	}
	if !verified {
		return s.markFailed(ctx, txRec.ID, "signature verification failed", ip)
	}
	if payload.Status == "failed" || payload.Status == "cancelled" {
		return s.markFailed(ctx, txRec.ID, "gateway reported failure", ip)
	}
	return s.CompleteTransaction(ctx, txRec.TransactionRef, payload.GatewayRef, true, payload.RawBody, ip)
}

func (s *PaymentService) CompleteTransaction(ctx context.Context, ref, gatewayRef string, verified bool, gwResp map[string]any, ip string) error {
	var completed *repository.GatewayTransactionRecord
	err := s.repos.Payments.WithTx(ctx, func(tx pgx.Tx) error {
		rec, err := s.repos.Payments.GetTransactionByRef(ctx, ref)
		if err != nil || rec == nil {
			return ErrNotFound
		}
		locked, err := s.repos.Payments.LockTransactionForUpdate(ctx, tx, rec.ID)
		if err != nil || locked == nil {
			return ErrNotFound
		}
		if locked.Status == model.GwStatusCompleted {
			completed = locked
			return nil
		}
		if locked.Status == model.GwStatusFailed || locked.Status == model.GwStatusRefunded {
			return fmt.Errorf("%w: transaction already finalized", ErrValidation)
		}
		now := time.Now()
		if err := s.repos.Payments.UpdateTransactionStatus(ctx, tx, locked.ID, repository.UpdateGatewayTxParams{
			Status: model.GwStatusProcessing, GatewayRef: gatewayRef, SignatureVerified: verified, GatewayResponse: gwResp,
		}); err != nil {
			return err
		}
		var paymentID *uuid.UUID
		switch locked.PaymentType {
		case model.GwPaymentStudentFee:
			pid, feeErr := s.completeStudentFee(ctx, locked, ip)
			if feeErr != nil {
				_ = s.repos.Payments.UpdateTransactionStatus(ctx, tx, locked.ID, repository.UpdateGatewayTxParams{
					Status: model.GwStatusFailed, ErrorMessage: feeErr.Error(), CompletedAt: &now,
				})
				completed, _ = s.repos.Payments.GetTransaction(ctx, locked.ID)
				return nil
			}
			paymentID = &pid
		case model.GwPaymentAdmission:
			if admErr := s.completeAdmissionFee(ctx, locked, gatewayRef, ip); admErr != nil {
				_ = s.repos.Payments.UpdateTransactionStatus(ctx, tx, locked.ID, repository.UpdateGatewayTxParams{
					Status: model.GwStatusFailed, ErrorMessage: admErr.Error(), CompletedAt: &now,
				})
				completed, _ = s.repos.Payments.GetTransaction(ctx, locked.ID)
				return nil
			}
		default:
			return fmt.Errorf("%w: unknown payment type", ErrValidation)
		}
		if err := s.repos.Payments.UpdateTransactionStatus(ctx, tx, locked.ID, repository.UpdateGatewayTxParams{
			Status: model.GwStatusCompleted, GatewayRef: gatewayRef, PaymentID: paymentID,
			SignatureVerified: verified, GatewayResponse: gwResp, CompletedAt: &now,
		}); err != nil {
			return err
		}
		completed, _ = s.repos.Payments.GetTransaction(ctx, locked.ID)
		return nil
	})
	if err != nil {
		return err
	}
	if completed != nil {
		s.audit.Log(ctx, nil, model.ActionUpdate, model.EntityGatewayTransaction, &completed.ID, ip, map[string]any{"status": completed.Status, "ref": ref})
		if completed.Status == model.GwStatusFailed && completed.StudentID != nil && s.notify != nil {
			go s.notify.OnPaymentFailed(context.Background(), *completed.StudentID, completed.Amount, completed.TransactionRef)
		}
	}
	return nil
}

func (s *PaymentService) completeStudentFee(ctx context.Context, txRec *repository.GatewayTransactionRecord, ip string) (uuid.UUID, error) {
	if s.fees == nil || txRec.StudentID == nil {
		return uuid.Nil, fmt.Errorf("invalid student fee transaction")
	}
	if txRec.PaymentID != nil {
		return *txRec.PaymentID, nil
	}
	method := gatewayToPaymentMethod(txRec.GatewaySlug)
	bills, _ := s.fees.PendingBillsForStudent(ctx, *txRec.StudentID)
	var allocations []dto.PaymentAllocationInput
	if len(txRec.BillIDs) > 0 {
		var err error
		allocations, _, _, err = s.buildAllocations(bills, txRec.Amount, txRec.BillIDs)
		if err != nil {
			return uuid.Nil, err
		}
	} else {
		remaining := txRec.Amount
		for _, b := range bills {
			if remaining <= 0 {
				break
			}
			due := b.TotalAmount - b.PaidAmount
			if due <= 0 {
				continue
			}
			alloc := math.Min(due, remaining)
			allocations = append(allocations, dto.PaymentAllocationInput{BillID: b.ID, Amount: alloc})
			remaining -= alloc
		}
	}
	resp, err := s.fees.CollectPayment(ctx, dto.PaymentRequest{
		StudentID: *txRec.StudentID, Amount: txRec.Amount, PaymentMethod: method,
		CollectionDate: time.Now(), Remarks: "Online payment " + txRec.TransactionRef,
	}, allocations, uuid.Nil, ip)
	if err != nil {
		return uuid.Nil, err
	}
	return resp.ID, nil
}

func (s *PaymentService) completeAdmissionFee(ctx context.Context, txRec *repository.GatewayTransactionRecord, gatewayRef, ip string) error {
	if s.admission == nil || txRec.AdmissionApplicationID == nil {
		return fmt.Errorf("invalid admission transaction")
	}
	_, err := s.admission.RecordPayment(ctx, *txRec.AdmissionApplicationID, dto.AdmissionPaymentRequest{
		Amount: txRec.Amount, PaymentReference: gatewayRef,
	}, uuid.Nil, ip)
	if err != nil {
		return err
	}
	if s.notify != nil {
		app, _ := s.repos.Admissions.GetByID(ctx, *txRec.AdmissionApplicationID)
		if app != nil && app.Email != "" {
			go s.notify.OnAdmissionPaymentReceived(context.Background(), app.Email, app.AppNumber, txRec.Amount, app.ReceiptNumber)
		}
	}
	return nil
}

func (s *PaymentService) markFailed(ctx context.Context, txID uuid.UUID, msg, ip string) error {
	err := s.repos.Payments.WithTx(ctx, func(tx pgx.Tx) error {
		return s.repos.Payments.UpdateTransactionStatus(ctx, tx, txID, repository.UpdateGatewayTxParams{
			Status: model.GwStatusFailed, ErrorMessage: msg, CompletedAt: ptrTime(time.Now()),
		})
	})
	if err != nil {
		return err
	}
	rec, _ := s.repos.Payments.GetTransaction(ctx, txID)
	if rec != nil && rec.StudentID != nil && s.notify != nil {
		go s.notify.OnPaymentFailed(context.Background(), *rec.StudentID, rec.Amount, rec.TransactionRef)
	}
	s.audit.Log(ctx, nil, model.ActionUpdate, model.EntityGatewayTransaction, &txID, ip, map[string]any{"status": "failed", "error": msg})
	return nil
}

func (s *PaymentService) DashboardStats(ctx context.Context) (*dto.PaymentDashboardStats, error) {
	rec, err := s.repos.Payments.DashboardStats(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.PaymentDashboardStats{
		TodayCollection: rec.TodayCollection, GatewayCollection: rec.GatewayCollection,
		FailedPayments: rec.FailedPayments, PendingPayments: rec.PendingPayments,
		TodayTransactions: rec.TodayTransactions,
	}, nil
}

func (s *PaymentService) SearchTransactions(ctx context.Context, f dto.PaymentReportFilter) (*dto.PaginatedGatewayTransactions, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 25
	}
	rf := repository.GatewayTxSearchParams{
		Query: f.Query, Status: f.Status, GatewaySlug: f.GatewaySlug, PaymentType: f.PaymentType,
		From: f.From, To: f.To, Limit: int32(f.PageSize), Offset: int32((f.Page - 1) * f.PageSize),
	}
	total, err := s.repos.Payments.CountTransactions(ctx, rf)
	if err != nil {
		return nil, err
	}
	recs, err := s.repos.Payments.SearchTransactions(ctx, rf)
	if err != nil {
		return nil, err
	}
	items := make([]dto.GatewayTransactionResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapGatewayTx(&r))
	}
	pages := int(total) / f.PageSize
	if int(total)%f.PageSize > 0 {
		pages++
	}
	return &dto.PaginatedGatewayTransactions{Items: items, Total: total, Page: f.Page, PageSize: f.PageSize, TotalPages: pages}, nil
}

func (s *PaymentService) StudentPaymentHistory(ctx context.Context, studentID uuid.UUID) ([]dto.GatewayTransactionResponse, error) {
	recs, err := s.repos.Payments.ListStudentTransactions(ctx, studentID, 20)
	if err != nil {
		return nil, err
	}
	items := make([]dto.GatewayTransactionResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapGatewayTx(&r))
	}
	return items, nil
}

func (s *PaymentService) GatewayCollectionReport(ctx context.Context, from, to time.Time) ([]dto.GatewayCollectionReport, error) {
	recs, err := s.repos.Payments.GatewayCollectionReport(ctx, from, to)
	if err != nil {
		return nil, err
	}
	items := make([]dto.GatewayCollectionReport, 0, len(recs))
	for _, r := range recs {
		items = append(items, dto.GatewayCollectionReport{
			GatewaySlug: r.GatewaySlug, GatewayName: r.GatewayName, Count: r.Count, TotalAmount: r.TotalAmount,
		})
	}
	return items, nil
}

func (s *PaymentService) RequestRefund(ctx context.Context, req dto.RefundRequest, actorID uuid.UUID, ip string) (*dto.RefundResponse, error) {
	pay, err := s.repos.Fees.GetPayment(ctx, req.PaymentID)
	if err != nil || pay == nil {
		return nil, ErrNotFound
	}
	if pay.Status != model.PaymentCompleted {
		return nil, fmt.Errorf("%w: only completed payments can be refunded", ErrValidation)
	}
	if req.Amount > pay.Amount {
		return nil, fmt.Errorf("%w: refund exceeds payment amount", ErrValidation)
	}
	var gwTxID *uuid.UUID
	if gwTx, _ := s.repos.Payments.GetTransactionByPaymentID(ctx, req.PaymentID); gwTx != nil {
		gwTxID = &gwTx.ID
	}
	rec, err := s.repos.Payments.CreateRefund(ctx, repository.CreateRefundParams{
		GatewayTransactionID: gwTxID, PaymentID: req.PaymentID, Amount: req.Amount,
		Reason: req.Reason, RequestedBy: &actorID,
	})
	if err != nil {
		return nil, err
	}
	resp := mapRefund(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityPaymentRefund, &rec.ID, ip, nil)
	return &resp, nil
}

func (s *PaymentService) ApproveRefund(ctx context.Context, id uuid.UUID, approve bool, actorID uuid.UUID, ip string) error {
	rec, err := s.repos.Payments.GetRefund(ctx, id)
	if err != nil || rec == nil {
		return ErrNotFound
	}
	if rec.Status != model.RefundRequested {
		return fmt.Errorf("%w: refund not in requested state", ErrValidation)
	}
	if !approve {
		return s.repos.Payments.UpdateRefundStatus(ctx, id, model.RefundRejected, &actorID)
	}
	if err := s.repos.Payments.UpdateRefundStatus(ctx, id, model.RefundApproved, &actorID); err != nil {
		return err
	}
	if s.fees != nil {
		if err := s.fees.RefundPayment(ctx, rec.PaymentID, actorID, ip); err != nil {
			return err
		}
	}
	if err := s.repos.Payments.UpdateRefundStatus(ctx, id, model.RefundProcessed, &actorID); err != nil {
		return err
	}
	pay, _ := s.repos.Fees.GetPayment(ctx, rec.PaymentID)
	if pay != nil && s.notify != nil {
		go s.notify.OnRefundProcessed(context.Background(), pay.StudentID, rec.Amount, rec.ID)
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityPaymentRefund, &id, ip, map[string]any{"status": "processed"})
	return nil
}

func (s *PaymentService) ListRefunds(ctx context.Context, status string, page, pageSize int) (*dto.PaginatedRefunds, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}
	total, err := s.repos.Payments.CountRefunds(ctx, status)
	if err != nil {
		return nil, err
	}
	recs, err := s.repos.Payments.SearchRefunds(ctx, status, int32(pageSize), int32((page-1)*pageSize))
	if err != nil {
		return nil, err
	}
	items := make([]dto.RefundResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapRefund(&r))
	}
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}
	return &dto.PaginatedRefunds{Items: items, Total: total, Page: page, PageSize: pageSize, TotalPages: pages}, nil
}

func (s *PaymentService) ParentPayNowData(ctx context.Context, studentID uuid.UUID) (*dto.ParentPayNowData, error) {
	if s.fees == nil {
		return nil, fmt.Errorf("fee service not configured")
	}
	st, err := s.repos.Students.GetByID(ctx, studentID)
	if err != nil || st == nil {
		return nil, ErrNotFound
	}
	bills, _ := s.fees.PendingBillsForStudent(ctx, studentID)
	gateways, _ := s.repos.Payments.ListActiveGateways(ctx)
	gwResp := make([]dto.PaymentGatewayResponse, 0, len(gateways))
	for _, g := range gateways {
		gwResp = append(gwResp, mapGateway(&g))
	}
	var due float64
	for _, b := range bills {
		due += b.TotalAmount - b.PaidAmount
	}
	return &dto.ParentPayNowData{
		StudentID: studentID, StudentName: st.FirstName + " " + st.LastName,
		CurrentDue: due, Gateways: gwResp, Bills: bills,
	}, nil
}

func (s *PaymentService) buildAllocations(bills []dto.StudentBillResponse, amount float64, billIDs []uuid.UUID) ([]dto.PaymentAllocationInput, float64, []uuid.UUID, error) {
	if amount <= 0 {
		var total float64
		ids := make([]uuid.UUID, 0, len(bills))
		allocs := make([]dto.PaymentAllocationInput, 0, len(bills))
		for _, b := range bills {
			due := b.TotalAmount - b.PaidAmount
			if due <= 0 {
				continue
			}
			allocs = append(allocs, dto.PaymentAllocationInput{BillID: b.ID, Amount: due})
			ids = append(ids, b.ID)
			total += due
		}
		return allocs, total, ids, nil
	}
	if len(billIDs) == 0 {
		remaining := amount
		allocs := make([]dto.PaymentAllocationInput, 0)
		ids := make([]uuid.UUID, 0)
		for _, b := range bills {
			if remaining <= 0 {
				break
			}
			due := b.TotalAmount - b.PaidAmount
			if due <= 0 {
				continue
			}
			alloc := math.Min(due, remaining)
			allocs = append(allocs, dto.PaymentAllocationInput{BillID: b.ID, Amount: alloc})
			ids = append(ids, b.ID)
			remaining -= alloc
		}
		if len(allocs) == 0 {
			return nil, 0, nil, fmt.Errorf("%w: no bills to allocate", ErrValidation)
		}
		return allocs, amount, ids, nil
	}
	billMap := map[uuid.UUID]dto.StudentBillResponse{}
	for _, b := range bills {
		billMap[b.ID] = b
	}
	var allocSum float64
	allocs := make([]dto.PaymentAllocationInput, 0, len(billIDs))
	for _, id := range billIDs {
		b, ok := billMap[id]
		if !ok {
			return nil, 0, nil, fmt.Errorf("%w: invalid bill", ErrValidation)
		}
		due := b.TotalAmount - b.PaidAmount
		alloc := math.Min(due, amount-allocSum)
		if alloc <= 0 {
			continue
		}
		allocs = append(allocs, dto.PaymentAllocationInput{BillID: id, Amount: alloc})
		allocSum += alloc
	}
	if math.Abs(allocSum-amount) > 0.01 {
		return nil, 0, nil, fmt.Errorf("%w: partial allocation must match amount", ErrValidation)
	}
	return allocs, amount, billIDs, nil
}

func mapGateway(r *repository.PaymentGatewayRecord) dto.PaymentGatewayResponse {
	return dto.PaymentGatewayResponse{
		ID: r.ID, Name: r.Name, Slug: r.Slug, IsActive: r.IsActive, IsSandbox: r.IsSandbox,
		MerchantID: r.MerchantID, StoreID: r.StoreID, CallbackURL: r.CallbackURL,
		SuccessURL: r.SuccessURL, FailURL: r.FailURL,
		HasCredentials: r.APIKey != "" && r.APISecret != "", UpdatedAt: r.UpdatedAt,
	}
}

func mapGatewayTx(r *repository.GatewayTransactionRecord) dto.GatewayTransactionResponse {
	return dto.GatewayTransactionResponse{
		ID: r.ID, TransactionRef: r.TransactionRef, GatewayRef: r.GatewayRef,
		GatewayName: r.GatewayName, GatewaySlug: r.GatewaySlug, PaymentType: r.PaymentType,
		Amount: r.Amount, Status: r.Status, StudentID: r.StudentID, StudentName: r.StudentName,
		PaymentID: r.PaymentID, ReceiptNumber: r.ReceiptNumber, SignatureVerified: r.SignatureVerified,
		ErrorMessage: r.ErrorMessage, CreatedAt: r.CreatedAt, CompletedAt: r.CompletedAt,
	}
}

func mapRefund(r *repository.PaymentRefundRecord) dto.RefundResponse {
	return dto.RefundResponse{
		ID: r.ID, PaymentID: r.PaymentID, Amount: r.Amount, Status: r.Status, Reason: r.Reason,
		RequestedByName: r.RequestedByName, CreatedAt: r.CreatedAt, ProcessedAt: r.ProcessedAt,
	}
}

func gatewayToPaymentMethod(slug string) string {
	switch slug {
	case model.GatewayBkash:
		return "bkash"
	case model.GatewayNagad:
		return "nagad"
	default:
		return "card"
	}
}

func randomKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func ptrTime(t time.Time) *time.Time { return &t }
