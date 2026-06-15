package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/repository"
)

type FeeService struct {
	repos  *repository.Repositories
	audit  *AuditService
	notify *NotificationService
}

func NewFeeService(repos *repository.Repositories, audit *AuditService) *FeeService {
	return &FeeService{repos: repos, audit: audit}
}

func (s *FeeService) SetNotifier(n *NotificationService) { s.notify = n }

// --- Fee Types ---

func (s *FeeService) CreateFeeType(ctx context.Context, req dto.FeeTypeRequest, actorID uuid.UUID, ip string) (*dto.FeeTypeResponse, error) {
	rec, err := s.repos.Fees.CreateFeeType(ctx, req.Name, req.Slug, req.Description, req.IsActive)
	if err != nil {
		return nil, err
	}
	resp := mapFeeType(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityFeeType, &rec.IDUUID, ip, nil)
	return &resp, nil
}

func (s *FeeService) UpdateFeeType(ctx context.Context, id uuid.UUID, req dto.FeeTypeRequest, actorID uuid.UUID, ip string) (*dto.FeeTypeResponse, error) {
	rec, err := s.repos.Fees.UpdateFeeType(ctx, id, req.Name, req.Slug, req.Description, req.IsActive)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp := mapFeeType(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityFeeType, &id, ip, nil)
	return &resp, nil
}

func (s *FeeService) DeleteFeeType(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.Fees.SoftDeleteFeeType(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityFeeType, &id, ip, nil)
	return nil
}

func (s *FeeService) ListFeeTypes(ctx context.Context, activeOnly bool) ([]dto.FeeTypeResponse, error) {
	recs, err := s.repos.Fees.ListFeeTypes(ctx, activeOnly)
	if err != nil {
		return nil, err
	}
	items := make([]dto.FeeTypeResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapFeeType(&r))
	}
	return items, nil
}

func (s *FeeService) GetFeeType(ctx context.Context, id uuid.UUID) (*dto.FeeTypeResponse, error) {
	rec, err := s.repos.Fees.GetFeeType(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp := mapFeeType(rec)
	return &resp, nil
}

// --- Fee Structures ---

func (s *FeeService) CreateFeeStructure(ctx context.Context, req dto.FeeStructureRequest, actorID uuid.UUID, ip string) (*dto.FeeStructureResponse, error) {
	p := mapStructureReq(req)
	rec, err := s.repos.Fees.CreateFeeStructure(ctx, p)
	if err != nil {
		return nil, err
	}
	resp := mapFeeStructure(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityFeeStructure, &rec.ID, ip, nil)
	return &resp, nil
}

func (s *FeeService) UpdateFeeStructure(ctx context.Context, id uuid.UUID, req dto.FeeStructureRequest, actorID uuid.UUID, ip string) (*dto.FeeStructureResponse, error) {
	p := mapStructureReq(req)
	rec, err := s.repos.Fees.UpdateFeeStructure(ctx, id, p)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp := mapFeeStructure(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityFeeStructure, &id, ip, nil)
	return &resp, nil
}

func (s *FeeService) DeleteFeeStructure(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.Fees.SoftDeleteFeeStructure(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityFeeStructure, &id, ip, nil)
	return nil
}

func (s *FeeService) ListFeeStructures(ctx context.Context, sessionID, classID *uuid.UUID) ([]dto.FeeStructureResponse, error) {
	recs, err := s.repos.Fees.ListFeeStructures(ctx, repository.FeeStructureFilter{SessionID: sessionID, ClassID: classID})
	if err != nil {
		return nil, err
	}
	items := make([]dto.FeeStructureResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapFeeStructure(&r))
	}
	return items, nil
}

// --- Discounts ---

