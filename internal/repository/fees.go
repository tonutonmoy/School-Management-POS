package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeeRepository interface {
	// Fee types
	CreateFeeType(ctx context.Context, name, slug, description string, isActive bool) (*FeeTypeRecord, error)
	UpdateFeeType(ctx context.Context, id uuid.UUID, name, slug, description string, isActive bool) (*FeeTypeRecord, error)
	SoftDeleteFeeType(ctx context.Context, id uuid.UUID) error
	ListFeeTypes(ctx context.Context, activeOnly bool) ([]FeeTypeRecord, error)
	GetFeeType(ctx context.Context, id uuid.UUID) (*FeeTypeRecord, error)

	// Fee structures
	CreateFeeStructure(ctx context.Context, p FeeStructureParams) (*FeeStructureRecord, error)
	UpdateFeeStructure(ctx context.Context, id uuid.UUID, p FeeStructureParams) (*FeeStructureRecord, error)
	SoftDeleteFeeStructure(ctx context.Context, id uuid.UUID) error
	ListFeeStructures(ctx context.Context, f FeeStructureFilter) ([]FeeStructureRecord, error)
	GetFeeStructure(ctx context.Context, id uuid.UUID) (*FeeStructureRecord, error)
	ListApplicableStructures(ctx context.Context, sessionID, classID, sectionID uuid.UUID, frequency string) ([]FeeStructureRecord, error)

	// Discounts
	CreateDiscount(ctx context.Context, p DiscountParams) (*DiscountRecord, error)
	UpdateDiscount(ctx context.Context, id uuid.UUID, p DiscountParams) (*DiscountRecord, error)
	SoftDeleteDiscount(ctx context.Context, id uuid.UUID) error
	ListDiscounts(ctx context.Context, studentID, sessionID *uuid.UUID) ([]DiscountRecord, error)
	GetActiveDiscounts(ctx context.Context, studentID, sessionID uuid.UUID) ([]DiscountRecord, error)

	// Bills
	NextFinanceNumber(ctx context.Context, entityType string, year int) (string, error)
	CreateBill(ctx context.Context, tx pgx.Tx, p CreateBillParams) (*BillRecord, error)
	CreateBillItem(ctx context.Context, tx pgx.Tx, billID, feeTypeID uuid.UUID, structureID *uuid.UUID, desc string, amount float64) error
	GetBill(ctx context.Context, id uuid.UUID) (*BillRecord, error)
	GetBillByStudentPeriod(ctx context.Context, studentID uuid.UUID, period string) (*BillRecord, error)
	SearchBills(ctx context.Context, f BillSearchParams) ([]BillRecord, error)
	CountBills(ctx context.Context, f BillSearchParams) (int64, error)
	ListBillItems(ctx context.Context, billID uuid.UUID) ([]BillItemRecord, error)
	CancelBill(ctx context.Context, id uuid.UUID) error
	UpdateBillPayment(ctx context.Context, tx pgx.Tx, billID uuid.UUID, paidDelta float64) error
	MarkOverdueBills(ctx context.Context, before time.Time) error

	// Payments (transactional)
	WithTx(ctx context.Context, fn func(pgx.Tx) error) error
	CreatePayment(ctx context.Context, tx pgx.Tx, p CreatePaymentParams) (*PaymentRecord, error)
	CreateAllocation(ctx context.Context, tx pgx.Tx, paymentID, billID uuid.UUID, amount float64) error
	CreateReceipt(ctx context.Context, tx pgx.Tx, p CreateReceiptParams) (*ReceiptRecord, error)
	ListPayments(ctx context.Context, f PaymentSearchParams) ([]PaymentRecord, error)
	CountPayments(ctx context.Context, f PaymentSearchParams) (int64, error)
	GetPayment(ctx context.Context, id uuid.UUID) (*PaymentRecord, error)
	GetReceipt(ctx context.Context, id uuid.UUID) (*ReceiptRecord, error)
	GetReceiptByToken(ctx context.Context, token string) (*ReceiptRecord, error)
	ListReceiptsByStudent(ctx context.Context, studentID uuid.UUID, limit int32) ([]ReceiptRecord, error)
	ListAllocationsByPayment(ctx context.Context, paymentID uuid.UUID) ([]AllocationRecord, error)
	RefundPayment(ctx context.Context, id uuid.UUID) error

	// Dashboard & reports
	SumCollection(ctx context.Context, from, to time.Time) (float64, error)
	SumOutstanding(ctx context.Context) (float64, error)
	CountStudentsWithDues(ctx context.Context) (int64, error)
	CollectionByMethod(ctx context.Context, from, to time.Time) ([]MethodAmountRecord, error)
	DailyCollectionTrend(ctx context.Context, from, to time.Time) ([]DailyAmountRecord, error)
	DailyDueTrend(ctx context.Context, from, to time.Time) ([]DailyAmountRecord, error)
	ListDueStudents(ctx context.Context, f BillSearchParams) ([]DueStudentRecord, error)
	FeeTypeCollection(ctx context.Context, from, to time.Time) ([]FeeTypeAmountRecord, error)
}

