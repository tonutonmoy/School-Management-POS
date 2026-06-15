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

func (h *Handler) registerHRRoutes(auth fiber.Router, mw *middleware.Middleware) {
	// Departments
	auth.Get("/departments", h.DepartmentList)
	auth.Get("/departments/new", mw.RequirePermission(model.PermDepartmentCreate), h.DepartmentCreatePage)
	auth.Post("/departments", mw.CSRFProtect(), mw.RequirePermission(model.PermDepartmentCreate), h.DepartmentCreate)
	auth.Get("/departments/:id/edit", mw.RequirePermission(model.PermDepartmentUpdate), h.DepartmentEditPage)
	auth.Post("/departments/:id", mw.CSRFProtect(), mw.RequirePermission(model.PermDepartmentUpdate), h.DepartmentUpdate)
	auth.Post("/departments/:id/delete", mw.CSRFProtect(), mw.RequirePermission(model.PermDepartmentDelete), h.DepartmentDelete)

	// Designations
	auth.Get("/designations", h.DesignationList)
	auth.Get("/designations/new", mw.RequirePermission(model.PermDesignationCreate), h.DesignationCreatePage)
	auth.Post("/designations", mw.CSRFProtect(), mw.RequirePermission(model.PermDesignationCreate), h.DesignationCreate)
	auth.Get("/designations/:id/edit", mw.RequirePermission(model.PermDesignationUpdate), h.DesignationEditPage)
	auth.Post("/designations/:id", mw.CSRFProtect(), mw.RequirePermission(model.PermDesignationUpdate), h.DesignationUpdate)
	auth.Post("/designations/:id/delete", mw.CSRFProtect(), mw.RequirePermission(model.PermDesignationDelete), h.DesignationDelete)

	// Teachers
	auth.Get("/teachers", mw.RequirePermission(model.PermTeacherView), h.TeacherList)
	auth.Get("/teachers/new", mw.RequirePermission(model.PermTeacherCreate), h.TeacherCreatePage)
	auth.Post("/teachers", mw.CSRFProtect(), mw.RequirePermission(model.PermTeacherCreate), h.TeacherCreate)
	auth.Get("/teachers/:id", mw.RequirePermission(model.PermTeacherView), h.TeacherProfile)
	auth.Get("/teachers/:id/edit", mw.RequirePermission(model.PermTeacherUpdate), h.TeacherEditPage)
	auth.Post("/teachers/:id", mw.CSRFProtect(), mw.RequirePermission(model.PermTeacherUpdate), h.TeacherUpdate)
	auth.Post("/teachers/:id/delete", mw.CSRFProtect(), mw.RequirePermission(model.PermTeacherDelete), h.TeacherDelete)
	auth.Get("/teachers/:id/assign", mw.RequirePermission(model.PermTeacherUpdate), h.TeacherAssignPage)
	auth.Post("/teachers/:id/assign", mw.CSRFProtect(), mw.RequirePermission(model.PermTeacherUpdate), h.TeacherAssign)
	auth.Post("/teachers/:id/documents", mw.CSRFProtect(), mw.RequirePermission(model.PermTeacherUpdate), h.TeacherUploadDocument)

	// Staff
	auth.Get("/staff", mw.RequirePermission(model.PermStaffView), h.StaffList)
	auth.Get("/staff/new", mw.RequirePermission(model.PermStaffCreate), h.StaffCreatePage)
	auth.Post("/staff", mw.CSRFProtect(), mw.RequirePermission(model.PermStaffCreate), h.StaffCreate)
	auth.Get("/staff/:id", mw.RequirePermission(model.PermStaffView), h.StaffProfile)
	auth.Get("/staff/:id/edit", mw.RequirePermission(model.PermStaffUpdate), h.StaffEditPage)
	auth.Post("/staff/:id", mw.CSRFProtect(), mw.RequirePermission(model.PermStaffUpdate), h.StaffUpdate)
	auth.Post("/staff/:id/delete", mw.CSRFProtect(), mw.RequirePermission(model.PermStaffDelete), h.StaffDelete)
	auth.Post("/staff/:id/documents", mw.CSRFProtect(), mw.RequirePermission(model.PermStaffUpdate), h.StaffUploadDocument)

	// Teacher portal
	auth.Get("/teacher/dashboard", h.TeacherPortalDashboard)

	// HR Reports & exports
	reports := auth.Group("/reports/hr", mw.RequirePermission(model.PermTeacherView))
	reports.Get("/teachers", h.ReportTeacherList)
	reports.Get("/staff", mw.RequirePermission(model.PermStaffView), h.ReportStaffList)
	reports.Get("/department", h.ReportDepartmentWise)
	reports.Get("/assignments", h.ReportTeacherAssignments)
	reports.Get("/teachers/export.csv", h.ExportTeachersCSV)
	reports.Get("/teachers/export.xlsx", h.ExportTeachersExcel)
	reports.Get("/staff/export.csv", h.ExportStaffCSV)
	reports.Get("/staff/export.xlsx", h.ExportStaffExcel)
}