func (s *FeeService) CreateDiscount(ctx context.Context, req dto.StudentDiscountRequest, actorID uuid.UUID, ip string) (*dto.StudentDiscountResponse, error) {
	rec, err := s.repos.Fees.CreateDiscount(ctx, repository.DiscountParams{
		StudentID: req.StudentID, SessionID: req.SessionID, DiscountType: req.DiscountType,
		DiscountValue: req.DiscountValue, Reason: req.Reason, Description: req.Description, IsActive: req.IsActive,
	})
	if err != nil {
		return nil, err
	}
	resp := mapDiscount(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityStudentDiscount, &rec.ID, ip, nil)
	return &resp, nil
}

func (s *FeeService) ListDiscounts(ctx context.Context, studentID, sessionID *uuid.UUID) ([]dto.StudentDiscountResponse, error) {
	recs, err := s.repos.Fees.ListDiscounts(ctx, studentID, sessionID)
	if err != nil {
		return nil, err
	}
	items := make([]dto.StudentDiscountResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapDiscount(&r))
	}
	return items, nil
}

// --- Bill Generation ---

func (s *FeeService) GenerateBills(ctx context.Context, req dto.GenerateBillsRequest, actorID uuid.UUID, ip string) (int, error) {
	_ = s.repos.Fees.MarkOverdueBills(ctx, time.Now())
	students, err := s.loadBillTargets(ctx, req)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, st := range students {
		if err := s.generateStudentBill(ctx, st, req, actorID); err != nil {
			continue
		}
		count++
	}
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityStudentBill, nil, ip, map[string]any{
		"period": req.BillPeriod, "count": count,
	})
	return count, nil
}

func (s *FeeService) loadBillTargets(ctx context.Context, req dto.GenerateBillsRequest) ([]repository.StudentRecord, error) {
	if req.StudentID != uuid.Nil {
		st, err := s.repos.Students.GetByID(ctx, req.StudentID)
		if err != nil || st == nil {
			return nil, ErrNotFound
		}
		return []repository.StudentRecord{*st}, nil
	}
	var classID, sessionID *uuid.UUID
	if req.ClassID != uuid.Nil {
		classID = &req.ClassID
	}
	if req.SessionID != uuid.Nil {
		sessionID = &req.SessionID
	}
	return s.repos.Students.ListForReport(ctx, classID, sessionID, model.StudentStatusActive)
}

func (s *FeeService) generateStudentBill(ctx context.Context, st repository.StudentRecord, req dto.GenerateBillsRequest, actorID uuid.UUID) error {
	existing, _ := s.repos.Fees.GetBillByStudentPeriod(ctx, st.ID, req.BillPeriod)
	if existing != nil {
		if !req.Regenerate {
			return nil
		}
		_ = s.repos.Fees.CancelBill(ctx, existing.ID)
	}
	structures, err := s.repos.Fees.ListApplicableStructures(ctx, st.SessionID, st.ClassID, st.SectionID, model.FreqMonthly)
	if err != nil || len(structures) == 0 {
		return fmt.Errorf("no fee structures")
	}
	var subtotal float64
	var maxDueDay int = 10
	for _, fs := range structures {
		subtotal += fs.Amount
		if fs.DueDay > maxDueDay {
			maxDueDay = fs.DueDay
		}
	}
	discountAmt := s.calcDiscount(ctx, st.ID, st.SessionID, subtotal)
	total := subtotal - discountAmt
	if total < 0 {
		total = 0
	}
	inv, err := s.repos.Fees.NextFinanceNumber(ctx, model.SeqInvoice, time.Now().Year())
	if err != nil {
		return err
	}
	dueDate := parseBillDueDate(req.BillPeriod, maxDueDay)
	bill, err := s.repos.Fees.CreateBill(ctx, nil, repository.CreateBillParams{
		InvoiceNumber: inv, StudentID: st.ID, SessionID: st.SessionID,
		ClassID: st.ClassID, SectionID: st.SectionID, BillPeriod: req.BillPeriod,
		DueDate: dueDate, Subtotal: subtotal, DiscountAmount: discountAmt,
		TotalAmount: total, Status: model.BillStatusPending,
	})
	if err != nil {
		return err
	}
	for _, fs := range structures {
		_ = s.repos.Fees.CreateBillItem(ctx, nil, bill.ID, fs.FeeTypeID, &fs.ID, fs.FeeTypeName+" ("+fs.Frequency+")", fs.Amount)
	}
	return nil
}

