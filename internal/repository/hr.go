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

type HRRepository interface {
	// Departments
	CreateDepartment(ctx context.Context, name, slug, description, deptType string) (*HRDepartmentRecord, error)
	UpdateDepartment(ctx context.Context, id uuid.UUID, name, slug, description, deptType string) (*HRDepartmentRecord, error)
	SoftDeleteDepartment(ctx context.Context, id uuid.UUID) error
	GetDepartmentByID(ctx context.Context, id uuid.UUID) (*HRDepartmentRecord, error)
	ListDepartments(ctx context.Context, deptType string) ([]HRDepartmentRecord, error)

	// Designations
	CreateDesignation(ctx context.Context, name, slug, category, description string) (*DesignationRecord, error)
	UpdateDesignation(ctx context.Context, id uuid.UUID, name, slug, category, description string) (*DesignationRecord, error)
	SoftDeleteDesignation(ctx context.Context, id uuid.UUID) error
	GetDesignationByID(ctx context.Context, id uuid.UUID) (*DesignationRecord, error)
	ListDesignations(ctx context.Context) ([]DesignationRecord, error)

	// Teachers
	NextEmployeeID(ctx context.Context, entityType string, year int) (string, error)
	CreateTeacher(ctx context.Context, p CreateTeacherParams) (*TeacherRecord, error)
	UpdateTeacher(ctx context.Context, id uuid.UUID, p UpdateTeacherParams) (*TeacherRecord, error)
	SoftDeleteTeacher(ctx context.Context, id uuid.UUID) error
	GetTeacherByID(ctx context.Context, id uuid.UUID) (*TeacherRecord, error)
	GetTeacherByUserID(ctx context.Context, userID uuid.UUID) (*TeacherRecord, error)
	GetTeacherByEmail(ctx context.Context, email string) (*TeacherRecord, error)
	SearchTeachers(ctx context.Context, f TeacherSearchParams) ([]TeacherRecord, error)
	CountTeachers(ctx context.Context, f TeacherSearchParams) (int64, error)
	CountActiveTeachers(ctx context.Context) (int64, error)
	ListTeachersReport(ctx context.Context, f TeacherSearchParams) ([]TeacherRecord, error)

	// Teacher assignments
	ClearTeacherAssignments(ctx context.Context, teacherID uuid.UUID) error
	AddTeacherAssignment(ctx context.Context, teacherID uuid.UUID, subjectID, classID, sectionID *uuid.UUID) error
	ListTeacherAssignments(ctx context.Context, teacherID uuid.UUID) ([]TeacherAssignmentRecord, error)
	CountTeacherClasses(ctx context.Context, teacherID uuid.UUID) (int64, error)
	CountTeacherSubjects(ctx context.Context, teacherID uuid.UUID) (int64, error)
	ListTodaySchedule(ctx context.Context, teacherID uuid.UUID, dayOfWeek int) ([]ScheduleRecord, error)

	// Teacher documents
	CreateTeacherDocument(ctx context.Context, teacherID uuid.UUID, docType, fileName, fileURL string) error
	ListTeacherDocuments(ctx context.Context, teacherID uuid.UUID) ([]DocumentRecord, error)

	// Staff
	CreateStaff(ctx context.Context, p CreateStaffParams) (*StaffRecord, error)
	UpdateStaff(ctx context.Context, id uuid.UUID, p UpdateStaffParams) (*StaffRecord, error)
	SoftDeleteStaff(ctx context.Context, id uuid.UUID) error
	GetStaffByID(ctx context.Context, id uuid.UUID) (*StaffRecord, error)
	SearchStaff(ctx context.Context, f StaffSearchParams) ([]StaffRecord, error)
	CountStaff(ctx context.Context, f StaffSearchParams) (int64, error)
	CountActiveStaff(ctx context.Context) (int64, error)
	ListStaffReport(ctx context.Context, f StaffSearchParams) ([]StaffRecord, error)
	CreateStaffDocument(ctx context.Context, staffID uuid.UUID, docType, fileName, fileURL string) error
	ListStaffDocuments(ctx context.Context, staffID uuid.UUID) ([]DocumentRecord, error)
}

type HRDepartmentRecord struct {
	ID           uuid.UUID
	Name         string
	Slug         string
	Description  string
	DeptType     string
	TeacherCount int64
	StaffCount   int64
}

type DesignationRecord struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Category    string
	Description string
}

