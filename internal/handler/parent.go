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

func (h *Handler) registerParentPortalRoutes(auth fiber.Router, mw *middleware.Middleware) {
	portal := auth.Group("/parent", mw.RequireParent())
	portal.Get("/dashboard", h.ParentDashboard)
	portal.Get("/profile", h.ParentProfilePage)
	portal.Post("/profile", mw.CSRFProtect(), h.ParentProfileUpdate)
	portal.Get("/children", h.ParentChildrenPage)
	portal.Get("/notices", h.ParentNoticesPage)
	portal.Get("/notices/:id", h.ParentNoticeView)
	portal.Post("/notices/:id/read", mw.CSRFProtect(), h.ParentNoticeMarkRead)
	portal.Get("/notifications", h.ParentNotificationsPage)
	portal.Post("/notifications/:id/read", mw.CSRFProtect(), h.ParentNotificationMarkRead)
	portal.Post("/notifications/read-all", mw.CSRFProtect(), h.ParentNotificationMarkAllRead)

	portal.Get("/students/:id/attendance", h.ParentPortalAttendance)
	portal.Get("/students/:id/fees", h.ParentPortalFees)
	portal.Get("/students/:id/results", h.ParentPortalResultsList)
	portal.Get("/students/:id/results/:examId", h.ParentPortalResultView)
	portal.Get("/students/:id/results/:examId/report-card", h.ParentPortalReportCardPDF)

	admin := auth.Group("/parents", mw.RequirePermission(model.PermParentView))
	admin.Get("/", h.ParentList)
	admin.Get("/new", h.ParentCreatePage)
	admin.Post("/", mw.CSRFProtect(), h.ParentCreate)
	admin.Get("/:id", h.ParentDetail)
	admin.Get("/:id/edit", h.ParentEditPage)
	admin.Post("/:id", mw.CSRFProtect(), h.ParentUpdate)
	admin.Post("/:id/delete", mw.CSRFProtect(), h.ParentDelete)
	admin.Post("/:id/link-student", mw.CSRFProtect(), h.ParentLinkStudent)
	admin.Post("/:id/unlink-student", mw.CSRFProtect(), h.ParentUnlinkStudent)

	notices := auth.Group("/notices")
	notices.Get("/", mw.RequirePermission(model.PermParentView), h.NoticeList)
	notices.Get("/new", mw.RequirePermission(model.PermNoticeCreate), h.NoticeCreatePage)
	notices.Post("/", mw.CSRFProtect(), mw.RequirePermission(model.PermNoticeCreate), h.NoticeCreate)
	notices.Get("/:id/edit", mw.RequirePermission(model.PermNoticeUpdate), h.NoticeEditPage)
	notices.Post("/:id", mw.CSRFProtect(), mw.RequirePermission(model.PermNoticeUpdate), h.NoticeUpdate)
	notices.Post("/:id/delete", mw.CSRFProtect(), mw.RequirePermission(model.PermNoticeDelete), h.NoticeDelete)

	comm := auth.Group("/communications", mw.RequirePermission(model.PermNotificationSend))
	comm.Get("/dashboard", h.CommunicationDashboard)
	comm.Get("/sms/export.csv", h.ExportSMSLogsCSV)
}

func (h *Handler) parentUser(c fiber.Ctx) *dto.AuthUser {
	return middleware.GetUser(c)
}

func (h *Handler) requireStudentAccess(c fiber.Ctx, studentID uuid.UUID, staffPerms ...string) error {
	user := h.parentUser(c)
	if !h.services.Parent.CanViewStudent(c.Context(), user, studentID, staffPerms...) {
		return c.Status(fiber.StatusForbidden).SendString("Forbidden")
	}
	return nil
}

func (h *Handler) ParentDashboard(c fiber.Ctx) error {
	user := h.parentUser(c)
	if user.RoleSlug != model.RoleParent {
		return c.Redirect().To("/parents")
	}
	stats, err := h.services.Parent.Dashboard(c.Context(), user.ID)
	if err != nil {
		return c.Status(500).SendString("Unable to load dashboard")
	}
	return h.render(c, fiber.StatusOK, web.ParentDashboardPage{
		Stats: stats, Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
	})
}

func (h *Handler) ParentProfilePage(c fiber.Ctx) error {
	user := h.parentUser(c)
	profile, err := h.services.Parent.GetProfile(c.Context(), user.ID)
	if err != nil {
		return c.Status(404).SendString("Profile not found")
	}
	return h.render(c, fiber.StatusOK, web.ParentProfilePage{
		Profile: profile, Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
	})
}

