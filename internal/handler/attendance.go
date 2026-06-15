package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/export"
	"github.com/school-management/pos/internal/middleware"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/validator"
	"github.com/school-management/pos/internal/web"
)

func (h *Handler) registerAttendanceRoutes(auth fiber.Router, mw *middleware.Middleware) {
	// Student attendance
	students := auth.Group("/attendance/students")
	students.Get("/", mw.RequirePermission(model.PermAttendanceStudentView), h.StudentAttendancePage)
	students.Post("/", mw.CSRFProtect(), mw.RequirePermission(model.PermAttendanceStudentMark), h.StudentAttendanceSave)

	// Teacher attendance
	teachers := auth.Group("/attendance/teachers")
	teachers.Get("/", mw.RequirePermission(model.PermAttendanceTeacherView), h.TeacherAttendancePage)
	teachers.Post("/", mw.CSRFProtect(), mw.RequirePermission(model.PermAttendanceTeacherMark), h.TeacherAttendanceSave)

	// Staff attendance
	staff := auth.Group("/attendance/staff")
	staff.Get("/", mw.RequirePermission(model.PermAttendanceStaffView), h.StaffAttendancePage)
	staff.Post("/", mw.CSRFProtect(), mw.RequirePermission(model.PermAttendanceStaffMark), h.StaffAttendanceSave)

	// Attendance dashboard
	auth.Get("/attendance/dashboard", mw.RequirePermission(model.PermAttendanceStudentView), h.AttendanceDashboard)

	// Leave management
	leave := auth.Group("/leave")
	leave.Get("/", h.LeaveList)
	leave.Get("/apply", h.LeaveApplyPage)
	leave.Post("/", mw.CSRFProtect(), h.LeaveApply)
	leave.Post("/:id/approve", mw.CSRFProtect(), mw.RequirePermission(model.PermLeaveApprove), h.LeaveApprove)
	leave.Post("/:id/reject", mw.CSRFProtect(), mw.RequirePermission(model.PermLeaveReject), h.LeaveReject)

	// Parent view foundation
	auth.Get("/parent/students/:id/attendance", mw.RequirePermission(model.PermAttendanceStudentView), h.ParentAttendanceSummary)

	// Reports
	reports := auth.Group("/reports/attendance")
	reports.Get("/students/daily", mw.RequirePermission(model.PermAttendanceStudentView), h.ReportStudentDaily)
	reports.Get("/students/monthly", mw.RequirePermission(model.PermAttendanceStudentView), h.ReportStudentMonthly)
	reports.Get("/students/by-class", mw.RequirePermission(model.PermAttendanceStudentView), h.ReportClassWiseAttendance)
	reports.Get("/students/:id/history", mw.RequirePermission(model.PermAttendanceStudentView), h.ReportStudentHistory)
	reports.Get("/teachers/daily", mw.RequirePermission(model.PermAttendanceTeacherView), h.ReportTeacherDaily)
	reports.Get("/teachers/monthly", mw.RequirePermission(model.PermAttendanceTeacherView), h.ReportTeacherMonthly)
	reports.Get("/staff/daily", mw.RequirePermission(model.PermAttendanceStaffView), h.ReportStaffDaily)
	reports.Get("/staff/monthly", mw.RequirePermission(model.PermAttendanceStaffView), h.ReportStaffMonthly)
	reports.Get("/students/export.csv", mw.RequirePermission(model.PermAttendanceStudentView), h.ExportStudentAttendanceCSV)
	reports.Get("/students/export.xlsx", mw.RequirePermission(model.PermAttendanceStudentView), h.ExportStudentAttendanceExcel)
	reports.Get("/teachers/export.csv", mw.RequirePermission(model.PermAttendanceTeacherView), h.ExportTeacherAttendanceCSV)
	reports.Get("/staff/export.csv", mw.RequirePermission(model.PermAttendanceStaffView), h.ExportStaffAttendanceCSV)
}

