package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/export"
	"github.com/school-management/pos/internal/middleware"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/pdf"
	"github.com/school-management/pos/internal/validator"
	"github.com/school-management/pos/internal/web"
)

func (h *Handler) registerExamRoutes(auth fiber.Router, mw *middleware.Middleware) {
	exams := auth.Group("/exams")
	exams.Get("/", mw.RequirePermission(model.PermMarksEntry), h.ExamList)
	exams.Get("/new", mw.RequirePermission(model.PermExamCreate), h.ExamCreatePage)
	exams.Post("/", mw.CSRFProtect(), mw.RequirePermission(model.PermExamCreate), h.ExamCreate)
	exams.Get("/dashboard", mw.RequirePermission(model.PermResultProcess), h.ExamDashboard)

	exam := exams.Group("/:id")
	exam.Get("/", mw.RequirePermission(model.PermMarksEntry), h.ExamDetail)
	exam.Get("/edit", mw.RequirePermission(model.PermExamUpdate), h.ExamEditPage)
	exam.Post("/", mw.CSRFProtect(), mw.RequirePermission(model.PermExamUpdate), h.ExamUpdate)
	exam.Post("/delete", mw.CSRFProtect(), mw.RequirePermission(model.PermExamDelete), h.ExamDelete)
	exam.Post("/publish", mw.CSRFProtect(), mw.RequirePermission(model.PermExamPublish), h.ExamPublish)
	exam.Post("/archive", mw.CSRFProtect(), mw.RequirePermission(model.PermExamUpdate), h.ExamArchive)

	exam.Post("/subjects", mw.CSRFProtect(), mw.RequirePermission(model.PermExamUpdate), h.ExamSubjectAdd)
	exam.Get("/marks", mw.RequirePermission(model.PermMarksEntry), h.MarksEntryPage)
	exam.Get("/marks/:subjectId", mw.RequirePermission(model.PermMarksEntry), h.MarksEntrySubject)
	exam.Post("/marks/:subjectId", mw.CSRFProtect(), mw.RequirePermission(model.PermMarksEntry), h.MarksSave)

	exam.Post("/process", mw.CSRFProtect(), mw.RequirePermission(model.PermResultProcess), h.ExamProcessResults)
	exam.Post("/publish-results", mw.CSRFProtect(), mw.RequirePermission(model.PermResultPublish), h.ExamPublishResults)
	exam.Get("/results", mw.RequirePermission(model.PermResultProcess), h.ExamResultsList)
	exam.Get("/results/:resultId", mw.RequirePermission(model.PermResultProcess), h.ExamResultDetail)

	exam.Get("/tabulation", mw.RequirePermission(model.PermResultProcess), h.TabulationPage)
	exam.Get("/tabulation/export.pdf", mw.RequirePermission(model.PermResultProcess), h.TabulationPDF)
	exam.Get("/tabulation/export.csv", mw.RequirePermission(model.PermResultProcess), h.TabulationCSV)
	exam.Get("/tabulation/export.xlsx", mw.RequirePermission(model.PermResultProcess), h.TabulationExcel)
	exam.Get("/merit-list", mw.RequirePermission(model.PermResultProcess), h.MeritListPage)

	auth.Get("/grading-systems", mw.RequirePermission(model.PermResultProcess), h.GradingSystemList)
	auth.Post("/grading-systems", mw.CSRFProtect(), mw.RequirePermission(model.PermResultProcess), h.GradingSystemCreate)

	auth.Get("/report-cards/:resultId/pdf", mw.RequirePermission(model.PermResultProcess), h.ReportCardPDF)

	auth.Get("/parent/students/:id/results", mw.RequirePermission(model.PermResultPublish), h.ParentResultsList)
	auth.Get("/parent/students/:id/results/:examId", mw.RequirePermission(model.PermResultPublish), h.ParentResultView)
	auth.Get("/parent/students/:id/results/:examId/report-card", mw.RequirePermission(model.PermResultPublish), h.ParentReportCardPDF)

	reports := auth.Group("/reports/exams", mw.RequirePermission(model.PermResultProcess))
	reports.Get("/summary", h.ExamReportSummary)
	reports.Get("/subject-performance", h.ExamReportSubjectPerf)
	reports.Get("/top-students", h.ExamReportTopStudents)
	reports.Get("/failed-students", h.ExamReportFailedStudents)
	reports.Get("/export.csv", h.ExamResultsExportCSV)
}

