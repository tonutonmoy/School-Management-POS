package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/middleware"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/validator"
	"github.com/school-management/pos/internal/web"
)

func (h *Handler) registerAcademicRoutes(auth fiber.Router, mw *middleware.Middleware) {
	auth.Get("/classes", h.ClassList)
	auth.Get("/classes/new", mw.RequirePermission(model.PermClassCreate), h.ClassCreatePage)
	auth.Post("/classes", mw.CSRFProtect(), mw.RequirePermission(model.PermClassCreate), h.ClassCreate)
	auth.Get("/classes/:id/edit", mw.RequirePermission(model.PermClassUpdate), h.ClassEditPage)
	auth.Post("/classes/:id", mw.CSRFProtect(), mw.RequirePermission(model.PermClassUpdate), h.ClassUpdate)
	auth.Post("/classes/:id/delete", mw.CSRFProtect(), mw.RequirePermission(model.PermClassDelete), h.ClassDelete)
	auth.Post("/classes/:id/subjects", mw.CSRFProtect(), mw.RequirePermission(model.PermClassUpdate), h.ClassAssignSubjects)

	auth.Get("/sections", h.SectionList)
	auth.Get("/sections/new", h.SectionCreatePage)
	auth.Post("/sections", mw.CSRFProtect(), h.SectionCreate)
	auth.Get("/sections/:id/edit", h.SectionEditPage)
	auth.Post("/sections/:id", mw.CSRFProtect(), h.SectionUpdate)
	auth.Post("/sections/:id/delete", mw.CSRFProtect(), h.SectionDelete)

	auth.Get("/subjects", h.SubjectList)
	auth.Get("/subjects/new", mw.RequirePermission(model.PermSubjectCreate), h.SubjectCreatePage)
	auth.Post("/subjects", mw.CSRFProtect(), mw.RequirePermission(model.PermSubjectCreate), h.SubjectCreate)
	auth.Get("/subjects/:id/edit", mw.RequirePermission(model.PermSubjectUpdate), h.SubjectEditPage)
	auth.Post("/subjects/:id", mw.CSRFProtect(), mw.RequirePermission(model.PermSubjectUpdate), h.SubjectUpdate)
	auth.Post("/subjects/:id/delete", mw.CSRFProtect(), mw.RequirePermission(model.PermSubjectDelete), h.SubjectDelete)
}

func (h *Handler) ClassList(c fiber.Ctx) error {
	classes, _ := h.services.Academic.ListClasses(c.Context())
	return h.render(c, fiber.StatusOK, web.ClassListPage{Classes: classes})
}

func (h *Handler) ClassCreatePage(c fiber.Ctx) error {
	subjects, _ := h.services.Academic.ListSubjects(c.Context())
	return h.render(c, fiber.StatusOK, web.ClassFormPage{Title: "Create Class", Subjects: subjects})
}

func (h *Handler) ClassCreate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	req := dto.ClassRequest{
		Name: c.FormValue("name"), Code: c.FormValue("code"),
		Description: c.FormValue("description"),
	}
	if v := c.FormValue("sort_order"); v != "" {
		if n, err := parseInt(v); err == nil {
			req.SortOrder = n
		}
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/classes/new")
	}
	cls, err := h.services.Academic.CreateClass(c.Context(), req, actor.ID, c.IP())
	if err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/classes/new")
	}
	permIDs := parseUUIDList(c)
	if len(permIDs) > 0 {
		_ = h.services.Academic.AssignSubjectsToClass(c.Context(), cls.ID, permIDs, actor.ID, c.IP())
	}
	return c.Redirect().To("/classes")
}

func (h *Handler) ClassEditPage(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	cls, err := h.services.Academic.GetClass(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Class not found")
	}
	allSubjects, _ := h.services.Academic.ListSubjects(c.Context())
	assigned, _ := h.services.Academic.ListSubjectsByClass(c.Context(), id)
	return h.render(c, fiber.StatusOK, web.ClassFormPage{Title: "Edit Class", Class: cls, Subjects: allSubjects, Assigned: assigned})
}

func (h *Handler) ClassUpdate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	req := dto.ClassRequest{Name: c.FormValue("name"), Code: c.FormValue("code"), Description: c.FormValue("description")}
	if v := c.FormValue("sort_order"); v != "" {
		if n, err := parseInt(v); err == nil {
			req.SortOrder = n
		}
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/classes/" + id.String() + "/edit")
	}
	if _, err := h.services.Academic.UpdateClass(c.Context(), id, req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/classes/" + id.String() + "/edit")
	}
	permIDs := parseUUIDList(c)
	_ = h.services.Academic.AssignSubjectsToClass(c.Context(), id, permIDs, actor.ID, c.IP())
	return c.Redirect().To("/classes")
}

