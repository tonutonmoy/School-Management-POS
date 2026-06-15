package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/repository"
)

type AttendanceService struct {
	repos  *repository.Repositories
	audit  *AuditService
	notify *NotificationService
}

func NewAttendanceService(repos *repository.Repositories, audit *AuditService) *AttendanceService {
	return &AttendanceService{repos: repos, audit: audit}
}

func (s *AttendanceService) SetNotifier(n *NotificationService) { s.notify = n }

func (s *AttendanceService) StudentSheet(ctx context.Context, sessionID, classID, sectionID uuid.UUID, date time.Time) ([]dto.StudentAttendanceRow, error) {
	rows, err := s.repos.Attendance.ListStudentAttendanceSheet(ctx, sessionID, classID, sectionID, date)
	if err != nil {
		return nil, err
	}
	items := make([]dto.StudentAttendanceRow, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.StudentAttendanceRow{
			StudentID: r.StudentID, AdmissionNumber: r.AdmissionNumber, RollNumber: r.RollNumber,
			FullName: r.FirstName + " " + r.LastName, PhotoURL: r.PhotoURL,
			Status: r.Status, Remarks: r.Remarks, RecordID: r.RecordID,
		})
	}
	return items, nil
}

func (s *AttendanceService) BulkMarkStudents(ctx context.Context, sessionID, classID, sectionID uuid.UUID, date time.Time, entries []dto.StudentAttendanceEntry, actorID uuid.UUID, ip string) error {
	for _, e := range entries {
		if !validAttendanceStatus(e.Status) {
			return fmt.Errorf("%w: invalid status for student %s", ErrValidation, e.StudentID)
		}
		if err := s.repos.Attendance.UpsertStudentAttendance(ctx, repository.UpsertStudentAttendanceParams{
			StudentID: e.StudentID, SessionID: sessionID, ClassID: classID, SectionID: sectionID,
			AttendanceDate: date, Status: e.Status, MarkedBy: actorID, Remarks: e.Remarks,
		}); err != nil {
			return err
		}
		if s.notify != nil && e.Status == model.AttendanceAbsent {
			go s.notify.OnAbsent(context.Background(), e.StudentID, date)
		}
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityStudentAttendance, nil, ip, map[string]any{
		"session_id": sessionID, "class_id": classID, "section_id": sectionID,
		"date": date.Format("2006-01-02"), "count": len(entries),
	})
	return nil
}

func (s *AttendanceService) TeacherSheet(ctx context.Context, date time.Time, query string) ([]dto.TeacherAttendanceRow, error) {
	rows, err := s.repos.Attendance.ListTeacherAttendanceSheet(ctx, date, query)
	if err != nil {
		return nil, err
	}
	return mapTeacherRows(rows), nil
}

func (s *AttendanceService) BulkMarkTeachers(ctx context.Context, date time.Time, entries []dto.StudentAttendanceEntry, actorID uuid.UUID, ip string) error {
	for _, e := range entries {
		if !validAttendanceStatus(e.Status) {
			return fmt.Errorf("%w: invalid status", ErrValidation)
		}
		if err := s.repos.Attendance.UpsertTeacherAttendance(ctx, repository.UpsertEmployeeAttendanceParams{
			EmployeeID: e.StudentID, AttendanceDate: date, Status: e.Status, MarkedBy: actorID, Remarks: e.Remarks,
		}); err != nil {
			return err
		}
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityTeacherAttendance, nil, ip, map[string]any{
		"date": date.Format("2006-01-02"), "count": len(entries),
	})
	return nil
}

func (s *AttendanceService) StaffSheet(ctx context.Context, date time.Time, query string) ([]dto.StaffAttendanceRow, error) {
	rows, err := s.repos.Attendance.ListStaffAttendanceSheet(ctx, date, query)
	if err != nil {
		return nil, err
	}
	return mapStaffRows(rows), nil
}