func (h *Handler) examFormData(c fiber.Ctx) *web.ExamFormData {
	sessions, _ := h.services.Sessions.List(c.Context())
	classes, _ := h.services.Academic.ListClasses(c.Context())
	sections, _ := h.services.Academic.ListSections(c.Context())
	subjects, _ := h.services.Academic.ListSubjects(c.Context())
	grading, _ := h.services.Exam.ListGradingSystems(c.Context())
	examList, _ := h.services.Exam.ListExams(c.Context(), dto.ExamSearchFilter{PerPage: 100, Page: 1})
	var exams []dto.ExamResponse
	if examList != nil {
		exams = examList.Items
	}
	return &web.ExamFormData{
		Sessions: sessions, Classes: classes, Sections: sections,
		Subjects: subjects, GradingSystems: grading, Exams: exams,
	}
}

func (h *Handler) ExamList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	filter := dto.ExamSearchFilter{
		Status: c.Query("status"), Query: c.Query("q"), Page: page, PerPage: 20,
	}
	if s := c.Query("session_id"); s != "" {
		filter.SessionID, _ = uuid.Parse(s)
	}
	if cl := c.Query("class_id"); cl != "" {
		filter.ClassID, _ = uuid.Parse(cl)
	}
	data, _ := h.services.Exam.ListExams(c.Context(), filter)
	return h.render(c, fiber.StatusOK, web.ExamListPage{
		Data: data, Filter: filter, FormData: h.examFormData(c),
		Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
	})
}

func (h *Handler) ExamCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.ExamFormPage{
		Title: "Create Exam", FormData: h.examFormData(c),
		Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
	})
}

func (h *Handler) ExamCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req, err := parseExamRequest(c)
	if err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/exams/new")
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/exams/new")
	}
	exam, err := h.services.Exam.CreateExam(c.Context(), req, user.ID, c.IP())
	if err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/exams/new")
	}
	return c.Redirect().To("/exams/" + exam.ID.String())
}

func (h *Handler) ExamEditPage(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	exam, err := h.services.Exam.GetExam(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.ExamFormPage{
		Title: "Edit Exam", Exam: exam, FormData: h.examFormData(c),
	})
}

func (h *Handler) ExamUpdate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	req, err := parseExamRequest(c)
	if err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/exams/" + id.String() + "/edit")
	}
	if _, err := h.services.Exam.UpdateExam(c.Context(), id, req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/exams/" + id.String())
}

func (h *Handler) ExamDelete(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	_ = h.services.Exam.DeleteExam(c.Context(), id, user.ID, c.IP())
	return c.Redirect().To("/exams")
}