func (h *Handler) ClassDelete(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	if err := h.services.Academic.DeleteClass(c.Context(), id, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/classes")
}

func (h *Handler) ClassAssignSubjects(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	permIDs := parseUUIDList(c)
	if err := h.services.Academic.AssignSubjectsToClass(c.Context(), id, permIDs, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/classes/" + id.String() + "/edit")
}

func (h *Handler) SectionList(c fiber.Ctx) error {
	sections, _ := h.services.Academic.ListSections(c.Context())
	return h.render(c, fiber.StatusOK, web.SectionListPage{Sections: sections})
}

func (h *Handler) SectionCreatePage(c fiber.Ctx) error {
	classes, _ := h.services.Academic.ListClasses(c.Context())
	return h.render(c, fiber.StatusOK, web.SectionFormPage{Title: "Create Section", Classes: classes})
}

func (h *Handler) SectionCreate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	classID, err := uuid.Parse(c.FormValue("class_id"))
	if err != nil {
		h.flash(c, "Invalid class", true)
		return c.Redirect().To("/sections/new")
	}
	req := dto.SectionRequest{ClassID: classID, Name: c.FormValue("name")}
	if v := c.FormValue("capacity"); v != "" {
		if n, err := parseInt(v); err == nil {
			req.Capacity = n
		}
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/sections/new")
	}
	if _, err := h.services.Academic.CreateSection(c.Context(), req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/sections/new")
	}
	return c.Redirect().To("/sections")
}

func (h *Handler) SectionEditPage(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	sections, _ := h.services.Academic.ListSections(c.Context())
	var current *dto.SectionResponse
	for i := range sections {
		if sections[i].ID == id {
			current = &sections[i]
			break
		}
	}
	if current == nil {
		return c.Status(404).SendString("Section not found")
	}
	classes, _ := h.services.Academic.ListClasses(c.Context())
	return h.render(c, fiber.StatusOK, web.SectionFormPage{Title: "Edit Section", Section: current, Classes: classes})
}

func (h *Handler) SectionUpdate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	classID, err := uuid.Parse(c.FormValue("class_id"))
	if err != nil {
		h.flash(c, "Invalid class", true)
		return c.Redirect().To("/sections/" + id.String() + "/edit")
	}
	req := dto.SectionRequest{ClassID: classID, Name: c.FormValue("name")}
	if v := c.FormValue("capacity"); v != "" {
		if n, err := parseInt(v); err == nil {
			req.Capacity = n
		}
	}
	if _, err := h.services.Academic.UpdateSection(c.Context(), id, req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/sections/" + id.String() + "/edit")
	}
	return c.Redirect().To("/sections")
}

func (h *Handler) SectionDelete(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	if err := h.services.Academic.DeleteSection(c.Context(), id, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/sections")
}

func (h *Handler) SubjectList(c fiber.Ctx) error {
	subjects, _ := h.services.Academic.ListSubjects(c.Context())
	return h.render(c, fiber.StatusOK, web.SubjectListPage{Subjects: subjects})
}

func (h *Handler) SubjectCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.SubjectFormPage{Title: "Create Subject"})
}

func (h *Handler) SubjectCreate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	req := dto.SubjectRequest{Name: c.FormValue("name"), Code: c.FormValue("code"), Description: c.FormValue("description")}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/subjects/new")
	}
	if _, err := h.services.Academic.CreateSubject(c.Context(), req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/subjects/new")
	}
	return c.Redirect().To("/subjects")
}

func (h *Handler) SubjectEditPage(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	subjects, _ := h.services.Academic.ListSubjects(c.Context())
	var current *dto.SubjectResponse
	for i := range subjects {
		if subjects[i].ID == id {
			current = &subjects[i]
			break
		}
	}
	if current == nil {
		return c.Status(404).SendString("Subject not found")
	}
	return h.render(c, fiber.StatusOK, web.SubjectFormPage{Title: "Edit Subject", Subject: current})
}

func (h *Handler) SubjectUpdate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	req := dto.SubjectRequest{Name: c.FormValue("name"), Code: c.FormValue("code"), Description: c.FormValue("description")}
	if _, err := h.services.Academic.UpdateSubject(c.Context(), id, req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/subjects/" + id.String() + "/edit")
	}
	return c.Redirect().To("/subjects")
}

func (h *Handler) SubjectDelete(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	if err := h.services.Academic.DeleteSubject(c.Context(), id, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/subjects")
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
