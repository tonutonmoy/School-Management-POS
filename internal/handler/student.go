package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/middleware"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/validator"
	"github.com/school-management/pos/internal/web"
)

func (h *Handler) registerStudentRoutes(auth fiber.Router, mw *middleware.Middleware) {
	students := auth.Group("/students")
	students.Get("/", mw.RequirePermission(model.PermStudentView), h.StudentList)
	students.Get("/new", mw.RequirePermission(model.PermStudentCreate), h.StudentCreatePage)
	students.Post("/", mw.CSRFProtect(), mw.RequirePermission(model.PermStudentCreate), h.StudentCreate)
	students.Get("/search", mw.RequirePermission(model.PermStudentView), h.StudentSearch)
	students.Get("/:id", mw.RequirePermission(model.PermStudentView), h.StudentProfile)
	students.Get("/:id/edit", mw.RequirePermission(model.PermStudentUpdate), h.StudentEditPage)
	students.Post("/:id", mw.CSRFProtect(), mw.RequirePermission(model.PermStudentUpdate), h.StudentUpdate)
	students.Post("/:id/delete", mw.CSRFProtect(), mw.RequirePermission(model.PermStudentDelete), h.StudentDelete)
	students.Get("/:id/promote", mw.RequirePermission(model.PermStudentUpdate), h.StudentPromotePage)
	students.Post("/:id/promote", mw.CSRFProtect(), mw.RequirePermission(model.PermStudentUpdate), h.StudentPromote)
	students.Get("/:id/transfer", mw.RequirePermission(model.PermStudentUpdate), h.StudentTransferPage)
	students.Post("/:id/transfer", mw.CSRFProtect(), mw.RequirePermission(model.PermStudentUpdate), h.StudentTransfer)
	students.Post("/:id/documents", mw.CSRFProtect(), mw.RequirePermission(model.PermStudentUpdate), h.StudentUploadDocument)
	students.Get("/:id/id-card", mw.RequirePermission(model.PermStudentView), h.StudentIDCard)

	auth.Get("/api/sections", h.SectionsHTMX)

	reports := auth.Group("/reports", mw.RequirePermission(model.PermStudentView))
	reports.Get("/students", h.ReportStudentList)
	reports.Get("/students/by-class", h.ReportClassWise)
	reports.Get("/admissions", h.ReportAdmissions)
}

func (h *Handler) studentFormData(c fiber.Ctx) (*web.StudentFormData, error) {
	sessions, _ := h.services.Sessions.List(c.Context())
	classes, _ := h.services.Academic.ListClasses(c.Context())
	departments, _ := h.services.Academic.ListDepartments(c.Context())
	return &web.StudentFormData{Sessions: sessions, Classes: classes, Departments: departments}, nil
}

func (h *Handler) parseStudentForm(c fiber.Ctx) (dto.StudentAdmissionRequest, error) {
	req := dto.StudentAdmissionRequest{
		RollNumber: c.FormValue("roll_number"), FirstName: c.FormValue("first_name"), LastName: c.FormValue("last_name"),
		Gender: c.FormValue("gender"), BloodGroup: c.FormValue("blood_group"), Religion: c.FormValue("religion"),
		Nationality: c.FormValue("nationality"), Phone: c.FormValue("phone"), Email: c.FormValue("email"),
		Address: c.FormValue("address"), Status: c.FormValue("status"),
		FatherName: c.FormValue("father_name"), FatherPhone: c.FormValue("father_phone"), FatherOccupation: c.FormValue("father_occupation"),
		MotherName: c.FormValue("mother_name"), MotherPhone: c.FormValue("mother_phone"), MotherOccupation: c.FormValue("mother_occupation"),
		GuardianName: c.FormValue("guardian_name"), GuardianPhone: c.FormValue("guardian_phone"),
	}
	if dob, err := parseDate(c.FormValue("date_of_birth")); err == nil {
		req.DateOfBirth = dob
	}
	if ad, err := parseDate(c.FormValue("admission_date")); err == nil {
		req.AdmissionDate = ad
	} else {
		req.AdmissionDate = time.Now()
	}
	req.SessionID, _ = uuid.Parse(c.FormValue("session_id"))
	req.ClassID, _ = uuid.Parse(c.FormValue("class_id"))
	req.SectionID, _ = uuid.Parse(c.FormValue("section_id"))
	if d := c.FormValue("department_id"); d != "" {
		req.DepartmentID, _ = uuid.Parse(d)
	}
	return req, nil
}