func (h *Handler) attendanceFormData(c fiber.Ctx) *web.AttendanceFormData {
	sessions, _ := h.services.Sessions.List(c.Context())
	classes, _ := h.services.Academic.ListClasses(c.Context())
	sections, _ := h.services.Academic.ListSections(c.Context())
	teachers, _ := h.services.HR.ListTeachersReport(c.Context(), dto.HRReportFilter{})
	staff, _ := h.services.HR.ListStaffReport(c.Context(), dto.HRReportFilter{})
	return &web.AttendanceFormData{
		Sessions: sessions, Classes: classes, Sections: sections, Teachers: teachers, Staff: staff,
	}
}

func (h *Handler) parseAttendanceDate(c fiber.Ctx) time.Time {
	if d, err := parseDate(c.Query("date", c.FormValue("date"))); err == nil {
		return d
	}
	return time.Now().Truncate(24 * time.Hour)
}

func (h *Handler) StudentAttendancePage(c fiber.Ctx) error {
	date := h.parseAttendanceDate(c)
	filter := dto.AttendanceFilter{Date: date}
	if s := c.Query("session_id"); s != "" {
		filter.SessionID, _ = uuid.Parse(s)
	}
	if cl := c.Query("class_id"); cl != "" {
		filter.ClassID, _ = uuid.Parse(cl)
	}
	if sec := c.Query("section_id"); sec != "" {
		filter.SectionID, _ = uuid.Parse(sec)
	}
	var rows []dto.StudentAttendanceRow
	if filter.SessionID != uuid.Nil && filter.ClassID != uuid.Nil && filter.SectionID != uuid.Nil {
		rows, _ = h.services.Attendance.StudentSheet(c.Context(), filter.SessionID, filter.ClassID, filter.SectionID, date)
	}
	return h.render(c, fiber.StatusOK, web.StudentAttendancePage{
		Rows: rows, Filter: filter, FormData: h.attendanceFormData(c),
		Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
	})
}

func (h *Handler) StudentAttendanceSave(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	date := h.parseAttendanceDate(c)
	sessionID, _ := uuid.Parse(c.FormValue("session_id"))
	classID, _ := uuid.Parse(c.FormValue("class_id"))
	sectionID, _ := uuid.Parse(c.FormValue("section_id"))
	if sessionID == uuid.Nil || classID == uuid.Nil || sectionID == uuid.Nil {
		h.flash(c, "Session, class and section are required", true)
		return c.Redirect().To("/attendance/students")
	}
	entries := parseBulkAttendanceEntries(c)
	if len(entries) == 0 {
		h.flash(c, "No attendance entries to save", true)
		return c.Redirect().To(h.studentAttendanceRedirect(sessionID, classID, sectionID, date))
	}
	if err := h.services.Attendance.BulkMarkStudents(c.Context(), sessionID, classID, sectionID, date, entries, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To(h.studentAttendanceRedirect(sessionID, classID, sectionID, date))
	}
	h.flash(c, "Student attendance saved", false)
	return c.Redirect().To(h.studentAttendanceRedirect(sessionID, classID, sectionID, date))
}

func (h *Handler) studentAttendanceRedirect(sessionID, classID, sectionID uuid.UUID, date time.Time) string {
	return "/attendance/students?session_id=" + sessionID.String() +
		"&class_id=" + classID.String() + "&section_id=" + sectionID.String() +
		"&date=" + date.Format("2006-01-02")
}

func (h *Handler) TeacherAttendancePage(c fiber.Ctx) error {
	date := h.parseAttendanceDate(c)
	query := c.Query("q")
	rows, _ := h.services.Attendance.TeacherSheet(c.Context(), date, query)
	return h.render(c, fiber.StatusOK, web.TeacherAttendancePage{
		Rows: rows, Date: date, Query: query,
		Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
	})
}

