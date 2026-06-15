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

type ExamRepository interface {
	WithTx(ctx context.Context, fn func(pgx.Tx) error) error

	// Exams
	CreateExam(ctx context.Context, p CreateExamParams) (*ExamRecord, error)
	UpdateExam(ctx context.Context, id uuid.UUID, p CreateExamParams) (*ExamRecord, error)
	SoftDeleteExam(ctx context.Context, id uuid.UUID) error
	GetExam(ctx context.Context, id uuid.UUID) (*ExamRecord, error)
	SearchExams(ctx context.Context, f ExamSearchParams) ([]ExamRecord, error)
	CountExams(ctx context.Context, f ExamSearchParams) (int64, error)
	UpdateExamStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateExamResultStatus(ctx context.Context, id uuid.UUID, status string) error
	CountExamsByStatus(ctx context.Context, status string) (int64, error)
	CountExamsByResultStatus(ctx context.Context, status string) (int64, error)

	// Exam subjects
	CreateExamSubject(ctx context.Context, examID uuid.UUID, p ExamSubjectParams) (*ExamSubjectRecord, error)
	UpdateExamSubject(ctx context.Context, id uuid.UUID, p ExamSubjectParams) (*ExamSubjectRecord, error)
	DeleteExamSubject(ctx context.Context, id uuid.UUID) error
	ListExamSubjects(ctx context.Context, examID uuid.UUID) ([]ExamSubjectRecord, error)
	GetExamSubject(ctx context.Context, id uuid.UUID) (*ExamSubjectRecord, error)

	// Grading
	ListGradingSystems(ctx context.Context) ([]GradingSystemRecord, error)
	GetDefaultGradingSystem(ctx context.Context) (*GradingSystemRecord, error)
	ListGradingScales(ctx context.Context, systemID uuid.UUID) ([]GradingScaleRecord, error)
	CreateGradingSystem(ctx context.Context, name string, scales []GradingScaleParams) (*GradingSystemRecord, error)

	// Marks
	UpsertStudentMark(ctx context.Context, p UpsertMarkParams) error
	ListMarksSheet(ctx context.Context, examSubjectID uuid.UUID) ([]MarkSheetRecord, error)
	ListMarksByExam(ctx context.Context, examID uuid.UUID) ([]StudentMarkRecord, error)

	// Results
	DeleteExamResults(ctx context.Context, examID uuid.UUID) error
	UpsertExamResult(ctx context.Context, tx pgx.Tx, p ExamResultParams) error
	UpdateResultPositions(ctx context.Context, tx pgx.Tx, examID uuid.UUID) error
	ListExamResults(ctx context.Context, f ResultSearchParams) ([]ExamResultRecord, error)
	CountExamResults(ctx context.Context, f ResultSearchParams) (int64, error)
	GetExamResult(ctx context.Context, id uuid.UUID) (*ExamResultRecord, error)
	GetStudentExamResult(ctx context.Context, examID, studentID uuid.UUID) (*ExamResultRecord, error)
	CountResultsPassed(ctx context.Context, examID uuid.UUID, passed bool) (int64, error)
	StudentCGPA(ctx context.Context, studentID, sessionID uuid.UUID) (float64, error)
	GPADistribution(ctx context.Context, examID uuid.UUID) ([]GradeCountRecord, error)
	SubjectPerformance(ctx context.Context, examID uuid.UUID) ([]SubjectPerfRecord, error)

	// Report cards
	CreateReportCard(ctx context.Context, examResultID, examID, studentID uuid.UUID, token string, generatedBy uuid.UUID) (*ReportCardRecord, error)
	GetReportCardByToken(ctx context.Context, token string) (*ReportCardRecord, error)
	GetReportCardByResult(ctx context.Context, resultID uuid.UUID) (*ReportCardRecord, error)
}

type CreateExamParams struct {
	Name, ExamType, Status string
	SessionID, ClassID     uuid.UUID
	GradingSystemID        *uuid.UUID
	StartDate, EndDate     time.Time
	TotalMarks, PassingMarks float64
}

type ExamSearchParams struct {
	SessionID, ClassID *uuid.UUID
	Status, Query      string
	Limit, Offset      int32
}