type CreateTeacherParams struct {
	EmployeeID     string
	FirstName      string
	LastName       string
	PhotoURL       string
	Gender         string
	DateOfBirth    *time.Time
	BloodGroup     string
	Religion       string
	Nationality    string
	Phone          string
	Email          string
	Address        string
	NationalID     string
	JoiningDate    time.Time
	DepartmentID   *uuid.UUID
	DesignationID  *uuid.UUID
	Qualification  string
	Experience     string
	Salary         float64
	EmploymentType string
	Status         string
}

type UpdateTeacherParams = CreateTeacherParams

type TeacherRecord struct {
	ID              uuid.UUID
	EmployeeID      string
	UserID          *uuid.UUID
	FirstName       string
	LastName        string
	PhotoURL        string
	Gender          string
	DateOfBirth     *time.Time
	BloodGroup      string
	Religion        string
	Nationality     string
	Phone           string
	Email           string
	Address         string
	NationalID      string
	JoiningDate     time.Time
	DepartmentID    *uuid.UUID
	DepartmentName  string
	DesignationID   *uuid.UUID
	DesignationName string
	Qualification   string
	Experience      string
	Salary          float64
	EmploymentType  string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type TeacherAssignmentRecord struct {
	ID          uuid.UUID
	SubjectID   *uuid.UUID
	SubjectName string
	ClassID     *uuid.UUID
	ClassName   string
	SectionID   *uuid.UUID
	SectionName string
}

type ScheduleRecord struct {
	ID          uuid.UUID
	SubjectName string
	ClassName   string
	SectionName string
	StartTime   string
	EndTime     string
	Room        string
}

type DocumentRecord struct {
	ID       uuid.UUID
	DocType  string
	FileName string
	FileURL  string
}

type TeacherSearchParams struct {
	Query         string
	DepartmentID  *uuid.UUID
	DesignationID *uuid.UUID
	Status        string
	Limit         int32
	Offset        int32
}

type CreateStaffParams struct {
	EmployeeID    string
	FirstName     string
	LastName      string
	PhotoURL      string
	Phone         string
	Email         string
	Address       string
	DepartmentID  *uuid.UUID
	DesignationID *uuid.UUID
	Salary        float64
	JoiningDate   time.Time
	Status        string
}

type UpdateStaffParams = CreateStaffParams

type StaffRecord struct {
	ID              uuid.UUID
	EmployeeID      string
	FirstName       string
	LastName        string
	PhotoURL        string
	Phone           string
	Email           string
	Address         string
	DepartmentID    *uuid.UUID
	DepartmentName  string
	DesignationID   *uuid.UUID
	DesignationName string
	Salary          float64
	JoiningDate     time.Time
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type StaffSearchParams struct {
	Query        string
	DepartmentID *uuid.UUID
	Status       string
	Limit        int32
	Offset       int32
}

type hrRepository struct{ pool *pgxpool.Pool }

func NewHRRepository(pool *pgxpool.Pool) HRRepository { return &hrRepository{pool: pool} }

func formatEmployeeID(entityType string, year, seq int) string {
	prefix := "EMP"
	if entityType == "teacher" {
		prefix = "TCH"
	} else if entityType == "staff" {
		prefix = "STF"
	}
	return fmt.Sprintf("%s-%d-%05d", prefix, year, seq)
}

func (r *hrRepository) NextEmployeeID(ctx context.Context, entityType string, year int) (string, error) {
	var seq int
	err := r.pool.QueryRow(ctx, `
INSERT INTO employee_sequences (entity_type, year, last_number) VALUES ($1, $2, 1)
ON CONFLICT (entity_type, year) DO UPDATE SET last_number = employee_sequences.last_number + 1
RETURNING last_number`, entityType, year).Scan(&seq)
	if err != nil {
		return "", err
	}
	return formatEmployeeID(entityType, year, seq), nil
}

func nullableUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func scanOptionalUUID(v pgtype.UUID) *uuid.UUID {
	if !v.Valid {
		return nil
	}
	id := uuid.UUID(v.Bytes)
	return &id
}

const teacherSelect = `
SELECT t.id, t.employee_id, t.user_id, t.first_name, t.last_name, t.photo_url, t.gender, t.date_of_birth,
       t.blood_group, t.religion, t.nationality, t.phone, t.email, t.address, t.national_id, t.joining_date,
       t.department_id, d.name, t.designation_id, des.name, t.qualification, t.experience, t.salary,
       t.employment_type, t.status, t.created_at, t.updated_at
FROM teachers t
LEFT JOIN departments d ON d.id = t.department_id
LEFT JOIN designations des ON des.id = t.designation_id`

func scanTeacher(row pgx.Row) (*TeacherRecord, error) {
	var t TeacherRecord
	var photo, blood, religion, nationality, phone, email, address, nid, qual, exp pgtype.Text
	var dob pgtype.Date
	var deptID, desID, userID pgtype.UUID
	var deptName, desName pgtype.Text
	var salary pgtype.Numeric
	if err := row.Scan(
		&t.ID, &t.EmployeeID, &userID, &t.FirstName, &t.LastName, &photo, &t.Gender, &dob,
		&blood, &religion, &nationality, &phone, &email, &address, &nid, &t.JoiningDate,
		&deptID, &deptName, &desID, &desName, &qual, &exp, &salary,
		&t.EmploymentType, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		return nil, err
	}
	t.PhotoURL = photo.String
	t.BloodGroup = blood.String
	t.Religion = religion.String
	t.Nationality = nationality.String
	t.Phone = phone.String
	t.Email = email.String
	t.Address = address.String
	t.NationalID = nid.String
	t.Qualification = qual.String
	t.Experience = exp.String
	t.DepartmentName = deptName.String
	t.DesignationName = desName.String
	t.DepartmentID = scanOptionalUUID(deptID)
	t.DesignationID = scanOptionalUUID(desID)
	t.UserID = scanOptionalUUID(userID)
	if dob.Valid {
		t.DateOfBirth = &dob.Time
	}
	if salary.Valid {
		f, _ := salary.Float64Value()
		t.Salary = f.Float64
	}
	return &t, nil
}

func (r *hrRepository) teacherWhere(f TeacherSearchParams) (string, []any) {
	args := []any{f.Query, uuidOrNil(f.DepartmentID), uuidOrNil(f.DesignationID), f.Status}
	where := ` WHERE t.deleted_at IS NULL
AND ($1::text = '' OR t.first_name ILIKE '%' || $1 || '%' OR t.last_name ILIKE '%' || $1 || '%'
     OR t.employee_id ILIKE '%' || $1 || '%' OR t.email ILIKE '%' || $1 || '%')
AND ($2::uuid IS NULL OR t.department_id = $2)
AND ($3::uuid IS NULL OR t.designation_id = $3)
AND ($4::text = '' OR t.status = $4)`
	return where, args
}

func (r *hrRepository) CreateDepartment(ctx context.Context, name, slug, description, deptType string) (*HRDepartmentRecord, error) {
	if deptType == "" {
		deptType = "employee"
	}
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO departments (name, slug, description, dept_type) VALUES ($1,$2,NULLIF($3,''),$4) RETURNING id`,
		name, slug, description, deptType).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetDepartmentByID(ctx, id)
}

func (r *hrRepository) UpdateDepartment(ctx context.Context, id uuid.UUID, name, slug, description, deptType string) (*HRDepartmentRecord, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE departments SET name=$2, slug=$3, description=NULLIF($4,''), dept_type=$5, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL`, id, name, slug, description, deptType)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	return r.GetDepartmentByID(ctx, id)
}

func (r *hrRepository) SoftDeleteDepartment(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE departments SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *hrRepository) GetDepartmentByID(ctx context.Context, id uuid.UUID) (*HRDepartmentRecord, error) {
	row := r.pool.QueryRow(ctx, `
SELECT d.id, d.name, d.slug, COALESCE(d.description,''), COALESCE(d.dept_type,'employee'),
       (SELECT COUNT(*) FROM teachers t WHERE t.department_id=d.id AND t.deleted_at IS NULL),
       (SELECT COUNT(*) FROM staffs s WHERE s.department_id=d.id AND s.deleted_at IS NULL)
FROM departments d WHERE d.id=$1 AND d.deleted_at IS NULL`, id)
	var d HRDepartmentRecord
	if err := row.Scan(&d.ID, &d.Name, &d.Slug, &d.Description, &d.DeptType, &d.TeacherCount, &d.StaffCount); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *hrRepository) ListDepartments(ctx context.Context, deptType string) ([]HRDepartmentRecord, error) {
	q := `
SELECT d.id, d.name, d.slug, COALESCE(d.description,''), COALESCE(d.dept_type,'employee'),
       (SELECT COUNT(*) FROM teachers t WHERE t.department_id=d.id AND t.deleted_at IS NULL),
       (SELECT COUNT(*) FROM staffs s WHERE s.department_id=d.id AND s.deleted_at IS NULL)
FROM departments d WHERE d.deleted_at IS NULL`
	args := []any{}
	if deptType != "" {
		q += ` AND d.dept_type = $1`
		args = append(args, deptType)
	}
	q += ` ORDER BY d.name`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []HRDepartmentRecord
	for rows.Next() {
		var d HRDepartmentRecord
		if err := rows.Scan(&d.ID, &d.Name, &d.Slug, &d.Description, &d.DeptType, &d.TeacherCount, &d.StaffCount); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *hrRepository) CreateDesignation(ctx context.Context, name, slug, category, description string) (*DesignationRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO designations (name, slug, category, description) VALUES ($1,$2,NULLIF($3,'general'),NULLIF($4,'')) RETURNING id`,
		name, slug, category, description).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetDesignationByID(ctx, id)
}

func (r *hrRepository) UpdateDesignation(ctx context.Context, id uuid.UUID, name, slug, category, description string) (*DesignationRecord, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE designations SET name=$2, slug=$3, category=$4, description=NULLIF($5,''), updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL`, id, name, slug, category, description)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	return r.GetDesignationByID(ctx, id)
}

func (r *hrRepository) SoftDeleteDesignation(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE designations SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *hrRepository) GetDesignationByID(ctx context.Context, id uuid.UUID) (*DesignationRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,name,slug,category,COALESCE(description,'') FROM designations WHERE id=$1 AND deleted_at IS NULL`, id)
	var d DesignationRecord
	if err := row.Scan(&d.ID, &d.Name, &d.Slug, &d.Category, &d.Description); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *hrRepository) ListDesignations(ctx context.Context) ([]DesignationRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,slug,category,COALESCE(description,'') FROM designations WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DesignationRecord
	for rows.Next() {
		var d DesignationRecord
		if err := rows.Scan(&d.ID, &d.Name, &d.Slug, &d.Category, &d.Description); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *hrRepository) CreateTeacher(ctx context.Context, p CreateTeacherParams) (*TeacherRecord, error) {
	var id uuid.UUID
	var dob pgtype.Date
	if p.DateOfBirth != nil {
		dob = pgtype.Date{Time: *p.DateOfBirth, Valid: true}
	}
	var salary pgtype.Numeric
	if p.Salary > 0 {
		_ = salary.Scan(fmt.Sprintf("%.2f", p.Salary))
	}
	err := r.pool.QueryRow(ctx, `
INSERT INTO teachers (employee_id, first_name, last_name, photo_url, gender, date_of_birth, blood_group, religion,
    nationality, phone, email, address, national_id, joining_date, department_id, designation_id, qualification,
    experience, salary, employment_type, status)
VALUES ($1,$2,$3,NULLIF($4,''),$5,$6,NULLIF($7,''),NULLIF($8,''),NULLIF($9,''),NULLIF($10,''),NULLIF($11,''),
NULLIF($12,''),NULLIF($13,''),$14,$15,$16,NULLIF($17,''),NULLIF($18,''),$19,$20,$21) RETURNING id`,
		p.EmployeeID, p.FirstName, p.LastName, p.PhotoURL, p.Gender, dob, p.BloodGroup, p.Religion, p.Nationality,
		p.Phone, p.Email, p.Address, p.NationalID, p.JoiningDate, nullableUUID(p.DepartmentID), nullableUUID(p.DesignationID),
		p.Qualification, p.Experience, salary, p.EmploymentType, p.Status,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetTeacherByID(ctx, id)
}

func (r *hrRepository) UpdateTeacher(ctx context.Context, id uuid.UUID, p UpdateTeacherParams) (*TeacherRecord, error) {
	var dob pgtype.Date
	if p.DateOfBirth != nil {
		dob = pgtype.Date{Time: *p.DateOfBirth, Valid: true}
	}
	var salary pgtype.Numeric
	if p.Salary > 0 {
		_ = salary.Scan(fmt.Sprintf("%.2f", p.Salary))
	}
	tag, err := r.pool.Exec(ctx, `
UPDATE teachers SET first_name=$2, last_name=$3,
    photo_url=CASE WHEN $4='' THEN photo_url ELSE $4 END,
    gender=$5, date_of_birth=$6, blood_group=NULLIF($7,''), religion=NULLIF($8,''), nationality=NULLIF($9,''),
    phone=NULLIF($10,''), email=NULLIF($11,''), address=NULLIF($12,''), national_id=NULLIF($13,''),
    joining_date=$14, department_id=$15, designation_id=$16, qualification=NULLIF($17,''),
    experience=NULLIF($18,''), salary=$19, employment_type=$20, status=$21, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL`,
		id, p.FirstName, p.LastName, p.PhotoURL, p.Gender, dob, p.BloodGroup, p.Religion, p.Nationality,
		p.Phone, p.Email, p.Address, p.NationalID, p.JoiningDate, nullableUUID(p.DepartmentID), nullableUUID(p.DesignationID),
		p.Qualification, p.Experience, salary, p.EmploymentType, p.Status)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	return r.GetTeacherByID(ctx, id)
}

func (r *hrRepository) SoftDeleteTeacher(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE teachers SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *hrRepository) GetTeacherByID(ctx context.Context, id uuid.UUID) (*TeacherRecord, error) {
	row := r.pool.QueryRow(ctx, teacherSelect+` WHERE t.id=$1 AND t.deleted_at IS NULL`, id)
	t, err := scanTeacher(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (r *hrRepository) GetTeacherByEmail(ctx context.Context, email string) (*TeacherRecord, error) {
	row := r.pool.QueryRow(ctx, teacherSelect+` WHERE LOWER(t.email)=LOWER($1) AND t.deleted_at IS NULL`, email)
	t, err := scanTeacher(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (r *hrRepository) CountActiveTeachers(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*)::bigint FROM teachers WHERE deleted_at IS NULL AND status='active'`).Scan(&count)
	return count, err
}

