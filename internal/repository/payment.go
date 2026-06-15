package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository interface {
	WithTx(ctx context.Context, fn func(pgx.Tx) error) error
	ListGateways(ctx context.Context) ([]PaymentGatewayRecord, error)
	ListActiveGateways(ctx context.Context) ([]PaymentGatewayRecord, error)
	GetGateway(ctx context.Context, id uuid.UUID) (*PaymentGatewayRecord, error)
	GetGatewayBySlug(ctx context.Context, slug string) (*PaymentGatewayRecord, error)
	UpdateGateway(ctx context.Context, id uuid.UUID, p UpdateGatewayParams) (*PaymentGatewayRecord, error)

	CreateTransaction(ctx context.Context, tx pgx.Tx, p CreateGatewayTxParams) (*GatewayTransactionRecord, error)
	GetTransaction(ctx context.Context, id uuid.UUID) (*GatewayTransactionRecord, error)
	GetTransactionByRef(ctx context.Context, ref string) (*GatewayTransactionRecord, error)
	GetTransactionByIdempotency(ctx context.Context, key string) (*GatewayTransactionRecord, error)
	GetTransactionByPaymentID(ctx context.Context, paymentID uuid.UUID) (*GatewayTransactionRecord, error)
	LockTransactionForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*GatewayTransactionRecord, error)
	UpdateTransactionStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, p UpdateGatewayTxParams) error
	SearchTransactions(ctx context.Context, f GatewayTxSearchParams) ([]GatewayTransactionRecord, error)
	CountTransactions(ctx context.Context, f GatewayTxSearchParams) (int64, error)
	ListStudentTransactions(ctx context.Context, studentID uuid.UUID, limit int32) ([]GatewayTransactionRecord, error)
	DashboardStats(ctx context.Context) (*PaymentDashboardStatsRecord, error)
	GatewayCollectionReport(ctx context.Context, from, to time.Time) ([]GatewayCollectionRecord, error)

	CreateRefund(ctx context.Context, p CreateRefundParams) (*PaymentRefundRecord, error)
	GetRefund(ctx context.Context, id uuid.UUID) (*PaymentRefundRecord, error)
	UpdateRefundStatus(ctx context.Context, id uuid.UUID, status string, approvedBy *uuid.UUID) error
	SearchRefunds(ctx context.Context, status string, limit, offset int32) ([]PaymentRefundRecord, error)
	CountRefunds(ctx context.Context, status string) (int64, error)
	NextTransactionRef(ctx context.Context) (string, error)
}

type PaymentGatewayRecord struct {
	ID          uuid.UUID
	Name, Slug  string
	IsActive, IsSandbox bool
	APIKey, APISecret, MerchantID, StoreID string
	CallbackURL, SuccessURL, FailURL string
	UpdatedAt   time.Time
}

type UpdateGatewayParams struct {
	Name, APIKey, APISecret, MerchantID, StoreID string
	CallbackURL, SuccessURL, FailURL string
	IsActive, IsSandbox bool
}

type CreateGatewayTxParams struct {
	GatewayID              uuid.UUID
	TransactionRef         string
	IdempotencyKey         string
	GatewayRef             string
	PaymentType            string
	ReferenceID            uuid.UUID
	StudentID, ParentID    *uuid.UUID
	AdmissionApplicationID *uuid.UUID
	BillIDs                []uuid.UUID
	Amount                 float64
	Currency               string
	Status                 string
	GatewayResponse        map[string]any
	IPAddress              string
}

type UpdateGatewayTxParams struct {
	Status            string
	GatewayRef        string
	PaymentID         *uuid.UUID
	SignatureVerified bool
	GatewayResponse   map[string]any
	ErrorMessage      string
	CompletedAt       *time.Time
}