func (h *Handler) TeacherAttendanceSave(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	date := h.parseAttendanceDate(c)
	entries := parseBulkAttendanceEntries(c)
	if len(entries) == 0 {
		h.flash(c, "No attendance entries to save", true)
		return c.Redirect().To("/attendance/teachers?date=" + date.Format("2006-01-02"))
	}
	if err := h.services.Attendance.BulkMarkTeachers(c.Context(), date, entries, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/attendance/teachers?date=" + date.Format("2006-01-02"))
	}
	h.flash(c, "Teacher attendance saved", false)
	return c.Redirect().To("/attendance/teachers?date=" + date.Format("2006-01-02"))
}

func (h *Handler) StaffAttendancePage(c fiber.Ctx) error {
	date := h.parseAttendanceDate(c)
	query := c.Query("q")
	rows, _ := h.services.Attendance.StaffSheet(c.Context(), date, query)
	return h.render(c, fiber.StatusOK, web.StaffAttendancePage{
		Rows: rows, Date: date, Query: query,
		Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
	})
}

func (h *Handler) StaffAttendanceSave(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	date := h.parseAttendanceDate(c)
	entries := parseBulkAttendanceEntries(c)
	if len(entries) == 0 {
		h.flash(c, "No attendance entries to save", true)
		return c.Redirect().To("/attendance/staff?date=" + date.Format("2006-01-02"))
	}
	if err := h.services.Attendance.BulkMarkStaff(c.Context(), date, entries, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/attendance/staff?date=" + date.Format("2006-01-02"))
	}
	h.flash(c, "Staff attendance saved", false)
	return c.Redirect().To("/attendance/staff?date=" + date.Format("2006-01-02"))
}

func (h *Handler) AttendanceDashboard(c fiber.Ctx) error {
	stats, err := h.services.Attendance.DashboardStats(c.Context())
	if err != nil {
		return c.Status(500).SendString("Failed to load attendance dashboard")
	}
	return h.render(c, fiber.StatusOK, web.AttendanceDashboardPage{Stats: stats})
}

func (h *Handler) LeaveList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	filter := dto.LeaveFilter{
		EntityType: c.Query("entity_type"), Status: c.Query("status"),
		LeaveType: c.Query("leave_type"), Query: c.Query("q"), Page: page, PerPage: 20,
	}
	data, err := h.services.Attendance.ListLeave(c.Context(), filter)
	if err != nil {
		return c.Status(500).SendString("Failed to load leave requests")
	}
	return h.render(c, fiber.StatusOK, web.LeaveListPage{Data: data, Filter: filter})
}

func (h *Handler) LeaveApplyPage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.LeaveApplyPage{FormData: h.attendanceFormData(c)})
}

func (h *Handler) LeaveApply(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req := dto.LeaveApplyRequest{
		EntityType: c.FormValue("entity_type"), LeaveType: c.FormValue("leave_type"),
		Reason: c.FormValue("reason"),
	}
	if t := c.FormValue("teacher_id"); t != "" {
		req.TeacherID, _ = uuid.Parse(t)
	}
	if s := c.FormValue("staff_id"); s != "" {
		req.StaffID, _ = uuid.Parse(s)
	}
	if sd, err := parseDate(c.FormValue("start_date")); err == nil {
		req.StartDate = sd
	}
	if ed, err := parseDate(c.FormValue("end_date")); err == nil {
		req.EndDate = ed
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/leave/apply")
	}
	if _, err := h.services.Attendance.ApplyLeave(c.Context(), req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/leave/apply")
	}
	h.flash(c, "Leave request submitted", false)
	return c.Redirect().To("/leave")
}

func (h *Handler) LeaveApprove(c fiber.Ctx) error {
	return h.leaveReview(c, true)
}

func (h *Handler) LeaveReject(c fiber.Ctx) error {
	return h.leaveReview(c, false)
}