func (r *hrRepository) CountActiveStaff(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*)::bigint FROM staffs WHERE deleted_at IS NULL AND status='active'`).Scan(&count)
	return count, err
}

func (r *hrRepository) GetTeacherByUserID(ctx context.Context, userID uuid.UUID) (*TeacherRecord, error) {
	row := r.pool.QueryRow(ctx, teacherSelect+` WHERE t.user_id=$1 AND t.deleted_at IS NULL`, userID)
	t, err := scanTeacher(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (r *hrRepository) SearchTeachers(ctx context.Context, f TeacherSearchParams) ([]TeacherRecord, error) {
	where, args := r.teacherWhere(f)
	args = append(args, f.Limit, f.Offset)
	rows, err := r.pool.Query(ctx, teacherSelect+where+` ORDER BY t.created_at DESC LIMIT $5 OFFSET $6`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TeacherRecord
	for rows.Next() {
		t, err := scanTeacher(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *t)
	}
	return items, rows.Err()
}

func (r *hrRepository) CountTeachers(ctx context.Context, f TeacherSearchParams) (int64, error) {
	where, args := r.teacherWhere(f)
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*)::bigint FROM teachers t`+where, args...).Scan(&count)
	return count, err
}

func (r *hrRepository) ListTeachersReport(ctx context.Context, f TeacherSearchParams) ([]TeacherRecord, error) {
	f.Limit = 10000
	f.Offset = 0
	return r.SearchTeachers(ctx, f)
}