func (h *Handler) ExamDetail(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	exam, err := h.services.Exam.GetExam(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	subjects, _ := h.services.Exam.ListExamSubjects(c.Context(), id)
	return h.render(c, fiber.StatusOK, web.ExamDetailPage{
		Exam: exam, Subjects: subjects, FormData: h.examFormData(c),
		Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
	})
}

func (h *Handler) ExamPublish(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	if err := h.services.Exam.PublishExam(c.Context(), id, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	} else {
		h.flash(c, "Exam published", false)
	}
	return c.Redirect().To("/exams/" + id.String())
}

func (h *Handler) ExamArchive(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	if err := h.services.Exam.ArchiveExam(c.Context(), id, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	} else {
		h.flash(c, "Exam archived", false)
	}
	return c.Redirect().To("/exams/" + id.String())
}

func (h *Handler) ExamSubjectAdd(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	examID, _ := parseUUIDParam(c, "id")
	req := dto.ExamSubjectRequest{
		SubjectID:      parseFormUUID(c, "subject_id"),
		FullMarks:      parseFormFloat(c, "full_marks"),
		PassMarks:      parseFormFloat(c, "pass_marks"),
		WrittenMarks:   parseFormFloat(c, "written_marks"),
		MCQMarks:       parseFormFloat(c, "mcq_marks"),
		PracticalMarks: parseFormFloat(c, "practical_marks"),
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/exams/" + examID.String())
	}
	if _, err := h.services.Exam.AddExamSubject(c.Context(), examID, req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/exams/" + examID.String())
}

func (h *Handler) MarksEntryPage(c fiber.Ctx) error {
	examID, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	exam, _ := h.services.Exam.GetExam(c.Context(), examID)
	subjects, _ := h.services.Exam.ListExamSubjects(c.Context(), examID)
	return h.render(c, fiber.StatusOK, web.MarksEntryPage{Exam: exam, Subjects: subjects})
}

func (h *Handler) MarksEntrySubject(c fiber.Ctx) error {
	examID, _ := parseUUIDParam(c, "id")
	subjectID, err := parseUUIDParam(c, "subjectId")
	if err != nil {
		return c.Status(400).SendString("Invalid subject")
	}
	exam, _ := h.services.Exam.GetExam(c.Context(), examID)
	subjects, _ := h.services.Exam.ListExamSubjects(c.Context(), examID)
	var active *dto.ExamSubjectResponse
	for i := range subjects {
		if subjects[i].ID == subjectID {
			active = &subjects[i]
			break
		}
	}
	rows, _ := h.services.Exam.MarksSheet(c.Context(), subjectID)
	return h.render(c, fiber.StatusOK, web.MarksEntryPage{
		Exam: exam, Subjects: subjects, ActiveSubject: active, Rows: rows,
		Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
	})
}

func (h *Handler) MarksSave(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	examID, _ := parseUUIDParam(c, "id")
	subjectID, _ := parseUUIDParam(c, "subjectId")
	entries := parseBulkMarkEntries(c)
	if err := h.services.Exam.BulkSaveMarks(c.Context(), examID, subjectID, entries, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	} else {
		h.flash(c, "Marks saved", false)
	}
	return c.Redirect().To("/exams/" + examID.String() + "/marks/" + subjectID.String())
}

func (h *Handler) ExamProcessResults(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	count, err := h.services.Exam.ProcessResults(c.Context(), id, user.ID, c.IP())
	if err != nil {
		h.flash(c, err.Error(), true)
	} else {
		h.flash(c, fmt.Sprintf("Processed results for %d students", count), false)
	}
	return c.Redirect().To("/exams/" + id.String())
}

func (h *Handler) ExamPublishResults(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	if err := h.services.Exam.PublishResults(c.Context(), id, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	} else {
		h.flash(c, "Results published", false)
	}
	return c.Redirect().To("/exams/" + id.String())
}

func (h *Handler) ExamResultsList(c fiber.Ctx) error {
	examID, _ := parseUUIDParam(c, "id")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	exam, _ := h.services.Exam.GetExam(c.Context(), examID)
	filter := dto.ExamReportFilter{ExamID: examID}
	if sec := c.Query("section_id"); sec != "" {
		filter.SectionID, _ = uuid.Parse(sec)
	}
	data, _ := h.services.Exam.ListResults(c.Context(), filter, page, 50)
	return h.render(c, fiber.StatusOK, web.ExamResultsPage{
		Exam: exam, Data: data, Filter: filter, FormData: h.examFormData(c),
	})
}

func (h *Handler) ExamResultDetail(c fiber.Ctx) error {
	examID, _ := parseUUIDParam(c, "id")
	resultID, err := parseUUIDParam(c, "resultId")
	if err != nil {
		return c.Status(400).SendString("Invalid result")
	}
	exam, _ := h.services.Exam.GetExam(c.Context(), examID)
	result, err := h.services.Exam.GetResult(c.Context(), resultID)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.ExamResultDetailPage{Exam: exam, Result: result})
}

func (h *Handler) ExamDashboard(c fiber.Ctx) error {
	stats, _ := h.services.Exam.DashboardStats(c.Context())
	var examStats *dto.ExamDashboardStats
	if eid := c.Query("exam_id"); eid != "" {
		if id, err := uuid.Parse(eid); err == nil {
			examStats, _ = h.services.Exam.DashboardStatsForExam(c.Context(), id)
		}
	}
	exams, _ := h.services.Exam.ListExams(c.Context(), dto.ExamSearchFilter{PerPage: 50, Page: 1})
	return h.render(c, fiber.StatusOK, web.ExamDashboardPage{
		Stats: stats, ExamStats: examStats, Exams: exams, FormData: h.examFormData(c),
	})
}

func (h *Handler) TabulationPage(c fiber.Ctx) error {
	examID, _ := parseUUIDParam(c, "id")
	exam, _ := h.services.Exam.GetExam(c.Context(), examID)
	var sectionID *uuid.UUID
	if sec := c.Query("section_id"); sec != "" {
		if id, err := uuid.Parse(sec); err == nil {
			sectionID = &id
		}
	}
	rows, _ := h.services.Exam.Tabulation(c.Context(), examID, sectionID)
	return h.render(c, fiber.StatusOK, web.TabulationPage{
		Exam: exam, Results: rows, SectionID: sectionID, FormData: h.examFormData(c),
	})
}

func (h *Handler) TabulationPDF(c fiber.Ctx) error {
	examID, _ := parseUUIDParam(c, "id")
	exam, _ := h.services.Exam.GetExam(c.Context(), examID)
	var sectionID *uuid.UUID
	if sec := c.Query("section_id"); sec != "" {
		if id, err := uuid.Parse(sec); err == nil {
			sectionID = &id
		}
	}
	rows, _ := h.services.Exam.Tabulation(c.Context(), examID, sectionID)
	data, err := pdf.GenerateTabulation(exam.Name, exam.ClassName, rows)
	if err != nil {
		return c.Status(500).SendString("PDF generation failed")
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", `attachment; filename="tabulation.pdf"`)
	return c.Send(data)
}

func (h *Handler) TabulationCSV(c fiber.Ctx) error {
	examID, _ := parseUUIDParam(c, "id")
	var sectionID *uuid.UUID
	if sec := c.Query("section_id"); sec != "" {
		if id, err := uuid.Parse(sec); err == nil {
			sectionID = &id
		}
	}
	rows, _ := h.services.Exam.Tabulation(c.Context(), examID, sectionID)
	data, err := export.ResultsCSV(rows)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", `attachment; filename="tabulation.csv"`)
	return c.Send(data)
}

func (h *Handler) TabulationExcel(c fiber.Ctx) error {
	examID, _ := parseUUIDParam(c, "id")
	exam, _ := h.services.Exam.GetExam(c.Context(), examID)
	var sectionID *uuid.UUID
	if sec := c.Query("section_id"); sec != "" {
		if id, err := uuid.Parse(sec); err == nil {
			sectionID = &id
		}
	}
	rows, _ := h.services.Exam.Tabulation(c.Context(), examID, sectionID)
	data, err := export.TabulationExcel(exam.Name, rows)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="tabulation.xlsx"`)
	return c.Send(data)
}

func (h *Handler) MeritListPage(c fiber.Ctx) error {
	examID, _ := parseUUIDParam(c, "id")
	exam, _ := h.services.Exam.GetExam(c.Context(), examID)
	filter := dto.ExamReportFilter{ExamID: examID}
	data, _ := h.services.Exam.ListResults(c.Context(), filter, 1, 100)
	return h.render(c, fiber.StatusOK, web.MeritListPage{Exam: exam, Results: data.Items})
}

func (h *Handler) GradingSystemList(c fiber.Ctx) error {
	systems, _ := h.services.Exam.ListGradingSystems(c.Context())
	return h.render(c, fiber.StatusOK, web.GradingSystemPage{
		Systems: systems, Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
	})
}

func (h *Handler) GradingSystemCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	name := c.FormValue("name")
	scales := defaultGradingScalesFromForm(c)
	if _, err := h.services.Exam.CreateGradingSystem(c.Context(), name, scales, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	} else {
		h.flash(c, "Grading system created", false)
	}
	return c.Redirect().To("/grading-systems")
}

func (h *Handler) ReportCardPDF(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	resultID, err := parseUUIDParam(c, "resultId")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	data, err := h.services.Exam.GenerateReportCard(c.Context(), resultID, user.ID, c.IP())
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	pdfData, err := pdf.GenerateReportCard(data)
	if err != nil {
		return c.Status(500).SendString("PDF generation failed")
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", `attachment; filename="report-card.pdf"`)
	return c.Send(pdfData)
}

func (h *Handler) ParentResultsList(c fiber.Ctx) error {
	studentID, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid student")
	}
	if err := h.requireStudentAccess(c, studentID, model.PermResultPublish); err != nil {
		return err
	}
	student, _ := h.services.Students.GetFull(c.Context(), studentID)
	filter := dto.ExamReportFilter{StudentID: studentID, PublishedOnly: true}
	data, _ := h.services.Exam.ListResults(c.Context(), filter, 1, 50)
	return h.render(c, fiber.StatusOK, web.ParentResultsPage{Student: student, Results: data.Items})
}

func (h *Handler) ParentResultView(c fiber.Ctx) error {
	studentID, _ := parseUUIDParam(c, "id")
	examID, _ := parseUUIDParam(c, "examId")
	if err := h.requireStudentAccess(c, studentID, model.PermResultPublish); err != nil {
		return err
	}
	result, err := h.services.Exam.ParentResult(c.Context(), examID, studentID)
	if err != nil {
		return c.Status(403).SendString("Result not available")
	}
	return h.render(c, fiber.StatusOK, web.ParentResultPage{Result: result, StudentID: studentID})
}

func (h *Handler) ParentReportCardPDF(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	studentID, _ := parseUUIDParam(c, "id")
	examID, _ := parseUUIDParam(c, "examId")
	if err := h.requireStudentAccess(c, studentID, model.PermResultPublish); err != nil {
		return err
	}
	result, err := h.services.Exam.ParentResult(c.Context(), examID, studentID)
	if err != nil {
		return c.Status(403).SendString("Result not available")
	}
	data, err := h.services.Exam.GenerateReportCard(c.Context(), result.ID, user.ID, c.IP())
	if err != nil {
		return c.Status(500).SendString("Failed")
	}
	pdfData, err := pdf.GenerateReportCard(data)
	if err != nil {
		return c.Status(500).SendString("PDF generation failed")
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", `attachment; filename="report-card.pdf"`)
	return c.Send(pdfData)
}

func (h *Handler) ExamReportSummary(c fiber.Ctx) error {
	examID, _ := uuid.Parse(c.Query("exam_id"))
	exam, _ := h.services.Exam.GetExam(c.Context(), examID)
	stats, _ := h.services.Exam.DashboardStatsForExam(c.Context(), examID)
	return h.render(c, fiber.StatusOK, web.ExamReportSummaryPage{Exam: exam, Stats: stats, FormData: h.examFormData(c)})
}

func (h *Handler) ExamReportSubjectPerf(c fiber.Ctx) error {
	examID, _ := uuid.Parse(c.Query("exam_id"))
	exam, _ := h.services.Exam.GetExam(c.Context(), examID)
	stats, _ := h.services.Exam.DashboardStatsForExam(c.Context(), examID)
	return h.render(c, fiber.StatusOK, web.ExamReportSubjectPage{Exam: exam, Stats: stats, FormData: h.examFormData(c)})
}

func (h *Handler) ExamReportTopStudents(c fiber.Ctx) error {
	examID, _ := uuid.Parse(c.Query("exam_id"))
	exam, _ := h.services.Exam.GetExam(c.Context(), examID)
	filter := dto.ExamReportFilter{ExamID: examID, PassedOnly: true}
	data, _ := h.services.Exam.ListResults(c.Context(), filter, 1, 20)
	return h.render(c, fiber.StatusOK, web.ExamReportTopPage{Exam: exam, Results: data.Items, FormData: h.examFormData(c)})
}

func (h *Handler) ExamReportFailedStudents(c fiber.Ctx) error {
	examID, _ := uuid.Parse(c.Query("exam_id"))
	exam, _ := h.services.Exam.GetExam(c.Context(), examID)
	filter := dto.ExamReportFilter{ExamID: examID, FailedOnly: true}
	data, _ := h.services.Exam.ListResults(c.Context(), filter, 1, 100)
	return h.render(c, fiber.StatusOK, web.ExamReportFailedPage{Exam: exam, Results: data.Items, FormData: h.examFormData(c)})
}

func (h *Handler) ExamResultsExportCSV(c fiber.Ctx) error {
	examID, _ := uuid.Parse(c.Query("exam_id"))
	filter := dto.ExamReportFilter{ExamID: examID}
	data, _ := h.services.Exam.ListResults(c.Context(), filter, 1, 1000)
	csvData, err := export.ResultsCSV(data.Items)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", `attachment; filename="exam-results.csv"`)
	return c.Send(csvData)
}

func parseExamRequest(c fiber.Ctx) (dto.ExamRequest, error) {
	start, err := parseDate(c.FormValue("start_date"))
	if err != nil {
		return dto.ExamRequest{}, fmt.Errorf("invalid start date")
	}
	end, err := parseDate(c.FormValue("end_date"))
	if err != nil {
		return dto.ExamRequest{}, fmt.Errorf("invalid end date")
	}
	return dto.ExamRequest{
		Name: c.FormValue("name"), ExamType: c.FormValue("exam_type"),
		SessionID: parseFormUUID(c, "session_id"), ClassID: parseFormUUID(c, "class_id"),
		StartDate: start, EndDate: end,
		TotalMarks: parseFormFloat(c, "total_marks"), PassingMarks: parseFormFloat(c, "passing_marks"),
		GradingSystemID: parseFormUUID(c, "grading_system_id"),
	}, nil
}

func parseFormUUID(c fiber.Ctx, key string) uuid.UUID {
	id, _ := uuid.Parse(c.FormValue(key))
	return id
}

func parseFormFloat(c fiber.Ctx, key string) float64 {
	v, _ := strconv.ParseFloat(c.FormValue(key), 64)
	return v
}

func parseBulkMarkEntries(c fiber.Ctx) []dto.StudentMarkEntry {
	var entries []dto.StudentMarkEntry
	seen := map[string]bool{}
	c.Request().PostArgs().VisitAll(func(key, value []byte) {
		k := string(key)
		if len(k) < 8 || k[:8] != "written_" {
			return
		}
		idStr := k[8:]
		if seen[idStr] {
			return
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return
		}
		seen[idStr] = true
		entries = append(entries, dto.StudentMarkEntry{
			StudentID: id,
			WrittenScore:   parseFormFloat(c, "written_"+idStr),
			MCQScore:       parseFormFloat(c, "mcq_"+idStr),
			PracticalScore: parseFormFloat(c, "practical_"+idStr),
			IsAbsent:       c.FormValue("absent_"+idStr) == "on" || c.FormValue("absent_"+idStr) == "true",
		})
	})
	return entries
}

func defaultGradingScalesFromForm(c fiber.Ctx) []dto.GradingScaleRequest {
	grades := []struct{ grade string; min, max, gpa float64 }{
		{"A+", 80, 100, 5.0}, {"A", 70, 79.99, 4.0}, {"A-", 60, 69.99, 3.5},
		{"B", 50, 59.99, 3.0}, {"C", 40, 49.99, 2.0}, {"D", 33, 39.99, 1.0}, {"F", 0, 32.99, 0.0},
	}
	if c.FormValue("use_default") != "on" {
		var scales []dto.GradingScaleRequest
		for i := 1; i <= 7; i++ {
			g := c.FormValue(fmt.Sprintf("grade_%d", i))
			if g == "" {
				continue
			}
			scales = append(scales, dto.GradingScaleRequest{
				Grade: g,
				MinPercentage: parseFormFloat(c, fmt.Sprintf("min_%d", i)),
				MaxPercentage: parseFormFloat(c, fmt.Sprintf("max_%d", i)),
				GPAPoint:      parseFormFloat(c, fmt.Sprintf("gpa_%d", i)),
			})
		}
		if len(scales) > 0 {
			return scales
		}
	}
	scales := make([]dto.GradingScaleRequest, len(grades))
	for i, g := range grades {
		scales[i] = dto.GradingScaleRequest{Grade: g.grade, MinPercentage: g.min, MaxPercentage: g.max, GPAPoint: g.gpa}
	}
	return scales
}