type ExamRecord struct {
	ID              uuid.UUID
	Name, ExamType  string
	SessionID       uuid.UUID
	SessionName     string
	ClassID         uuid.UUID
	ClassName       string
	StartDate, EndDate time.Time
	TotalMarks, PassingMarks float64
	GradingSystemID *uuid.UUID
	Status, ResultStatus string
	SubjectCount    int64
}

type ExamSubjectParams struct {
	SubjectID uuid.UUID
	FullMarks, PassMarks, WrittenMarks, MCQMarks, PracticalMarks float64
}

type ExamSubjectRecord struct {
	ID, ExamID, SubjectID uuid.UUID
	SubjectName, SubjectCode string
	FullMarks, PassMarks, WrittenMarks, MCQMarks, PracticalMarks float64
}

type GradingSystemRecord struct {
	ID        uuid.UUID
	Name      string
	IsDefault bool
}

type GradingScaleRecord struct {
	Grade         string
	MinPercentage, MaxPercentage, GPAPoint float64
	SortOrder     int
}

type GradingScaleParams struct {
	Grade         string
	MinPercentage, MaxPercentage, GPAPoint float64
	SortOrder     int
}

type UpsertMarkParams struct {
	ExamID, ExamSubjectID, StudentID, EnteredBy uuid.UUID
	WrittenScore, MCQScore, PracticalScore, TotalScore float64
	IsAbsent bool
}

type MarkSheetRecord struct {
	StudentID       uuid.UUID
	StudentName, RollNumber, AdmissionNo string
	WrittenScore, MCQScore, PracticalScore, TotalScore float64
	IsAbsent        bool
	RecordID        *uuid.UUID
}

type StudentMarkRecord struct {
	ID, ExamID, ExamSubjectID, StudentID uuid.UUID
	SubjectName string
	WrittenScore, MCQScore, PracticalScore, TotalScore float64
	IsAbsent    bool
}

type ExamResultParams struct {
	ExamID, StudentID, SessionID, ClassID, SectionID uuid.UUID
	TotalObtained, TotalFull, Percentage, GPA, CGPA float64
	Grade          string
	IsPassed       bool
	ResultStatus   string
}

type ExamResultRecord struct {
	ID, ExamID, StudentID uuid.UUID
	ExamName, StudentName, AdmissionNo, RollNumber, ClassName, SectionName string
	TotalObtained, TotalFull, Percentage, GPA, CGPA float64
	Grade          string
	IsPassed       bool
	ClassPosition, SectionPosition, MeritPosition *int
	ResultStatus   string
	ProcessedAt    *time.Time
}

type ResultSearchParams struct {
	ExamID, ClassID, SectionID, StudentID *uuid.UUID
	PassedOnly, FailedOnly, PublishedOnly bool
	Limit, Offset int32
}

type GradeCountRecord struct {
	Grade string
	Count int64
}

type SubjectPerfRecord struct {
	SubjectName string
	AvgScore, PassRate float64
}

type ReportCardRecord struct {
	ID, ExamResultID, ExamID, StudentID uuid.UUID
	CardToken  string
	GeneratedAt time.Time
}

type examRepository struct{ pool *pgxpool.Pool }

func NewExamRepository(pool *pgxpool.Pool) ExamRepository { return &examRepository{pool: pool} }

func (r *examRepository) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
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

func (r *examRepository) examSelect() string {
	return `
SELECT e.id, e.name, e.exam_type, e.session_id, sess.name, e.class_id, c.name,
    e.start_date, e.end_date, e.total_marks, e.passing_marks, e.grading_system_id,
    e.status, e.result_status,
    (SELECT COUNT(*) FROM exam_subjects es WHERE es.exam_id = e.id)
FROM exams e
JOIN academic_sessions sess ON sess.id = e.session_id
JOIN classes c ON c.id = e.class_id
WHERE e.deleted_at IS NULL`
}

func scanExam(row pgx.Row) (*ExamRecord, error) {
	var rec ExamRecord
	var gsID *uuid.UUID
	err := row.Scan(&rec.ID, &rec.Name, &rec.ExamType, &rec.SessionID, &rec.SessionName,
		&rec.ClassID, &rec.ClassName, &rec.StartDate, &rec.EndDate,
		&rec.TotalMarks, &rec.PassingMarks, &gsID, &rec.Status, &rec.ResultStatus, &rec.SubjectCount)
	rec.GradingSystemID = gsID
	return &rec, err
}