type GatewayTransactionRecord struct {
	ID                     uuid.UUID
	GatewayID              uuid.UUID
	GatewayName, GatewaySlug string
	TransactionRef, IdempotencyKey, GatewayRef string
	PaymentType              string
	ReferenceID              uuid.UUID
	StudentID, ParentID      *uuid.UUID
	StudentName              string
	AdmissionApplicationID   *uuid.UUID
	BillIDs                  []uuid.UUID
	Amount                   float64
	Currency, Status         string
	PaymentID                *uuid.UUID
	ReceiptNumber            string
	SignatureVerified        bool
	ErrorMessage             string
	IPAddress                string
	CompletedAt              *time.Time
	CreatedAt, UpdatedAt     time.Time
}

type GatewayTxSearchParams struct {
	Query, Status, GatewaySlug, PaymentType string
	StudentID                               *uuid.UUID
	From, To                                time.Time
	Limit, Offset                           int32
}

type PaymentDashboardStatsRecord struct {
	TodayCollection   float64
	GatewayCollection float64
	FailedPayments    int64
	PendingPayments   int64
	TodayTransactions int64
}

type GatewayCollectionRecord struct {
	GatewaySlug, GatewayName string
	Count                    int64
	TotalAmount              float64
}

type CreateRefundParams struct {
	GatewayTransactionID *uuid.UUID
	PaymentID            uuid.UUID
	Amount               float64
	Reason               string
	RequestedBy          *uuid.UUID
}

type PaymentRefundRecord struct {
	ID                     uuid.UUID
	GatewayTransactionID   *uuid.UUID
	PaymentID              uuid.UUID
	PaymentNumber          string
	Amount                 float64
	Status, Reason         string
	RequestedByName, ApprovedByName string
	CreatedAt              time.Time
	ProcessedAt            *time.Time
}

type paymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) PaymentRepository {
	return &paymentRepository{pool: pool}
}

