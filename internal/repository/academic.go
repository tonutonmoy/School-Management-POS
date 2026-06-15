package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AcademicRepository interface {
	// Departments
	ListDepartments(ctx context.Context) ([]DepartmentRecord, error)
	GetDepartmentByID(ctx context.Context, id uuid.UUID) (*DepartmentRecord, error)

	// Classes
	CreateClass(ctx context.Context, name, code, description string, sortOrder int) (*ClassRecord, error)
	UpdateClass(ctx context.Context, id uuid.UUID, name, code, description string, sortOrder int) (*ClassRecord, error)
	SoftDeleteClass(ctx context.Context, id uuid.UUID) error
	GetClassByID(ctx context.Context, id uuid.UUID) (*ClassRecord, error)
	ListClasses(ctx context.Context) ([]ClassRecord, error)

	// Sections
	CreateSection(ctx context.Context, classID uuid.UUID, name string, capacity int) (*SectionRecord, error)
	UpdateSection(ctx context.Context, id, classID uuid.UUID, name string, capacity int) (*SectionRecord, error)
	SoftDeleteSection(ctx context.Context, id uuid.UUID) error
	GetSectionByID(ctx context.Context, id uuid.UUID) (*SectionRecord, error)
	ListSectionsByClass(ctx context.Context, classID uuid.UUID) ([]SectionRecord, error)
	ListSections(ctx context.Context) ([]SectionRecord, error)

	// Subjects
	CreateSubject(ctx context.Context, name, code, description string) (*SubjectRecord, error)
	UpdateSubject(ctx context.Context, id uuid.UUID, name, code, description string) (*SubjectRecord, error)
	SoftDeleteSubject(ctx context.Context, id uuid.UUID) error
	GetSubjectByID(ctx context.Context, id uuid.UUID) (*SubjectRecord, error)
	ListSubjects(ctx context.Context) ([]SubjectRecord, error)
	AssignSubjectToClass(ctx context.Context, classID, subjectID uuid.UUID) error
	ClearClassSubjects(ctx context.Context, classID uuid.UUID) error
	ListSubjectsByClass(ctx context.Context, classID uuid.UUID) ([]SubjectRecord, error)
}

type DepartmentRecord struct {
	ID   uuid.UUID
	Name string
	Slug string
}

type ClassRecord struct {
	ID          uuid.UUID
	Name        string
	Code        string
	Description string
	SortOrder   int
}

type SectionRecord struct {
	ID        uuid.UUID
	ClassID   uuid.UUID
	ClassName string
	Name      string
	Capacity  int
}

type SubjectRecord struct {
	ID          uuid.UUID
	Name        string
	Code        string
	Description string
}

type academicRepository struct {
	pool *pgxpool.Pool
}

func NewAcademicRepository(pool *pgxpool.Pool) AcademicRepository {
	return &academicRepository{pool: pool}
}

func (r *academicRepository) ListDepartments(ctx context.Context) ([]DepartmentRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,slug FROM departments WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DepartmentRecord
	for rows.Next() {
		var d DepartmentRecord
		if err := rows.Scan(&d.ID, &d.Name, &d.Slug); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *academicRepository) GetDepartmentByID(ctx context.Context, id uuid.UUID) (*DepartmentRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,name,slug FROM departments WHERE id=$1 AND deleted_at IS NULL`, id)
	var d DepartmentRecord
	if err := row.Scan(&d.ID, &d.Name, &d.Slug); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *academicRepository) CreateClass(ctx context.Context, name, code, description string, sortOrder int) (*ClassRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO classes (name,code,description,sort_order) VALUES ($1,$2,NULLIF($3,''),$4) RETURNING id`,
		name, code, description, sortOrder).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetClassByID(ctx, id)
}

func (r *academicRepository) UpdateClass(ctx context.Context, id uuid.UUID, name, code, description string, sortOrder int) (*ClassRecord, error) {
	tag, err := r.pool.Exec(ctx, `
UPDATE classes SET name=$2,code=$3,description=NULLIF($4,''),sort_order=$5,updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL`, id, name, code, description, sortOrder)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	return r.GetClassByID(ctx, id)
}