func (s *AttendanceService) BulkMarkStaff(ctx context.Context, date time.Time, entries []dto.StudentAttendanceEntry, actorID uuid.UUID, ip string) error {
	for _, e := range entries {
		if !validAttendanceStatus(e.Status) {
			return fmt.Errorf("%w: invalid status", ErrValidation)
		}
		if err := s.repos.Attendance.UpsertStaffAttendance(ctx, repository.UpsertEmployeeAttendanceParams{
			EmployeeID: e.StudentID, AttendanceDate: date, Status: e.Status, MarkedBy: actorID, Remarks: e.Remarks,
		}); err != nil {
			return err
		}
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityStaffAttendance, nil, ip, map[string]any{
		"date": date.Format("2006-01-02"), "count": len(entries),
	})
	return nil
}

func (s *AttendanceService) DashboardStats(ctx context.Context) (*dto.AttendanceDashboardStats, error) {
	today := time.Now().Truncate(24 * time.Hour)
	present, _ := s.repos.Attendance.CountStudentAttendanceByStatus(ctx, today, model.AttendancePresent)
	absent, _ := s.repos.Attendance.CountStudentAttendanceByStatus(ctx, today, model.AttendanceAbsent)
	teacherPresent, _ := s.repos.Attendance.CountTeacherAttendanceByStatus(ctx, today, model.AttendancePresent)
	staffPresent, _ := s.repos.Attendance.CountStaffAttendanceByStatus(ctx, today, model.AttendancePresent)

	from := today.AddDate(0, -1, 0)
	trend, _ := s.repos.Attendance.MonthlyStudentTrend(ctx, from, today)
	classWise, _ := s.repos.Attendance.ClassWiseAttendanceToday(ctx, today)

	stats := &dto.AttendanceDashboardStats{
		StudentPresentToday: present,
		StudentAbsentToday:  absent,
		TeacherPresentToday: teacherPresent,
		StaffPresentToday:   staffPresent,
	}
	for _, t := range trend {
		stats.MonthlyTrend = append(stats.MonthlyTrend, dto.AttendanceTrendPoint{
			Date: t.Date, Present: t.Present, Absent: t.Absent, Late: t.Late, Leave: t.Leave,
		})
	}
	for _, c := range classWise {
		stats.ClassWiseToday = append(stats.ClassWiseToday, dto.ClassAttendanceSummary{
			ClassID: c.ClassID, ClassName: c.ClassName,
			Present: c.Present, Absent: c.Absent, Late: c.Late, Leave: c.Leave, Total: c.Total,
		})
	}
	return stats, nil
}

func (s *AttendanceService) ApplyLeave(ctx context.Context, req dto.LeaveApplyRequest, actorID uuid.UUID, ip string) (*dto.LeaveRequestResponse, error) {
	p := repository.CreateLeaveParams{
		EntityType: req.EntityType, LeaveType: req.LeaveType,
		StartDate: req.StartDate, EndDate: req.EndDate, Reason: req.Reason, AppliedBy: actorID,
	}
	if req.EntityType == model.LeaveEntityTeacher {
		if req.TeacherID == uuid.Nil {
			return nil, fmt.Errorf("%w: teacher required", ErrValidation)
		}
		p.TeacherID = &req.TeacherID
	} else {
		if req.StaffID == uuid.Nil {
			return nil, fmt.Errorf("%w: staff required", ErrValidation)
		}
		p.StaffID = &req.StaffID
	}
	rec, err := s.repos.Attendance.CreateLeaveRequest(ctx, p)
	if err != nil {
		return nil, err
	}
	resp := mapLeave(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityLeaveRequest, &rec.ID, ip, map[string]any{
		"entity_type": req.EntityType, "leave_type": req.LeaveType,
	})
	return &resp, nil
}

func (s *AttendanceService) ApproveLeave(ctx context.Context, id uuid.UUID, remarks string, actorID uuid.UUID, ip string) (*dto.LeaveRequestResponse, error) {
	return s.reviewLeave(ctx, id, model.LeaveStatusApproved, remarks, actorID, ip)
}

func (s *AttendanceService) RejectLeave(ctx context.Context, id uuid.UUID, remarks string, actorID uuid.UUID, ip string) (*dto.LeaveRequestResponse, error) {
	return s.reviewLeave(ctx, id, model.LeaveStatusRejected, remarks, actorID, ip)
}

