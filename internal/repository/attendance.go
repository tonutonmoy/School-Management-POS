package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AttendanceRepository interface {
	// Student attendance
	UpsertStudentAttendance(ctx context.Context, p UpsertStudentAttendanceParams) error
	ListStudentAttendanceSheet(ctx context.Context, sessionID, classID, sectionID uuid.UUID, date time.Time) ([]StudentAttendanceSheetRow, error)
	ListStudentAttendanceReport(ctx context.Context, f AttendanceReportParams) ([]StudentAttendanceRecord, error)
	CountStudentAttendanceByStatus(ctx context.Context, date time.Time, status string) (int64, error)
	StudentAttendanceSummary(ctx context.Context, studentID uuid.UUID, from, to *time.Time) (*AttendanceSummaryRecord, error)
	StudentAttendanceHistory(ctx context.Context, studentID uuid.UUID, from, to time.Time, limit, offset int32) ([]StudentAttendanceRecord, int64, error)
	MonthlyStudentTrend(ctx context.Context, from, to time.Time) ([]DailyAttendanceCount, error)
	ClassWiseAttendanceToday(ctx context.Context, date time.Time) ([]ClassAttendanceCount, error)

	// Teacher attendance
	UpsertTeacherAttendance(ctx context.Context, p UpsertEmployeeAttendanceParams) error
	ListTeacherAttendanceSheet(ctx context.Context, date time.Time, query string) ([]EmployeeAttendanceSheetRow, error)
	ListTeacherAttendanceReport(ctx context.Context, f AttendanceReportParams) ([]EmployeeAttendanceRecord, error)
	CountTeacherAttendanceByStatus(ctx context.Context, date time.Time, status string) (int64, error)

	// Staff attendance
	UpsertStaffAttendance(ctx context.Context, p UpsertEmployeeAttendanceParams) error
	ListStaffAttendanceSheet(ctx context.Context, date time.Time, query string) ([]EmployeeAttendanceSheetRow, error)
	ListStaffAttendanceReport(ctx context.Context, f AttendanceReportParams) ([]EmployeeAttendanceRecord, error)
	CountStaffAttendanceByStatus(ctx context.Context, date time.Time, status string) (int64, error)

	// Leave
	CreateLeaveRequest(ctx context.Context, p CreateLeaveParams) (*LeaveRequestRecord, error)
	UpdateLeaveStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy uuid.UUID, remarks string) (*LeaveRequestRecord, error)
	GetLeaveByID(ctx context.Context, id uuid.UUID) (*LeaveRequestRecord, error)
	SearchLeaveRequests(ctx context.Context, f LeaveSearchParams) ([]LeaveRequestRecord, error)
	CountLeaveRequests(ctx context.Context, f LeaveSearchParams) (int64, error)
}

type UpsertStudentAttendanceParams struct {
	StudentID      uuid.UUID
	SessionID      uuid.UUID
	ClassID        uuid.UUID
	SectionID      uuid.UUID
	AttendanceDate time.Time
	Status         string
	MarkedBy       uuid.UUID
	Remarks        string
}

type UpsertEmployeeAttendanceParams struct {
	EmployeeID     uuid.UUID
	AttendanceDate time.Time
	Status         string
	MarkedBy       uuid.UUID
	Remarks        string
}

type StudentAttendanceSheetRow struct {
	StudentID       uuid.UUID
	AdmissionNumber string
	RollNumber      string
	FirstName       string
	LastName        string
	PhotoURL        string
	Status          string
	Remarks         string
	RecordID        *uuid.UUID
}

type EmployeeAttendanceSheetRow struct {
	EmployeeID uuid.UUID
	EmpCode    string
	Name       string
	PhotoURL   string
	Department string
	Status     string
	Remarks    string
	RecordID   *uuid.UUID
}

type StudentAttendanceRecord struct {
	ID              uuid.UUID
	StudentID       uuid.UUID
	StudentName     string
	AdmissionNumber string
	RollNumber      string
	SessionID       uuid.UUID
	ClassID         uuid.UUID
	ClassName       string
	SectionID       uuid.UUID
	SectionName     string
	AttendanceDate  time.Time
	Status          string
	Remarks         string
}