func (h *Handler) hrFormData(c fiber.Ctx) *web.HRFormData {
	depts, _ := h.services.HR.ListDepartments(c.Context(), "")
	desigs, _ := h.services.HR.ListDesignations(c.Context())
	classes, _ := h.services.Academic.ListClasses(c.Context())
	subjects, _ := h.services.Academic.ListSubjects(c.Context())
	sections, _ := h.services.Academic.ListSections(c.Context())
	return &web.HRFormData{Departments: depts, Designations: desigs, Classes: classes, Subjects: subjects, Sections: sections}
}

func (h *Handler) parseTeacherForm(c fiber.Ctx) dto.TeacherRequest {
	req := dto.TeacherRequest{
		FirstName: c.FormValue("first_name"), LastName: c.FormValue("last_name"), Gender: c.FormValue("gender"),
		BloodGroup: c.FormValue("blood_group"), Religion: c.FormValue("religion"), Nationality: c.FormValue("nationality"),
		Phone: c.FormValue("phone"), Email: c.FormValue("email"), Address: c.FormValue("address"),
		NationalID: c.FormValue("national_id"), Qualification: c.FormValue("qualification"),
		Experience: c.FormValue("experience"), EmploymentType: c.FormValue("employment_type"), Status: c.FormValue("status"),
	}
	if dob, err := parseDate(c.FormValue("date_of_birth")); err == nil {
		req.DateOfBirth = dob
	}
	if jd, err := parseDate(c.FormValue("joining_date")); err == nil {
		req.JoiningDate = jd
	} else {
		req.JoiningDate = time.Now()
	}
	if s, err := strconv.ParseFloat(c.FormValue("salary"), 64); err == nil {
		req.Salary = s
	}
	if d := c.FormValue("department_id"); d != "" {
		req.DepartmentID, _ = uuid.Parse(d)
	}
	if d := c.FormValue("designation_id"); d != "" {
		req.DesignationID, _ = uuid.Parse(d)
	}
	return req
}

func (h *Handler) parseStaffForm(c fiber.Ctx) dto.StaffRequest {
	req := dto.StaffRequest{
		FirstName: c.FormValue("first_name"), LastName: c.FormValue("last_name"),
		Phone: c.FormValue("phone"), Email: c.FormValue("email"), Address: c.FormValue("address"),
		Status: c.FormValue("status"),
	}
	if jd, err := parseDate(c.FormValue("joining_date")); err == nil {
		req.JoiningDate = jd
	} else {
		req.JoiningDate = time.Now()
	}
	if s, err := strconv.ParseFloat(c.FormValue("salary"), 64); err == nil {
		req.Salary = s
	}
	if d := c.FormValue("department_id"); d != "" {
		req.DepartmentID, _ = uuid.Parse(d)
	}
	if d := c.FormValue("designation_id"); d != "" {
		req.DesignationID, _ = uuid.Parse(d)
	}
	return req
}