func (h *Handler) uploadPhoto(c fiber.Ctx) string {
	file, err := c.FormFile("photo")
	if err != nil || file == nil {
		return ""
	}
	f, err := file.Open()
	if err != nil {
		return ""
	}
	defer f.Close()
	url, _ := h.storage.Upload(c.Context(), file.Filename, f, file.Header.Get("Content-Type"))
	return url
}

func (h *Handler) StudentList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	filter := dto.StudentSearchFilter{
		AdmissionNumber: c.Query("admission_number"),
		RollNumber:      c.Query("roll_number"),
		Name:            c.Query("name"),
		Page:            page,
		PageSize:        20,
	}
	if v := c.Query("class_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.ClassID = &id
		}
	}
	if v := c.Query("section_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.SectionID = &id
		}
	}
	if v := c.Query("session_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.SessionID = &id
		}
	}
	data, _ := h.services.Students.Search(c.Context(), filter)
	formData, _ := h.studentFormData(c)
	return h.render(c, fiber.StatusOK, web.StudentListPage{Data: data, Filter: filter, FormData: formData})
}

func (h *Handler) StudentSearch(c fiber.Ctx) error {
	return h.StudentList(c)
}

func (h *Handler) StudentCreatePage(c fiber.Ctx) error {
	formData, _ := h.studentFormData(c)
	return h.render(c, fiber.StatusOK, web.StudentFormPage{Title: "New Admission", FormData: formData})
}

func (h *Handler) StudentCreate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	req, _ := h.parseStudentForm(c)
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/students/new")
	}
	photo := h.uploadPhoto(c)
	student, err := h.services.Students.Admit(c.Context(), req, photo, actor.ID, c.IP())
	if err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/students/new")
	}
	h.uploadStudentDocuments(c, student.ID, actor.ID)
	return c.Redirect().To("/students/" + student.ID.String())
}

func (h *Handler) uploadStudentDocuments(c fiber.Ctx, studentID uuid.UUID, actorID uuid.UUID) {
	docs := map[string]string{
		model.DocBirthCertificate:  "birth_certificate",
		model.DocPreviousMarksheet:   "previous_marksheet",
		model.DocPassportPhoto:     "passport_photo",
		model.DocOther:             "other_document",
	}
	for docType, field := range docs {
		file, err := c.FormFile(field)
		if err != nil || file == nil {
			continue
		}
		f, err := file.Open()
		if err != nil {
			continue
		}
		url, err := h.storage.Upload(c.Context(), file.Filename, f, file.Header.Get("Content-Type"))
		f.Close()
		if err == nil {
			_ = h.services.Students.AddDocument(c.Context(), studentID, docType, file.Filename, url, actorID, c.IP())
		}
	}
}

func (h *Handler) StudentProfile(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	student, err := h.services.Students.GetFull(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Student not found")
	}
	return h.render(c, fiber.StatusOK, web.StudentProfilePage{Student: student})
}

func (h *Handler) StudentEditPage(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	student, err := h.services.Students.GetFull(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Student not found")
	}
	formData, _ := h.studentFormData(c)
	if student.ClassID != uuid.Nil {
		sections, _ := h.services.Academic.ListSectionsByClass(c.Context(), student.ClassID)
		formData.Sections = sections
	}
	return h.render(c, fiber.StatusOK, web.StudentFormPage{Title: "Edit Student", Student: student, FormData: formData})
}

func (h *Handler) StudentUpdate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	req, _ := h.parseStudentForm(c)
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/students/" + id.String() + "/edit")
	}
	photo := h.uploadPhoto(c)
	if _, err := h.services.Students.Update(c.Context(), id, req, photo, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/students/" + id.String() + "/edit")
	}
	h.uploadStudentDocuments(c, id, actor.ID)
	return c.Redirect().To("/students/" + id.String())
}

func (h *Handler) StudentDelete(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	if err := h.services.Students.Delete(c.Context(), id, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/students")
}

func (h *Handler) StudentPromotePage(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	student, err := h.services.Students.GetFull(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Student not found")
	}
	formData, _ := h.studentFormData(c)
	return h.render(c, fiber.StatusOK, web.StudentPromotePage{Student: student, FormData: formData, Action: "promote"})
}

func (h *Handler) StudentPromote(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	req := dto.PromoteStudentRequest{Notes: c.FormValue("notes")}
	req.ToSessionID, _ = uuid.Parse(c.FormValue("to_session_id"))
	req.ToClassID, _ = uuid.Parse(c.FormValue("to_class_id"))
	req.ToSectionID, _ = uuid.Parse(c.FormValue("to_section_id"))
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/students/" + id.String() + "/promote")
	}
	if _, err := h.services.Students.Promote(c.Context(), id, req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/students/" + id.String() + "/promote")
	}
	return c.Redirect().To("/students/" + id.String())
}