func (s *AttendanceService) reviewLeave(ctx context.Context, id uuid.UUID, status, remarks string, actorID uuid.UUID, ip string) (*dto.LeaveRequestResponse, error) {
	existing, err := s.repos.Attendance.GetLeaveByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if existing.Status != model.LeaveStatusPending {
		return nil, fmt.Errorf("%w: leave already reviewed", ErrValidation)
	}
	rec, err := s.repos.Attendance.UpdateLeaveStatus(ctx, id, status, actorID, remarks)
	if err != nil {
		return nil, err
	}
	resp := mapLeave(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityLeaveRequest, &id, ip, map[string]any{"status": status})
	return &resp, nil
}

func (s *AttendanceService) ListLeave(ctx context.Context, f dto.LeaveFilter) (*dto.PaginatedLeaveRequests, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 {
		f.PerPage = 20
	}
	params := repository.LeaveSearchParams{
		EntityType: f.EntityType, Status: f.Status, LeaveType: f.LeaveType, Query: f.Query,
		Limit: int32(f.PerPage), Offset: int32((f.Page - 1) * f.PerPage),
	}
	total, err := s.repos.Attendance.CountLeaveRequests(ctx, params)
	if err != nil {
		return nil, err
	}
	recs, err := s.repos.Attendance.SearchLeaveRequests(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.LeaveRequestResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapLeave(&r))
	}
	return &dto.PaginatedLeaveRequests{
		Items: items, Total: total, Page: f.Page, PerPage: f.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(f.PerPage))),
	}, nil
}

func (s *AttendanceService) StudentAttendanceReport(ctx context.Context, f dto.AttendanceReportFilter) ([]dto.StudentAttendanceResponse, error) {
	params := mapReportParams(f)
	recs, err := s.repos.Attendance.ListStudentAttendanceReport(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapStudentAttendance(recs), nil
}

func (s *AttendanceService) TeacherAttendanceReport(ctx context.Context, f dto.AttendanceReportFilter) ([]dto.StudentAttendanceResponse, error) {
	params := mapReportParams(f)
	recs, err := s.repos.Attendance.ListTeacherAttendanceReport(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapEmployeeAttendance(recs), nil
}

func (s *AttendanceService) StaffAttendanceReport(ctx context.Context, f dto.AttendanceReportFilter) ([]dto.StudentAttendanceResponse, error) {
	params := mapReportParams(f)
	recs, err := s.repos.Attendance.ListStaffAttendanceReport(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapEmployeeAttendance(recs), nil
}

func (s *AttendanceService) StudentHistory(ctx context.Context, studentID uuid.UUID, from, to time.Time, page, perPage int) ([]dto.StudentAttendanceResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 30
	}
	recs, total, err := s.repos.Attendance.StudentAttendanceHistory(ctx, studentID, from, to, int32(perPage), int32((page-1)*perPage))
	if err != nil {
		return nil, 0, err
	}
	return mapStudentAttendance(recs), total, nil
}

func (s *AttendanceService) ParentSummary(ctx context.Context, studentID uuid.UUID) (*dto.StudentAttendanceSummary, error) {
	student, err := s.repos.Students.GetByID(ctx, studentID)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, ErrNotFound
	}
	summary, err := s.repos.Attendance.StudentAttendanceSummary(ctx, studentID, nil, nil)
	if err != nil {
		return nil, err
	}
	total := summary.PresentDays + summary.AbsentDays + summary.LateDays + summary.LeaveDays
	pct := 0.0
	if total > 0 {
		pct = float64(summary.PresentDays+summary.LateDays) / float64(total) * 100
	}
	return &dto.StudentAttendanceSummary{
		StudentID: studentID, StudentName: student.FirstName + " " + student.LastName,
		PresentDays: summary.PresentDays, AbsentDays: summary.AbsentDays,
		LateDays: summary.LateDays, LeaveDays: summary.LeaveDays,
		TotalMarked: total, AttendancePct: pct,
	}, nil
}

func (s *AttendanceService) ClassWiseReport(ctx context.Context, date time.Time) ([]dto.ClassAttendanceSummary, error) {
	recs, err := s.repos.Attendance.ClassWiseAttendanceToday(ctx, date)
	if err != nil {
		return nil, err
	}
	items := make([]dto.ClassAttendanceSummary, 0, len(recs))
	for _, c := range recs {
		items = append(items, dto.ClassAttendanceSummary{
			ClassID: c.ClassID, ClassName: c.ClassName,
			Present: c.Present, Absent: c.Absent, Late: c.Late, Leave: c.Leave, Total: c.Total,
		})
	}
	return items, nil
}

func validAttendanceStatus(s string) bool {
	switch s {
	case model.AttendancePresent, model.AttendanceAbsent, model.AttendanceLate, model.AttendanceLeave:
		return true
	}
	return false
}

func mapReportParams(f dto.AttendanceReportFilter) repository.AttendanceReportParams {
	p := repository.AttendanceReportParams{From: f.From, To: f.To, Status: f.Status}
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
	if f.TeacherID != uuid.Nil {
		p.TeacherID = &f.TeacherID
	}
	if f.StaffID != uuid.Nil {
		p.StaffID = &f.StaffID
	}
	return p
}

func mapStudentAttendance(recs []repository.StudentAttendanceRecord) []dto.StudentAttendanceResponse {
	items := make([]dto.StudentAttendanceResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, dto.StudentAttendanceResponse{
			ID: r.ID, StudentID: r.StudentID, StudentName: r.StudentName,
			AdmissionNumber: r.AdmissionNumber, RollNumber: r.RollNumber,
			SessionID: r.SessionID, ClassID: r.ClassID, ClassName: r.ClassName,
			SectionID: r.SectionID, SectionName: r.SectionName,
			AttendanceDate: r.AttendanceDate, Status: r.Status, Remarks: r.Remarks,
		})
	}
	return items
}

