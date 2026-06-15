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

type StudentRepository interface {
	NextAdmissionNumber(ctx context.Context, year int) (string, error)
	Create(ctx context.Context, params CreateStudentParams) (*StudentRecord, error)
	Update(ctx context.Context, id uuid.UUID, params UpdateStudentParams) (*StudentRecord, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*StudentRecord, error)
	UpsertParents(ctx context.Context, params StudentParentParams) error
	GetParents(ctx context.Context, studentID uuid.UUID) (*StudentParentRecord, error)
	CreateDocument(ctx context.Context, studentID uuid.UUID, docType, fileName, fileURL string) (*StudentDocumentRecord, error)
	ListDocuments(ctx context.Context, studentID uuid.UUID) ([]StudentDocumentRecord, error)
	CreatePromotion(ctx context.Context, params PromotionParams) error
	Search(ctx context.Context, filter StudentSearchParams) ([]StudentRecord, error)
	CountSearch(ctx context.Context, filter StudentSearchParams) (int64, error)
	CountAll(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
	CountNewAdmissionsThisMonth(ctx context.Context) (int64, error)
	CountByClass(ctx context.Context) ([]ClassCountRecord, error)
	ListForReport(ctx context.Context, classID, sessionID *uuid.UUID, status string) ([]StudentRecord, error)
	ListAdmissionsReport(ctx context.Context, from, to time.Time) ([]StudentRecord, error)
}

type CreateStudentParams struct {
	AdmissionNumber string
	RollNumber      string
	FirstName       string
	LastName        string
	DateOfBirth     time.Time
	Gender          string
	BloodGroup      string
	Religion        string
	Nationality     string
	PhotoURL        string
	Phone           string
	Email           string
	Address         string
	SessionID       uuid.UUID
	ClassID         uuid.UUID
	SectionID       uuid.UUID
	DepartmentID    *uuid.UUID
	AdmissionDate   time.Time
	Status          string
}

type UpdateStudentParams struct {
	RollNumber    string
	FirstName     string
	LastName      string
	DateOfBirth   time.Time
	Gender        string
	BloodGroup    string
	Religion      string
	Nationality   string
	PhotoURL      string
	Phone         string
	Email         string
	Address       string
	SessionID     uuid.UUID
	ClassID       uuid.UUID
	SectionID     uuid.UUID
	DepartmentID  *uuid.UUID
	AdmissionDate time.Time
	Status        string
}

type StudentParentParams struct {
	StudentID        uuid.UUID
	FatherName       string
	FatherPhone      string
	FatherOccupation string
	MotherName       string
	MotherPhone      string
	MotherOccupation string
	GuardianName     string
	GuardianPhone    string
}

type StudentParentRecord struct {
	FatherName       string
	FatherPhone      string
	FatherOccupation string
	MotherName       string
	MotherPhone      string
	MotherOccupation string
	GuardianName     string
	GuardianPhone    string
}

type StudentDocumentRecord struct {
	ID       uuid.UUID
	DocType  string
	FileName string
	FileURL  string
}

type PromotionParams struct {
	StudentID     uuid.UUID
	PromotionType string
	FromSessionID *uuid.UUID
	ToSessionID   uuid.UUID
	FromClassID   *uuid.UUID
	ToClassID     uuid.UUID
	FromSectionID *uuid.UUID
	ToSectionID   uuid.UUID
	PromotionDate time.Time
	Notes         string
	CreatedBy     uuid.UUID
}

type StudentSearchParams struct {
	AdmissionNumber string
	RollNumber      string
	Name            string
	ClassID         *uuid.UUID
	SectionID       *uuid.UUID
	SessionID       *uuid.UUID
	Limit           int32
	Offset          int32
}

type StudentRecord struct {
	ID              uuid.UUID
	AdmissionNumber string
	RollNumber      string
	FirstName       string
	LastName        string
	DateOfBirth     time.Time
	Gender          string
	BloodGroup      string
	Religion        string
	Nationality     string
	PhotoURL        string
	Phone           string
	Email           string
	Address         string
	SessionID       uuid.UUID
	SessionName     string
	ClassID         uuid.UUID
	ClassName       string
	SectionID       uuid.UUID
	SectionName     string
	DepartmentID    *uuid.UUID
	DepartmentName  string
	AdmissionDate   time.Time
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ClassCountRecord struct {
	ClassID      uuid.UUID
	ClassName    string
	StudentCount int64
}

type studentRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentRepository{pool: pool}
}

func (r *studentRepository) NextAdmissionNumber(ctx context.Context, year int) (string, error) {
	var seq int
	err := r.pool.QueryRow(ctx, `
INSERT INTO admission_sequences (year, last_number) VALUES ($1, 1)
ON CONFLICT (year) DO UPDATE SET last_number = admission_sequences.last_number + 1
RETURNING last_number`, year).Scan(&seq)
	if err != nil {
		return "", err
	}
	return formatAdmissionNumber(year, seq), nil
}

func (r *studentRepository) Create(ctx context.Context, params CreateStudentParams) (*StudentRecord, error) {
	var id uuid.UUID
	var deptID pgtype.UUID
	if params.DepartmentID != nil {
		deptID = pgtype.UUID{Bytes: *params.DepartmentID, Valid: true}
	}
	err := r.pool.QueryRow(ctx, `
INSERT INTO students (
    admission_number, roll_number, first_name, last_name, date_of_birth, gender,
    blood_group, religion, nationality, photo_url, phone, email, address,
    session_id, class_id, section_id, department_id, admission_date, status
) VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,''),NULLIF($8,''),NULLIF($9,''),NULLIF($10,''),
NULLIF($11,''),NULLIF($12,''),NULLIF($13,''),$14,$15,$16,$17,$18,$19) RETURNING id`,
		params.AdmissionNumber, params.RollNumber, params.FirstName, params.LastName,
		params.DateOfBirth, params.Gender, params.BloodGroup, params.Religion, params.Nationality,
		params.PhotoURL, params.Phone, params.Email, params.Address,
		params.SessionID, params.ClassID, params.SectionID, deptID, params.AdmissionDate, params.Status,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *studentRepository) Update(ctx context.Context, id uuid.UUID, params UpdateStudentParams) (*StudentRecord, error) {
	var deptID pgtype.UUID
	if params.DepartmentID != nil {
		deptID = pgtype.UUID{Bytes: *params.DepartmentID, Valid: true}
	}
	tag, err := r.pool.Exec(ctx, `
UPDATE students SET
    roll_number=NULLIF($2,''), first_name=$3, last_name=$4, date_of_birth=$5, gender=$6,
    blood_group=NULLIF($7,''), religion=NULLIF($8,''), nationality=NULLIF($9,''),
    photo_url=CASE WHEN $10='' THEN photo_url ELSE $10 END,
    phone=NULLIF($11,''), email=NULLIF($12,''), address=NULLIF($13,''),
    session_id=$14, class_id=$15, section_id=$16, department_id=$17,
    admission_date=$18, status=$19, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL`,
		id, params.RollNumber, params.FirstName, params.LastName, params.DateOfBirth, params.Gender,
		params.BloodGroup, params.Religion, params.Nationality, params.PhotoURL,
		params.Phone, params.Email, params.Address,
		params.SessionID, params.ClassID, params.SectionID, deptID, params.AdmissionDate, params.Status)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	return r.GetByID(ctx, id)
}

func (r *studentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE students SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func scanStudent(row pgx.Row) (*StudentRecord, error) {
	var s StudentRecord
	var roll, blood, religion, nationality, photo, phone, email, address, deptName pgtype.Text
	var deptID pgtype.UUID
	if err := row.Scan(
		&s.ID, &s.AdmissionNumber, &roll, &s.FirstName, &s.LastName, &s.DateOfBirth, &s.Gender,
		&blood, &religion, &nationality, &photo, &phone, &email, &address,
		&s.SessionID, &s.ClassID, &s.SectionID, &deptID, &s.AdmissionDate, &s.Status,
		&s.CreatedAt, &s.UpdatedAt,
		&s.SessionName, &s.ClassName, &s.SectionName, &deptName,
	); err != nil {
		return nil, err
	}
	s.RollNumber = roll.String
	s.BloodGroup = blood.String
	s.Religion = religion.String
	s.Nationality = nationality.String
	s.PhotoURL = photo.String
	s.Phone = phone.String
	s.Email = email.String
	s.Address = address.String
	s.DepartmentName = deptName.String
	if deptID.Valid {
		id := uuid.UUID(deptID.Bytes)
		s.DepartmentID = &id
	}
	return &s, nil
}

const studentSelect = `
SELECT st.id, st.admission_number, st.roll_number, st.first_name, st.last_name, st.date_of_birth, st.gender,
       st.blood_group, st.religion, st.nationality, st.photo_url, st.phone, st.email, st.address,
       st.session_id, st.class_id, st.section_id, st.department_id, st.admission_date, st.status,
       st.created_at, st.updated_at,
       sess.name, c.name, sec.name, d.name
FROM students st
JOIN academic_sessions sess ON sess.id = st.session_id
JOIN classes c ON c.id = st.class_id
JOIN sections sec ON sec.id = st.section_id
LEFT JOIN departments d ON d.id = st.department_id`

func (r *studentRepository) GetByID(ctx context.Context, id uuid.UUID) (*StudentRecord, error) {
	row := r.pool.QueryRow(ctx, studentSelect+` WHERE st.id=$1 AND st.deleted_at IS NULL`, id)
	s, err := scanStudent(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *studentRepository) UpsertParents(ctx context.Context, params StudentParentParams) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO student_parents (student_id,father_name,father_phone,father_occupation,mother_name,mother_phone,mother_occupation,guardian_name,guardian_phone)
VALUES ($1,NULLIF($2,''),NULLIF($3,''),NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),NULLIF($8,''),NULLIF($9,''))
ON CONFLICT (student_id) DO UPDATE SET
father_name=NULLIF($2,''),father_phone=NULLIF($3,''),father_occupation=NULLIF($4,''),
mother_name=NULLIF($5,''),mother_phone=NULLIF($6,''),mother_occupation=NULLIF($7,''),
guardian_name=NULLIF($8,''),guardian_phone=NULLIF($9,''),updated_at=NOW()`,
		params.StudentID, params.FatherName, params.FatherPhone, params.FatherOccupation,
		params.MotherName, params.MotherPhone, params.MotherOccupation, params.GuardianName, params.GuardianPhone)
	return err
}

func (r *studentRepository) GetParents(ctx context.Context, studentID uuid.UUID) (*StudentParentRecord, error) {
	row := r.pool.QueryRow(ctx, `
SELECT father_name,father_phone,father_occupation,mother_name,mother_phone,mother_occupation,guardian_name,guardian_phone
FROM student_parents WHERE student_id=$1`, studentID)
	var p StudentParentRecord
	var fields [8]pgtype.Text
	if err := row.Scan(&fields[0], &fields[1], &fields[2], &fields[3], &fields[4], &fields[5], &fields[6], &fields[7]); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	p.FatherName = fields[0].String
	p.FatherPhone = fields[1].String
	p.FatherOccupation = fields[2].String
	p.MotherName = fields[3].String
	p.MotherPhone = fields[4].String
	p.MotherOccupation = fields[5].String
	p.GuardianName = fields[6].String
	p.GuardianPhone = fields[7].String
	return &p, nil
}

func (r *studentRepository) CreateDocument(ctx context.Context, studentID uuid.UUID, docType, fileName, fileURL string) (*StudentDocumentRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO student_documents (student_id,doc_type,file_name,file_url) VALUES ($1,$2,$3,$4) RETURNING id`,
		studentID, docType, fileName, fileURL).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &StudentDocumentRecord{ID: id, DocType: docType, FileName: fileName, FileURL: fileURL}, nil
}

func (r *studentRepository) ListDocuments(ctx context.Context, studentID uuid.UUID) ([]StudentDocumentRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id,doc_type,file_name,file_url FROM student_documents WHERE student_id=$1 AND deleted_at IS NULL ORDER BY created_at`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []StudentDocumentRecord
	for rows.Next() {
		var d StudentDocumentRecord
		if err := rows.Scan(&d.ID, &d.DocType, &d.FileName, &d.FileURL); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *studentRepository) CreatePromotion(ctx context.Context, params PromotionParams) error {
	var fromSess, fromClass, fromSec pgtype.UUID
	if params.FromSessionID != nil {
		fromSess = pgtype.UUID{Bytes: *params.FromSessionID, Valid: true}
	}
	if params.FromClassID != nil {
		fromClass = pgtype.UUID{Bytes: *params.FromClassID, Valid: true}
	}
	if params.FromSectionID != nil {
		fromSec = pgtype.UUID{Bytes: *params.FromSectionID, Valid: true}
	}
	var createdBy pgtype.UUID
	if params.CreatedBy != uuid.Nil {
		createdBy = pgtype.UUID{Bytes: params.CreatedBy, Valid: true}
	}
	_, err := r.pool.Exec(ctx, `
INSERT INTO student_promotions (student_id,promotion_type,from_session_id,to_session_id,from_class_id,to_class_id,from_section_id,to_section_id,promotion_date,notes,created_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NULLIF($10,''),$11)`,
		params.StudentID, params.PromotionType, fromSess, params.ToSessionID,
		fromClass, params.ToClassID, fromSec, params.ToSectionID,
		params.PromotionDate, params.Notes, createdBy)
	return err
}

func (r *studentRepository) searchWhere(filter StudentSearchParams) (string, []any) {
	args := []any{
		filter.AdmissionNumber, filter.RollNumber, filter.Name,
		uuidOrNil(filter.ClassID), uuidOrNil(filter.SectionID), uuidOrNil(filter.SessionID),
	}
	where := ` WHERE st.deleted_at IS NULL
AND ($1::text = '' OR st.admission_number ILIKE '%' || $1 || '%')
AND ($2::text = '' OR st.roll_number ILIKE '%' || $2 || '%')
AND ($3::text = '' OR st.first_name ILIKE '%' || $3 || '%' OR st.last_name ILIKE '%' || $3 || '%' OR (st.first_name || ' ' || st.last_name) ILIKE '%' || $3 || '%')
AND ($4::uuid IS NULL OR st.class_id = $4)
AND ($5::uuid IS NULL OR st.section_id = $5)
AND ($6::uuid IS NULL OR st.session_id = $6)`
	return where, args
}

func uuidOrNil(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}

func (r *studentRepository) Search(ctx context.Context, filter StudentSearchParams) ([]StudentRecord, error) {
	where, args := r.searchWhere(filter)
	args = append(args, filter.Limit, filter.Offset)
	q := studentSelect + where + ` ORDER BY st.created_at DESC LIMIT $7 OFFSET $8`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []StudentRecord
	for rows.Next() {
		s, err := scanStudent(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *s)
	}
	return items, rows.Err()
}

func (r *studentRepository) CountSearch(ctx context.Context, filter StudentSearchParams) (int64, error) {
	where, args := r.searchWhere(filter)
	q := `SELECT COUNT(*)::bigint FROM students st` + where
	var count int64
	err := r.pool.QueryRow(ctx, q, args...).Scan(&count)
	return count, err
}

func (r *studentRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*)::bigint FROM students WHERE deleted_at IS NULL`).Scan(&count)
	return count, err
}