func (s *FeeService) calcDiscount(ctx context.Context, studentID, sessionID uuid.UUID, subtotal float64) float64 {
	discounts, _ := s.repos.Fees.GetActiveDiscounts(ctx, studentID, sessionID)
	var total float64
	for _, d := range discounts {
		if !d.IsActive {
			continue
		}
		if d.DiscountType == model.DiscountPercentage {
			total += subtotal * d.DiscountValue / 100
		} else {
			total += d.DiscountValue
		}
	}
	if total > subtotal {
		return subtotal
	}
	return total
}

func parseBillDueDate(period string, dueDay int) time.Time {
	// period format: YYYY-MM
	t, err := time.Parse("2006-01", period)
	if err != nil {
		return time.Now().AddDate(0, 0, dueDay)
	}
	return time.Date(t.Year(), t.Month(), dueDay, 0, 0, 0, 0, time.Local)
}

// --- Bills ---

func (s *FeeService) ListBills(ctx context.Context, f dto.BillSearchFilter) (*dto.PaginatedBills, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 {
		f.PerPage = 20
	}
	params := mapBillFilter(f)
	total, err := s.repos.Fees.CountBills(ctx, params)
	if err != nil {
		return nil, err
	}
	recs, err := s.repos.Fees.SearchBills(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.StudentBillResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, s.mapBill(ctx, &r))
	}
	return &dto.PaginatedBills{
		Items: items, Total: total, Page: f.Page, PerPage: f.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(f.PerPage))),
	}, nil
}

func (s *FeeService) GetBill(ctx context.Context, id uuid.UUID) (*dto.StudentBillResponse, error) {
	rec, err := s.repos.Fees.GetBill(ctx, id)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	resp := s.mapBill(ctx, rec)
	return &resp, nil
}

func (s *FeeService) mapBill(ctx context.Context, r *repository.BillRecord) dto.StudentBillResponse {
	items, _ := s.repos.Fees.ListBillItems(ctx, r.ID)
	billItems := make([]dto.BillItemResponse, 0, len(items))
	for _, it := range items {
		billItems = append(billItems, dto.BillItemResponse{
			ID: it.ID, FeeTypeName: it.FeeTypeName, Description: it.Description, Amount: it.Amount,
		})
	}
	due := r.TotalAmount - r.PaidAmount
	if due < 0 {
		due = 0
	}
	return dto.StudentBillResponse{
		ID: r.ID, InvoiceNumber: r.InvoiceNumber, StudentID: r.StudentID, StudentName: r.StudentName,
		AdmissionNo: r.AdmissionNo, ClassName: r.ClassName, SectionName: r.SectionName,
		BillPeriod: r.BillPeriod, DueDate: r.DueDate, Subtotal: r.Subtotal,
		DiscountAmount: r.DiscountAmount, TotalAmount: r.TotalAmount, PaidAmount: r.PaidAmount,
		DueAmount: due, Status: r.Status, Items: billItems, GeneratedAt: r.GeneratedAt,
	}
}

// --- Payment Collection (transaction-safe) ---