func (h *Handler) ParentProfileUpdate(c fiber.Ctx) error {
	user := h.parentUser(c)
	req := dto.UpdateParentProfileRequest{
		Phone: c.FormValue("phone"), Address: c.FormValue("address"), Occupation: c.FormValue("occupation"),
	}
	if _, err := h.services.Parent.UpdateProfile(c.Context(), user.ID, req); err != nil {
		h.flash(c, "Unable to update profile", true)
		return c.Redirect().To("/parent/profile")
	}
	h.flash(c, "Profile updated", false)
	return c.Redirect().To("/parent/profile")
}

func (h *Handler) ParentChildrenPage(c fiber.Ctx) error {
	user := h.parentUser(c)
	children, err := h.services.Parent.ListChildren(c.Context(), user.ID)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.ParentChildrenPage{Children: children})
}

func (h *Handler) ParentPortalAttendance(c fiber.Ctx) error {
	studentID, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	if err := h.requireStudentAccess(c, studentID, model.PermAttendanceStudentView); err != nil {
		return err
	}
	user := h.parentUser(c)
	month := time.Now()
	if m, e := parseDate(c.Query("month")); e == nil {
		month = m
	}
	var view *dto.ParentAttendanceView
	if user.RoleSlug == model.RoleParent {
		view, err = h.services.Parent.AttendanceView(c.Context(), user.ID, studentID, month)
	} else {
		view, err = h.services.Parent.StaffAttendanceView(c.Context(), studentID, month)
	}
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.ParentPortalAttendancePage{View: view})
}

func (h *Handler) ParentPortalFees(c fiber.Ctx) error {
	studentID, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	if err := h.requireStudentAccess(c, studentID, model.PermFeeReportView); err != nil {
		return err
	}
	summary, err := h.services.Fees.ParentFeeSummary(c.Context(), studentID)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	gateways, _ := h.services.Payment.ListGateways(c.Context())
	active := make([]dto.PaymentGatewayResponse, 0)
	for _, g := range gateways {
		if g.IsActive {
			active = append(active, g)
		}
	}
	online, _ := h.services.Payment.StudentPaymentHistory(c.Context(), studentID)
	return h.render(c, fiber.StatusOK, web.ParentPortalFeePage{Summary: summary, Gateways: active, Online: online})
}

func (h *Handler) ParentPortalResultsList(c fiber.Ctx) error {
	studentID, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	if err := h.requireStudentAccess(c, studentID, model.PermResultPublish); err != nil {
		return err
	}
	student, err := h.services.Students.GetFull(c.Context(), studentID)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	data, err := h.services.Exam.ListResults(c.Context(), dto.ExamReportFilter{StudentID: studentID, PublishedOnly: true}, 1, 50)
	if err != nil {
		return c.Status(500).SendString("Error")
	}
	return h.render(c, fiber.StatusOK, web.ParentPortalResultsPage{Student: student, Results: data.Items})
}

func (h *Handler) ParentPortalResultView(c fiber.Ctx) error {
	studentID, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	examID, err := parseUUIDParam(c, "examId")
	if err != nil {
		return c.Status(400).SendString("Invalid exam")
	}
	if err := h.requireStudentAccess(c, studentID, model.PermResultPublish); err != nil {
		return err
	}
	result, err := h.services.Exam.ParentResult(c.Context(), examID, studentID)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.ParentPortalResultPage{Result: result, StudentID: studentID})
}

func (h *Handler) ParentPortalReportCardPDF(c fiber.Ctx) error {
	studentID, _ := parseUUIDParam(c, "id")
	if err := h.requireStudentAccess(c, studentID, model.PermResultPublish); err != nil {
		return err
	}
	return h.ParentReportCardPDF(c)
}