type EmployeeAttendanceRecord struct {
	ID             uuid.UUID
	EmployeeID     uuid.UUID
	EmployeeCode   string
	EmployeeName   string
	DepartmentName string
	AttendanceDate time.Time
	Status         string
	Remarks        string
}

type AttendanceReportParams struct {
	SessionID *uuid.UUID
	ClassID   *uuid.UUID
	SectionID *uuid.UUID
	StudentID *uuid.UUID
	TeacherID *uuid.UUID
	StaffID   *uuid.UUID
	From      time.Time
	To        time.Time
	Status    string
	Limit     int32
	Offset    int32
}

type AttendanceSummaryRecord struct {
	PresentDays int64
	AbsentDays  int64
	LateDays    int64
	LeaveDays   int64
}

type DailyAttendanceCount struct {
	Date    time.Time
	Present int64
	Absent  int64
	Late    int64
	Leave   int64
}

type ClassAttendanceCount struct {
	ClassID   uuid.UUID
	ClassName string
	Present   int64
	Absent    int64
	Late      int64
	Leave     int64
	Total     int64
}

type CreateLeaveParams struct {
	EntityType string
	TeacherID  *uuid.UUID
	StaffID    *uuid.UUID
	LeaveType  string
	StartDate  time.Time
	EndDate    time.Time
	Reason     string
	AppliedBy  uuid.UUID
}

type LeaveRequestRecord struct {
	ID            uuid.UUID
	EntityType    string
	TeacherID     *uuid.UUID
	StaffID       *uuid.UUID
	EmployeeName  string
	EmployeeID    string
	LeaveType     string
	StartDate     time.Time
	EndDate       time.Time
	Reason        string
	Status        string
	ReviewRemarks string
	ReviewedAt    *time.Time
	CreatedAt     time.Time
}

type LeaveSearchParams struct {
	EntityType string
	Status     string
	LeaveType  string
	Query      string
	Limit      int32
	Offset     int32
}

type attendanceRepository struct {
	pool *pgxpool.Pool
}

func NewAttendanceRepository(pool *pgxpool.Pool) AttendanceRepository {
	return &attendanceRepository{pool: pool}
}

func (r *attendanceRepository) UpsertStudentAttendance(ctx context.Context, p UpsertStudentAttendanceParams) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO student_attendance (student_id, session_id, class_id, section_id, attendance_date, status, marked_by, remarks)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (student_id, attendance_date) WHERE deleted_at IS NULL
DO UPDATE SET status = EXCLUDED.status, remarks = EXCLUDED.remarks, marked_by = EXCLUDED.marked_by,
    session_id = EXCLUDED.session_id, class_id = EXCLUDED.class_id, section_id = EXCLUDED.section_id,
    updated_at = NOW()`,
		p.StudentID, p.SessionID, p.ClassID, p.SectionID, p.AttendanceDate, p.Status, p.MarkedBy, p.Remarks)
	return err
}

func (r *attendanceRepository) ListStudentAttendanceSheet(ctx context.Context, sessionID, classID, sectionID uuid.UUID, date time.Time) ([]StudentAttendanceSheetRow, error) {
	rows, err := r.pool.Query(ctx, `
SELECT s.id, s.admission_number, s.roll_number, s.first_name, s.last_name, s.photo_url,
    COALESCE(sa.status, ''), COALESCE(sa.remarks, ''), sa.id
FROM students s
LEFT JOIN student_attendance sa ON sa.student_id = s.id AND sa.attendance_date = $4 AND sa.deleted_at IS NULL
WHERE s.deleted_at IS NULL AND s.status = 'active'
    AND s.session_id = $1 AND s.class_id = $2 AND s.section_id = $3