func (r *hrRepository) ClearTeacherAssignments(ctx context.Context, teacherID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM teacher_assignments WHERE teacher_id=$1`, teacherID)
	return err
}

func (r *hrRepository) AddTeacherAssignment(ctx context.Context, teacherID uuid.UUID, subjectID, classID, sectionID *uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO teacher_assignments (teacher_id, subject_id, class_id, section_id) VALUES ($1,$2,$3,$4)
ON CONFLICT DO NOTHING`, teacherID, nullableUUID(subjectID), nullableUUID(classID), nullableUUID(sectionID))
	return err
}

func (r *hrRepository) ListTeacherAssignments(ctx context.Context, teacherID uuid.UUID) ([]TeacherAssignmentRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT ta.id, ta.subject_id, sub.name, ta.class_id, c.name, ta.section_id, sec.name
FROM teacher_assignments ta
LEFT JOIN subjects sub ON sub.id = ta.subject_id
LEFT JOIN classes c ON c.id = ta.class_id
LEFT JOIN sections sec ON sec.id = ta.section_id
WHERE ta.teacher_id=$1 ORDER BY c.sort_order, sub.name`, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TeacherAssignmentRecord
	for rows.Next() {
		var a TeacherAssignmentRecord
		var subID, classID, secID pgtype.UUID
		var subName, className, secName pgtype.Text
		if err := rows.Scan(&a.ID, &subID, &subName, &classID, &className, &secID, &secName); err != nil {
			return nil, err
		}
		a.SubjectID = scanOptionalUUID(subID)
		a.ClassID = scanOptionalUUID(classID)
		a.SectionID = scanOptionalUUID(secID)
		a.SubjectName = subName.String
		a.ClassName = className.String
		a.SectionName = secName.String
		items = append(items, a)
	}
	return items, rows.Err()
}

func (r *hrRepository) CountTeacherClasses(ctx context.Context, teacherID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT class_id)::bigint FROM teacher_assignments WHERE teacher_id=$1 AND class_id IS NOT NULL`, teacherID).Scan(&count)
	return count, err
}

