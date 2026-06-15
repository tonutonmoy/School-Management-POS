package handler

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/export"
	"github.com/school-management/pos/internal/middleware"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/validator"
	"github.com/school-management/pos/internal/web"
)

func (h *Handler) registerAdmissionRoutes(app, auth fiber.Router, mw *middleware.Middleware) {
	pub := app.Group("/admission")
	pub.Get("/apply", h.PublicAdmissionForm)
	pub.Post("/apply", mw.CSRFProtect(), h.PublicAdmissionSubmit)
	pub.Get("/track", h.PublicAdmissionTrackPage)
	pub.Post("/track", mw.CSRFProtect(), h.PublicAdmissionTrack)
	pub.Get("/success", h.PublicAdmissionSuccess)

	admin := auth.Group("/admissions", mw.RequirePermission(model.PermAdmissionReview))
	admin.Get("/dashboard", h.AdmissionDashboard)
	admin.Get("/", h.AdmissionList)
	admin.Get("/export.csv", h.AdmissionExportCSV)
	admin.Get("/:id", h.AdmissionDetail)
	admin.Post("/:id/review", mw.CSRFProtect(), h.AdmissionUnderReview)
	admin.Post("/:id/approve", mw.CSRFProtect(), mw.RequirePermission(model.PermAdmissionApprove), h.AdmissionApprove)
	admin.Post("/:id/reject", mw.CSRFProtect(), mw.RequirePermission(model.PermAdmissionReject), h.AdmissionReject)
	admin.Post("/:id/payment", mw.CSRFProtect(), h.AdmissionRecordPayment)
	admin.Post("/:id/admit", mw.CSRFProtect(), mw.RequirePermission(model.PermAdmissionApprove), h.AdmissionConvertStudent)
}

func (h *Handler) admissionFormData(c fiber.Ctx) *web.AdmissionFormData {
	sessions, _ := h.services.Sessions.List(c.Context())
	classes, _ := h.services.Academic.ListClasses(c.Context())
	return &web.AdmissionFormData{Sessions: sessions, Classes: classes}
}

func (h *Handler) PublicAdmissionForm(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.PublicAdmissionFormPage{
		FormData: h.admissionFormData(c), Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
	})
}

func (h *Handler) parseAdmissionForm(c fiber.Ctx) dto.AdmissionApplicationRequest {
	req := dto.AdmissionApplicationRequest{
		FirstName: c.FormValue("first_name"), LastName: c.FormValue("last_name"),
		Gender: c.FormValue("gender"), BloodGroup: c.FormValue("blood_group"),
		Religion: c.FormValue("religion"), Nationality: c.FormValue("nationality"),
		Phone: c.FormValue("phone"), Email: c.FormValue("email"), Address: c.FormValue("address"),
		FatherName: c.FormValue("father_name"), FatherPhone: c.FormValue("father_phone"),
		FatherOccupation: c.FormValue("father_occupation"),
		MotherName: c.FormValue("mother_name"), MotherPhone: c.FormValue("mother_phone"),
		MotherOccupation: c.FormValue("mother_occupation"),
		GuardianName: c.FormValue("guardian_name"), GuardianPhone: c.FormValue("guardian_phone"),
		PreviousSchool: c.FormValue("previous_school"), PreviousClass: c.FormValue("previous_class"),
		PreviousBoard: c.FormValue("previous_board"),
	}
	if dob, err := parseDate(c.FormValue("date_of_birth")); err == nil {
		req.DateOfBirth = dob
	}
	req.SessionID, _ = uuid.Parse(c.FormValue("session_id"))
	req.ClassID, _ = uuid.Parse(c.FormValue("class_id"))
	if s := c.FormValue("section_id"); s != "" {
		req.SectionID, _ = uuid.Parse(s)
	}
	if fee, err := strconv.ParseFloat(c.FormValue("admission_fee"), 64); err == nil {
		req.AdmissionFee = fee
	}
	return req
}

func (h *Handler) PublicAdmissionSubmit(c fiber.Ctx) error {
	req := h.parseAdmissionForm(c)
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/admission/apply")
	}
	app, err := h.services.Admission.Submit(c.Context(), req)
	if err != nil {
		h.flash(c, "Unable to submit application", true)
		return c.Redirect().To("/admission/apply")
	}
	for _, docType := range []string{"birth_certificate", "previous_marksheet", "passport_photo"} {
		file, ferr := c.FormFile(docType)
		if ferr != nil || file == nil {
			continue
		}
		f, _ := file.Open()
		if f == nil {
			continue
		}
		url, _ := h.storage.Upload(c.Context(), file.Filename, f, file.Header.Get("Content-Type"))
		f.Close()
		if url != "" {
			_ = h.services.Admission.AddDocument(c.Context(), app.ID, docType, file.Filename, url)
		}
	}
	c.Cookie(&fiber.Cookie{Name: "last_app_no", Value: app.ApplicationNumber, Path: "/", MaxAge: 300})
	c.Cookie(&fiber.Cookie{Name: "last_app_token", Value: app.TrackingToken, Path: "/", MaxAge: 300})
	return c.Redirect().To("/admission/success")
}