ORDER BY s.roll_number NULLS LAST, s.first_name, s.last_name`,
		sessionID, classID, sectionID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []StudentAttendanceSheetRow
	for rows.Next() {
		var row StudentAttendanceSheetRow
		var recordID *uuid.UUID
		if err := rows.Scan(&row.StudentID, &row.AdmissionNumber, &row.RollNumber, &row.FirstName, &row.LastName,
			&row.PhotoURL, &row.Status, &row.Remarks, &recordID); err != nil {
			return nil, err
		}
		row.RecordID = recordID
		items = append(items, row)
	}
	return items, rows.Err()
}

func (r *attendanceRepository) ListStudentAttendanceReport(ctx context.Context, f AttendanceReportParams) ([]StudentAttendanceRecord, error) {
	q := `
SELECT sa.id, sa.student_id, s.first_name || ' ' || s.last_name, s.admission_number, COALESCE(s.roll_number, ''),
    sa.session_id, sa.class_id, c.name, sa.section_id, sec.name, sa.attendance_date, sa.status, COALESCE(sa.remarks, '')
FROM student_attendance sa
JOIN students s ON s.id = sa.student_id
JOIN classes c ON c.id = sa.class_id
JOIN sections sec ON sec.id = sa.section_id
WHERE sa.deleted_at IS NULL`
	args := []any{}
	n := 1
	if !f.From.IsZero() {
		q += fmt.Sprintf(" AND sa.attendance_date >= $%d", n)
		args = append(args, f.From)
		n++
	}
	if !f.To.IsZero() {
		q += fmt.Sprintf(" AND sa.attendance_date <= $%d", n)
		args = append(args, f.To)
		n++
	}
	if f.SessionID != nil {
		q += fmt.Sprintf(" AND sa.session_id = $%d", n)
		args = append(args, *f.SessionID)
		n++
	}
	if f.ClassID != nil {
		q += fmt.Sprintf(" AND sa.class_id = $%d", n)
		args = append(args, *f.ClassID)
		n++
	}
	if f.SectionID != nil {
		q += fmt.Sprintf(" AND sa.section_id = $%d", n)
		args = append(args, *f.SectionID)
		n++
	}
	if f.StudentID != nil {
		q += fmt.Sprintf(" AND sa.student_id = $%d", n)
		args = append(args, *f.StudentID)
		n++
	}
	if f.Status != "" {
		q += fmt.Sprintf(" AND sa.status = $%d", n)
		args = append(args, f.Status)
		n++
	}
	q += " ORDER BY sa.attendance_date DESC, s.first_name"
	if f.Limit > 0 {
		q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", n, n+1)
		args = append(args, f.Limit, f.Offset)
	}
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStudentAttendanceRecords(rows)
}

func scanStudentAttendanceRecords(rows pgx.Rows) ([]StudentAttendanceRecord, error) {
	var items []StudentAttendanceRecord
	for rows.Next() {
		var rec StudentAttendanceRecord
		if err := rows.Scan(&rec.ID, &rec.StudentID, &rec.StudentName, &rec.AdmissionNumber, &rec.RollNumber,
			&rec.SessionID, &rec.ClassID, &rec.ClassName, &rec.SectionID, &rec.SectionName,
			&rec.AttendanceDate, &rec.Status, &rec.Remarks); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *attendanceRepository) CountStudentAttendanceByStatus(ctx context.Context, date time.Time, status string) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `
SELECT COUNT(*) FROM student_attendance
WHERE deleted_at IS NULL AND attendance_date = $1 AND status = $2`, date, status).Scan(&count)
	return count, err
}

func (r *attendanceRepository) StudentAttendanceSummary(ctx context.Context, studentID uuid.UUID, from, to *time.Time) (*AttendanceSummaryRecord, error) {
	q := `
SELECT
    COUNT(*) FILTER (WHERE status = 'present'),
    COUNT(*) FILTER (WHERE status = 'absent'),
    COUNT(*) FILTER (WHERE status = 'late'),
    COUNT(*) FILTER (WHERE status = 'leave')