func (h *Handler) uploadEmployeePhoto(c fiber.Ctx) string {
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

func (h *Handler) uploadEmployeeDoc(c fiber.Ctx, field string) (docType, fileName, fileURL string) {
	docType = c.FormValue("doc_type")
	if docType == "" {
		docType = field
	}
	file, err := c.FormFile(field)
	if err != nil || file == nil {
		return "", "", ""
	}
	f, err := file.Open()
	if err != nil {
		return "", "", ""
	}
	defer f.Close()
	url, err := h.storage.Upload(c.Context(), file.Filename, f, file.Header.Get("Content-Type"))
	if err != nil {
		return "", "", ""
	}
	return docType, file.Filename, url
}

// Departments
func (h *Handler) DepartmentList(c fiber.Ctx) error {
	depts, _ := h.services.HR.ListDepartments(c.Context(), c.Query("type"))
	return h.render(c, fiber.StatusOK, web.DepartmentListPage{Departments: depts})
}

func (h *Handler) DepartmentCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.DepartmentFormPage{Title: "Create Department"})
}

func (h *Handler) DepartmentCreate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	req := dto.DepartmentRequest{Name: c.FormValue("name"), Slug: c.FormValue("slug"), Description: c.FormValue("description"), DeptType: c.FormValue("dept_type")}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/departments/new")
	}
	if _, err := h.services.HR.CreateDepartment(c.Context(), req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/departments/new")
	}
	return c.Redirect().To("/departments")
}

func (h *Handler) DepartmentEditPage(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	dept, err := h.services.HR.GetDepartment(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.DepartmentFormPage{Title: "Edit Department", Department: dept})
}

func (h *Handler) DepartmentUpdate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	req := dto.DepartmentRequest{Name: c.FormValue("name"), Slug: c.FormValue("slug"), Description: c.FormValue("description"), DeptType: c.FormValue("dept_type")}
	if _, err := h.services.HR.UpdateDepartment(c.Context(), id, req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/departments")
}

func (h *Handler) DepartmentDelete(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	_ = h.services.HR.DeleteDepartment(c.Context(), id, actor.ID, c.IP())
	return c.Redirect().To("/departments")
}

// Designations
func (h *Handler) DesignationList(c fiber.Ctx) error {
	items, _ := h.services.HR.ListDesignations(c.Context())
	return h.render(c, fiber.StatusOK, web.DesignationListPage{Designations: items})
}

func (h *Handler) DesignationCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.DesignationFormPage{Title: "Create Designation"})
}

func (h *Handler) DesignationCreate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	req := dto.DesignationRequest{Name: c.FormValue("name"), Slug: c.FormValue("slug"), Category: c.FormValue("category"), Description: c.FormValue("description")}
	if _, err := h.services.HR.CreateDesignation(c.Context(), req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/designations/new")
	}
	return c.Redirect().To("/designations")
}

func (h *Handler) DesignationEditPage(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	items, _ := h.services.HR.ListDesignations(c.Context())
	var current *dto.DesignationResponse
	for i := range items {
		if items[i].ID == id {
			current = &items[i]
			break
		}
	}
	if current == nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.DesignationFormPage{Title: "Edit Designation", Designation: current})
}

func (h *Handler) DesignationUpdate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	req := dto.DesignationRequest{Name: c.FormValue("name"), Slug: c.FormValue("slug"), Category: c.FormValue("category"), Description: c.FormValue("description")}
	if _, err := h.services.HR.UpdateDesignation(c.Context(), id, req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/designations")
}

func (h *Handler) DesignationDelete(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	_ = h.services.HR.DeleteDesignation(c.Context(), id, actor.ID, c.IP())
	return c.Redirect().To("/designations")
}

// Teachers
func (h *Handler) TeacherList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	filter := dto.TeacherSearchFilter{Query: c.Query("q"), Status: c.Query("status"), Page: page, PageSize: 20}
	if v := c.Query("department_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.DepartmentID = &id
		}
	}
	data, _ := h.services.HR.SearchTeachers(c.Context(), filter)
	return h.render(c, fiber.StatusOK, web.TeacherListPage{Data: data, Filter: filter, FormData: h.hrFormData(c)})
}

func (h *Handler) TeacherCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.TeacherFormPage{Title: "Add Teacher", FormData: h.hrFormData(c)})
}