func (r *paymentRepository) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *paymentRepository) ListGateways(ctx context.Context) ([]PaymentGatewayRecord, error) {
	rows, err := r.pool.Query(ctx, gatewaySelect()+` ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGateways(rows)
}

func (r *paymentRepository) ListActiveGateways(ctx context.Context) ([]PaymentGatewayRecord, error) {
	rows, err := r.pool.Query(ctx, gatewaySelect()+` AND is_active=true ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGateways(rows)
}

func gatewaySelect() string {
	return `
SELECT id, name, slug, is_active, is_sandbox, COALESCE(api_key,''), COALESCE(api_secret,''),
    COALESCE(merchant_id,''), COALESCE(store_id,''), COALESCE(callback_url,''), COALESCE(success_url,''), COALESCE(fail_url,''), updated_at
FROM payment_gateways WHERE deleted_at IS NULL`
}

func scanGateways(rows pgx.Rows) ([]PaymentGatewayRecord, error) {
	var items []PaymentGatewayRecord
	for rows.Next() {
		var rec PaymentGatewayRecord
		if err := rows.Scan(&rec.ID, &rec.Name, &rec.Slug, &rec.IsActive, &rec.IsSandbox,
			&rec.APIKey, &rec.APISecret, &rec.MerchantID, &rec.StoreID,
			&rec.CallbackURL, &rec.SuccessURL, &rec.FailURL, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *paymentRepository) GetGateway(ctx context.Context, id uuid.UUID) (*PaymentGatewayRecord, error) {
	row := r.pool.QueryRow(ctx, gatewaySelect()+` AND id=$1`, id)
	return scanGateway(row)
}

func (r *paymentRepository) GetGatewayBySlug(ctx context.Context, slug string) (*PaymentGatewayRecord, error) {
	row := r.pool.QueryRow(ctx, gatewaySelect()+` AND slug=$1`, slug)
	return scanGateway(row)
}

func scanGateway(row pgx.Row) (*PaymentGatewayRecord, error) {
	var rec PaymentGatewayRecord
	if err := row.Scan(&rec.ID, &rec.Name, &rec.Slug, &rec.IsActive, &rec.IsSandbox,
		&rec.APIKey, &rec.APISecret, &rec.MerchantID, &rec.StoreID,
		&rec.CallbackURL, &rec.SuccessURL, &rec.FailURL, &rec.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *paymentRepository) UpdateGateway(ctx context.Context, id uuid.UUID, p UpdateGatewayParams) (*PaymentGatewayRecord, error) {
	_, err := r.pool.Exec(ctx, `
UPDATE payment_gateways SET name=$2, is_active=$3, is_sandbox=$4, api_key=$5, api_secret=$6,
    merchant_id=$7, store_id=$8, callback_url=$9, success_url=$10, fail_url=$11, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL`,
		id, p.Name, p.IsActive, p.IsSandbox, p.APIKey, p.APISecret, p.MerchantID, p.StoreID,
		p.CallbackURL, p.SuccessURL, p.FailURL)
	if err != nil {
		return nil, err
	}
	return r.GetGateway(ctx, id)
}

func (r *paymentRepository) NextTransactionRef(ctx context.Context) (string, error) {
	year := time.Now().Year()
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*)+1 FROM gateway_transactions WHERE EXTRACT(YEAR FROM created_at)=$1`, year).Scan(&n)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("GW-%d-%06d", year, n), nil
}

func (r *paymentRepository) CreateTransaction(ctx context.Context, tx pgx.Tx, p CreateGatewayTxParams) (*GatewayTransactionRecord, error) {
	billJSON, _ := json.Marshal(p.BillIDs)
	respJSON, _ := json.Marshal(p.GatewayResponse)
	var id uuid.UUID
	err := tx.QueryRow(ctx, `
INSERT INTO gateway_transactions (gateway_id, transaction_ref, idempotency_key, gateway_ref, payment_type, reference_id,
    student_id, parent_id, admission_application_id, bill_ids, amount, currency, status, gateway_response, ip_address)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING id`,
		p.GatewayID, p.TransactionRef, p.IdempotencyKey, p.GatewayRef, p.PaymentType, p.ReferenceID,
		p.StudentID, p.ParentID, p.AdmissionApplicationID, billJSON, p.Amount, p.Currency, p.Status, respJSON, p.IPAddress,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.getTransaction(ctx, tx, id)
}

func (r *paymentRepository) getTransaction(ctx context.Context, q pgxQuerier, id uuid.UUID) (*GatewayTransactionRecord, error) {
	row := q.QueryRow(ctx, txSelect()+` WHERE gt.id=$1`, id)
	return scanGatewayTx(row)
}

func (r *paymentRepository) GetTransaction(ctx context.Context, id uuid.UUID) (*GatewayTransactionRecord, error) {
	return r.getTransaction(ctx, r.pool, id)
}

func (r *paymentRepository) GetTransactionByRef(ctx context.Context, ref string) (*GatewayTransactionRecord, error) {
	row := r.pool.QueryRow(ctx, txSelect()+` WHERE gt.transaction_ref=$1`, ref)
	return scanGatewayTx(row)
}

func (r *paymentRepository) GetTransactionByIdempotency(ctx context.Context, key string) (*GatewayTransactionRecord, error) {
	row := r.pool.QueryRow(ctx, txSelect()+` WHERE gt.idempotency_key=$1`, key)
	return scanGatewayTx(row)
}

func (r *paymentRepository) GetTransactionByPaymentID(ctx context.Context, paymentID uuid.UUID) (*GatewayTransactionRecord, error) {
	row := r.pool.QueryRow(ctx, txSelect()+` WHERE gt.payment_id=$1`, paymentID)
	return scanGatewayTx(row)
}

func (r *paymentRepository) LockTransactionForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*GatewayTransactionRecord, error) {
	row := tx.QueryRow(ctx, txSelect()+` WHERE gt.id=$1 FOR UPDATE`, id)
	return scanGatewayTx(row)
}

func (r *paymentRepository) UpdateTransactionStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, p UpdateGatewayTxParams) error {
	respJSON, _ := json.Marshal(p.GatewayResponse)
	_, err := tx.Exec(ctx, `
UPDATE gateway_transactions SET status=$2, gateway_ref=COALESCE(NULLIF($3,''), gateway_ref),
    payment_id=COALESCE($4, payment_id), signature_verified=$5, gateway_response=COALESCE($6, gateway_response),
    error_message=$7, completed_at=COALESCE($8, completed_at), updated_at=NOW()
WHERE id=$1`,
		id, p.Status, p.GatewayRef, p.PaymentID, p.SignatureVerified, respJSON, p.ErrorMessage, p.CompletedAt)
	return err
}

func txSelect() string {
	return `
SELECT gt.id, gt.gateway_id, pg.name, pg.slug, gt.transaction_ref, gt.idempotency_key, COALESCE(gt.gateway_ref,''),
    gt.payment_type, gt.reference_id, gt.student_id, gt.parent_id, gt.admission_application_id, gt.bill_ids,
    gt.amount, gt.currency, gt.status, gt.payment_id, COALESCE(r.receipt_number,''), gt.signature_verified,
    COALESCE(gt.error_message,''), COALESCE(gt.ip_address,''), gt.completed_at, gt.created_at, gt.updated_at,
    COALESCE(s.first_name||' '||s.last_name,'')
FROM gateway_transactions gt
JOIN payment_gateways pg ON pg.id = gt.gateway_id
LEFT JOIN payments p ON p.id = gt.payment_id
LEFT JOIN receipts r ON r.payment_id = p.id
LEFT JOIN students s ON s.id = gt.student_id`
}

type pgxQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func scanGatewayTx(row pgx.Row) (*GatewayTransactionRecord, error) {
	var rec GatewayTransactionRecord
	var billJSON []byte
	var studentID, parentID, admID, payID *uuid.UUID
	var completedAt *time.Time
	if err := row.Scan(&rec.ID, &rec.GatewayID, &rec.GatewayName, &rec.GatewaySlug,
		&rec.TransactionRef, &rec.IdempotencyKey, &rec.GatewayRef, &rec.PaymentType, &rec.ReferenceID,
		&studentID, &parentID, &admID, &billJSON, &rec.Amount, &rec.Currency, &rec.Status,
		&payID, &rec.ReceiptNumber, &rec.SignatureVerified, &rec.ErrorMessage, &rec.IPAddress,
		&completedAt, &rec.CreatedAt, &rec.UpdatedAt, &rec.StudentName); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	rec.StudentID, rec.ParentID, rec.AdmissionApplicationID, rec.PaymentID = studentID, parentID, admID, payID
	rec.CompletedAt = completedAt
	if len(billJSON) > 0 {
		_ = json.Unmarshal(billJSON, &rec.BillIDs)
	}
	return &rec, nil
}

func (r *paymentRepository) SearchTransactions(ctx context.Context, f GatewayTxSearchParams) ([]GatewayTransactionRecord, error) {
	where, args := txFilter(f)
	q := txSelect() + where + fmt.Sprintf(" ORDER BY gt.created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, f.Limit, f.Offset)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GatewayTransactionRecord
	for rows.Next() {
		rec, err := scanGatewayTxRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func scanGatewayTxRow(rows pgx.Rows) (*GatewayTransactionRecord, error) {
	var rec GatewayTransactionRecord
	var billJSON []byte
	var studentID, parentID, admID, payID *uuid.UUID
	var completedAt *time.Time
	if err := rows.Scan(&rec.ID, &rec.GatewayID, &rec.GatewayName, &rec.GatewaySlug,
		&rec.TransactionRef, &rec.IdempotencyKey, &rec.GatewayRef, &rec.PaymentType, &rec.ReferenceID,
		&studentID, &parentID, &admID, &billJSON, &rec.Amount, &rec.Currency, &rec.Status,
		&payID, &rec.ReceiptNumber, &rec.SignatureVerified, &rec.ErrorMessage, &rec.IPAddress,
		&completedAt, &rec.CreatedAt, &rec.UpdatedAt, &rec.StudentName); err != nil {
		return nil, err
	}
	rec.StudentID, rec.ParentID, rec.AdmissionApplicationID, rec.PaymentID = studentID, parentID, admID, payID
	rec.CompletedAt = completedAt
	if len(billJSON) > 0 {
		_ = json.Unmarshal(billJSON, &rec.BillIDs)
	}
	return &rec, nil
}

func txFilter(f GatewayTxSearchParams) (string, []any) {
	var parts []string
	var args []any
	if f.Status != "" {
		args = append(args, f.Status)
		parts = append(parts, fmt.Sprintf("gt.status=$%d", len(args)))
	}
	if f.GatewaySlug != "" {
		args = append(args, f.GatewaySlug)
		parts = append(parts, fmt.Sprintf("pg.slug=$%d", len(args)))
	}
	if f.PaymentType != "" {
		args = append(args, f.PaymentType)
		parts = append(parts, fmt.Sprintf("gt.payment_type=$%d", len(args)))
	}
	if f.StudentID != nil {
		args = append(args, *f.StudentID)
		parts = append(parts, fmt.Sprintf("gt.student_id=$%d", len(args)))
	}
	if !f.From.IsZero() {
		args = append(args, f.From)
		parts = append(parts, fmt.Sprintf("gt.created_at >= $%d", len(args)))
	}
	if !f.To.IsZero() {
		args = append(args, f.To)
		parts = append(parts, fmt.Sprintf("gt.created_at <= $%d", len(args)))
	}
	if f.Query != "" {
		args = append(args, "%"+strings.ToLower(f.Query)+"%")
		parts = append(parts, fmt.Sprintf("(LOWER(gt.transaction_ref) LIKE $%d OR LOWER(COALESCE(gt.gateway_ref,'')) LIKE $%d OR LOWER(COALESCE(s.first_name||' '||s.last_name,'')) LIKE $%d)", len(args), len(args), len(args)))
	}
	if len(parts) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

func (r *paymentRepository) CountTransactions(ctx context.Context, f GatewayTxSearchParams) (int64, error) {
	where, args := txFilter(f)
	q := `SELECT COUNT(*) FROM gateway_transactions gt JOIN payment_gateways pg ON pg.id=gt.gateway_id LEFT JOIN students s ON s.id=gt.student_id` + where
	var n int64
	err := r.pool.QueryRow(ctx, q, args...).Scan(&n)
	return n, err
}

func (r *paymentRepository) ListStudentTransactions(ctx context.Context, studentID uuid.UUID, limit int32) ([]GatewayTransactionRecord, error) {
	sid := studentID
	return r.SearchTransactions(ctx, GatewayTxSearchParams{StudentID: &sid, Limit: limit, Offset: 0})
}

func (r *paymentRepository) DashboardStats(ctx context.Context) (*PaymentDashboardStatsRecord, error) {
	row := r.pool.QueryRow(ctx, `
SELECT
    COALESCE(SUM(CASE WHEN status='completed' AND completed_at::date=CURRENT_DATE THEN amount ELSE 0 END),0),
    COALESCE(SUM(CASE WHEN status='completed' THEN amount ELSE 0 END),0),
    COUNT(*) FILTER (WHERE status='failed'),
    COUNT(*) FILTER (WHERE status IN ('pending','processing')),
    COUNT(*) FILTER (WHERE created_at::date=CURRENT_DATE)
FROM gateway_transactions`)
	var s PaymentDashboardStatsRecord
	if err := row.Scan(&s.TodayCollection, &s.GatewayCollection, &s.FailedPayments, &s.PendingPayments, &s.TodayTransactions); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *paymentRepository) GatewayCollectionReport(ctx context.Context, from, to time.Time) ([]GatewayCollectionRecord, error) {
	q := `
SELECT pg.slug, pg.name, COUNT(*)::bigint, COALESCE(SUM(gt.amount),0)
FROM gateway_transactions gt
JOIN payment_gateways pg ON pg.id=gt.gateway_id
WHERE gt.status='completed'`
	var args []any
	if !from.IsZero() {
		args = append(args, from)
		q += fmt.Sprintf(" AND gt.completed_at >= $%d", len(args))
	}
	if !to.IsZero() {
		args = append(args, to)
		q += fmt.Sprintf(" AND gt.completed_at <= $%d", len(args))
	}
	q += " GROUP BY pg.slug, pg.name ORDER BY SUM(gt.amount) DESC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GatewayCollectionRecord
	for rows.Next() {
		var rec GatewayCollectionRecord
		if err := rows.Scan(&rec.GatewaySlug, &rec.GatewayName, &rec.Count, &rec.TotalAmount); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *paymentRepository) CreateRefund(ctx context.Context, p CreateRefundParams) (*PaymentRefundRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO payment_refunds (gateway_transaction_id, payment_id, amount, reason, requested_by)
VALUES ($1,$2,$3,$4,$5) RETURNING id`, p.GatewayTransactionID, p.PaymentID, p.Amount, p.Reason, p.RequestedBy).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetRefund(ctx, id)
}

func (r *paymentRepository) GetRefund(ctx context.Context, id uuid.UUID) (*PaymentRefundRecord, error) {
	return scanRefund(r.pool.QueryRow(ctx, refundSelect()+` WHERE pr.id=$1`, id))
}

func refundSelect() string {
	return `
SELECT pr.id, pr.gateway_transaction_id, pr.payment_id, COALESCE(p.payment_number,''), pr.amount, pr.status, pr.reason,
    COALESCE(ru.first_name||' '||ru.last_name,''), COALESCE(au.first_name||' '||au.last_name,''), pr.created_at, pr.processed_at
FROM payment_refunds pr
JOIN payments p ON p.id = pr.payment_id
LEFT JOIN users ru ON ru.id = pr.requested_by
LEFT JOIN users au ON au.id = pr.approved_by`
}

func scanRefund(row pgx.Row) (*PaymentRefundRecord, error) {
	var rec PaymentRefundRecord
	var gwTxID *uuid.UUID
	if err := row.Scan(&rec.ID, &gwTxID, &rec.PaymentID, &rec.PaymentNumber, &rec.Amount, &rec.Status, &rec.Reason,
		&rec.RequestedByName, &rec.ApprovedByName, &rec.CreatedAt, &rec.ProcessedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	rec.GatewayTransactionID = gwTxID
	return &rec, nil
}

func (r *paymentRepository) UpdateRefundStatus(ctx context.Context, id uuid.UUID, status string, approvedBy *uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
UPDATE payment_refunds SET status=$2, approved_by=COALESCE($3, approved_by),
    processed_at=CASE WHEN $2 IN ('processed','rejected') THEN NOW() ELSE processed_at END, updated_at=NOW()
WHERE id=$1`, id, status, approvedBy)
	return err
}

func (r *paymentRepository) SearchRefunds(ctx context.Context, status string, limit, offset int32) ([]PaymentRefundRecord, error) {
	q := refundSelect()
	var args []any
	if status != "" {
		args = append(args, status)
		q += " WHERE pr.status=$1"
	}
	q += fmt.Sprintf(" ORDER BY pr.created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PaymentRefundRecord
	for rows.Next() {
		var rec PaymentRefundRecord
		var gwTxID *uuid.UUID
		if err := rows.Scan(&rec.ID, &gwTxID, &rec.PaymentID, &rec.PaymentNumber, &rec.Amount, &rec.Status, &rec.Reason,
			&rec.RequestedByName, &rec.ApprovedByName, &rec.CreatedAt, &rec.ProcessedAt); err != nil {
			return nil, err
		}
		rec.GatewayTransactionID = gwTxID
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *paymentRepository) CountRefunds(ctx context.Context, status string) (int64, error) {
	q := `SELECT COUNT(*) FROM payment_refunds`
	var args []any
	if status != "" {
		args = append(args, status)
		q += " WHERE status=$1"
	}
	var n int64
	err := r.pool.QueryRow(ctx, q, args...).Scan(&n)
	return n, err
}