func (r *studentRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*)::bigint FROM students WHERE deleted_at IS NULL AND status='active'`).Scan(&count)
	return count, err
}

func (r *studentRepository) CountNewAdmissionsThisMonth(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `
SELECT COUNT(*)::bigint FROM students WHERE deleted_at IS NULL
AND admission_date >= date_trunc('month', CURRENT_DATE)::date`).Scan(&count)
	return count, err
}

func (r *studentRepository) CountByClass(ctx context.Context) ([]ClassCountRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT c.id, c.name, COUNT(st.id)::bigint
FROM classes c
LEFT JOIN students st ON st.class_id=c.id AND st.deleted_at IS NULL AND st.status='active'
WHERE c.deleted_at IS NULL
GROUP BY c.id, c.name ORDER BY c.sort_order, c.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ClassCountRecord
	for rows.Next() {
		var rec ClassCountRecord
		if err := rows.Scan(&rec.ClassID, &rec.ClassName, &rec.StudentCount); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *studentRepository) ListForReport(ctx context.Context, classID, sessionID *uuid.UUID, status string) ([]StudentRecord, error) {
	q := studentSelect + ` WHERE st.deleted_at IS NULL
AND ($1::uuid IS NULL OR st.class_id=$1)
AND ($2::uuid IS NULL OR st.session_id=$2)
AND ($3::text='' OR st.status=$3)
ORDER BY c.sort_order, sec.name, st.roll_number, st.last_name`
	rows, err := r.pool.Query(ctx, q, uuidOrNil(classID), uuidOrNil(sessionID), status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []StudentRecord
	for rows.Next() {
		s, err := scanStudent(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *s)
	}
	return items, rows.Err()
}

func (r *studentRepository) ListAdmissionsReport(ctx context.Context, from, to time.Time) ([]StudentRecord, error) {
	q := studentSelect + ` WHERE st.deleted_at IS NULL AND st.admission_date >= $1 AND st.admission_date <= $2 ORDER BY st.admission_date DESC`
	rows, err := r.pool.Query(ctx, q, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []StudentRecord
	for rows.Next() {
		s, err := scanStudent(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *s)
	}
	return items, rows.Err()
}

func formatAdmissionNumber(year int, seq int) string {
	return fmt.Sprintf("ADM-%d-%05d", year, seq)
}