func (h *Handler) TeacherCreate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	req := h.parseTeacherForm(c)
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/teachers/new")
	}
	photo := h.uploadEmployeePhoto(c)
	if _, err := h.services.HR.CreateTeacher(c.Context(), req, photo, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/teachers/new")
	}
	return c.Redirect().To("/teachers")
}

func (h *Handler) TeacherProfile(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	teacher, err := h.services.HR.GetTeacher(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.TeacherProfilePage{Teacher: teacher})
}

func (h *Handler) TeacherEditPage(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	teacher, err := h.services.HR.GetTeacher(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.TeacherFormPage{Title: "Edit Teacher", Teacher: teacher, FormData: h.hrFormData(c)})
}

func (h *Handler) TeacherUpdate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	req := h.parseTeacherForm(c)
	photo := h.uploadEmployeePhoto(c)
	if _, err := h.services.HR.UpdateTeacher(c.Context(), id, req, photo, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/teachers/" + id.String())
}

func (h *Handler) TeacherDelete(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	_ = h.services.HR.DeleteTeacher(c.Context(), id, actor.ID, c.IP())
	return c.Redirect().To("/teachers")
}

func (h *Handler) TeacherAssignPage(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	teacher, err := h.services.HR.GetTeacher(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.TeacherAssignPage{Teacher: teacher, FormData: h.hrFormData(c)})
}

func (h *Handler) TeacherAssign(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	req := dto.TeacherAssignmentRequest{
		SubjectIDs: parseUUIDListByField(c, "subject_ids"),
		ClassIDs:   parseUUIDListByField(c, "class_ids"),
		SectionIDs: parseUUIDListByField(c, "section_ids"),
	}
	if err := h.services.HR.AssignTeacher(c.Context(), id, req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/teachers/" + id.String())
}

func (h *Handler) TeacherUploadDocument(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	docType, fileName, url := h.uploadEmployeeDoc(c, "document")
	if url != "" {
		_ = h.services.HR.AddTeacherDocument(c.Context(), id, docType, fileName, url, actor.ID, c.IP())
	}
	return c.Redirect().To("/teachers/" + id.String())
}

// Staff
func (h *Handler) StaffList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	filter := dto.StaffSearchFilter{Query: c.Query("q"), Status: c.Query("status"), Page: page, PageSize: 20}
	if v := c.Query("department_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.DepartmentID = &id
		}
	}
	data, _ := h.services.HR.SearchStaff(c.Context(), filter)
	return h.render(c, fiber.StatusOK, web.StaffListPage{Data: data, Filter: filter, FormData: h.hrFormData(c)})
}

func (h *Handler) StaffCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.StaffFormPage{Title: "Add Staff", FormData: h.hrFormData(c)})
}

func (h *Handler) StaffCreate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	req := h.parseStaffForm(c)
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/staff/new")
	}
	photo := h.uploadEmployeePhoto(c)
	if _, err := h.services.HR.CreateStaff(c.Context(), req, photo, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/staff/new")
	}
	return c.Redirect().To("/staff")
}

func (h *Handler) StaffProfile(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	staff, err := h.services.HR.GetStaff(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.StaffProfilePage{Staff: staff})
}

func (h *Handler) StaffEditPage(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	staff, err := h.services.HR.GetStaff(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.StaffFormPage{Title: "Edit Staff", Staff: staff, FormData: h.hrFormData(c)})
}

func (h *Handler) StaffUpdate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	req := h.parseStaffForm(c)
	photo := h.uploadEmployeePhoto(c)
	if _, err := h.services.HR.UpdateStaff(c.Context(), id, req, photo, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/staff/" + id.String())
}

func (h *Handler) StaffDelete(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	_ = h.services.HR.DeleteStaff(c.Context(), id, actor.ID, c.IP())
	return c.Redirect().To("/staff")
}

func (h *Handler) StaffUploadDocument(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	docType, fileName, url := h.uploadEmployeeDoc(c, "document")
	if url != "" {
		_ = h.services.HR.AddStaffDocument(c.Context(), id, docType, fileName, url, actor.ID, c.IP())
	}
	return c.Redirect().To("/staff/" + id.String())
}