func (r *hrRepository) CountTeacherSubjects(ctx context.Context, teacherID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT subject_id)::bigint FROM teacher_assignments WHERE teacher_id=$1 AND subject_id IS NOT NULL`, teacherID).Scan(&count)
	return count, err
}

func (r *hrRepository) ListTodaySchedule(ctx context.Context, teacherID uuid.UUID, dayOfWeek int) ([]ScheduleRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT ts.id, COALESCE(sub.name,''), COALESCE(c.name,''), COALESCE(sec.name,''),
       to_char(ts.start_time,'HH24:MI'), to_char(ts.end_time,'HH24:MI'), COALESCE(ts.room,'')
FROM teacher_schedules ts
LEFT JOIN subjects sub ON sub.id = ts.subject_id
LEFT JOIN classes c ON c.id = ts.class_id
LEFT JOIN sections sec ON sec.id = ts.section_id
WHERE ts.teacher_id=$1 AND ts.day_of_week=$2 ORDER BY ts.start_time`, teacherID, dayOfWeek)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ScheduleRecord
	for rows.Next() {
		var s ScheduleRecord
		if err := rows.Scan(&s.ID, &s.SubjectName, &s.ClassName, &s.SectionName, &s.StartTime, &s.EndTime, &s.Room); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *hrRepository) CreateTeacherDocument(ctx context.Context, teacherID uuid.UUID, docType, fileName, fileURL string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO teacher_documents (teacher_id, doc_type, file_name, file_url) VALUES ($1,$2,$3,$4)`, teacherID, docType, fileName, fileURL)
	return err
}

func (r *hrRepository) ListTeacherDocuments(ctx context.Context, teacherID uuid.UUID) ([]DocumentRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, doc_type, file_name, file_url FROM teacher_documents WHERE teacher_id=$1 AND deleted_at IS NULL ORDER BY created_at`, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDocuments(rows)
}