func (r *examRepository) CreateExam(ctx context.Context, p CreateExamParams) (*ExamRecord, error) {
	var id uuid.UUID
	var gsID pgtype.UUID
	if p.GradingSystemID != nil {
		gsID = pgtype.UUID{Bytes: *p.GradingSystemID, Valid: true}
	}
	status := p.Status
	if status == "" {
		status = "draft"
	}
	err := r.pool.QueryRow(ctx, `
INSERT INTO exams (name, exam_type, session_id, class_id, start_date, end_date, total_marks, passing_marks, grading_system_id, status)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		p.Name, p.ExamType, p.SessionID, p.ClassID, p.StartDate, p.EndDate,
		p.TotalMarks, p.PassingMarks, gsID, status).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetExam(ctx, id)
}

func (r *examRepository) UpdateExam(ctx context.Context, id uuid.UUID, p CreateExamParams) (*ExamRecord, error) {
	var gsID pgtype.UUID
	if p.GradingSystemID != nil {
		gsID = pgtype.UUID{Bytes: *p.GradingSystemID, Valid: true}
	}
	_, err := r.pool.Exec(ctx, `
UPDATE exams SET name=$2, exam_type=$3, session_id=$4, class_id=$5, start_date=$6, end_date=$7,
    total_marks=$8, passing_marks=$9, grading_system_id=$10, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL`,
		id, p.Name, p.ExamType, p.SessionID, p.ClassID, p.StartDate, p.EndDate,
		p.TotalMarks, p.PassingMarks, gsID)
	if err != nil {
		return nil, err
	}
	return r.GetExam(ctx, id)
}

func (r *examRepository) SoftDeleteExam(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE exams SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *examRepository) GetExam(ctx context.Context, id uuid.UUID) (*ExamRecord, error) {
	rec, err := scanExam(r.pool.QueryRow(ctx, r.examSelect()+` AND e.id=$1`, id))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *examRepository) examSearchQuery(f ExamSearchParams, count bool) (string, []any) {
	q := r.examSelect()
	args := []any{}
	n := 1
	if f.SessionID != nil {
		q += fmt.Sprintf(" AND e.session_id=$%d", n)
		args = append(args, *f.SessionID)
		n++
	}
	if f.ClassID != nil {
		q += fmt.Sprintf(" AND e.class_id=$%d", n)
		args = append(args, *f.ClassID)
		n++
	}
	if f.Status != "" {
		q += fmt.Sprintf(" AND e.status=$%d", n)
		args = append(args, f.Status)
		n++
	}
	if f.Query != "" {
		q += fmt.Sprintf(" AND e.name ILIKE $%d", n)
		args = append(args, "%"+f.Query+"%")
		n++
	}
	if count {
		return "SELECT COUNT(*) FROM (" + q + ") sub", args
	}
	q += " ORDER BY e.start_date DESC"
	if f.Limit > 0 {
		q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", n, n+1)
		args = append(args, f.Limit, f.Offset)
	}
	return q, args
}

func (r *examRepository) SearchExams(ctx context.Context, f ExamSearchParams) ([]ExamRecord, error) {
	q, args := r.examSearchQuery(f, false)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExamRecord
	for rows.Next() {
		rec, err := scanExam(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func (r *examRepository) CountExams(ctx context.Context, f ExamSearchParams) (int64, error) {
	q, args := r.examSearchQuery(f, true)
	var count int64
	return count, r.pool.QueryRow(ctx, q, args...).Scan(&count)
}

func (r *examRepository) UpdateExamStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE exams SET status=$2, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id, status)
	return err
}

func (r *examRepository) UpdateExamResultStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.WithTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `UPDATE exams SET result_status=$2, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id, status); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `UPDATE exam_results SET result_status=$2, updated_at=NOW() WHERE exam_id=$1`, id, status)
		return err
	})
}

func (r *examRepository) CountExamsByResultStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM exams WHERE deleted_at IS NULL AND result_status=$1`, status).Scan(&count)
	return count, err
}

func (r *examRepository) CountExamsByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM exams WHERE deleted_at IS NULL AND status=$1`, status).Scan(&count)
	return count, err
}