type FeeTypeRecord struct {
	IDUUID      uuid.UUID
	Name, Slug, Description string
	IsActive    bool
}

type FeeStructureParams struct {
	FeeTypeID, SessionID, ClassID uuid.UUID
	SectionID                     *uuid.UUID
	Amount                        float64
	DueDay                        int
	Frequency                     string
	IsActive                      bool
}

type FeeStructureFilter struct {
	SessionID, ClassID *uuid.UUID
	ActiveOnly         bool
}

type FeeStructureRecord struct {
	ID, FeeTypeID, SessionID, ClassID uuid.UUID
	SectionID                         *uuid.UUID
	FeeTypeName, SessionName, ClassName, SectionName, Frequency string
	Amount                            float64
	DueDay                            int
	IsActive                          bool
}

type DiscountParams struct {
	StudentID, SessionID          uuid.UUID
	DiscountType, Reason, Description string
	DiscountValue                 float64
	IsActive                      bool
}

type DiscountRecord struct {
	ID, StudentID, SessionID uuid.UUID
	StudentName, DiscountType, Reason, Description string
	DiscountValue float64
	IsActive      bool
}

type CreateBillParams struct {
	InvoiceNumber                          string
	StudentID, SessionID, ClassID, SectionID uuid.UUID
	BillPeriod                             string
	DueDate                                time.Time
	Subtotal, DiscountAmount, TotalAmount  float64
	Status                                 string
}

type BillRecord struct {
	ID, StudentID, SessionID, ClassID, SectionID uuid.UUID
	InvoiceNumber, StudentName, AdmissionNo, ClassName, SectionName, BillPeriod, Status string
	DueDate, GeneratedAt time.Time
	Subtotal, DiscountAmount, TotalAmount, PaidAmount float64
}

type BillItemRecord struct {
	ID, FeeTypeID uuid.UUID
	FeeTypeName, Description string
	Amount float64
}

type BillSearchParams struct {
	SessionID, ClassID, SectionID, StudentID *uuid.UUID
	Status, Query                            string
	OverdueOnly                              bool
	Limit, Offset                            int32
}

type CreatePaymentParams struct {
	PaymentNumber              string
	StudentID                  uuid.UUID
	Amount                     float64
	PaymentMethod              string
	CollectedBy                uuid.UUID
	CollectionDate             time.Time
	Remarks                    string
}

type PaymentRecord struct {
	ID, StudentID uuid.UUID
	PaymentNumber, StudentName, PaymentMethod, CollectorName, Remarks, Status string
	Amount         float64
	CollectionDate time.Time
	ReceiptID      *uuid.UUID
	ReceiptNumber  string
}

type PaymentSearchParams struct {
	StudentID *uuid.UUID
	From, To  time.Time
	Method    string
	Limit, Offset int32
}

type CreateReceiptParams struct {
	ReceiptNumber       string
	PaymentID, StudentID uuid.UUID
	TotalAmount         float64
	QRToken             string
	IssuedBy            uuid.UUID
}

type ReceiptRecord struct {
	ID, PaymentID, StudentID uuid.UUID
	ReceiptNumber, PaymentNumber, StudentName, AdmissionNo, ClassName, SectionName string
	QRToken, CollectorName string
	TotalAmount float64
	IssuedAt    time.Time
}

type AllocationRecord struct {
	BillID uuid.UUID
	InvoiceNumber string
	Amount float64
}

type DueStudentRecord struct {
	StudentID uuid.UUID
	StudentName, AdmissionNo, ClassName, SectionName string
	TotalDue, OverdueAmount float64
	BillCount int64
}

type MethodAmountRecord struct {
	Method string
	Amount float64
	Count  int64
}

type DailyAmountRecord struct {
	Date   time.Time
	Amount float64
}

type FeeTypeAmountRecord struct {
	FeeTypeName string
	Amount      float64
}

type feeRepository struct{ pool *pgxpool.Pool }

func NewFeeRepository(pool *pgxpool.Pool) FeeRepository { return &feeRepository{pool: pool} }

func (r *feeRepository) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
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

func (r *feeRepository) NextFinanceNumber(ctx context.Context, entityType string, year int) (string, error) {
	var seq int
	prefix := map[string]string{"invoice": "INV", "payment": "PAY", "receipt": "RCP"}[entityType]
	if prefix == "" {
		prefix = "DOC"
	}
	err := r.pool.QueryRow(ctx, `
INSERT INTO finance_sequences (entity_type, year, last_number) VALUES ($1, $2, 1)
ON CONFLICT (entity_type, year) DO UPDATE SET last_number = finance_sequences.last_number + 1
RETURNING last_number`, entityType, year).Scan(&seq)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%d-%05d", prefix, year, seq), nil
}