FROM student_attendance WHERE student_id = $1 AND deleted_at IS NULL`
	args := []any{studentID}
	n := 2
	if from != nil {
		q += fmt.Sprintf(" AND attendance_date >= $%d", n)
		args = append(args, *from)
		n++
	}
	if to != nil {
		q += fmt.Sprintf(" AND attendance_date <= $%d", n)
		args = append(args, *to)
	}
	var rec AttendanceSummaryRecord
	err := r.pool.QueryRow(ctx, q, args...).Scan(&rec.PresentDays, &rec.AbsentDays, &rec.LateDays, &rec.LeaveDays)
	return &rec, err
}

func (r *attendanceRepository) StudentAttendanceHistory(ctx context.Context, studentID uuid.UUID, from, to time.Time, limit, offset int32) ([]StudentAttendanceRecord, int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `
SELECT COUNT(*) FROM student_attendance
WHERE student_id = $1 AND deleted_at IS NULL AND attendance_date BETWEEN $2 AND $3`,
		studentID, from, to).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	recs, err := r.ListStudentAttendanceReport(ctx, AttendanceReportParams{
		StudentID: &studentID, From: from, To: to, Limit: limit, Offset: offset,
	})
	return recs, total, err
}

func (r *attendanceRepository) MonthlyStudentTrend(ctx context.Context, from, to time.Time) ([]DailyAttendanceCount, error) {
	rows, err := r.pool.Query(ctx, `
SELECT attendance_date,
    COUNT(*) FILTER (WHERE status = 'present'),
    COUNT(*) FILTER (WHERE status = 'absent'),
    COUNT(*) FILTER (WHERE status = 'late'),
    COUNT(*) FILTER (WHERE status = 'leave')
FROM student_attendance
WHERE deleted_at IS NULL AND attendance_date BETWEEN $1 AND $2
GROUP BY attendance_date ORDER BY attendance_date`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DailyAttendanceCount
	for rows.Next() {
		var d DailyAttendanceCount
		if err := rows.Scan(&d.Date, &d.Present, &d.Absent, &d.Late, &d.Leave); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *attendanceRepository) ClassWiseAttendanceToday(ctx context.Context, date time.Time) ([]ClassAttendanceCount, error) {
	rows, err := r.pool.Query(ctx, `
SELECT c.id, c.name,
    COUNT(*) FILTER (WHERE sa.status = 'present'),
    COUNT(*) FILTER (WHERE sa.status = 'absent'),
    COUNT(*) FILTER (WHERE sa.status = 'late'),
    COUNT(*) FILTER (WHERE sa.status = 'leave'),
    COUNT(*)
FROM student_attendance sa
JOIN classes c ON c.id = sa.class_id
WHERE sa.deleted_at IS NULL AND sa.attendance_date = $1
GROUP BY c.id, c.name ORDER BY c.name`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ClassAttendanceCount
	for rows.Next() {
		var c ClassAttendanceCount
		if err := rows.Scan(&c.ClassID, &c.ClassName, &c.Present, &c.Absent, &c.Late, &c.Leave, &c.Total); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *attendanceRepository) UpsertTeacherAttendance(ctx context.Context, p UpsertEmployeeAttendanceParams) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO teacher_attendance (teacher_id, attendance_date, status, marked_by, remarks)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (teacher_id, attendance_date) WHERE deleted_at IS NULL
DO UPDATE SET status = EXCLUDED.status, remarks = EXCLUDED.remarks, marked_by = EXCLUDED.marked_by, updated_at = NOW()`,
		p.EmployeeID, p.AttendanceDate, p.Status, p.MarkedBy, p.Remarks)
	return err
}

func (r *attendanceRepository) ListTeacherAttendanceSheet(ctx context.Context, date time.Time, query string) ([]EmployeeAttendanceSheetRow, error) {
	q := `
SELECT t.id, t.employee_id, t.first_name || ' ' || t.last_name, COALESCE(t.photo_url, ''),
    COALESCE(d.name, ''), COALESCE(ta.status, ''), COALESCE(ta.remarks, ''), ta.id
FROM teachers t
LEFT JOIN departments d ON d.id = t.department_id
LEFT JOIN teacher_attendance ta ON ta.teacher_id = t.id AND ta.attendance_date = $1 AND ta.deleted_at IS NULL
WHERE t.deleted_at IS NULL AND t.status = 'active'`
	args := []any{date}
	if query != "" {
		q += " AND (t.first_name ILIKE $2 OR t.last_name ILIKE $2 OR t.employee_id ILIKE $2)"
		args = append(args, "%"+query+"%")
	}
	q += " ORDER BY t.first_name, t.last_name"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEmployeeSheet(rows)
}

