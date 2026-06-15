package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdmissionRepository interface {
	NextApplicationNumber(ctx context.Context, year int) (string, error)
	Create(ctx context.Context, params CreateAdmissionParams) (*AdmissionRecord, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AdmissionRecord, error)
	GetByTracking(ctx context.Context, appNo, token string) (*AdmissionRecord, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, notes string, reviewerID *uuid.UUID) error
	UpdatePayment(ctx context.Context, id uuid.UUID, status, ref, receipt string, amount float64) error
	SetAdmitted(ctx context.Context, id, studentID uuid.UUID, parentUserID *uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, f AdmissionSearchParams) ([]AdmissionRecord, error)
	Count(ctx context.Context, f AdmissionSearchParams) (int64, error)
	Stats(ctx context.Context) (*AdmissionStatsRecord, error)
	AddDocument(ctx context.Context, appID uuid.UUID, docType, fileName, fileURL string) (*AdmissionDocumentRecord, error)
	ListDocuments(ctx context.Context, appID uuid.UUID) ([]AdmissionDocumentRecord, error)
	ExportList(ctx context.Context, f AdmissionSearchParams) ([]AdmissionRecord, error)
}

type CreateAdmissionParams struct {
	ApplicationNumber string
	TrackingToken     string
	FirstName, LastName string
	DateOfBirth       time.Time
	Gender, BloodGroup, Religion, Nationality string
	Phone, Email, Address string
	FatherName, FatherPhone, FatherOccupation string
	MotherName, MotherPhone, MotherOccupation string
	GuardianName, GuardianPhone string
	PreviousSchool, PreviousClass, PreviousBoard string
	SessionID, ClassID, SectionID *uuid.UUID
	AdmissionFeeAmount float64
}

type AdmissionRecord struct {
	IDUUID          uuid.UUID
	AppNumber       string
	Token           string
	StatusVal       string
	FirstName, LastName string
	DateOfBirth     time.Time
	Gender, BloodGroup, Religion, Nationality string
	Phone, Email, Address string
	FatherName, FatherPhone, FatherOccupation string
	MotherName, MotherPhone, MotherOccupation string
	GuardianName, GuardianPhone string
	PreviousSchool, PreviousClass, PreviousBoard string
	SessionID, ClassID, SectionID, StudentID *uuid.UUID
	SessionName, ClassName, SectionName string
	AdmissionFeeAmount float64
	PaymentStatus, PaymentReference, ReceiptNumber string
	ReviewNotes, ReviewedByName string
	ReviewedBy *uuid.UUID
	ReviewedAt *time.Time
	CreatedAt, UpdatedAt time.Time
}

type AdmissionDocumentRecord struct {
	ID         uuid.UUID
	AppID      uuid.UUID
	DocType    string
	FileName   string
	FileURL    string
	CreatedAt  time.Time
}

type AdmissionSearchParams struct {
	Query, Status, PaymentStatus string
	SessionID, ClassID *uuid.UUID
	From, To time.Time
	Limit, Offset int32
}

type AdmissionStatsRecord struct {
	Total, Pending, UnderReview, Approved, Admitted, Rejected, Today int64
}

type admissionRepo struct{ pool *pgxpool.Pool }

func NewAdmissionRepository(pool *pgxpool.Pool) AdmissionRepository {
	return &admissionRepo{pool: pool}
}