func scanDocuments(rows pgx.Rows) ([]DocumentRecord, error) {
	var items []DocumentRecord
	for rows.Next() {
		var d DocumentRecord
		if err := rows.Scan(&d.ID, &d.DocType, &d.FileName, &d.FileURL); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

const staffSelect = `
SELECT s.id, s.employee_id, s.first_name, s.last_name, s.photo_url, s.phone, s.email, s.address,
       s.department_id, d.name, s.designation_id, des.name, s.salary, s.joining_date, s.status, s.created_at, s.updated_at
FROM staffs s
LEFT JOIN departments d ON d.id = s.department_id
LEFT JOIN designations des ON des.id = s.designation_id`

func scanStaff(row pgx.Row) (*StaffRecord, error) {
	var s StaffRecord
	var photo, phone, email, address, deptName, desName pgtype.Text
	var deptID, desID pgtype.UUID
	var salary pgtype.Numeric
	if err := row.Scan(
		&s.ID, &s.EmployeeID, &s.FirstName, &s.LastName, &photo, &phone, &email, &address,
		&deptID, &deptName, &desID, &desName, &salary, &s.JoiningDate, &s.Status, &s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		return nil, err
	}
	s.PhotoURL = photo.String
	s.Phone = phone.String
	s.Email = email.String
	s.Address = address.String
	s.DepartmentName = deptName.String
	s.DesignationName = desName.String
	s.DepartmentID = scanOptionalUUID(deptID)
	s.DesignationID = scanOptionalUUID(desID)
	if salary.Valid {
		f, _ := salary.Float64Value()
		s.Salary = f.Float64
	}
	return &s, nil
}

func (r *hrRepository) staffWhere(f StaffSearchParams) (string, []any) {
	args := []any{f.Query, uuidOrNil(f.DepartmentID), f.Status}
	where := ` WHERE s.deleted_at IS NULL
AND ($1::text = '' OR s.first_name ILIKE '%' || $1 || '%' OR s.last_name ILIKE '%' || $1 || '%'
     OR s.employee_id ILIKE '%' || $1 || '%' OR s.email ILIKE '%' || $1 || '%')
AND ($2::uuid IS NULL OR s.department_id = $2)
AND ($3::text = '' OR s.status = $3)`
	return where, args
}

func (r *hrRepository) CreateStaff(ctx context.Context, p CreateStaffParams) (*StaffRecord, error) {
	var id uuid.UUID
	var salary pgtype.Numeric
	if p.Salary > 0 {
		_ = salary.Scan(fmt.Sprintf("%.2f", p.Salary))
	}
	err := r.pool.QueryRow(ctx, `
INSERT INTO staffs (employee_id, first_name, last_name, photo_url, phone, email, address, department_id, designation_id, salary, joining_date, status)
VALUES ($1,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),$8,$9,$10,$11,$12) RETURNING id`,
		p.EmployeeID, p.FirstName, p.LastName, p.PhotoURL, p.Phone, p.Email, p.Address,
		nullableUUID(p.DepartmentID), nullableUUID(p.DesignationID), salary, p.JoiningDate, p.Status,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetStaffByID(ctx, id)
}

func (r *hrRepository) UpdateStaff(ctx context.Context, id uuid.UUID, p UpdateStaffParams) (*StaffRecord, error) {
	var salary pgtype.Numeric
	if p.Salary > 0 {
		_ = salary.Scan(fmt.Sprintf("%.2f", p.Salary))
	}
	tag, err := r.pool.Exec(ctx, `
UPDATE staffs SET first_name=$2, last_name=$3,
    photo_url=CASE WHEN $4='' THEN photo_url ELSE $4 END,
    phone=NULLIF($5,''), email=NULLIF($6,''), address=NULLIF($7,''),
    department_id=$8, designation_id=$9, salary=$10, joining_date=$11, status=$12, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL`,
		id, p.FirstName, p.LastName, p.PhotoURL, p.Phone, p.Email, p.Address,
		nullableUUID(p.DepartmentID), nullableUUID(p.DesignationID), salary, p.JoiningDate, p.Status)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	return r.GetStaffByID(ctx, id)
}

func (r *hrRepository) SoftDeleteStaff(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE staffs SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *hrRepository) GetStaffByID(ctx context.Context, id uuid.UUID) (*StaffRecord, error) {
	row := r.pool.QueryRow(ctx, staffSelect+` WHERE s.id=$1 AND s.deleted_at IS NULL`, id)
	s, err := scanStaff(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *hrRepository) SearchStaff(ctx context.Context, f StaffSearchParams) ([]StaffRecord, error) {
	where, args := r.staffWhere(f)
	args = append(args, f.Limit, f.Offset)
	rows, err := r.pool.Query(ctx, staffSelect+where+` ORDER BY s.created_at DESC LIMIT $4 OFFSET $5`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []StaffRecord
	for rows.Next() {
		s, err := scanStaff(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *s)
	}
	return items, rows.Err()
}

func (r *hrRepository) CountStaff(ctx context.Context, f StaffSearchParams) (int64, error) {
	where, args := r.staffWhere(f)
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*)::bigint FROM staffs s`+where, args...).Scan(&count)
	return count, err
}

func (r *hrRepository) ListStaffReport(ctx context.Context, f StaffSearchParams) ([]StaffRecord, error) {
	f.Limit = 10000
	f.Offset = 0
	return r.SearchStaff(ctx, f)
}

func (r *hrRepository) CreateStaffDocument(ctx context.Context, staffID uuid.UUID, docType, fileName, fileURL string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO staff_documents (staff_id, doc_type, file_name, file_url) VALUES ($1,$2,$3,$4)`, staffID, docType, fileName, fileURL)
	return err
}

func (r *hrRepository) ListStaffDocuments(ctx context.Context, staffID uuid.UUID) ([]DocumentRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, doc_type, file_name, file_url FROM staff_documents WHERE staff_id=$1 AND deleted_at IS NULL ORDER BY created_at`, staffID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDocuments(rows)
}