func (h *Handler) TeacherPortalDashboard(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	dash, err := h.services.HR.TeacherPortal(c.Context(), user.ID, user.Email)
	if err != nil {
		return c.Redirect().To("/dashboard")
	}
	return h.render(c, fiber.StatusOK, web.TeacherPortalPage{Dashboard: dash})
}

// Reports
func (h *Handler) ReportTeacherList(c fiber.Ctx) error {
	filter := parseHRReportFilter(c)
	teachers, _ := h.services.HR.ListTeachersReport(c.Context(), filter)
	return h.render(c, fiber.StatusOK, web.HRTeacherReportPage{Teachers: teachers, Filter: filter})
}

func (h *Handler) ReportStaffList(c fiber.Ctx) error {
	filter := parseHRReportFilter(c)
	staff, _ := h.services.HR.ListStaffReport(c.Context(), filter)
	return h.render(c, fiber.StatusOK, web.HRStaffReportPage{Staff: staff, Filter: filter})
}

func (h *Handler) ReportDepartmentWise(c fiber.Ctx) error {
	filter := parseHRReportFilter(c)
	depts, _ := h.services.HR.ListDepartments(c.Context(), "")
	var groups []web.DepartmentReportGroup
	for _, d := range depts {
		if filter.DepartmentID != nil && d.ID != *filter.DepartmentID {
			continue
		}
		f := dto.HRReportFilter{DepartmentID: &d.ID, Status: filter.Status}
		teachers, _ := h.services.HR.ListTeachersReport(c.Context(), f)
		staff, _ := h.services.HR.ListStaffReport(c.Context(), f)
		groups = append(groups, web.DepartmentReportGroup{Department: d, Teachers: teachers, Staff: staff})
	}
	return h.render(c, fiber.StatusOK, web.DepartmentReportPage{Groups: groups})
}

func (h *Handler) ReportTeacherAssignments(c fiber.Ctx) error {
	filter := parseHRReportFilter(c)
	teachers, _ := h.services.HR.ListTeachersReport(c.Context(), filter)
	return h.render(c, fiber.StatusOK, web.AssignmentReportPage{Teachers: teachers})
}

func parseHRReportFilter(c fiber.Ctx) dto.HRReportFilter {
	filter := dto.HRReportFilter{Status: c.Query("status")}
	if v := c.Query("department_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.DepartmentID = &id
		}
	}
	if v := c.Query("designation_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.DesignationID = &id
		}
	}
	return filter
}

func parseUUIDListByField(c fiber.Ctx, field string) []uuid.UUID {
	var ids []uuid.UUID
	c.Request().PostArgs().VisitAll(func(key, value []byte) {
		if string(key) == field {
			if id, err := uuid.Parse(string(value)); err == nil {
				ids = append(ids, id)
			}
		}
	})
	return ids
}

func (h *Handler) ExportTeachersCSV(c fiber.Ctx) error {
	teachers, _ := h.services.HR.ListTeachersReport(c.Context(), parseHRReportFilter(c))
	data, err := export.TeachersCSV(teachers)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", `attachment; filename="teachers.csv"`)
	return c.Send(data)
}

func (h *Handler) ExportTeachersExcel(c fiber.Ctx) error {
	teachers, _ := h.services.HR.ListTeachersReport(c.Context(), parseHRReportFilter(c))
	data, err := export.TeachersExcel(teachers)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="teachers.xlsx"`)
	return c.Send(data)
}

func (h *Handler) ExportStaffCSV(c fiber.Ctx) error {
	staff, _ := h.services.HR.ListStaffReport(c.Context(), parseHRReportFilter(c))
	data, err := export.StaffCSV(staff)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", `attachment; filename="staff.csv"`)
	return c.Send(data)
}

func (h *Handler) ExportStaffExcel(c fiber.Ctx) error {
	staff, _ := h.services.HR.ListStaffReport(c.Context(), parseHRReportFilter(c))
	data, err := export.StaffExcel(staff)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="staff.xlsx"`)
	return c.Send(data)
}