func (h *Handler) leaveReview(c fiber.Ctx, approve bool) error {
	user := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	remarks := c.FormValue("review_remarks")
	var svcErr error
	if approve {
		_, svcErr = h.services.Attendance.ApproveLeave(c.Context(), id, remarks, user.ID, c.IP())
	} else {
		_, svcErr = h.services.Attendance.RejectLeave(c.Context(), id, remarks, user.ID, c.IP())
	}
	if svcErr != nil {
		h.flash(c, svcErr.Error(), true)
	}
	return c.Redirect().To("/leave")
}

func (h *Handler) ParentAttendanceSummary(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	if err := h.requireStudentAccess(c, id, model.PermAttendanceStudentView); err != nil {
		return err
	}
	summary, err := h.services.Attendance.ParentSummary(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Student not found")
	}
	return h.render(c, fiber.StatusOK, web.ParentAttendancePage{Summary: summary})
}

func (h *Handler) parseAttendanceReportFilter(c fiber.Ctx) dto.AttendanceReportFilter {
	f := dto.AttendanceReportFilter{Status: c.Query("status")}
	if d, err := parseDate(c.Query("date")); err == nil {
		f.From, f.To = d, d
	}
	if from, err := parseDate(c.Query("from")); err == nil {
		f.From = from
	}
	if to, err := parseDate(c.Query("to")); err == nil {
		f.To = to
	}
	if f.From.IsZero() && f.To.IsZero() {
		today := time.Now().Truncate(24 * time.Hour)
		f.From, f.To = today, today
	}
	if s := c.Query("session_id"); s != "" {
		f.SessionID, _ = uuid.Parse(s)
	}
	if cl := c.Query("class_id"); cl != "" {
		f.ClassID, _ = uuid.Parse(cl)
	}
	if sec := c.Query("section_id"); sec != "" {
		f.SectionID, _ = uuid.Parse(sec)
	}
	if st := c.Query("student_id"); st != "" {
		f.StudentID, _ = uuid.Parse(st)
	}
	return f
}

func (h *Handler) ReportStudentDaily(c fiber.Ctx) error {
	f := h.parseAttendanceReportFilter(c)
	recs, _ := h.services.Attendance.StudentAttendanceReport(c.Context(), f)
	return h.render(c, fiber.StatusOK, web.StudentAttendanceReportPage{
		Title: "Daily Student Attendance", Records: recs, Filter: f, FormData: h.attendanceFormData(c),
	})
}

func (h *Handler) ReportStudentMonthly(c fiber.Ctx) error {
	f := h.parseAttendanceReportFilter(c)
	if f.From.IsZero() {
		now := time.Now()
		f.From = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		f.To = now
	}
	recs, _ := h.services.Attendance.StudentAttendanceReport(c.Context(), f)
	return h.render(c, fiber.StatusOK, web.StudentAttendanceReportPage{
		Title: "Monthly Student Attendance", Records: recs, Filter: f, FormData: h.attendanceFormData(c),
	})
}

func (h *Handler) ReportClassWiseAttendance(c fiber.Ctx) error {
	date := h.parseAttendanceDate(c)
	groups, _ := h.services.Attendance.ClassWiseReport(c.Context(), date)
	return h.render(c, fiber.StatusOK, web.ClassWiseAttendanceReportPage{Groups: groups, Date: date})
}

func (h *Handler) ReportStudentHistory(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	now := time.Now()
	from := time.Date(now.Year(), now.Month()-6, 1, 0, 0, 0, 0, now.Location())
	if f, err := parseDate(c.Query("from")); err == nil {
		from = f
	}
	to := now
	if t, err := parseDate(c.Query("to")); err == nil {
		to = t
	}
	recs, total, _ := h.services.Attendance.StudentHistory(c.Context(), id, from, to, page, 30)
	summary, _ := h.services.Attendance.ParentSummary(c.Context(), id)
	return h.render(c, fiber.StatusOK, web.StudentHistoryReportPage{
		Records: recs, Summary: summary, Total: total, Page: page,
		StudentID: id, From: from, To: to,
	})
}