func (h *Handler) PublicAdmissionSuccess(c fiber.Ctx) error {
	appNo := c.Cookies("last_app_no")
	token := c.Cookies("last_app_token")
	var app *dto.AdmissionApplicationResponse
	if appNo != "" && token != "" {
		app, _ = h.services.Admission.Track(c.Context(), appNo, token)
	}
	return h.render(c, fiber.StatusOK, web.PublicAdmissionSuccessPage{Application: app})
}

func (h *Handler) PublicAdmissionTrackPage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.PublicAdmissionTrackPage{Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type")})
}

func (h *Handler) PublicAdmissionTrack(c fiber.Ctx) error {
	req := dto.AdmissionTrackRequest{
		ApplicationNumber: c.FormValue("application_number"),
		TrackingToken:     c.FormValue("tracking_token"),
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/admission/track")
	}
	app, err := h.services.Admission.Track(c.Context(), req.ApplicationNumber, req.TrackingToken)
	if err != nil {
		h.flash(c, "Application not found", true)
		return c.Redirect().To("/admission/track")
	}
	gateways, _ := h.services.Payment.ListGateways(c.Context())
	active := make([]dto.PaymentGatewayResponse, 0)
	for _, g := range gateways {
		if g.IsActive {
			active = append(active, g)
		}
	}
	return h.render(c, fiber.StatusOK, web.PublicAdmissionTrackResultPage{Application: app, Gateways: active})
}

func (h *Handler) AdmissionDashboard(c fiber.Ctx) error {
	stats, _ := h.services.Admission.DashboardStats(c.Context())
	return h.render(c, fiber.StatusOK, web.AdmissionDashboardPage{Stats: stats})
}

func (h *Handler) AdmissionList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	f := dto.AdmissionSearchFilter{
		Query: c.Query("q"), Status: c.Query("status"), PaymentStatus: c.Query("payment"),
		Page: page, PageSize: 20,
	}
	data, err := h.services.Admission.List(c.Context(), f)
	if err != nil {
		return c.Status(500).SendString("Error")
	}
	return h.render(c, fiber.StatusOK, web.AdmissionListPage{Data: data})
}

func (h *Handler) AdmissionDetail(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	app, err := h.services.Admission.Get(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	sections := []dto.SectionResponse{}
	if app.ClassID != nil {
		secs, _ := h.services.Academic.ListSectionsByClass(c.Context(), *app.ClassID)
		sections = secs
	}
	return h.render(c, fiber.StatusOK, web.AdmissionDetailPage{Application: app, Sections: sections})
}

func (h *Handler) AdmissionUnderReview(c fiber.Ctx) error {
	return h.admissionStatusAction(c, model.AdmissionUnderReview, h.services.Admission.UnderReview)
}

func (h *Handler) AdmissionApprove(c fiber.Ctx) error {
	return h.admissionStatusAction(c, model.AdmissionApproved, h.services.Admission.Approve)
}

func (h *Handler) AdmissionReject(c fiber.Ctx) error {
	return h.admissionStatusAction(c, model.AdmissionRejected, h.services.Admission.Reject)
}

func (h *Handler) admissionStatusAction(c fiber.Ctx, _ string, fn func(context.Context, uuid.UUID, string, uuid.UUID, string) error) error {
	id, _ := parseUUIDParam(c, "id")
	user := middleware.GetUser(c)
	notes := c.FormValue("review_notes")
	if err := fn(c.Context(), id, notes, user.ID, c.IP()); err != nil {
		h.flash(c, "Action failed", true)
	} else {
		h.flash(c, "Status updated", false)
	}
	return c.Redirect().To("/admissions/" + id.String())
}

func (h *Handler) AdmissionRecordPayment(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	user := middleware.GetUser(c)
	amount, _ := strconv.ParseFloat(c.FormValue("amount"), 64)
	req := dto.AdmissionPaymentRequest{PaymentReference: c.FormValue("payment_reference"), Amount: amount}
	if _, err := h.services.Admission.RecordPayment(c.Context(), id, req, user.ID, c.IP()); err != nil {
		h.flash(c, "Payment update failed", true)
	} else {
		h.flash(c, "Payment recorded", false)
	}
	return c.Redirect().To("/admissions/" + id.String())
}

func (h *Handler) AdmissionConvertStudent(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	user := middleware.GetUser(c)
	sectionID, _ := uuid.Parse(c.FormValue("section_id"))
	student, err := h.services.Admission.ConvertToStudent(c.Context(), id, sectionID, user.ID, c.IP())
	if err != nil {
		h.flash(c, "Admission failed: "+err.Error(), true)
		return c.Redirect().To("/admissions/" + id.String())
	}
	h.flash(c, "Student admitted: "+student.AdmissionNumber, false)
	return c.Redirect().To("/students/" + student.ID.String())
}

func (h *Handler) AdmissionExportCSV(c fiber.Ctx) error {
	f := dto.AdmissionSearchFilter{Query: c.Query("q"), Status: c.Query("status")}
	recs, err := h.services.Admission.ExportCSV(c.Context(), f)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	data, _ := export.AdmissionApplicationsCSV(recs)
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=admission-applications.csv")
	return c.Send(data)
}