func (r *examRepository) CreateExamSubject(ctx context.Context, examID uuid.UUID, p ExamSubjectParams) (*ExamSubjectRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO exam_subjects (exam_id, subject_id, full_marks, pass_marks, written_marks, mcq_marks, practical_marks)
VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		examID, p.SubjectID, p.FullMarks, p.PassMarks, p.WrittenMarks, p.MCQMarks, p.PracticalMarks).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetExamSubject(ctx, id)
}

func (r *examRepository) UpdateExamSubject(ctx context.Context, id uuid.UUID, p ExamSubjectParams) (*ExamSubjectRecord, error) {
	_, err := r.pool.Exec(ctx, `
UPDATE exam_subjects SET subject_id=$2, full_marks=$3, pass_marks=$4, written_marks=$5, mcq_marks=$6, practical_marks=$7, updated_at=NOW()
WHERE id=$1`, id, p.SubjectID, p.FullMarks, p.PassMarks, p.WrittenMarks, p.MCQMarks, p.PracticalMarks)
	if err != nil {
		return nil, err
	}
	return r.GetExamSubject(ctx, id)
}

func (r *examRepository) DeleteExamSubject(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM exam_subjects WHERE id=$1`, id)
	return err
}

func (r *examRepository) examSubjectSelect() string {
	return `
SELECT es.id, es.exam_id, es.subject_id, s.name, s.code,
    es.full_marks, es.pass_marks, es.written_marks, es.mcq_marks, es.practical_marks
FROM exam_subjects es JOIN subjects s ON s.id = es.subject_id`
}

func scanExamSubject(row pgx.Row) (*ExamSubjectRecord, error) {
	var rec ExamSubjectRecord
	err := row.Scan(&rec.ID, &rec.ExamID, &rec.SubjectID, &rec.SubjectName, &rec.SubjectCode,
		&rec.FullMarks, &rec.PassMarks, &rec.WrittenMarks, &rec.MCQMarks, &rec.PracticalMarks)
	return &rec, err
}

func (r *examRepository) ListExamSubjects(ctx context.Context, examID uuid.UUID) ([]ExamSubjectRecord, error) {
	rows, err := r.pool.Query(ctx, r.examSubjectSelect()+` WHERE es.exam_id=$1 ORDER BY s.name`, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExamSubjectRecord
	for rows.Next() {
		rec, err := scanExamSubject(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func (r *examRepository) GetExamSubject(ctx context.Context, id uuid.UUID) (*ExamSubjectRecord, error) {
	rec, err := scanExamSubject(r.pool.QueryRow(ctx, r.examSubjectSelect()+` WHERE es.id=$1`, id))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *examRepository) ListGradingSystems(ctx context.Context) ([]GradingSystemRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, is_default FROM grading_systems WHERE deleted_at IS NULL ORDER BY is_default DESC, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GradingSystemRecord
	for rows.Next() {
		var rec GradingSystemRecord
		if err := rows.Scan(&rec.ID, &rec.Name, &rec.IsDefault); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *examRepository) GetDefaultGradingSystem(ctx context.Context) (*GradingSystemRecord, error) {
	var rec GradingSystemRecord
	err := r.pool.QueryRow(ctx, `SELECT id, name, is_default FROM grading_systems WHERE is_default=true AND deleted_at IS NULL LIMIT 1`).Scan(&rec.ID, &rec.Name, &rec.IsDefault)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &rec, err
}

func (r *examRepository) ListGradingScales(ctx context.Context, systemID uuid.UUID) ([]GradingScaleRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT grade, min_percentage, max_percentage, gpa_point, sort_order
FROM grading_scales WHERE system_id=$1 ORDER BY sort_order`, systemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GradingScaleRecord
	for rows.Next() {
		var rec GradingScaleRecord
		if err := rows.Scan(&rec.Grade, &rec.MinPercentage, &rec.MaxPercentage, &rec.GPAPoint, &rec.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *examRepository) CreateGradingSystem(ctx context.Context, name string, scales []GradingScaleParams) (*GradingSystemRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `INSERT INTO grading_systems (name) VALUES ($1) RETURNING id`, name).Scan(&id)
	if err != nil {
		return nil, err
	}
	for _, s := range scales {
		_, err := r.pool.Exec(ctx, `
INSERT INTO grading_scales (system_id, grade, min_percentage, max_percentage, gpa_point, sort_order)
VALUES ($1,$2,$3,$4,$5,$6)`, id, s.Grade, s.MinPercentage, s.MaxPercentage, s.GPAPoint, s.SortOrder)
		if err != nil {
			return nil, err
		}
	}
	return &GradingSystemRecord{ID: id, Name: name}, nil
}

func (r *examRepository) UpsertStudentMark(ctx context.Context, p UpsertMarkParams) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO student_marks (exam_id, exam_subject_id, student_id, written_score, mcq_score, practical_score, total_score, is_absent, entered_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (exam_subject_id, student_id) DO UPDATE SET
    written_score=EXCLUDED.written_score, mcq_score=EXCLUDED.mcq_score, practical_score=EXCLUDED.practical_score,
    total_score=EXCLUDED.total_score, is_absent=EXCLUDED.is_absent, entered_by=EXCLUDED.entered_by, updated_at=NOW()`,
		p.ExamID, p.ExamSubjectID, p.StudentID, p.WrittenScore, p.MCQScore, p.PracticalScore, p.TotalScore, p.IsAbsent, p.EnteredBy)
	return err
}

func (r *examRepository) ListMarksSheet(ctx context.Context, examSubjectID uuid.UUID) ([]MarkSheetRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT s.id, s.first_name||' '||s.last_name, COALESCE(s.roll_number,''), s.admission_number,
    COALESCE(sm.written_score,0), COALESCE(sm.mcq_score,0), COALESCE(sm.practical_score,0),
    COALESCE(sm.total_score,0), COALESCE(sm.is_absent,false), sm.id
FROM exam_subjects es
JOIN exams e ON e.id = es.exam_id
JOIN students s ON s.class_id = e.class_id AND s.session_id = e.session_id AND s.deleted_at IS NULL AND s.status='active'
LEFT JOIN student_marks sm ON sm.exam_subject_id = es.id AND sm.student_id = s.id
WHERE es.id = $1
ORDER BY s.roll_number NULLS LAST, s.first_name`, examSubjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []MarkSheetRecord
	for rows.Next() {
		var rec MarkSheetRecord
		var recordID *uuid.UUID
		if err := rows.Scan(&rec.StudentID, &rec.StudentName, &rec.RollNumber, &rec.AdmissionNo,
			&rec.WrittenScore, &rec.MCQScore, &rec.PracticalScore, &rec.TotalScore, &rec.IsAbsent, &recordID); err != nil {
			return nil, err
		}
		rec.RecordID = recordID
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *examRepository) ListMarksByExam(ctx context.Context, examID uuid.UUID) ([]StudentMarkRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT sm.id, sm.exam_id, sm.exam_subject_id, sm.student_id, sub.name,
    sm.written_score, sm.mcq_score, sm.practical_score, sm.total_score, sm.is_absent
FROM student_marks sm
JOIN exam_subjects es ON es.id = sm.exam_subject_id
JOIN subjects sub ON sub.id = es.subject_id
WHERE sm.exam_id = $1`, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []StudentMarkRecord
	for rows.Next() {
		var rec StudentMarkRecord
		if err := rows.Scan(&rec.ID, &rec.ExamID, &rec.ExamSubjectID, &rec.StudentID, &rec.SubjectName,
			&rec.WrittenScore, &rec.MCQScore, &rec.PracticalScore, &rec.TotalScore, &rec.IsAbsent); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *examRepository) DeleteExamResults(ctx context.Context, examID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM exam_results WHERE exam_id=$1`, examID)
	return err
}

func (r *examRepository) UpsertExamResult(ctx context.Context, tx pgx.Tx, p ExamResultParams) error {
	_, err := tx.Exec(ctx, `
INSERT INTO exam_results (exam_id, student_id, session_id, class_id, section_id, total_obtained, total_full, percentage, gpa, cgpa, grade, is_passed, result_status, processed_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW())
ON CONFLICT (exam_id, student_id) DO UPDATE SET
    total_obtained=EXCLUDED.total_obtained, total_full=EXCLUDED.total_full, percentage=EXCLUDED.percentage,
    gpa=EXCLUDED.gpa, cgpa=EXCLUDED.cgpa, grade=EXCLUDED.grade, is_passed=EXCLUDED.is_passed,
    result_status=EXCLUDED.result_status, processed_at=NOW(), updated_at=NOW()`,
		p.ExamID, p.StudentID, p.SessionID, p.ClassID, p.SectionID,
		p.TotalObtained, p.TotalFull, p.Percentage, p.GPA, p.CGPA, p.Grade, p.IsPassed, p.ResultStatus)
	return err
}

func (r *examRepository) UpdateResultPositions(ctx context.Context, tx pgx.Tx, examID uuid.UUID) error {
	queries := []string{
		`WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY total_obtained DESC, gpa DESC) AS pos
			FROM exam_results WHERE exam_id = $1
		) UPDATE exam_results er SET merit_position = r.pos FROM ranked r WHERE er.id = r.id`,
		`WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (PARTITION BY section_id ORDER BY total_obtained DESC, gpa DESC) AS pos
			FROM exam_results WHERE exam_id = $1
		) UPDATE exam_results er SET section_position = r.pos FROM ranked r WHERE er.id = r.id`,
		`WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (PARTITION BY class_id ORDER BY total_obtained DESC, gpa DESC) AS pos
			FROM exam_results WHERE exam_id = $1
		) UPDATE exam_results er SET class_position = r.pos FROM ranked r WHERE er.id = r.id`,
	}
	for _, q := range queries {
		if _, err := tx.Exec(ctx, q, examID); err != nil {
			return err
		}
	}
	return nil
}

func (r *examRepository) resultSelect() string {
	return `
SELECT er.id, er.exam_id, e.name, er.student_id, s.first_name||' '||s.last_name,
    s.admission_number, COALESCE(s.roll_number,''), c.name, sec.name,
    er.total_obtained, er.total_full, er.percentage, er.gpa, er.cgpa, er.grade, er.is_passed,
    er.class_position, er.section_position, er.merit_position, er.result_status, er.processed_at
FROM exam_results er
JOIN exams e ON e.id = er.exam_id
JOIN students s ON s.id = er.student_id
JOIN classes c ON c.id = er.class_id
JOIN sections sec ON sec.id = er.section_id`
}

func scanResult(row pgx.Row) (*ExamResultRecord, error) {
	var rec ExamResultRecord
	var cp, sp, mp *int
	err := row.Scan(&rec.ID, &rec.ExamID, &rec.ExamName, &rec.StudentID, &rec.StudentName,
		&rec.AdmissionNo, &rec.RollNumber, &rec.ClassName, &rec.SectionName,
		&rec.TotalObtained, &rec.TotalFull, &rec.Percentage, &rec.GPA, &rec.CGPA, &rec.Grade, &rec.IsPassed,
		&cp, &sp, &mp, &rec.ResultStatus, &rec.ProcessedAt)
	rec.ClassPosition, rec.SectionPosition, rec.MeritPosition = cp, sp, mp
	return &rec, err
}

func (r *examRepository) resultSearchQuery(f ResultSearchParams, count bool) (string, []any) {
	q := r.resultSelect() + ` WHERE 1=1`
	args := []any{}
	n := 1
	if f.ExamID != nil {
		q += fmt.Sprintf(" AND er.exam_id=$%d", n)
		args = append(args, *f.ExamID)
		n++
	}
	if f.SectionID != nil {
		q += fmt.Sprintf(" AND er.section_id=$%d", n)
		args = append(args, *f.SectionID)
		n++
	}
	if f.StudentID != nil {
		q += fmt.Sprintf(" AND er.student_id=$%d", n)
		args = append(args, *f.StudentID)
		n++
	}
	if f.PassedOnly {
		q += " AND er.is_passed = true"
	}
	if f.FailedOnly {
		q += " AND er.is_passed = false"
	}
	if f.PublishedOnly {
		q += " AND er.result_status = 'published'"
	}
	if count {
		return "SELECT COUNT(*) FROM (" + q + ") sub", args
	}
	q += " ORDER BY er.merit_position NULLS LAST, er.total_obtained DESC"
	if f.Limit > 0 {
		q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", n, n+1)
		args = append(args, f.Limit, f.Offset)
	}
	return q, args
}

func (r *examRepository) ListExamResults(ctx context.Context, f ResultSearchParams) ([]ExamResultRecord, error) {
	q, args := r.resultSearchQuery(f, false)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExamResultRecord
	for rows.Next() {
		rec, err := scanResult(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func (r *examRepository) CountExamResults(ctx context.Context, f ResultSearchParams) (int64, error) {
	q, args := r.resultSearchQuery(f, true)
	var count int64
	return count, r.pool.QueryRow(ctx, q, args...).Scan(&count)
}

func (r *examRepository) GetExamResult(ctx context.Context, id uuid.UUID) (*ExamResultRecord, error) {
	rec, err := scanResult(r.pool.QueryRow(ctx, r.resultSelect()+` WHERE er.id=$1`, id))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *examRepository) GetStudentExamResult(ctx context.Context, examID, studentID uuid.UUID) (*ExamResultRecord, error) {
	rec, err := scanResult(r.pool.QueryRow(ctx, r.resultSelect()+` WHERE er.exam_id=$1 AND er.student_id=$2`, examID, studentID))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *examRepository) CountResultsPassed(ctx context.Context, examID uuid.UUID, passed bool) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM exam_results WHERE exam_id=$1 AND is_passed=$2`, examID, passed).Scan(&count)
	return count, err
}

func (r *examRepository) StudentCGPA(ctx context.Context, studentID, sessionID uuid.UUID) (float64, error) {
	var cgpa float64
	err := r.pool.QueryRow(ctx, `
SELECT COALESCE(AVG(gpa),0) FROM exam_results er
JOIN exams e ON e.id = er.exam_id
WHERE er.student_id=$1 AND er.session_id=$2 AND er.result_status='published'`, studentID, sessionID).Scan(&cgpa)
	return cgpa, err
}

func (r *examRepository) GPADistribution(ctx context.Context, examID uuid.UUID) ([]GradeCountRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT COALESCE(grade,'F'), COUNT(*) FROM exam_results WHERE exam_id=$1 GROUP BY grade ORDER BY grade`, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GradeCountRecord
	for rows.Next() {
		var rec GradeCountRecord
		if err := rows.Scan(&rec.Grade, &rec.Count); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *examRepository) SubjectPerformance(ctx context.Context, examID uuid.UUID) ([]SubjectPerfRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT sub.name,
    COALESCE(AVG(sm.total_score),0),
    COALESCE(100.0 * COUNT(*) FILTER (WHERE sm.total_score >= es.pass_marks) / NULLIF(COUNT(*),0),0)
FROM exam_subjects es
JOIN subjects sub ON sub.id = es.subject_id
LEFT JOIN student_marks sm ON sm.exam_subject_id = es.id AND NOT sm.is_absent
WHERE es.exam_id = $1
GROUP BY sub.name ORDER BY sub.name`, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SubjectPerfRecord
	for rows.Next() {
		var rec SubjectPerfRecord
		if err := rows.Scan(&rec.SubjectName, &rec.AvgScore, &rec.PassRate); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *examRepository) CreateReportCard(ctx context.Context, examResultID, examID, studentID uuid.UUID, token string, generatedBy uuid.UUID) (*ReportCardRecord, error) {
	var rec ReportCardRecord
	err := r.pool.QueryRow(ctx, `
INSERT INTO report_cards (exam_result_id, exam_id, student_id, card_token, generated_by)
VALUES ($1,$2,$3,$4,$5) RETURNING id, exam_result_id, exam_id, student_id, card_token, generated_at`,
		examResultID, examID, studentID, token, generatedBy).
		Scan(&rec.ID, &rec.ExamResultID, &rec.ExamID, &rec.StudentID, &rec.CardToken, &rec.GeneratedAt)
	return &rec, err
}

func (r *examRepository) GetReportCardByToken(ctx context.Context, token string) (*ReportCardRecord, error) {
	var rec ReportCardRecord
	err := r.pool.QueryRow(ctx, `SELECT id, exam_result_id, exam_id, student_id, card_token, generated_at FROM report_cards WHERE card_token=$1`, token).
		Scan(&rec.ID, &rec.ExamResultID, &rec.ExamID, &rec.StudentID, &rec.CardToken, &rec.GeneratedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &rec, err
}

func (r *examRepository) GetReportCardByResult(ctx context.Context, resultID uuid.UUID) (*ReportCardRecord, error) {
	var rec ReportCardRecord
	err := r.pool.QueryRow(ctx, `SELECT id, exam_result_id, exam_id, student_id, card_token, generated_at FROM report_cards WHERE exam_result_id=$1`, resultID).
		Scan(&rec.ID, &rec.ExamResultID, &rec.ExamID, &rec.StudentID, &rec.CardToken, &rec.GeneratedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &rec, err
}