func (r *attendanceRepository) ListTeacherAttendanceReport(ctx context.Context, f AttendanceReportParams) ([]EmployeeAttendanceRecord, error) {
	return r.listEmployeeAttendanceReport(ctx, "teacher", f)
}

func (r *attendanceRepository) CountTeacherAttendanceByStatus(ctx context.Context, date time.Time, status string) (int64, error) {
	return r.countEmployeeAttendanceByStatus(ctx, "teacher", date, status)
}

func (r *attendanceRepository) UpsertStaffAttendance(ctx context.Context, p UpsertEmployeeAttendanceParams) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO staff_attendance (staff_id, attendance_date, status, marked_by, remarks)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (staff_id, attendance_date) WHERE deleted_at IS NULL
DO UPDATE SET status = EXCLUDED.status, remarks = EXCLUDED.remarks, marked_by = EXCLUDED.marked_by, updated_at = NOW()`,
		p.EmployeeID, p.AttendanceDate, p.Status, p.MarkedBy, p.Remarks)
	return err
}

func (r *attendanceRepository) ListStaffAttendanceSheet(ctx context.Context, date time.Time, query string) ([]EmployeeAttendanceSheetRow, error) {
	q := `
SELECT s.id, s.employee_id, s.name, COALESCE(s.photo_url, ''),
    COALESCE(d.name, ''), COALESCE(sa.status, ''), COALESCE(sa.remarks, ''), sa.id
FROM staffs s
LEFT JOIN departments d ON d.id = s.department_id
LEFT JOIN staff_attendance sa ON sa.staff_id = s.id AND sa.attendance_date = $1 AND sa.deleted_at IS NULL
WHERE s.deleted_at IS NULL AND s.status = 'active'`
	args := []any{date}
	if query != "" {
		q += " AND (s.name ILIKE $2 OR s.employee_id ILIKE $2)"
		args = append(args, "%"+query+"%")
	}
	q += " ORDER BY s.name"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEmployeeSheet(rows)
}

func (r *attendanceRepository) ListStaffAttendanceReport(ctx context.Context, f AttendanceReportParams) ([]EmployeeAttendanceRecord, error) {
	return r.listEmployeeAttendanceReport(ctx, "staff", f)
}

func (r *attendanceRepository) CountStaffAttendanceByStatus(ctx context.Context, date time.Time, status string) (int64, error) {
	return r.countEmployeeAttendanceByStatus(ctx, "staff", date, status)
}

func scanEmployeeSheet(rows pgx.Rows) ([]EmployeeAttendanceSheetRow, error) {
	var items []EmployeeAttendanceSheetRow
	for rows.Next() {
		var row EmployeeAttendanceSheetRow
		var recordID *uuid.UUID
		if err := rows.Scan(&row.EmployeeID, &row.EmpCode, &row.Name, &row.PhotoURL,
			&row.Department, &row.Status, &row.Remarks, &recordID); err != nil {
			return nil, err
		}
		row.RecordID = recordID
		items = append(items, row)
	}
	return items, rows.Err()
}

func (r *attendanceRepository) listEmployeeAttendanceReport(ctx context.Context, entity string, f AttendanceReportParams) ([]EmployeeAttendanceRecord, error) {
	var table, idCol, nameExpr, codeCol, deptJoin string
	switch entity {
	case "teacher":
		table, idCol = "teacher_attendance", "teacher_id"
		nameExpr = "t.first_name || ' ' || t.last_name"
		codeCol = "t.employee_id"
		deptJoin = "JOIN teachers t ON t.id = ta.teacher_id LEFT JOIN departments d ON d.id = t.department_id"
	case "staff":
		table, idCol = "staff_attendance", "staff_id"
		nameExpr = "s.name"
		codeCol = "s.employee_id"
		deptJoin = "JOIN staffs s ON s.id = ta.staff_id LEFT JOIN departments d ON d.id = s.department_id"
	default:
		return nil, fmt.Errorf("unknown entity: %s", entity)
	}
	q := fmt.Sprintf(`