func (h *Handler) ReportTeacherDaily(c fiber.Ctx) error {
	f := h.parseAttendanceReportFilter(c)
	recs, _ := h.services.Attendance.TeacherAttendanceReport(c.Context(), f)
	return h.render(c, fiber.StatusOK, web.EmployeeAttendanceReportPage{
		Title: "Daily Teacher Attendance", Records: recs, Filter: f, Entity: "teacher",
	})
}

func (h *Handler) ReportTeacherMonthly(c fiber.Ctx) error {
	f := h.parseAttendanceReportFilter(c)
	if f.From.IsZero() {
		now := time.Now()
		f.From = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		f.To = now
	}
	recs, _ := h.services.Attendance.TeacherAttendanceReport(c.Context(), f)
	return h.render(c, fiber.StatusOK, web.EmployeeAttendanceReportPage{
		Title: "Monthly Teacher Attendance", Records: recs, Filter: f, Entity: "teacher",
	})
}

func (h *Handler) ReportStaffDaily(c fiber.Ctx) error {
	f := h.parseAttendanceReportFilter(c)
	recs, _ := h.services.Attendance.StaffAttendanceReport(c.Context(), f)
	return h.render(c, fiber.StatusOK, web.EmployeeAttendanceReportPage{
		Title: "Daily Staff Attendance", Records: recs, Filter: f, Entity: "staff",
	})
}

func (h *Handler) ReportStaffMonthly(c fiber.Ctx) error {
	f := h.parseAttendanceReportFilter(c)
	if f.From.IsZero() {
		now := time.Now()
		f.From = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		f.To = now
	}
	recs, _ := h.services.Attendance.StaffAttendanceReport(c.Context(), f)
	return h.render(c, fiber.StatusOK, web.EmployeeAttendanceReportPage{
		Title: "Monthly Staff Attendance", Records: recs, Filter: f, Entity: "staff",
	})
}

func (h *Handler) ExportStudentAttendanceCSV(c fiber.Ctx) error {
	f := h.parseAttendanceReportFilter(c)
	recs, err := h.services.Attendance.StudentAttendanceReport(c.Context(), f)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	data, err := export.StudentAttendanceCSV(recs)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=student-attendance.csv")
	return c.Send(data)
}

func (h *Handler) ExportStudentAttendanceExcel(c fiber.Ctx) error {
	f := h.parseAttendanceReportFilter(c)
	recs, err := h.services.Attendance.StudentAttendanceReport(c.Context(), f)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	data, err := export.StudentAttendanceExcel(recs)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=student-attendance.xlsx")
	return c.Send(data)
}

func (h *Handler) ExportTeacherAttendanceCSV(c fiber.Ctx) error {
	f := h.parseAttendanceReportFilter(c)
	recs, err := h.services.Attendance.TeacherAttendanceReport(c.Context(), f)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	data, _ := export.EmployeeAttendanceCSV(recs)
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=teacher-attendance.csv")
	return c.Send(data)
}

func (h *Handler) ExportStaffAttendanceCSV(c fiber.Ctx) error {
	f := h.parseAttendanceReportFilter(c)
	recs, err := h.services.Attendance.StaffAttendanceReport(c.Context(), f)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	data, _ := export.EmployeeAttendanceCSV(recs)
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=staff-attendance.csv")
	return c.Send(data)
}

func parseBulkAttendanceEntries(c fiber.Ctx) []dto.StudentAttendanceEntry {
	var entries []dto.StudentAttendanceEntry
	seen := map[string]bool{}
	c.Request().PostArgs().VisitAll(func(key, value []byte) {
		k := string(key)
		if len(k) < 8 || k[:8] != "status_" {
			return
		}
		idStr := k[8:]
		if seen[idStr] {
			return
		}
		status := string(value)
		if status == "" {
			return
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return
		}
		seen[idStr] = true
		entries = append(entries, dto.StudentAttendanceEntry{
			StudentID: id, Status: status, Remarks: c.FormValue("remarks_" + idStr),
		})
	})
	return entries
}