func mapEmployeeAttendance(recs []repository.EmployeeAttendanceRecord) []dto.StudentAttendanceResponse {
	items := make([]dto.StudentAttendanceResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, dto.StudentAttendanceResponse{
			ID: r.ID, StudentID: r.EmployeeID, StudentName: r.EmployeeName,
			AdmissionNumber: r.EmployeeCode, ClassName: r.DepartmentName,
			AttendanceDate: r.AttendanceDate, Status: r.Status, Remarks: r.Remarks,
		})
	}
	return items
}

func mapTeacherRows(rows []repository.EmployeeAttendanceSheetRow) []dto.TeacherAttendanceRow {
	items := make([]dto.TeacherAttendanceRow, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.TeacherAttendanceRow{
			TeacherID: r.EmployeeID, EmployeeID: r.EmpCode, FullName: r.Name,
			PhotoURL: r.PhotoURL, Department: r.Department,
			Status: r.Status, Remarks: r.Remarks, RecordID: r.RecordID,
		})
	}
	return items
}

func mapStaffRows(rows []repository.EmployeeAttendanceSheetRow) []dto.StaffAttendanceRow {
	items := make([]dto.StaffAttendanceRow, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.StaffAttendanceRow{
			StaffID: r.EmployeeID, EmployeeID: r.EmpCode, Name: r.Name,
			PhotoURL: r.PhotoURL, Department: r.Department,
			Status: r.Status, Remarks: r.Remarks, RecordID: r.RecordID,
		})
	}
	return items
}

func mapLeave(r *repository.LeaveRequestRecord) dto.LeaveRequestResponse {
	return dto.LeaveRequestResponse{
		ID: r.ID, EntityType: r.EntityType, EmployeeName: r.EmployeeName, EmployeeID: r.EmployeeID,
		LeaveType: r.LeaveType, StartDate: r.StartDate, EndDate: r.EndDate,
		Reason: r.Reason, Status: r.Status, ReviewRemarks: r.ReviewRemarks,
		ReviewedAt: r.ReviewedAt, CreatedAt: r.CreatedAt,
	}
}