func (s *FeeService) CollectPayment(ctx context.Context, req dto.PaymentRequest, allocations []dto.PaymentAllocationInput, actorID uuid.UUID, ip string) (*dto.PaymentResponse, error) {
	if len(allocations) == 0 {
		return nil, fmt.Errorf("%w: at least one bill allocation required", ErrValidation)
	}
	var allocSum float64
	for _, a := range allocations {
		allocSum += a.Amount
	}
	if math.Abs(allocSum-req.Amount) > 0.01 {
		return nil, fmt.Errorf("%w: allocation total must match payment amount", ErrValidation)
	}
	year := time.Now().Year()
	payNum, err := s.repos.Fees.NextFinanceNumber(ctx, model.SeqPayment, year)
	if err != nil {
		return nil, err
	}
	rcpNum, err := s.repos.Fees.NextFinanceNumber(ctx, model.SeqReceipt, year)
	if err != nil {
		return nil, err
	}
	qrToken, err := generateQRToken()
	if err != nil {
		return nil, err
	}
	var payment *repository.PaymentRecord
	err = s.repos.Fees.WithTx(ctx, func(tx pgx.Tx) error {
		p, err := s.repos.Fees.CreatePayment(ctx, tx, repository.CreatePaymentParams{
			PaymentNumber: payNum, StudentID: req.StudentID, Amount: req.Amount,
			PaymentMethod: req.PaymentMethod, CollectedBy: actorID,
			CollectionDate: req.CollectionDate, Remarks: req.Remarks,
		})
		if err != nil {
			return err
		}
		payment = p
		for _, a := range allocations {
			bill, err := s.repos.Fees.GetBill(ctx, a.BillID)
			if err != nil || bill == nil {
				return ErrNotFound
			}
			due := bill.TotalAmount - bill.PaidAmount
			if a.Amount > due+0.01 {
				return fmt.Errorf("%w: allocation exceeds bill due", ErrValidation)
			}
			if err := s.repos.Fees.CreateAllocation(ctx, tx, p.ID, a.BillID, a.Amount); err != nil {
				return err
			}
			if err := s.repos.Fees.UpdateBillPayment(ctx, tx, a.BillID, a.Amount); err != nil {
				return err
			}
		}
		_, err = s.repos.Fees.CreateReceipt(ctx, tx, repository.CreateReceiptParams{
			ReceiptNumber: rcpNum, PaymentID: p.ID, StudentID: req.StudentID,
			TotalAmount: req.Amount, QRToken: qrToken, IssuedBy: actorID,
		})
		if err != nil {
			return err
		}
		journalAllocs := make([]repository.FeeAllocationJournal, len(allocations))
		for i, a := range allocations {
			journalAllocs[i] = repository.FeeAllocationJournal{BillID: a.BillID, Amount: a.Amount}
		}
		return s.repos.Accounting.PostFeePaymentJournal(ctx, tx, repository.FeePaymentJournalParams{
			PaymentID: p.ID, Amount: req.Amount, PaymentMethod: req.PaymentMethod,
			CollectionDate: req.CollectionDate, Allocations: journalAllocs, ActorID: actorID,
		})
	})
	if err != nil {
		return nil, err
	}
	resp := mapPayment(payment)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityPayment, &payment.ID, ip, map[string]any{
		"amount": req.Amount, "method": req.PaymentMethod,
	})
	if s.notify != nil {
		go s.notify.OnPaymentReceived(context.Background(), req.StudentID, payment.ID, req.Amount, payNum)
	}
	return &resp, nil
}