func scanFeeType(row pgx.Row) (*FeeTypeRecord, error) {
	var rec FeeTypeRecord
	err := row.Scan(&rec.IDUUID, &rec.Name, &rec.Slug, &rec.Description, &rec.IsActive)
	return &rec, err
}

func (r *feeRepository) CreateFeeType(ctx context.Context, name, slug, description string, isActive bool) (*FeeTypeRecord, error) {
	return scanFeeType(r.pool.QueryRow(ctx, `
INSERT INTO fee_types (name, slug, description, is_active) VALUES ($1,$2,$3,$4)
RETURNING id, name, slug, COALESCE(description,''), is_active`, name, slug, description, isActive))
}

func (r *feeRepository) UpdateFeeType(ctx context.Context, id uuid.UUID, name, slug, description string, isActive bool) (*FeeTypeRecord, error) {
	return scanFeeType(r.pool.QueryRow(ctx, `
UPDATE fee_types SET name=$2, slug=$3, description=$4, is_active=$5, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL RETURNING id, name, slug, COALESCE(description,''), is_active`,
		id, name, slug, description, isActive))
}

func (r *feeRepository) SoftDeleteFeeType(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE fee_types SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *feeRepository) ListFeeTypes(ctx context.Context, activeOnly bool) ([]FeeTypeRecord, error) {
	q := `SELECT id, name, slug, COALESCE(description,''), is_active FROM fee_types WHERE deleted_at IS NULL`
	if activeOnly {
		q += ` AND is_active = true`
	}
	q += ` ORDER BY name`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FeeTypeRecord
	for rows.Next() {
		var rec FeeTypeRecord
		if err := rows.Scan(&rec.IDUUID, &rec.Name, &rec.Slug, &rec.Description, &rec.IsActive); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *feeRepository) GetFeeType(ctx context.Context, id uuid.UUID) (*FeeTypeRecord, error) {
	rec, err := scanFeeType(r.pool.QueryRow(ctx, `
SELECT id, name, slug, COALESCE(description,''), is_active FROM fee_types WHERE id=$1 AND deleted_at IS NULL`, id))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func nullUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func (r *feeRepository) CreateFeeStructure(ctx context.Context, p FeeStructureParams) (*FeeStructureRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO fee_structures (fee_type_id, session_id, class_id, section_id, amount, due_day, frequency, is_active)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		p.FeeTypeID, p.SessionID, p.ClassID, nullUUID(p.SectionID), p.Amount, p.DueDay, p.Frequency, p.IsActive).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetFeeStructure(ctx, id)
}

func (r *feeRepository) UpdateFeeStructure(ctx context.Context, id uuid.UUID, p FeeStructureParams) (*FeeStructureRecord, error) {
	_, err := r.pool.Exec(ctx, `
UPDATE fee_structures SET fee_type_id=$2, session_id=$3, class_id=$4, section_id=$5, amount=$6, due_day=$7, frequency=$8, is_active=$9, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL`,
		id, p.FeeTypeID, p.SessionID, p.ClassID, nullUUID(p.SectionID), p.Amount, p.DueDay, p.Frequency, p.IsActive)
	if err != nil {
		return nil, err
	}
	return r.GetFeeStructure(ctx, id)
}

func (r *feeRepository) SoftDeleteFeeStructure(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE fee_structures SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *feeRepository) feeStructureSelect() string {
	return `
SELECT fs.id, fs.fee_type_id, ft.name, fs.session_id, sess.name, fs.class_id, c.name,
    fs.section_id, COALESCE(sec.name,''), fs.amount, fs.due_day, fs.frequency, fs.is_active
FROM fee_structures fs
JOIN fee_types ft ON ft.id = fs.fee_type_id
JOIN academic_sessions sess ON sess.id = fs.session_id
JOIN classes c ON c.id = fs.class_id
LEFT JOIN sections sec ON sec.id = fs.section_id
WHERE fs.deleted_at IS NULL`
}

func scanFeeStructure(row pgx.Row) (*FeeStructureRecord, error) {
	var rec FeeStructureRecord
	var secID *uuid.UUID
	err := row.Scan(&rec.ID, &rec.FeeTypeID, &rec.FeeTypeName, &rec.SessionID, &rec.SessionName,
		&rec.ClassID, &rec.ClassName, &secID, &rec.SectionName, &rec.Amount, &rec.DueDay, &rec.Frequency, &rec.IsActive)
	rec.SectionID = secID
	return &rec, err
}

func (r *feeRepository) GetFeeStructure(ctx context.Context, id uuid.UUID) (*FeeStructureRecord, error) {
	rec, err := scanFeeStructure(r.pool.QueryRow(ctx, r.feeStructureSelect()+` AND fs.id=$1`, id))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *feeRepository) ListFeeStructures(ctx context.Context, f FeeStructureFilter) ([]FeeStructureRecord, error) {
	q := r.feeStructureSelect()
	args := []any{}
	n := 1
	if f.SessionID != nil {
		q += fmt.Sprintf(" AND fs.session_id=$%d", n)
		args = append(args, *f.SessionID)
		n++
	}
	if f.ClassID != nil {
		q += fmt.Sprintf(" AND fs.class_id=$%d", n)
		args = append(args, *f.ClassID)
		n++
	}
	if f.ActiveOnly {
		q += " AND fs.is_active = true"
	}
	q += " ORDER BY sess.name, c.name, ft.name"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FeeStructureRecord
	for rows.Next() {
		rec, err := scanFeeStructure(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func (r *feeRepository) ListApplicableStructures(ctx context.Context, sessionID, classID, sectionID uuid.UUID, frequency string) ([]FeeStructureRecord, error) {
	q := r.feeStructureSelect() + ` AND fs.session_id=$1 AND fs.class_id=$2 AND fs.is_active=true
AND (fs.section_id IS NULL OR fs.section_id=$3)`
	args := []any{sessionID, classID, sectionID}
	if frequency != "" {
		q += " AND fs.frequency=$4"
		args = append(args, frequency)
	}
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FeeStructureRecord
	for rows.Next() {
		rec, err := scanFeeStructure(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func (r *feeRepository) CreateDiscount(ctx context.Context, p DiscountParams) (*DiscountRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO student_discounts (student_id, session_id, discount_type, discount_value, reason, description, is_active)
VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		p.StudentID, p.SessionID, p.DiscountType, p.DiscountValue, p.Reason, p.Description, p.IsActive).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.getDiscount(ctx, id)
}

func (r *feeRepository) getDiscount(ctx context.Context, id uuid.UUID) (*DiscountRecord, error) {
	var rec DiscountRecord
	err := r.pool.QueryRow(ctx, `
SELECT sd.id, sd.student_id, s.first_name||' '||s.last_name, sd.session_id, sd.discount_type, sd.discount_value, sd.reason, COALESCE(sd.description,''), sd.is_active
FROM student_discounts sd JOIN students s ON s.id=sd.student_id
WHERE sd.id=$1 AND sd.deleted_at IS NULL`, id).Scan(
		&rec.ID, &rec.StudentID, &rec.StudentName, &rec.SessionID, &rec.DiscountType,
		&rec.DiscountValue, &rec.Reason, &rec.Description, &rec.IsActive)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &rec, err
}

func (r *feeRepository) UpdateDiscount(ctx context.Context, id uuid.UUID, p DiscountParams) (*DiscountRecord, error) {
	_, err := r.pool.Exec(ctx, `
UPDATE student_discounts SET student_id=$2, session_id=$3, discount_type=$4, discount_value=$5, reason=$6, description=$7, is_active=$8, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL`,
		id, p.StudentID, p.SessionID, p.DiscountType, p.DiscountValue, p.Reason, p.Description, p.IsActive)
	if err != nil {
		return nil, err
	}
	return r.getDiscount(ctx, id)
}

func (r *feeRepository) SoftDeleteDiscount(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE student_discounts SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *feeRepository) ListDiscounts(ctx context.Context, studentID, sessionID *uuid.UUID) ([]DiscountRecord, error) {
	q := `
SELECT sd.id, sd.student_id, s.first_name||' '||s.last_name, sd.session_id, sd.discount_type, sd.discount_value, sd.reason, COALESCE(sd.description,''), sd.is_active
FROM student_discounts sd JOIN students s ON s.id=sd.student_id WHERE sd.deleted_at IS NULL`
	args := []any{}
	n := 1
	if studentID != nil {
		q += fmt.Sprintf(" AND sd.student_id=$%d", n)
		args = append(args, *studentID)
		n++
	}
	if sessionID != nil {
		q += fmt.Sprintf(" AND sd.session_id=$%d", n)
		args = append(args, *sessionID)
	}
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DiscountRecord
	for rows.Next() {
		var rec DiscountRecord
		if err := rows.Scan(&rec.ID, &rec.StudentID, &rec.StudentName, &rec.SessionID, &rec.DiscountType,
			&rec.DiscountValue, &rec.Reason, &rec.Description, &rec.IsActive); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *feeRepository) GetActiveDiscounts(ctx context.Context, studentID, sessionID uuid.UUID) ([]DiscountRecord, error) {
	return r.ListDiscounts(ctx, &studentID, &sessionID)
}

func (r *feeRepository) billSelect() string {
	return `
SELECT b.id, b.invoice_number, b.student_id, s.first_name||' '||s.last_name, s.admission_number,
    b.session_id, b.class_id, c.name, b.section_id, sec.name, b.bill_period, b.due_date,
    b.subtotal, b.discount_amount, b.total_amount, b.paid_amount, b.status, b.generated_at
FROM student_bills b
JOIN students s ON s.id=b.student_id
JOIN classes c ON c.id=b.class_id
JOIN sections sec ON sec.id=b.section_id
WHERE b.deleted_at IS NULL`
}

func scanBill(row pgx.Row) (*BillRecord, error) {
	var rec BillRecord
	err := row.Scan(&rec.ID, &rec.InvoiceNumber, &rec.StudentID, &rec.StudentName, &rec.AdmissionNo,
		&rec.SessionID, &rec.ClassID, &rec.ClassName, &rec.SectionID, &rec.SectionName,
		&rec.BillPeriod, &rec.DueDate, &rec.Subtotal, &rec.DiscountAmount, &rec.TotalAmount,
		&rec.PaidAmount, &rec.Status, &rec.GeneratedAt)
	return &rec, err
}

func (r *feeRepository) CreateBill(ctx context.Context, tx pgx.Tx, p CreateBillParams) (*BillRecord, error) {
	var id uuid.UUID
	q := `
INSERT INTO student_bills (invoice_number, student_id, session_id, class_id, section_id, bill_period, due_date, subtotal, discount_amount, total_amount, status)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`
	var err error
	if tx != nil {
		err = tx.QueryRow(ctx, q, p.InvoiceNumber, p.StudentID, p.SessionID, p.ClassID, p.SectionID,
			p.BillPeriod, p.DueDate, p.Subtotal, p.DiscountAmount, p.TotalAmount, p.Status).Scan(&id)
	} else {
		err = r.pool.QueryRow(ctx, q, p.InvoiceNumber, p.StudentID, p.SessionID, p.ClassID, p.SectionID,
			p.BillPeriod, p.DueDate, p.Subtotal, p.DiscountAmount, p.TotalAmount, p.Status).Scan(&id)
	}
	if err != nil {
		return nil, err
	}
	return r.GetBill(ctx, id)
}

func (r *feeRepository) CreateBillItem(ctx context.Context, tx pgx.Tx, billID, feeTypeID uuid.UUID, structureID *uuid.UUID, desc string, amount float64) error {
	q := `INSERT INTO bill_items (bill_id, fee_type_id, fee_structure_id, description, amount) VALUES ($1,$2,$3,$4,$5)`
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, q, billID, feeTypeID, nullUUID(structureID), desc, amount)
	} else {
		_, err = r.pool.Exec(ctx, q, billID, feeTypeID, nullUUID(structureID), desc, amount)
	}
	return err
}

func (r *feeRepository) GetBill(ctx context.Context, id uuid.UUID) (*BillRecord, error) {
	rec, err := scanBill(r.pool.QueryRow(ctx, r.billSelect()+` AND b.id=$1`, id))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *feeRepository) GetBillByStudentPeriod(ctx context.Context, studentID uuid.UUID, period string) (*BillRecord, error) {
	rec, err := scanBill(r.pool.QueryRow(ctx, r.billSelect()+` AND b.student_id=$1 AND b.bill_period=$2 AND b.status != 'cancelled'`, studentID, period))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *feeRepository) billSearchQuery(f BillSearchParams, count bool) (string, []any) {
	base := r.billSelect()
	args := []any{}
	n := 1
	if f.SessionID != nil {
		base += fmt.Sprintf(" AND b.session_id=$%d", n)
		args = append(args, *f.SessionID)
		n++
	}
	if f.ClassID != nil {
		base += fmt.Sprintf(" AND b.class_id=$%d", n)
		args = append(args, *f.ClassID)
		n++
	}
	if f.SectionID != nil {
		base += fmt.Sprintf(" AND b.section_id=$%d", n)
		args = append(args, *f.SectionID)
		n++
	}
	if f.StudentID != nil {
		base += fmt.Sprintf(" AND b.student_id=$%d", n)
		args = append(args, *f.StudentID)
		n++
	}
	if f.Status != "" {
		base += fmt.Sprintf(" AND b.status=$%d", n)
		args = append(args, f.Status)
		n++
	}
	if f.OverdueOnly {
		base += " AND b.status IN ('pending','partial','overdue') AND b.due_date < CURRENT_DATE"
	}
	if f.Query != "" {
		base += fmt.Sprintf(" AND (s.first_name ILIKE $%d OR s.last_name ILIKE $%d OR s.admission_number ILIKE $%d OR b.invoice_number ILIKE $%d)", n, n, n, n)
		args = append(args, "%"+f.Query+"%")
		n++
	}
	if count {
		return "SELECT COUNT(*) FROM (" + base + ") sub", args
	}
	base += " ORDER BY b.due_date DESC, b.generated_at DESC"
	if f.Limit > 0 {
		base += fmt.Sprintf(" LIMIT $%d OFFSET $%d", n, n+1)
		args = append(args, f.Limit, f.Offset)
	}
	return base, args
}

func (r *feeRepository) SearchBills(ctx context.Context, f BillSearchParams) ([]BillRecord, error) {
	q, args := r.billSearchQuery(f, false)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BillRecord
	for rows.Next() {
		rec, err := scanBill(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func (r *feeRepository) CountBills(ctx context.Context, f BillSearchParams) (int64, error) {
	q, args := r.billSearchQuery(f, true)
	var count int64
	return count, r.pool.QueryRow(ctx, q, args...).Scan(&count)
}

func (r *feeRepository) ListBillItems(ctx context.Context, billID uuid.UUID) ([]BillItemRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT bi.id, bi.fee_type_id, ft.name, bi.description, bi.amount
FROM bill_items bi JOIN fee_types ft ON ft.id=bi.fee_type_id WHERE bi.bill_id=$1`, billID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BillItemRecord
	for rows.Next() {
		var rec BillItemRecord
		if err := rows.Scan(&rec.ID, &rec.FeeTypeID, &rec.FeeTypeName, &rec.Description, &rec.Amount); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *feeRepository) CancelBill(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE student_bills SET status='cancelled', updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *feeRepository) UpdateBillPayment(ctx context.Context, tx pgx.Tx, billID uuid.UUID, paidDelta float64) error {
	q := `
UPDATE student_bills SET paid_amount = paid_amount + $2,
    status = CASE
        WHEN paid_amount + $2 >= total_amount THEN 'paid'
        WHEN paid_amount + $2 > 0 THEN 'partial'
        WHEN due_date < CURRENT_DATE THEN 'overdue'
        ELSE 'pending'
    END,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND status NOT IN ('cancelled','paid')`
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, q, billID, paidDelta)
	} else {
		_, err = r.pool.Exec(ctx, q, billID, paidDelta)
	}
	return err
}

func (r *feeRepository) MarkOverdueBills(ctx context.Context, before time.Time) error {
	_, err := r.pool.Exec(ctx, `
UPDATE student_bills SET status='overdue', updated_at=NOW()
WHERE deleted_at IS NULL AND status IN ('pending','partial') AND due_date < $1`, before)
	return err
}

func (r *feeRepository) CreatePayment(ctx context.Context, tx pgx.Tx, p CreatePaymentParams) (*PaymentRecord, error) {
	var id uuid.UUID
	q := `
INSERT INTO payments (payment_number, student_id, amount, payment_method, collected_by, collection_date, remarks)
VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`
	err := tx.QueryRow(ctx, q, p.PaymentNumber, p.StudentID, p.Amount, p.PaymentMethod, p.CollectedBy, p.CollectionDate, p.Remarks).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetPayment(ctx, id)
}

func (r *feeRepository) CreateAllocation(ctx context.Context, tx pgx.Tx, paymentID, billID uuid.UUID, amount float64) error {
	_, err := tx.Exec(ctx, `INSERT INTO payment_allocations (payment_id, bill_id, amount) VALUES ($1,$2,$3)`, paymentID, billID, amount)
	return err
}

func (r *feeRepository) CreateReceipt(ctx context.Context, tx pgx.Tx, p CreateReceiptParams) (*ReceiptRecord, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `
INSERT INTO receipts (receipt_number, payment_id, student_id, total_amount, qr_token, issued_by)
VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		p.ReceiptNumber, p.PaymentID, p.StudentID, p.TotalAmount, p.QRToken, p.IssuedBy).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetReceipt(ctx, id)
}

func (r *feeRepository) paymentSelect() string {
	return `
SELECT p.id, p.payment_number, p.student_id, s.first_name||' '||s.last_name, p.amount, p.payment_method,
    COALESCE(u.first_name||' '||u.last_name,''), p.collection_date, COALESCE(p.remarks,''), p.status,
    r.id, COALESCE(r.receipt_number,'')
FROM payments p
JOIN students s ON s.id=p.student_id
LEFT JOIN users u ON u.id=p.collected_by
LEFT JOIN receipts r ON r.payment_id=p.id
WHERE p.deleted_at IS NULL`
}

func scanPayment(row pgx.Row) (*PaymentRecord, error) {
	var rec PaymentRecord
	err := row.Scan(&rec.ID, &rec.PaymentNumber, &rec.StudentID, &rec.StudentName, &rec.Amount,
		&rec.PaymentMethod, &rec.CollectorName, &rec.CollectionDate, &rec.Remarks, &rec.Status,
		&rec.ReceiptID, &rec.ReceiptNumber)
	return &rec, err
}

func (r *feeRepository) GetPayment(ctx context.Context, id uuid.UUID) (*PaymentRecord, error) {
	rec, err := scanPayment(r.pool.QueryRow(ctx, r.paymentSelect()+` AND p.id=$1`, id))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *feeRepository) ListPayments(ctx context.Context, f PaymentSearchParams) ([]PaymentRecord, error) {
	q := r.paymentSelect()
	args := []any{}
	n := 1
	if f.StudentID != nil {
		q += fmt.Sprintf(" AND p.student_id=$%d", n)
		args = append(args, *f.StudentID)
		n++
	}
	if !f.From.IsZero() {
		q += fmt.Sprintf(" AND p.collection_date >= $%d", n)
		args = append(args, f.From)
		n++
	}
	if !f.To.IsZero() {
		q += fmt.Sprintf(" AND p.collection_date <= $%d", n)
		args = append(args, f.To)
		n++
	}
	if f.Method != "" {
		q += fmt.Sprintf(" AND p.payment_method=$%d", n)
		args = append(args, f.Method)
		n++
	}
	q += " ORDER BY p.collection_date DESC, p.created_at DESC"
	if f.Limit > 0 {
		q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", n, n+1)
		args = append(args, f.Limit, f.Offset)
	}
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PaymentRecord
	for rows.Next() {
		rec, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func (r *feeRepository) CountPayments(ctx context.Context, f PaymentSearchParams) (int64, error) {
	items, err := r.ListPayments(ctx, PaymentSearchParams{StudentID: f.StudentID, From: f.From, To: f.To, Method: f.Method, Limit: 0})
	if err != nil {
		return 0, err
	}
	return int64(len(items)), nil
}

func (r *feeRepository) receiptSelect() string {
	return `
SELECT r.id, r.receipt_number, r.payment_id, p.payment_number, r.student_id,
    s.first_name||' '||s.last_name, s.admission_number, c.name, sec.name,
    r.total_amount, r.qr_token, r.issued_at, COALESCE(u.first_name||' '||u.last_name,'')
FROM receipts r
JOIN payments p ON p.id=r.payment_id
JOIN students s ON s.id=r.student_id
JOIN classes c ON c.id=s.class_id
JOIN sections sec ON sec.id=s.section_id
LEFT JOIN users u ON u.id=r.issued_by`
}

func scanReceipt(row pgx.Row) (*ReceiptRecord, error) {
	var rec ReceiptRecord
	err := row.Scan(&rec.ID, &rec.ReceiptNumber, &rec.PaymentID, &rec.PaymentNumber, &rec.StudentID,
		&rec.StudentName, &rec.AdmissionNo, &rec.ClassName, &rec.SectionName,
		&rec.TotalAmount, &rec.QRToken, &rec.IssuedAt, &rec.CollectorName)
	return &rec, err
}

func (r *feeRepository) GetReceipt(ctx context.Context, id uuid.UUID) (*ReceiptRecord, error) {
	rec, err := scanReceipt(r.pool.QueryRow(ctx, r.receiptSelect()+` WHERE r.id=$1`, id))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *feeRepository) GetReceiptByToken(ctx context.Context, token string) (*ReceiptRecord, error) {
	rec, err := scanReceipt(r.pool.QueryRow(ctx, r.receiptSelect()+` WHERE r.qr_token=$1`, token))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *feeRepository) ListReceiptsByStudent(ctx context.Context, studentID uuid.UUID, limit int32) ([]ReceiptRecord, error) {
	q := r.receiptSelect() + ` WHERE r.student_id=$1 ORDER BY r.issued_at DESC`
	args := []any{studentID}
	if limit > 0 {
		q += ` LIMIT $2`
		args = append(args, limit)
	}
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReceiptRecord
	for rows.Next() {
		rec, err := scanReceipt(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func (r *feeRepository) ListAllocationsByPayment(ctx context.Context, paymentID uuid.UUID) ([]AllocationRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT pa.bill_id, b.invoice_number, pa.amount
FROM payment_allocations pa JOIN student_bills b ON b.id=pa.bill_id
WHERE pa.payment_id=$1`, paymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AllocationRecord
	for rows.Next() {
		var rec AllocationRecord
		if err := rows.Scan(&rec.BillID, &rec.InvoiceNumber, &rec.Amount); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *feeRepository) RefundPayment(ctx context.Context, id uuid.UUID) error {
	return r.WithTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE payments SET status='refunded', updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
		return err
	})
}

func (r *feeRepository) SumCollection(ctx context.Context, from, to time.Time) (float64, error) {
	var sum float64
	err := r.pool.QueryRow(ctx, `
SELECT COALESCE(SUM(amount),0) FROM payments
WHERE deleted_at IS NULL AND status='completed' AND collection_date BETWEEN $1 AND $2`, from, to).Scan(&sum)
	return sum, err
}

func (r *feeRepository) SumOutstanding(ctx context.Context) (float64, error) {
	var sum float64
	err := r.pool.QueryRow(ctx, `
SELECT COALESCE(SUM(total_amount - paid_amount),0) FROM student_bills
WHERE deleted_at IS NULL AND status IN ('pending','partial','overdue')`).Scan(&sum)
	return sum, err
}

func (r *feeRepository) CountStudentsWithDues(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `
SELECT COUNT(DISTINCT student_id) FROM student_bills
WHERE deleted_at IS NULL AND status IN ('pending','partial','overdue') AND total_amount > paid_amount`).Scan(&count)
	return count, err
}

func (r *feeRepository) CollectionByMethod(ctx context.Context, from, to time.Time) ([]MethodAmountRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT payment_method, COALESCE(SUM(amount),0), COUNT(*)
FROM payments WHERE deleted_at IS NULL AND status='completed' AND collection_date BETWEEN $1 AND $2
GROUP BY payment_method ORDER BY SUM(amount) DESC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []MethodAmountRecord
	for rows.Next() {
		var rec MethodAmountRecord
		if err := rows.Scan(&rec.Method, &rec.Amount, &rec.Count); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *feeRepository) DailyCollectionTrend(ctx context.Context, from, to time.Time) ([]DailyAmountRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT collection_date, COALESCE(SUM(amount),0) FROM payments
WHERE deleted_at IS NULL AND status='completed' AND collection_date BETWEEN $1 AND $2
GROUP BY collection_date ORDER BY collection_date`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDailyAmounts(rows)
}

func (r *feeRepository) DailyDueTrend(ctx context.Context, from, to time.Time) ([]DailyAmountRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT due_date, COALESCE(SUM(total_amount - paid_amount),0) FROM student_bills
WHERE deleted_at IS NULL AND status IN ('pending','partial','overdue') AND due_date BETWEEN $1 AND $2
GROUP BY due_date ORDER BY due_date`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDailyAmounts(rows)
}

func scanDailyAmounts(rows pgx.Rows) ([]DailyAmountRecord, error) {
	var items []DailyAmountRecord
	for rows.Next() {
		var rec DailyAmountRecord
		if err := rows.Scan(&rec.Date, &rec.Amount); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *feeRepository) ListDueStudents(ctx context.Context, f BillSearchParams) ([]DueStudentRecord, error) {
	q := `
SELECT s.id, s.first_name||' '||s.last_name, s.admission_number, c.name, sec.name,
    COALESCE(SUM(b.total_amount - b.paid_amount),0),
    COALESCE(SUM(CASE WHEN b.due_date < CURRENT_DATE THEN b.total_amount - b.paid_amount ELSE 0 END),0),
    COUNT(b.id)
FROM student_bills b
JOIN students s ON s.id=b.student_id
JOIN classes c ON c.id=s.class_id
JOIN sections sec ON sec.id=s.section_id
WHERE b.deleted_at IS NULL AND b.status IN ('pending','partial','overdue') AND b.total_amount > b.paid_amount`
	args := []any{}
	n := 1
	if f.ClassID != nil {
		q += fmt.Sprintf(" AND b.class_id=$%d", n)
		args = append(args, *f.ClassID)
		n++
	}
	if f.SectionID != nil {
		q += fmt.Sprintf(" AND b.section_id=$%d", n)
		args = append(args, *f.SectionID)
		n++
	}
	q += ` GROUP BY s.id, s.first_name, s.last_name, s.admission_number, c.name, sec.name ORDER BY SUM(b.total_amount - b.paid_amount) DESC`
	if f.Limit > 0 {
		q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", n, n+1)
		args = append(args, f.Limit, f.Offset)
	}
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DueStudentRecord
	for rows.Next() {
		var rec DueStudentRecord
		if err := rows.Scan(&rec.StudentID, &rec.StudentName, &rec.AdmissionNo, &rec.ClassName, &rec.SectionName,
			&rec.TotalDue, &rec.OverdueAmount, &rec.BillCount); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *feeRepository) FeeTypeCollection(ctx context.Context, from, to time.Time) ([]FeeTypeAmountRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT ft.name, COALESCE(SUM(pa.amount),0)
FROM payment_allocations pa
JOIN payments p ON p.id=pa.payment_id
JOIN bill_items bi ON bi.bill_id=pa.bill_id
JOIN fee_types ft ON ft.id=bi.fee_type_id
WHERE p.deleted_at IS NULL AND p.status='completed' AND p.collection_date BETWEEN $1 AND $2
GROUP BY ft.name ORDER BY SUM(pa.amount) DESC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FeeTypeAmountRecord
	for rows.Next() {
		var rec FeeTypeAmountRecord
		if err := rows.Scan(&rec.FeeTypeName, &rec.Amount); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}