func (r *admissionRepo) NextApplicationNumber(ctx context.Context, year int) (string, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		INSERT INTO application_sequences (year, last_number) VALUES ($1, 1)
		ON CONFLICT (year) DO UPDATE SET last_number = application_sequences.last_number + 1
		RETURNING last_number`, year).Scan(&n)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("APP-%d-%05d", year, n), nil
}

func generateTrackingToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (r *admissionRepo) Create(ctx context.Context, params CreateAdmissionParams) (*AdmissionRecord, error) {
	token := params.TrackingToken
	if token == "" {
		var err error
		token, err = generateTrackingToken()
		if err != nil {
			return nil, err
		}
	}
	row := r.pool.QueryRow(ctx, `
		INSERT INTO admission_applications (
			application_number, tracking_token, first_name, last_name, date_of_birth, gender,
			blood_group, religion, nationality, phone, email, address,
			father_name, father_phone, father_occupation, mother_name, mother_phone, mother_occupation,
			guardian_name, guardian_phone, previous_school, previous_class, previous_board,
			session_id, class_id, section_id, admission_fee_amount
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27)
		RETURNING id`, params.ApplicationNumber, token, params.FirstName, params.LastName, params.DateOfBirth,
		params.Gender, params.BloodGroup, params.Religion, params.Nationality, params.Phone, params.Email, params.Address,
		params.FatherName, params.FatherPhone, params.FatherOccupation, params.MotherName, params.MotherPhone, params.MotherOccupation,
		params.GuardianName, params.GuardianPhone, params.PreviousSchool, params.PreviousClass, params.PreviousBoard,
		params.SessionID, params.ClassID, params.SectionID, params.AdmissionFeeAmount)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func admissionSelectSQL() string {
	return `
		SELECT a.id, a.application_number, a.tracking_token, a.status,
			a.first_name, a.last_name, a.date_of_birth, a.gender, COALESCE(a.blood_group,''), COALESCE(a.religion,''),
			COALESCE(a.nationality,''), COALESCE(a.phone,''), COALESCE(a.email,''), COALESCE(a.address,''),
			COALESCE(a.father_name,''), COALESCE(a.father_phone,''), COALESCE(a.father_occupation,''),
			COALESCE(a.mother_name,''), COALESCE(a.mother_phone,''), COALESCE(a.mother_occupation,''),
			COALESCE(a.guardian_name,''), COALESCE(a.guardian_phone,''),
			COALESCE(a.previous_school,''), COALESCE(a.previous_class,''), COALESCE(a.previous_board,''),
			a.session_id, a.class_id, a.section_id, a.student_id,
			COALESCE(sess.name,''), COALESCE(c.name,''), COALESCE(sec.name,''),
			a.admission_fee_amount, a.payment_status, COALESCE(a.payment_reference,''), COALESCE(a.receipt_number,''),
			COALESCE(a.review_notes,''), a.reviewed_by, COALESCE(u.first_name||' '||u.last_name,''), a.reviewed_at,
			a.created_at, a.updated_at
		FROM admission_applications a
		LEFT JOIN academic_sessions sess ON sess.id = a.session_id
		LEFT JOIN classes c ON c.id = a.class_id
		LEFT JOIN sections sec ON sec.id = a.section_id
		LEFT JOIN users u ON u.id = a.reviewed_by`
}

func scanAdmission(row pgx.Row) (*AdmissionRecord, error) {
	var rec AdmissionRecord
	var sessID, classID, secID, studentID, reviewedBy *uuid.UUID
	if err := row.Scan(&rec.IDUUID, &rec.AppNumber, &rec.Token, &rec.StatusVal,
		&rec.FirstName, &rec.LastName, &rec.DateOfBirth, &rec.Gender, &rec.BloodGroup, &rec.Religion,
		&rec.Nationality, &rec.Phone, &rec.Email, &rec.Address,
		&rec.FatherName, &rec.FatherPhone, &rec.FatherOccupation,
		&rec.MotherName, &rec.MotherPhone, &rec.MotherOccupation,
		&rec.GuardianName, &rec.GuardianPhone,
		&rec.PreviousSchool, &rec.PreviousClass, &rec.PreviousBoard,
		&sessID, &classID, &secID, &studentID,
		&rec.SessionName, &rec.ClassName, &rec.SectionName,
		&rec.AdmissionFeeAmount, &rec.PaymentStatus, &rec.PaymentReference, &rec.ReceiptNumber,
		&rec.ReviewNotes, &reviewedBy, &rec.ReviewedByName, &rec.ReviewedAt,
		&rec.CreatedAt, &rec.UpdatedAt); err != nil {
		return nil, err
	}
	rec.SessionID, rec.ClassID, rec.SectionID, rec.StudentID = sessID, classID, secID, studentID
	rec.ReviewedBy = reviewedBy
	return &rec, nil
}

func (r *admissionRepo) GetByID(ctx context.Context, id uuid.UUID) (*AdmissionRecord, error) {
	row := r.pool.QueryRow(ctx, admissionSelectSQL()+` WHERE a.id = $1 AND a.deleted_at IS NULL`, id)
	rec, err := scanAdmission(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *admissionRepo) GetByTracking(ctx context.Context, appNo, token string) (*AdmissionRecord, error) {
	row := r.pool.QueryRow(ctx, admissionSelectSQL()+` WHERE a.application_number = $1 AND a.tracking_token = $2 AND a.deleted_at IS NULL`,
		appNo, token)
	rec, err := scanAdmission(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *admissionRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status, notes string, reviewerID *uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE admission_applications SET status = $2, review_notes = $3, reviewed_by = $4, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`, id, status, notes, reviewerID)
	return err
}

func (r *admissionRepo) UpdatePayment(ctx context.Context, id uuid.UUID, status, ref, receipt string, amount float64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE admission_applications SET payment_status = $2, payment_reference = $3, receipt_number = $4,
			admission_fee_amount = CASE WHEN $5 > 0 THEN $5 ELSE admission_fee_amount END, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`, id, status, ref, receipt, amount)
	return err
}

func (r *admissionRepo) SetAdmitted(ctx context.Context, id, studentID uuid.UUID, parentUserID *uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE admission_applications SET status = 'admitted', student_id = $2, parent_user_id = $3, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`, id, studentID, parentUserID)
	return err
}