SELECT ta.id, ta.%s, %s, %s, COALESCE(d.name, ''), ta.attendance_date, ta.status, COALESCE(ta.remarks, '')
FROM %s ta %s WHERE ta.deleted_at IS NULL`, idCol, nameExpr, codeCol, table, deptJoin)
	args := []any{}
	n := 1
	if !f.From.IsZero() {
		q += fmt.Sprintf(" AND ta.attendance_date >= $%d", n)
		args = append(args, f.From)
		n++
	}
	if !f.To.IsZero() {
		q += fmt.Sprintf(" AND ta.attendance_date <= $%d", n)
		args = append(args, f.To)
		n++
	}
	if entity == "teacher" && f.TeacherID != nil {
		q += fmt.Sprintf(" AND ta.teacher_id = $%d", n)
		args = append(args, *f.TeacherID)
		n++
	}
	if entity == "staff" && f.StaffID != nil {
		q += fmt.Sprintf(" AND ta.staff_id = $%d", n)
		args = append(args, *f.StaffID)
		n++
	}
	if f.Status != "" {
		q += fmt.Sprintf(" AND ta.status = $%d", n)
		args = append(args, f.Status)
		n++
	}
	q += " ORDER BY ta.attendance_date DESC"
	if f.Limit > 0 {
		q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", n, n+1)
		args = append(args, f.Limit, f.Offset)
	}
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []EmployeeAttendanceRecord
	for rows.Next() {
		var rec EmployeeAttendanceRecord
		if err := rows.Scan(&rec.ID, &rec.EmployeeID, &rec.EmployeeName, &rec.EmployeeCode,
			&rec.DepartmentName, &rec.AttendanceDate, &rec.Status, &rec.Remarks); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *attendanceRepository) countEmployeeAttendanceByStatus(ctx context.Context, entity string, date time.Time, status string) (int64, error) {
	table := "teacher_attendance"
	if entity == "staff" {
		table = "staff_attendance"
	}
	var count int64
	err := r.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s WHERE deleted_at IS NULL AND attendance_date = $1 AND status = $2`, table),
		date, status).Scan(&count)
	return count, err
}

func (r *attendanceRepository) CreateLeaveRequest(ctx context.Context, p CreateLeaveParams) (*LeaveRequestRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO leave_requests (entity_type, teacher_id, staff_id, leave_type, start_date, end_date, reason, applied_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		p.EntityType, p.TeacherID, p.StaffID, p.LeaveType, p.StartDate, p.EndDate, p.Reason, p.AppliedBy).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetLeaveByID(ctx, id)
}

func (r *attendanceRepository) UpdateLeaveStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy uuid.UUID, remarks string) (*LeaveRequestRecord, error) {
	_, err := r.pool.Exec(ctx, `
UPDATE leave_requests SET status = $2, reviewed_by = $3, reviewed_at = NOW(), review_remarks = $4, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL`, id, status, reviewedBy, remarks)
	if err != nil {
		return nil, err
	}
	return r.GetLeaveByID(ctx, id)
}