func (h *Handler) ParentNoticesPage(c fiber.Ctx) error {
	user := h.parentUser(c)
	parent, _ := h.services.Parent.GetProfile(c.Context(), user.ID)
	if parent == nil {
		return c.Status(404).SendString("Not found")
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	data, err := h.services.Notice.ListForParent(c.Context(), parent.ID, dto.NoticeFilter{
		Query: c.Query("q"), NoticeType: c.Query("type"), Page: page, PageSize: 20,
	})
	if err != nil {
		return c.Status(500).SendString("Error")
	}
	return h.render(c, fiber.StatusOK, web.ParentNoticesPage{Data: data})
}

func (h *Handler) ParentNoticeView(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	notice, err := h.services.Notice.Get(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	user := h.parentUser(c)
	if profile, _ := h.services.Parent.GetProfile(c.Context(), user.ID); profile != nil {
		_ = h.services.Notice.MarkRead(c.Context(), id, profile.ID)
	}
	return h.render(c, fiber.StatusOK, web.ParentNoticeDetailPage{Notice: notice})
}

func (h *Handler) ParentNoticeMarkRead(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	user := h.parentUser(c)
	if profile, _ := h.services.Parent.GetProfile(c.Context(), user.ID); profile != nil {
		_ = h.services.Notice.MarkRead(c.Context(), id, profile.ID)
	}
	return c.Redirect().To("/parent/notices")
}

func (h *Handler) ParentNotificationsPage(c fiber.Ctx) error {
	user := h.parentUser(c)
	profile, err := h.services.Parent.GetProfile(c.Context(), user.ID)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	data, err := h.services.Notification.ListForParent(c.Context(), profile.ID, dto.NotificationFilter{
		UnreadOnly: c.Query("unread") == "1", Page: page, PageSize: 20,
	})
	if err != nil {
		return c.Status(500).SendString("Error")
	}
	return h.render(c, fiber.StatusOK, web.ParentNotificationsPage{Data: data})
}

func (h *Handler) ParentNotificationMarkRead(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	user := h.parentUser(c)
	if profile, _ := h.services.Parent.GetProfile(c.Context(), user.ID); profile != nil {
		_ = h.services.Notification.MarkRead(c.Context(), id, profile.ID)
	}
	return c.Redirect().To("/parent/notifications")
}

func (h *Handler) ParentNotificationMarkAllRead(c fiber.Ctx) error {
	user := h.parentUser(c)
	if profile, _ := h.services.Parent.GetProfile(c.Context(), user.ID); profile != nil {
		_ = h.services.Notification.MarkAllRead(c.Context(), profile.ID)
	}
	h.flash(c, "All notifications marked read", false)
	return c.Redirect().To("/parent/notifications")
}

// Admin parent management

func (h *Handler) ParentList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	data, err := h.services.Parent.List(c.Context(), page, 20)
	if err != nil {
		return c.Status(500).SendString("Error")
	}
	students, _ := h.services.Students.Search(c.Context(), dto.StudentSearchFilter{Page: 1, PageSize: 500})
	return h.render(c, fiber.StatusOK, web.ParentListPage{Data: data, Students: students.Items})
}

func (h *Handler) ParentCreatePage(c fiber.Ctx) error {
	students, _ := h.services.Students.Search(c.Context(), dto.StudentSearchFilter{Page: 1, PageSize: 500})
	return h.render(c, fiber.StatusOK, web.ParentFormPage{Title: "Create Parent", Students: students.Items})
}

func (h *Handler) ParentCreate(c fiber.Ctx) error {
	user := h.parentUser(c)
	req := dto.CreateParentRequest{
		Email: c.FormValue("email"), Password: c.FormValue("password"),
		FirstName: c.FormValue("first_name"), LastName: c.FormValue("last_name"),
		Phone: c.FormValue("phone"), Address: c.FormValue("address"), Occupation: c.FormValue("occupation"),
	}
	if sid := c.FormValue("student_id"); sid != "" {
		if id, err := uuid.Parse(sid); err == nil {
			req.StudentLinks = []dto.ParentLinkInput{{StudentID: id, Relationship: c.FormValue("relationship"), IsPrimary: true}}
		}
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/parents/new")
	}
	if _, err := h.services.Parent.Create(c.Context(), req, user.ID, c.IP()); err != nil {
		h.flash(c, "Unable to create parent account", true)
		return c.Redirect().To("/parents/new")
	}
	h.flash(c, "Parent account created", false)
	return c.Redirect().To("/parents")
}

func (h *Handler) ParentDetail(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	parent, err := h.services.Parent.Get(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	children, _ := h.services.Parent.ListChildren(c.Context(), parent.UserID)
	students, _ := h.services.Students.Search(c.Context(), dto.StudentSearchFilter{Page: 1, PageSize: 500})
	return h.render(c, fiber.StatusOK, web.ParentDetailPage{Parent: parent, Children: children, Students: students.Items})
}

func (h *Handler) ParentEditPage(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	parent, err := h.services.Parent.Get(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.ParentFormPage{Title: "Edit Parent", Parent: parent})
}

func (h *Handler) ParentUpdate(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	user := h.parentUser(c)
	req := dto.UpdateParentRequest{
		FirstName: c.FormValue("first_name"), LastName: c.FormValue("last_name"),
		Phone: c.FormValue("phone"), Address: c.FormValue("address"), Occupation: c.FormValue("occupation"),
		IsActive: c.FormValue("is_active") == "on" || c.FormValue("is_active") == "true",
	}
	if _, err := h.services.Parent.Update(c.Context(), id, req, user.ID, c.IP()); err != nil {
		h.flash(c, "Update failed", true)
		return c.Redirect().To("/parents/" + id.String() + "/edit")
	}
	h.flash(c, "Parent updated", false)
	return c.Redirect().To("/parents/" + id.String())
}

func (h *Handler) ParentDelete(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	user := h.parentUser(c)
	if err := h.services.Parent.Delete(c.Context(), id, user.ID, c.IP()); err != nil {
		h.flash(c, "Delete failed", true)
	} else {
		h.flash(c, "Parent deleted", false)
	}
	return c.Redirect().To("/parents")
}

func (h *Handler) ParentLinkStudent(c fiber.Ctx) error {
	parentID, _ := parseUUIDParam(c, "id")
	studentID, err := uuid.Parse(c.FormValue("student_id"))
	if err != nil {
		h.flash(c, "Invalid student", true)
		return c.Redirect().To("/parents/" + parentID.String())
	}
	user := h.parentUser(c)
	link := dto.ParentLinkInput{StudentID: studentID, Relationship: c.FormValue("relationship"), IsPrimary: c.FormValue("is_primary") == "on"}
	if err := h.services.Parent.LinkStudent(c.Context(), parentID, link, user.ID, c.IP()); err != nil {
		h.flash(c, "Link failed", true)
	} else {
		h.flash(c, "Student linked", false)
	}
	return c.Redirect().To("/parents/" + parentID.String())
}

func (h *Handler) ParentUnlinkStudent(c fiber.Ctx) error {
	parentID, _ := parseUUIDParam(c, "id")
	studentID, _ := uuid.Parse(c.FormValue("student_id"))
	user := h.parentUser(c)
	_ = h.services.Parent.UnlinkStudent(c.Context(), parentID, studentID, user.ID, c.IP())
	h.flash(c, "Student unlinked", false)
	return c.Redirect().To("/parents/" + parentID.String())
}

// Notices (admin)

func (h *Handler) NoticeList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	data, err := h.services.Notice.List(c.Context(), dto.NoticeFilter{
		Query: c.Query("q"), NoticeType: c.Query("type"), Page: page, PageSize: 20,
	})
	if err != nil {
		return c.Status(500).SendString("Error")
	}
	return h.render(c, fiber.StatusOK, web.NoticeListPage{Data: data})
}

func (h *Handler) NoticeCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.NoticeFormPage{Title: "Create Notice"})
}