func (h *Handler) StudentTransferPage(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	student, err := h.services.Students.GetFull(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Student not found")
	}
	formData, _ := h.studentFormData(c)
	return h.render(c, fiber.StatusOK, web.StudentPromotePage{Student: student, FormData: formData, Action: "transfer"})
}

func (h *Handler) StudentTransfer(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	req := dto.TransferStudentRequest{Notes: c.FormValue("notes")}
	req.ToSessionID, _ = uuid.Parse(c.FormValue("to_session_id"))
	req.ToClassID, _ = uuid.Parse(c.FormValue("to_class_id"))
	req.ToSectionID, _ = uuid.Parse(c.FormValue("to_section_id"))
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/students/" + id.String() + "/transfer")
	}
	if _, err := h.services.Students.Transfer(c.Context(), id, req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/students/" + id.String() + "/transfer")
	}
	return c.Redirect().To("/students/" + id.String())
}

func (h *Handler) StudentUploadDocument(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	docType := c.FormValue("doc_type")
	if docType == "" {
		docType = model.DocOther
	}
	file, err := c.FormFile("document")
	if err != nil || file == nil {
		h.flash(c, "No file uploaded", true)
		return c.Redirect().To("/students/" + id.String())
	}
	f, err := file.Open()
	if err != nil {
		h.flash(c, "Upload failed", true)
		return c.Redirect().To("/students/" + id.String())
	}
	defer f.Close()
	url, err := h.storage.Upload(c.Context(), file.Filename, f, file.Header.Get("Content-Type"))
	if err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/students/" + id.String())
	}
	_ = h.services.Students.AddDocument(c.Context(), id, docType, file.Filename, url, actor.ID, c.IP())
	return c.Redirect().To("/students/" + id.String())
}

func (h *Handler) StudentIDCard(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	data, err := h.services.Students.IDCardData(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Student not found")
	}
	return h.render(c, fiber.StatusOK, web.StudentIDCardPage{Data: data})
}

func (h *Handler) SectionsHTMX(c fiber.Ctx) error {
	classIDStr := c.Query("class_id")
	if classIDStr == "" {
		classIDStr = c.FormValue("class_id")
	}
	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		return c.SendString(`<select name="section_id" required class="w-full rounded-lg border px-3 py-2"><option value="">Select class first</option></select>`)
	}
	html, err := h.services.Students.SectionsHTMX(c.Context(), classID, c.Query("field", "section_id"))
	if err != nil {
		return c.Status(500).SendString("Error loading sections")
	}
	return c.SendString(html)
}

func (h *Handler) ReportStudentList(c fiber.Ctx) error {
	filter := parseReportFilter(c)
	students, _ := h.services.Students.StudentListReport(c.Context(), filter)
	return h.render(c, fiber.StatusOK, web.StudentReportPage{Title: "Student List Report", Students: students, Filter: filter})
}

func (h *Handler) ReportClassWise(c fiber.Ctx) error {
	filter := parseReportFilter(c)
	students, _ := h.services.Students.StudentListReport(c.Context(), filter)
	byClass := groupStudentsByClass(students)
	return h.render(c, fiber.StatusOK, web.ClassWiseReportPage{Groups: byClass, Filter: filter})
}

func (h *Handler) ReportAdmissions(c fiber.Ctx) error {
	from, to := time.Now().AddDate(0, -1, 0), time.Now()
	if v := c.Query("from"); v != "" {
		if t, err := parseDate(v); err == nil {
			from = t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := parseDate(v); err == nil {
			to = t
		}
	}
	students, _ := h.services.Students.AdmissionReport(c.Context(), from, to)
	return h.render(c, fiber.StatusOK, web.AdmissionReportPage{Students: students, From: from, To: to})
}

func parseReportFilter(c fiber.Ctx) dto.ReportFilter {
	filter := dto.ReportFilter{Status: c.Query("status")}
	if v := c.Query("class_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.ClassID = &id
		}
	}
	if v := c.Query("session_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.SessionID = &id
		}
	}
	return filter
}

func groupStudentsByClass(students []dto.StudentResponse) []web.ClassStudentGroup {
	groups := map[string]*web.ClassStudentGroup{}
	var order []string
	for _, s := range students {
		key := s.ClassName
		if _, ok := groups[key]; !ok {
			groups[key] = &web.ClassStudentGroup{ClassName: key}
			order = append(order, key)
		}
		groups[key].Students = append(groups[key].Students, s)
	}
	var result []web.ClassStudentGroup
	for _, k := range order {
		result = append(result, *groups[k])
	}
	return result
}