func (r *attendanceRepository) GetLeaveByID(ctx context.Context, id uuid.UUID) (*LeaveRequestRecord, error) {
	var rec LeaveRequestRecord
	var teacherID, staffID *uuid.UUID
	err := r.pool.QueryRow(ctx, `
SELECT lr.id, lr.entity_type, lr.teacher_id, lr.staff_id, lr.leave_type, lr.start_date, lr.end_date,
    COALESCE(lr.reason, ''), lr.status, COALESCE(lr.review_remarks, ''), lr.reviewed_at, lr.created_at,
    CASE WHEN lr.entity_type = 'teacher' THEN t.first_name || ' ' || t.last_name ELSE s.name END,
    CASE WHEN lr.entity_type = 'teacher' THEN t.employee_id ELSE s.employee_id END
FROM leave_requests lr
LEFT JOIN teachers t ON t.id = lr.teacher_id
LEFT JOIN staffs s ON s.id = lr.staff_id
WHERE lr.id = $1 AND lr.deleted_at IS NULL`, id).Scan(
		&rec.ID, &rec.EntityType, &teacherID, &staffID, &rec.LeaveType, &rec.StartDate, &rec.EndDate,
		&rec.Reason, &rec.Status, &rec.ReviewRemarks, &rec.ReviewedAt, &rec.CreatedAt,
		&rec.EmployeeName, &rec.EmployeeID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rec.TeacherID = teacherID
	rec.StaffID = staffID
	return &rec, nil
}

func (r *attendanceRepository) SearchLeaveRequests(ctx context.Context, f LeaveSearchParams) ([]LeaveRequestRecord, error) {
	q, args := r.leaveSearchQuery(f, true)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLeaveRecords(rows)
}

func (r *attendanceRepository) CountLeaveRequests(ctx context.Context, f LeaveSearchParams) (int64, error) {
	q, args := r.leaveSearchQuery(f, false)
	var count int64
	err := r.pool.QueryRow(ctx, q, args...).Scan(&count)
	return count, err
}

func (r *attendanceRepository) leaveSearchQuery(f LeaveSearchParams, withLimit bool) (string, []any) {
	q := `
SELECT lr.id, lr.entity_type, lr.teacher_id, lr.staff_id, lr.leave_type, lr.start_date, lr.end_date,
    COALESCE(lr.reason, ''), lr.status, COALESCE(lr.review_remarks, ''), lr.reviewed_at, lr.created_at,
    CASE WHEN lr.entity_type = 'teacher' THEN t.first_name || ' ' || t.last_name ELSE s.name END,
    CASE WHEN lr.entity_type = 'teacher' THEN t.employee_id ELSE s.employee_id END
FROM leave_requests lr
LEFT JOIN teachers t ON t.id = lr.teacher_id
LEFT JOIN staffs s ON s.id = lr.staff_id
WHERE lr.deleted_at IS NULL`
	args := []any{}
	n := 1
	if f.EntityType != "" {
		q += fmt.Sprintf(" AND lr.entity_type = $%d", n)
		args = append(args, f.EntityType)
		n++
	}
	if f.Status != "" {
		q += fmt.Sprintf(" AND lr.status = $%d", n)
		args = append(args, f.Status)
		n++
	}
	if f.LeaveType != "" {
		q += fmt.Sprintf(" AND lr.leave_type = $%d", n)
		args = append(args, f.LeaveType)
		n++
	}
	if f.Query != "" {
		q += fmt.Sprintf(` AND (
            t.first_name ILIKE $%d OR t.last_name ILIKE $%d OR t.employee_id ILIKE $%d OR
            s.name ILIKE $%d OR s.employee_id ILIKE $%d)`, n, n, n, n, n)
		args = append(args, "%"+f.Query+"%")
		n++
	}
	if withLimit {
		q += " ORDER BY lr.created_at DESC"
		if f.Limit > 0 {
			q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", n, n+1)
			args = append(args, f.Limit, f.Offset)
		}
	} else {
		q = "SELECT COUNT(*) FROM (" + q + ") sub"
	}
	return q, args
}

func scanLeaveRecords(rows pgx.Rows) ([]LeaveRequestRecord, error) {
	var items []LeaveRequestRecord
	for rows.Next() {
		var rec LeaveRequestRecord
		var teacherID, staffID *uuid.UUID
		if err := rows.Scan(&rec.ID, &rec.EntityType, &teacherID, &staffID, &rec.LeaveType, &rec.StartDate, &rec.EndDate,
			&rec.Reason, &rec.Status, &rec.ReviewRemarks, &rec.ReviewedAt, &rec.CreatedAt,
			&rec.EmployeeName, &rec.EmployeeID); err != nil {
			return nil, err
		}
		rec.TeacherID = teacherID
		rec.StaffID = staffID
		items = append(items, rec)
	}
	return items, rows.Err()
}