func (h *Handler) parseNoticeRequest(c fiber.Ctx) dto.NoticeRequest {
	req := dto.NoticeRequest{
		Title: c.FormValue("title"), Body: c.FormValue("body"),
		NoticeType: c.FormValue("notice_type"), TargetAudience: c.FormValue("target_audience"),
		IsPublished: c.FormValue("is_published") == "on" || c.FormValue("is_published") == "true",
	}
	if t, err := parseDate(c.FormValue("publish_at")); err == nil {
		req.PublishAt = t
	}
	if exp := c.FormValue("expires_at"); exp != "" {
		if t, err := parseDate(exp); err == nil {
			req.ExpiresAt = &t
		}
	}
	return req
}

func (h *Handler) NoticeCreate(c fiber.Ctx) error {
	user := h.parentUser(c)
	req := h.parseNoticeRequest(c)
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/notices/new")
	}
	if _, err := h.services.Notice.Create(c.Context(), req, user.ID, c.IP()); err != nil {
		h.flash(c, "Create failed", true)
		return c.Redirect().To("/notices/new")
	}
	h.flash(c, "Notice published", false)
	return c.Redirect().To("/notices")
}

func (h *Handler) NoticeEditPage(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	notice, err := h.services.Notice.Get(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.NoticeFormPage{Title: "Edit Notice", Notice: notice})
}

func (h *Handler) NoticeUpdate(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	user := h.parentUser(c)
	req := h.parseNoticeRequest(c)
	if _, err := h.services.Notice.Update(c.Context(), id, req, user.ID, c.IP()); err != nil {
		h.flash(c, "Update failed", true)
		return c.Redirect().To("/notices/" + id.String() + "/edit")
	}
	h.flash(c, "Notice updated", false)
	return c.Redirect().To("/notices")
}

func (h *Handler) NoticeDelete(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	user := h.parentUser(c)
	_ = h.services.Notice.Delete(c.Context(), id, user.ID, c.IP())
	h.flash(c, "Notice deleted", false)
	return c.Redirect().To("/notices")
}

func (h *Handler) CommunicationDashboard(c fiber.Ctx) error {
	stats, err := h.services.Notification.CommunicationDashboard(c.Context())
	if err != nil {
		return c.Status(500).SendString("Error")
	}
	return h.render(c, fiber.StatusOK, web.CommunicationDashboardPage{Stats: stats})
}

func (h *Handler) ExportSMSLogsCSV(c fiber.Ctx) error {
	to := time.Now()
	from := to.AddDate(0, -1, 0)
	if f, err := parseDate(c.Query("from")); err == nil {
		from = f
	}
	if t, err := parseDate(c.Query("to")); err == nil {
		to = t.Add(24 * time.Hour)
	}
	recs, err := h.services.Notification.ExportSMSLogs(c.Context(), from, to)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	data, _ := export.SMSLogsCSV(recs)
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename="+export.FormatSMSExportFilename(from, to))
	return c.Send(data)
}