func (r *academicRepository) SoftDeleteClass(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE classes SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *academicRepository) GetClassByID(ctx context.Context, id uuid.UUID) (*ClassRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,name,code,COALESCE(description,''),sort_order FROM classes WHERE id=$1 AND deleted_at IS NULL`, id)
	var c ClassRecord
	if err := row.Scan(&c.ID, &c.Name, &c.Code, &c.Description, &c.SortOrder); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *academicRepository) ListClasses(ctx context.Context) ([]ClassRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,code,COALESCE(description,''),sort_order FROM classes WHERE deleted_at IS NULL ORDER BY sort_order,name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ClassRecord
	for rows.Next() {
		var c ClassRecord
		if err := rows.Scan(&c.ID, &c.Name, &c.Code, &c.Description, &c.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *academicRepository) CreateSection(ctx context.Context, classID uuid.UUID, name string, capacity int) (*SectionRecord, error) {
	var id uuid.UUID
	var capVal pgtype.Int4
	if capacity > 0 {
		capVal = pgtype.Int4{Int32: int32(capacity), Valid: true}
	}
	err := r.pool.QueryRow(ctx, `INSERT INTO sections (class_id,name,capacity) VALUES ($1,$2,$3) RETURNING id`, classID, name, capVal).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetSectionByID(ctx, id)
}

func (r *academicRepository) UpdateSection(ctx context.Context, id, classID uuid.UUID, name string, capacity int) (*SectionRecord, error) {
	var capVal pgtype.Int4
	if capacity > 0 {
		capVal = pgtype.Int4{Int32: int32(capacity), Valid: true}
	}
	tag, err := r.pool.Exec(ctx, `UPDATE sections SET class_id=$2,name=$3,capacity=$4,updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`,
		id, classID, name, capVal)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	return r.GetSectionByID(ctx, id)
}

func (r *academicRepository) SoftDeleteSection(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE sections SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *academicRepository) GetSectionByID(ctx context.Context, id uuid.UUID) (*SectionRecord, error) {
	row := r.pool.QueryRow(ctx, `
SELECT s.id,s.class_id,c.name,s.name,s.capacity FROM sections s
JOIN classes c ON c.id=s.class_id WHERE s.id=$1 AND s.deleted_at IS NULL`, id)
	var sec SectionRecord
	var cap pgtype.Int4
	if err := row.Scan(&sec.ID, &sec.ClassID, &sec.ClassName, &sec.Name, &cap); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if cap.Valid {
		sec.Capacity = int(cap.Int32)
	}
	return &sec, nil
}

func (r *academicRepository) ListSectionsByClass(ctx context.Context, classID uuid.UUID) ([]SectionRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT s.id,s.class_id,c.name,s.name,s.capacity FROM sections s
JOIN classes c ON c.id=s.class_id WHERE s.class_id=$1 AND s.deleted_at IS NULL ORDER BY s.name`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSections(rows)
}

func (r *academicRepository) ListSections(ctx context.Context) ([]SectionRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT s.id,s.class_id,c.name,s.name,s.capacity FROM sections s
JOIN classes c ON c.id=s.class_id WHERE s.deleted_at IS NULL ORDER BY c.sort_order,s.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSections(rows)
}

func scanSections(rows pgx.Rows) ([]SectionRecord, error) {
	var items []SectionRecord
	for rows.Next() {
		var sec SectionRecord
		var cap pgtype.Int4
		if err := rows.Scan(&sec.ID, &sec.ClassID, &sec.ClassName, &sec.Name, &cap); err != nil {
			return nil, err
		}
		if cap.Valid {
			sec.Capacity = int(cap.Int32)
		}
		items = append(items, sec)
	}
	return items, rows.Err()
}

func (r *academicRepository) CreateSubject(ctx context.Context, name, code, description string) (*SubjectRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `INSERT INTO subjects (name,code,description) VALUES ($1,$2,NULLIF($3,'')) RETURNING id`, name, code, description).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetSubjectByID(ctx, id)
}

func (r *academicRepository) UpdateSubject(ctx context.Context, id uuid.UUID, name, code, description string) (*SubjectRecord, error) {
	tag, err := r.pool.Exec(ctx, `UPDATE subjects SET name=$2,code=$3,description=NULLIF($4,''),updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id, name, code, description)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, nil
	}
	return r.GetSubjectByID(ctx, id)
}

func (r *academicRepository) SoftDeleteSubject(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE subjects SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *academicRepository) GetSubjectByID(ctx context.Context, id uuid.UUID) (*SubjectRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,name,code,COALESCE(description,'') FROM subjects WHERE id=$1 AND deleted_at IS NULL`, id)
	var s SubjectRecord
	if err := row.Scan(&s.ID, &s.Name, &s.Code, &s.Description); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *academicRepository) ListSubjects(ctx context.Context) ([]SubjectRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,code,COALESCE(description,'') FROM subjects WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SubjectRecord
	for rows.Next() {
		var s SubjectRecord
		if err := rows.Scan(&s.ID, &s.Name, &s.Code, &s.Description); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *academicRepository) AssignSubjectToClass(ctx context.Context, classID, subjectID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO class_subjects (class_id,subject_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, classID, subjectID)
	return err
}

func (r *academicRepository) ClearClassSubjects(ctx context.Context, classID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM class_subjects WHERE class_id=$1`, classID)
	return err
}

func (r *academicRepository) ListSubjectsByClass(ctx context.Context, classID uuid.UUID) ([]SubjectRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT s.id,s.name,s.code,COALESCE(s.description,'') FROM subjects s
JOIN class_subjects cs ON cs.subject_id=s.id WHERE cs.class_id=$1 AND s.deleted_at IS NULL ORDER BY s.name`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SubjectRecord
	for rows.Next() {
		var s SubjectRecord
		if err := rows.Scan(&s.ID, &s.Name, &s.Code, &s.Description); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}