func (s *FeeService) RefundPayment(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	err := s.repos.Fees.WithTx(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE payments SET status='refunded', updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL AND status='completed'`, id)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		return s.repos.Accounting.PostFeeRefundJournal(ctx, tx, id, actorID)
	})
	if err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityPayment, &id, ip, map[string]any{"status": "refunded"})
	return nil
}

func (s *FeeService) ListPayments(ctx context.Context, f dto.FinanceReportFilter, page, perPage int) ([]dto.PaymentResponse, error) {
	params := repository.PaymentSearchParams{From: f.From, To: f.To, Method: f.Method}
	if f.StudentID != uuid.Nil {
		params.StudentID = &f.StudentID
	}
	if perPage > 0 {
		params.Limit = int32(perPage)
		params.Offset = int32((page - 1) * perPage)
	}
	recs, err := s.repos.Fees.ListPayments(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.PaymentResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapPayment(&r))
	}
	return items, nil
}

// --- Receipts ---

func (s *FeeService) GetReceipt(ctx context.Context, id uuid.UUID) (*dto.ReceiptResponse, error) {
	rec, err := s.repos.Fees.GetReceipt(ctx, id)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	return s.buildReceiptResponse(ctx, rec)
}

func (s *FeeService) VerifyReceipt(ctx context.Context, token string) (*dto.ReceiptResponse, error) {
	rec, err := s.repos.Fees.GetReceiptByToken(ctx, token)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	return s.buildReceiptResponse(ctx, rec)
}

func (s *FeeService) buildReceiptResponse(ctx context.Context, rec *repository.ReceiptRecord) (*dto.ReceiptResponse, error) {
	allocs, _ := s.repos.Fees.ListAllocationsByPayment(ctx, rec.PaymentID)
	details := make([]dto.PaymentAllocationDetail, 0, len(allocs))
	for _, a := range allocs {
		details = append(details, dto.PaymentAllocationDetail{InvoiceNumber: a.InvoiceNumber, Amount: a.Amount})
	}
	school, _ := s.repos.Schools.Get(ctx)
	resp := &dto.ReceiptResponse{
		ID: rec.ID, ReceiptNumber: rec.ReceiptNumber, PaymentID: rec.PaymentID,
		PaymentNumber: rec.PaymentNumber, StudentID: rec.StudentID, StudentName: rec.StudentName,
		AdmissionNo: rec.AdmissionNo, ClassName: rec.ClassName, SectionName: rec.SectionName,
		TotalAmount: rec.TotalAmount, QRToken: rec.QRToken, IssuedAt: rec.IssuedAt,
		CollectorName: rec.CollectorName, Allocations: details,
	}
	if school != nil {
		resp.SchoolName = school.Name
		resp.SchoolLogo = school.LogoURL
	}
	return resp, nil
}

// --- Dashboard ---

func (s *FeeService) DashboardStats(ctx context.Context) (*dto.FinanceDashboardStats, error) {
	today := time.Now().Truncate(24 * time.Hour)
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	todayCol, _ := s.repos.Fees.SumCollection(ctx, today, today)
	monthCol, _ := s.repos.Fees.SumCollection(ctx, monthStart, today)
	outstanding, _ := s.repos.Fees.SumOutstanding(ctx)
	withDues, _ := s.repos.Fees.CountStudentsWithDues(ctx)
	trendFrom := today.AddDate(0, -1, 0)
	colTrend, _ := s.repos.Fees.DailyCollectionTrend(ctx, trendFrom, today)
	dueTrend, _ := s.repos.Fees.DailyDueTrend(ctx, trendFrom, today)
	methods, _ := s.repos.Fees.CollectionByMethod(ctx, monthStart, today)

	stats := &dto.FinanceDashboardStats{
		TodayCollection: todayCol, MonthlyCollection: monthCol,
		OutstandingDues: outstanding, StudentsWithDues: withDues,
	}
	for _, t := range colTrend {
		stats.CollectionTrend = append(stats.CollectionTrend, dto.FinanceTrendPoint{
			Label: t.Date.Format("Jan 2"), Amount: t.Amount,
		})
	}
	for _, t := range dueTrend {
		stats.DueTrend = append(stats.DueTrend, dto.FinanceTrendPoint{
			Label: t.Date.Format("Jan 2"), Amount: t.Amount,
		})
	}
	for _, m := range methods {
		stats.PaymentMethodBreakdown = append(stats.PaymentMethodBreakdown, dto.PaymentMethodStat{
			Method: m.Method, Amount: m.Amount, Count: m.Count,
		})
	}
	return stats, nil
}

// --- Due Management ---

func (s *FeeService) ListDueStudents(ctx context.Context, f dto.BillSearchFilter) ([]dto.DueStudentRow, error) {
	params := mapBillFilter(f)
	params.Limit = 100
	recs, err := s.repos.Fees.ListDueStudents(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.DueStudentRow, 0, len(recs))
	for _, r := range recs {
		items = append(items, dto.DueStudentRow{
			StudentID: r.StudentID, StudentName: r.StudentName, AdmissionNo: r.AdmissionNo,
			ClassName: r.ClassName, SectionName: r.SectionName,
			TotalDue: r.TotalDue, OverdueAmount: r.OverdueAmount, BillCount: r.BillCount,
		})
	}
	return items, nil
}

func (s *FeeService) ListOverdueBills(ctx context.Context, f dto.BillSearchFilter) (*dto.PaginatedBills, error) {
	f.Status = ""
	params := mapBillFilter(f)
	params.OverdueOnly = true
	recs, err := s.repos.Fees.SearchBills(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.StudentBillResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, s.mapBill(ctx, &r))
	}
	return &dto.PaginatedBills{Items: items, Total: int64(len(items)), Page: 1, PerPage: len(items), TotalPages: 1}, nil
}

// --- Reports ---

func (s *FeeService) FeeTypeCollection(ctx context.Context, from, to time.Time) ([]dto.FinanceTrendPoint, error) {
	recs, err := s.repos.Fees.FeeTypeCollection(ctx, from, to)
	if err != nil {
		return nil, err
	}
	items := make([]dto.FinanceTrendPoint, 0, len(recs))
	for _, r := range recs {
		items = append(items, dto.FinanceTrendPoint{Label: r.FeeTypeName, Amount: r.Amount})
	}
	return items, nil
}

func (s *FeeService) StudentLedger(ctx context.Context, studentID uuid.UUID, from, to time.Time) ([]dto.StudentLedgerEntry, error) {
	bills, _ := s.repos.Fees.SearchBills(ctx, repository.BillSearchParams{
		StudentID: &studentID, Limit: 500,
	})
	payments, _ := s.repos.Fees.ListPayments(ctx, repository.PaymentSearchParams{
		StudentID: &studentID, From: from, To: to, Limit: 500,
	})
	var entries []dto.StudentLedgerEntry
	var balance float64
	for _, b := range bills {
		if !from.IsZero() && b.GeneratedAt.Before(from) {
			continue
		}
		entries = append(entries, dto.StudentLedgerEntry{
			Date: b.GeneratedAt, Type: "bill", Reference: b.InvoiceNumber,
			Description: "Bill " + b.BillPeriod, Debit: b.TotalAmount, Balance: balance + b.TotalAmount,
		})
		balance += b.TotalAmount
		entries[len(entries)-1].Balance = balance
	}
	for _, p := range payments {
		entries = append(entries, dto.StudentLedgerEntry{
			Date: p.CollectionDate, Type: "payment", Reference: p.PaymentNumber,
			Description: "Payment via " + p.PaymentMethod, Credit: p.Amount, Balance: balance - p.Amount,
		})
		balance -= p.Amount
		entries[len(entries)-1].Balance = balance
	}
	return entries, nil
}

// --- Parent View ---

func (s *FeeService) ParentFeeSummary(ctx context.Context, studentID uuid.UUID) (*dto.ParentFeeSummary, error) {
	st, err := s.repos.Students.GetByID(ctx, studentID)
	if err != nil || st == nil {
		return nil, ErrNotFound
	}
	bills, _ := s.repos.Fees.SearchBills(ctx, repository.BillSearchParams{
		StudentID: &studentID, Limit: 50,
	})
	var currentDue float64
	for _, b := range bills {
		if b.Status != model.BillStatusCancelled && b.Status != model.BillStatusPaid {
			currentDue += b.TotalAmount - b.PaidAmount
		}
	}
	payments, _ := s.repos.Fees.ListPayments(ctx, repository.PaymentSearchParams{StudentID: &studentID, Limit: 10})
	var totalPaid float64
	payResp := make([]dto.PaymentResponse, 0, len(payments))
	for _, p := range payments {
		if p.Status == model.PaymentCompleted {
			totalPaid += p.Amount
		}
		payResp = append(payResp, mapPayment(&p))
	}
	receipts, _ := s.repos.Fees.ListReceiptsByStudent(ctx, studentID, 10)
	rcpResp := make([]dto.ReceiptResponse, 0, len(receipts))
	for _, r := range receipts {
		if resp, err := s.buildReceiptResponse(ctx, &r); err == nil {
			rcpResp = append(rcpResp, *resp)
		}
	}
	return &dto.ParentFeeSummary{
		StudentID: studentID, StudentName: st.FirstName + " " + st.LastName,
		CurrentDue: currentDue, TotalPaid: totalPaid,
		RecentPayments: payResp, Receipts: rcpResp,
	}, nil
}

func (s *FeeService) PendingBillsForStudent(ctx context.Context, studentID uuid.UUID) ([]dto.StudentBillResponse, error) {
	recs, err := s.repos.Fees.SearchBills(ctx, repository.BillSearchParams{StudentID: &studentID, Limit: 50})
	if err != nil {
		return nil, err
	}
	items := make([]dto.StudentBillResponse, 0)
	for _, r := range recs {
		if r.Status == model.BillStatusPending || r.Status == model.BillStatusPartial || r.Status == model.BillStatusOverdue {
			if r.TotalAmount > r.PaidAmount {
				items = append(items, s.mapBill(ctx, &r))
			}
		}
	}
	return items, nil
}

// --- Helpers ---

func mapFeeType(r *repository.FeeTypeRecord) dto.FeeTypeResponse {
	return dto.FeeTypeResponse{ID: r.IDUUID, Name: r.Name, Slug: r.Slug, Description: r.Description, IsActive: r.IsActive}
}

func mapFeeStructure(r *repository.FeeStructureRecord) dto.FeeStructureResponse {
	return dto.FeeStructureResponse{
		ID: r.ID, FeeTypeID: r.FeeTypeID, FeeTypeName: r.FeeTypeName,
		SessionID: r.SessionID, SessionName: r.SessionName,
		ClassID: r.ClassID, ClassName: r.ClassName,
		SectionID: r.SectionID, SectionName: r.SectionName,
		Amount: r.Amount, DueDay: r.DueDay, Frequency: r.Frequency, IsActive: r.IsActive,
	}
}

func mapStructureReq(req dto.FeeStructureRequest) repository.FeeStructureParams {
	p := repository.FeeStructureParams{
		FeeTypeID: req.FeeTypeID, SessionID: req.SessionID, ClassID: req.ClassID,
		Amount: req.Amount, DueDay: req.DueDay, Frequency: req.Frequency, IsActive: req.IsActive,
	}
	if req.SectionID != uuid.Nil {
		p.SectionID = &req.SectionID
	}
	return p
}

func mapDiscount(r *repository.DiscountRecord) dto.StudentDiscountResponse {
	return dto.StudentDiscountResponse{
		ID: r.ID, StudentID: r.StudentID, StudentName: r.StudentName, SessionID: r.SessionID,
		DiscountType: r.DiscountType, DiscountValue: r.DiscountValue,
		Reason: r.Reason, Description: r.Description, IsActive: r.IsActive,
	}
}

func mapPayment(r *repository.PaymentRecord) dto.PaymentResponse {
	return dto.PaymentResponse{
		ID: r.ID, PaymentNumber: r.PaymentNumber, StudentID: r.StudentID, StudentName: r.StudentName,
		Amount: r.Amount, PaymentMethod: r.PaymentMethod, CollectorName: r.CollectorName,
		CollectionDate: r.CollectionDate, Remarks: r.Remarks, Status: r.Status,
		ReceiptID: r.ReceiptID, ReceiptNumber: r.ReceiptNumber,
	}
}

func mapBillFilter(f dto.BillSearchFilter) repository.BillSearchParams {
	p := repository.BillSearchParams{Status: f.Status, Query: f.Query}
	if f.SessionID != uuid.Nil {
		p.SessionID = &f.SessionID
	}
	if f.ClassID != uuid.Nil {
		p.ClassID = &f.ClassID
	}
	if f.SectionID != uuid.Nil {
		p.SectionID = &f.SectionID
	}
	if f.StudentID != uuid.Nil {
		p.StudentID = &f.StudentID
	}
	if f.PerPage > 0 {
		p.Limit = int32(f.PerPage)
		p.Offset = int32((f.Page - 1) * f.PerPage)
	}
	return p
}

func generateQRToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