func (r *admissionRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE admission_applications SET deleted_at = NOW() WHERE id = $1`, id)
	return err
}

func buildAdmissionQuery(f AdmissionSearchParams, countOnly bool) (string, []any) {
	var sb strings.Builder
	args := []any{}
	n := 1
	if countOnly {
		sb.WriteString(`SELECT COUNT(*) FROM admission_applications a WHERE a.deleted_at IS NULL`)
	} else {
		sb.WriteString(admissionSelectSQL() + ` WHERE a.deleted_at IS NULL`)
	}
	if f.Query != "" {
		sb.WriteString(fmt.Sprintf(" AND (a.application_number ILIKE $%d OR a.first_name ILIKE $%d OR a.last_name ILIKE $%d OR a.email ILIKE $%d)", n, n, n, n))
		args = append(args, "%"+f.Query+"%")
		n++
	}
	if f.Status != "" {
		sb.WriteString(fmt.Sprintf(" AND a.status = $%d", n))
		args = append(args, f.Status)
		n++
	}
	if f.PaymentStatus != "" {
		sb.WriteString(fmt.Sprintf(" AND a.payment_status = $%d", n))
		args = append(args, f.PaymentStatus)
		n++
	}
	if f.SessionID != nil {
		sb.WriteString(fmt.Sprintf(" AND a.session_id = $%d", n))
		args = append(args, *f.SessionID)
		n++
	}
	if f.ClassID != nil {
		sb.WriteString(fmt.Sprintf(" AND a.class_id = $%d", n))
		args = append(args, *f.ClassID)
		n++
	}
	if !f.From.IsZero() {
		sb.WriteString(fmt.Sprintf(" AND a.created_at >= $%d", n))
		args = append(args, f.From)
		n++
	}
	if !f.To.IsZero() {
		sb.WriteString(fmt.Sprintf(" AND a.created_at < $%d", n))
		args = append(args, f.To)
		n++
	}
	if !countOnly {
		sb.WriteString(" ORDER BY a.created_at DESC")
		if f.Limit > 0 {
			sb.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", n, n+1))
			args = append(args, f.Limit, f.Offset)
		}
	}
	return sb.String(), args
}

func (r *admissionRepo) Search(ctx context.Context, f AdmissionSearchParams) ([]AdmissionRecord, error) {
	q, args := buildAdmissionQuery(f, false)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AdmissionRecord
	for rows.Next() {
		rec, err := scanAdmission(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func (r *admissionRepo) Count(ctx context.Context, f AdmissionSearchParams) (int64, error) {
	q, args := buildAdmissionQuery(f, true)
	var n int64
	err := r.pool.QueryRow(ctx, q, args...).Scan(&n)
	return n, err
}

func (r *admissionRepo) Stats(ctx context.Context) (*AdmissionStatsRecord, error) {
	var s AdmissionStatsRecord
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM admission_applications WHERE deleted_at IS NULL),
			(SELECT COUNT(*) FROM admission_applications WHERE deleted_at IS NULL AND status = 'pending'),
			(SELECT COUNT(*) FROM admission_applications WHERE deleted_at IS NULL AND status = 'under_review'),
			(SELECT COUNT(*) FROM admission_applications WHERE deleted_at IS NULL AND status = 'approved'),
			(SELECT COUNT(*) FROM admission_applications WHERE deleted_at IS NULL AND status = 'admitted'),
			(SELECT COUNT(*) FROM admission_applications WHERE deleted_at IS NULL AND status = 'rejected'),
			(SELECT COUNT(*) FROM admission_applications WHERE deleted_at IS NULL AND created_at >= CURRENT_DATE)`).
		Scan(&s.Total, &s.Pending, &s.UnderReview, &s.Approved, &s.Admitted, &s.Rejected, &s.Today)
	return &s, err
}

func (r *admissionRepo) AddDocument(ctx context.Context, appID uuid.UUID, docType, fileName, fileURL string) (*AdmissionDocumentRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO admission_documents (application_id, doc_type, file_name, file_url)
		VALUES ($1, $2, $3, $4) RETURNING id, application_id, doc_type, file_name, file_url, created_at`,
		appID, docType, fileName, fileURL)
	var rec AdmissionDocumentRecord
	err := row.Scan(&rec.ID, &rec.AppID, &rec.DocType, &rec.FileName, &rec.FileURL, &rec.CreatedAt)
	return &rec, err
}

func (r *admissionRepo) ListDocuments(ctx context.Context, appID uuid.UUID) ([]AdmissionDocumentRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, application_id, doc_type, file_name, file_url, created_at
		FROM admission_documents WHERE application_id = $1 AND deleted_at IS NULL ORDER BY created_at`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AdmissionDocumentRecord
	for rows.Next() {
		var rec AdmissionDocumentRecord
		if err := rows.Scan(&rec.ID, &rec.AppID, &rec.DocType, &rec.FileName, &rec.FileURL, &rec.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *admissionRepo) ExportList(ctx context.Context, f AdmissionSearchParams) ([]AdmissionRecord, error) {
	f.Limit = 10000
	f.Offset = 0
	return r.Search(ctx, f)
}
