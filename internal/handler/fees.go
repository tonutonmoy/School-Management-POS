package handler

import (
	"fmt"
	"strconv"
	"time"

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

func (h *Handler) registerFeeRoutes(auth fiber.Router, mw *middleware.Middleware) {
	// Fee types
	ft := auth.Group("/fees/types")
	ft.Get("/", mw.RequirePermission(model.PermFeeReportView), h.FeeTypeList)
	ft.Get("/new", mw.RequirePermission(model.PermFeeCreate), h.FeeTypeCreatePage)
	ft.Post("/", mw.CSRFProtect(), mw.RequirePermission(model.PermFeeCreate), h.FeeTypeCreate)
	ft.Get("/:id/edit", mw.RequirePermission(model.PermFeeUpdate), h.FeeTypeEditPage)
	ft.Post("/:id", mw.CSRFProtect(), mw.RequirePermission(model.PermFeeUpdate), h.FeeTypeUpdate)
	ft.Post("/:id/delete", mw.CSRFProtect(), mw.RequirePermission(model.PermFeeDelete), h.FeeTypeDelete)

	// Fee structures
	fs := auth.Group("/fees/structures")
	fs.Get("/", mw.RequirePermission(model.PermFeeReportView), h.FeeStructureList)
	fs.Get("/new", mw.RequirePermission(model.PermFeeCreate), h.FeeStructureCreatePage)
	fs.Post("/", mw.CSRFProtect(), mw.RequirePermission(model.PermFeeCreate), h.FeeStructureCreate)
	fs.Get("/:id/edit", mw.RequirePermission(model.PermFeeUpdate), h.FeeStructureEditPage)
	fs.Post("/:id", mw.CSRFProtect(), mw.RequirePermission(model.PermFeeUpdate), h.FeeStructureUpdate)
	fs.Post("/:id/delete", mw.CSRFProtect(), mw.RequirePermission(model.PermFeeDelete), h.FeeStructureDelete)

	// Discounts
	auth.Get("/fees/discounts", mw.RequirePermission(model.PermFeeUpdate), h.DiscountList)
	auth.Get("/fees/discounts/new", mw.RequirePermission(model.PermFeeUpdate), h.DiscountCreatePage)
	auth.Post("/fees/discounts", mw.CSRFProtect(), mw.RequirePermission(model.PermFeeUpdate), h.DiscountCreate)

	// Billing
	bills := auth.Group("/fees/bills")
	bills.Get("/", mw.RequirePermission(model.PermFeeReportView), h.BillList)
	bills.Get("/generate", mw.RequirePermission(model.PermFeeCreate), h.BillGeneratePage)
	bills.Post("/generate", mw.CSRFProtect(), mw.RequirePermission(model.PermFeeCreate), h.BillGenerate)
	bills.Get("/:id", mw.RequirePermission(model.PermFeeReportView), h.BillDetail)

	// Collection
	collect := auth.Group("/fees/collect", mw.RequirePermission(model.PermFeeCollect))
	collect.Get("/", h.FeeCollectPage)
	collect.Post("/", mw.CSRFProtect(), h.FeeCollect)
	collect.Get("/student/:id/bills", h.StudentPendingBillsHTMX)

	// Finance dashboard
	auth.Get("/finance/dashboard", mw.RequirePermission(model.PermFeeReportView), h.FinanceDashboard)

	// Due management
	auth.Get("/fees/dues", mw.RequirePermission(model.PermFeeReportView), h.DueList)
	auth.Get("/fees/overdue", mw.RequirePermission(model.PermFeeReportView), h.OverdueList)

	// Receipts
	auth.Get("/receipts/:id", h.ReceiptView)
	auth.Get("/receipts/:id/pdf", h.ReceiptPDF)
	auth.Get("/receipts/verify/:token", h.ReceiptVerify)

	// Parent view
	auth.Get("/parent/students/:id/fees", mw.RequirePermission(model.PermFeeReportView), h.ParentFeeView)

	// Reports
	reports := auth.Group("/reports/fees", mw.RequirePermission(model.PermFeeReportView))
	reports.Get("/collection/daily", h.ReportDailyCollection)
	reports.Get("/collection/monthly", h.ReportMonthlyCollection)
	reports.Get("/collection/yearly", h.ReportYearlyCollection)
	reports.Get("/students/:id/ledger", h.ReportStudentLedger)
	reports.Get("/students/:id/payments", h.ReportPaymentHistory)
	reports.Get("/students/:id/dues", h.ReportDueStatement)
	reports.Get("/income", h.ReportIncomeSummary)
	reports.Get("/by-method", h.ReportByMethod)
	reports.Get("/by-fee-type", h.ReportByFeeType)
	reports.Get("/collection/export.csv", h.ExportCollectionCSV)
	reports.Get("/collection/export.xlsx", h.ExportCollectionExcel)
	reports.Get("/bills/export.csv", h.ExportBillsCSV)

	// Refund
	auth.Post("/fees/payments/:id/refund", mw.CSRFProtect(), mw.RequirePermission(model.PermFeeRefund), h.PaymentRefund)
}

func (h *Handler) feeFormData(c fiber.Ctx) *web.FeeFormData {
	sessions, _ := h.services.Sessions.List(c.Context())
	classes, _ := h.services.Academic.ListClasses(c.Context())
	sections, _ := h.services.Academic.ListSections(c.Context())
	feeTypes, _ := h.services.Fees.ListFeeTypes(c.Context(), false)
	return &web.FeeFormData{Sessions: sessions, Classes: classes, Sections: sections, FeeTypes: feeTypes}
}

func (h *Handler) FeeTypeList(c fiber.Ctx) error {
	types, _ := h.services.Fees.ListFeeTypes(c.Context(), false)
	return h.render(c, fiber.StatusOK, web.FeeTypeListPage{Types: types})
}

func (h *Handler) FeeTypeCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.FeeTypeFormPage{Title: "Create Fee Type"})
}

func (h *Handler) FeeTypeCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req := dto.FeeTypeRequest{
		Name: c.FormValue("name"), Slug: c.FormValue("slug"), Description: c.FormValue("description"),
		IsActive: c.FormValue("is_active") == "on" || c.FormValue("is_active") == "true",
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/fees/types/new")
	}
	if _, err := h.services.Fees.CreateFeeType(c.Context(), req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/fees/types/new")
	}
	return c.Redirect().To("/fees/types")
}

func (h *Handler) FeeTypeEditPage(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	ft, err := h.services.Fees.GetFeeType(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.FeeTypeFormPage{Title: "Edit Fee Type", FeeType: ft})
}

func (h *Handler) FeeTypeUpdate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	req := dto.FeeTypeRequest{
		Name: c.FormValue("name"), Slug: c.FormValue("slug"), Description: c.FormValue("description"),
		IsActive: c.FormValue("is_active") == "on" || c.FormValue("is_active") == "true",
	}
	if _, err := h.services.Fees.UpdateFeeType(c.Context(), id, req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/fees/types")
}

func (h *Handler) FeeTypeDelete(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	_ = h.services.Fees.DeleteFeeType(c.Context(), id, user.ID, c.IP())
	return c.Redirect().To("/fees/types")
}

func (h *Handler) FeeStructureList(c fiber.Ctx) error {
	var sessionID, classID *uuid.UUID
	if s := c.Query("session_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			sessionID = &id
		}
	}
	if cl := c.Query("class_id"); cl != "" {
		if id, err := uuid.Parse(cl); err == nil {
			classID = &id
		}
	}
	structs, _ := h.services.Fees.ListFeeStructures(c.Context(), sessionID, classID)
	return h.render(c, fiber.StatusOK, web.FeeStructureListPage{Structures: structs, FormData: h.feeFormData(c)})
}

func (h *Handler) FeeStructureCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.FeeStructureFormPage{Title: "Create Fee Structure", FormData: h.feeFormData(c)})
}

func (h *Handler) FeeStructureCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req, err := h.parseFeeStructureForm(c)
	if err != nil || len(h.validate.Validate(req)) > 0 {
		h.flash(c, "Invalid form data", true)
		return c.Redirect().To("/fees/structures/new")
	}
	if _, err := h.services.Fees.CreateFeeStructure(c.Context(), req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/fees/structures/new")
	}
	return c.Redirect().To("/fees/structures")
}

func (h *Handler) FeeStructureEditPage(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	// load via list filter - use Get from service - we need GetFeeStructure - add to service or use list
	structs, _ := h.services.Fees.ListFeeStructures(c.Context(), nil, nil)
	var found *dto.FeeStructureResponse
	for i := range structs {
		if structs[i].ID == id {
			found = &structs[i]
			break
		}
	}
	if found == nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.FeeStructureFormPage{Title: "Edit Fee Structure", Structure: found, FormData: h.feeFormData(c)})
}

func (h *Handler) FeeStructureUpdate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	req, _ := h.parseFeeStructureForm(c)
	if _, err := h.services.Fees.UpdateFeeStructure(c.Context(), id, req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/fees/structures")
}

func (h *Handler) FeeStructureDelete(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	_ = h.services.Fees.DeleteFeeStructure(c.Context(), id, user.ID, c.IP())
	return c.Redirect().To("/fees/structures")
}

func (h *Handler) parseFeeStructureForm(c fiber.Ctx) (dto.FeeStructureRequest, error) {
	req := dto.FeeStructureRequest{
		Frequency: c.FormValue("frequency"),
		IsActive:  c.FormValue("is_active") == "on" || c.FormValue("is_active") == "true",
	}
	req.FeeTypeID, _ = uuid.Parse(c.FormValue("fee_type_id"))
	req.SessionID, _ = uuid.Parse(c.FormValue("session_id"))
	req.ClassID, _ = uuid.Parse(c.FormValue("class_id"))
	if s := c.FormValue("section_id"); s != "" {
		req.SectionID, _ = uuid.Parse(s)
	}
	req.Amount, _ = strconv.ParseFloat(c.FormValue("amount"), 64)
	req.DueDay, _ = strconv.Atoi(c.FormValue("due_day"))
	if req.DueDay == 0 {
		req.DueDay = 10
	}
	return req, nil
}

func (h *Handler) DiscountList(c fiber.Ctx) error {
	discounts, _ := h.services.Fees.ListDiscounts(c.Context(), nil, nil)
	return h.render(c, fiber.StatusOK, web.DiscountListPage{Discounts: discounts})
}

func (h *Handler) DiscountCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.DiscountFormPage{FormData: h.feeFormData(c)})
}

func (h *Handler) DiscountCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req := dto.StudentDiscountRequest{
		DiscountType: c.FormValue("discount_type"), Reason: c.FormValue("reason"),
		Description: c.FormValue("description"),
		IsActive: c.FormValue("is_active") == "on" || c.FormValue("is_active") == "true",
	}
	req.StudentID, _ = uuid.Parse(c.FormValue("student_id"))
	req.SessionID, _ = uuid.Parse(c.FormValue("session_id"))
	req.DiscountValue, _ = strconv.ParseFloat(c.FormValue("discount_value"), 64)
	if _, err := h.services.Fees.CreateDiscount(c.Context(), req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/fees/discounts")
}

func (h *Handler) BillList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	filter := dto.BillSearchFilter{
		Status: c.Query("status"), Query: c.Query("q"), Page: page, PerPage: 20,
	}
	if s := c.Query("session_id"); s != "" {
		filter.SessionID, _ = uuid.Parse(s)
	}
	if cl := c.Query("class_id"); cl != "" {
		filter.ClassID, _ = uuid.Parse(cl)
	}
	data, _ := h.services.Fees.ListBills(c.Context(), filter)
	return h.render(c, fiber.StatusOK, web.BillListPage{Data: data, Filter: filter, FormData: h.feeFormData(c)})
}

func (h *Handler) BillGeneratePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.BillGeneratePage{FormData: h.feeFormData(c)})
}

func (h *Handler) BillGenerate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req := dto.GenerateBillsRequest{
		BillPeriod: c.FormValue("bill_period"),
		Regenerate: c.FormValue("regenerate") == "on" || c.FormValue("regenerate") == "true",
	}
	req.SessionID, _ = uuid.Parse(c.FormValue("session_id"))
	if cl := c.FormValue("class_id"); cl != "" {
		req.ClassID, _ = uuid.Parse(cl)
	}
	if sec := c.FormValue("section_id"); sec != "" {
		req.SectionID, _ = uuid.Parse(sec)
	}
	if st := c.FormValue("student_id"); st != "" {
		req.StudentID, _ = uuid.Parse(st)
	}
	if req.BillPeriod == "" {
		req.BillPeriod = time.Now().Format("2006-01")
	}
	count, err := h.services.Fees.GenerateBills(c.Context(), req, user.ID, c.IP())
	if err != nil {
		h.flash(c, err.Error(), true)
	} else {
		h.flash(c, fmt.Sprintf("Generated %d bills", count), false)
	}
	return c.Redirect().To("/fees/bills")
}

func (h *Handler) BillDetail(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	bill, err := h.services.Fees.GetBill(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.BillDetailPage{Bill: bill})
}

func (h *Handler) FeeCollectPage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.FeeCollectPage{Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type")})
}

func (h *Handler) FeeCollect(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req := dto.PaymentRequest{PaymentMethod: c.FormValue("payment_method"), Remarks: c.FormValue("remarks")}
	req.StudentID, _ = uuid.Parse(c.FormValue("student_id"))
	req.Amount, _ = strconv.ParseFloat(c.FormValue("amount"), 64)
	if d, err := parseDate(c.FormValue("collection_date")); err == nil {
		req.CollectionDate = d
	} else {
		req.CollectionDate = time.Now()
	}
	allocations := parsePaymentAllocations(c)
	if _, err := h.services.Fees.CollectPayment(c.Context(), req, allocations, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/fees/collect")
	}
	h.flash(c, "Payment collected successfully", false)
	return c.Redirect().To("/fees/collect")
}

func (h *Handler) StudentPendingBillsHTMX(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.SendString("")
	}
	bills, _ := h.services.Fees.PendingBillsForStudent(c.Context(), id)
	html := `<div class="mt-4 space-y-2">`
	for _, b := range bills {
		html += fmt.Sprintf(`<label class="flex items-center gap-3 rounded-lg border p-3">
			<input type="checkbox" name="bill_%s" value="%.2f" class="bill-check" data-bill-id="%s" data-due="%.2f"/>
			<span class="flex-1"><strong>%s</strong> — Due: %.2f (Period: %s)</span>
			<input type="number" step="0.01" name="alloc_%s" placeholder="Amount" class="w-24 rounded border px-2 py-1 text-sm alloc-amount"/>
		</label>`, b.ID, b.DueAmount, b.ID, b.DueAmount, b.InvoiceNumber, b.DueAmount, b.BillPeriod, b.ID)
	}
	html += `</div><script>
	document.querySelectorAll('.bill-check').forEach(cb => {
		cb.addEventListener('change', function() {
			const alloc = document.querySelector('[name="alloc_' + this.dataset.billId + '"]');
			if (this.checked) alloc.value = this.dataset.due;
			else alloc.value = '';
		});
	});
	</script>`
	if len(bills) == 0 {
		html = `<p class="text-sm text-slate-500">No pending bills for this student.</p>`
	}
	_ = bills
	return c.SendString(html)
}

func (h *Handler) FinanceDashboard(c fiber.Ctx) error {
	stats, _ := h.services.Fees.DashboardStats(c.Context())
	return h.render(c, fiber.StatusOK, web.FinanceDashboardPage{Stats: stats})
}

func (h *Handler) DueList(c fiber.Ctx) error {
	filter := dto.BillSearchFilter{}
	if cl := c.Query("class_id"); cl != "" {
		filter.ClassID, _ = uuid.Parse(cl)
	}
	dues, _ := h.services.Fees.ListDueStudents(c.Context(), filter)
	return h.render(c, fiber.StatusOK, web.DueListPage{Rows: dues, FormData: h.feeFormData(c)})
}

func (h *Handler) OverdueList(c fiber.Ctx) error {
	data, _ := h.services.Fees.ListOverdueBills(c.Context(), dto.BillSearchFilter{})
	return h.render(c, fiber.StatusOK, web.OverdueListPage{Data: data})
}

func (h *Handler) ReceiptView(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	receipt, err := h.services.Fees.GetReceipt(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.ReceiptPage{Receipt: receipt})
}

func (h *Handler) ReceiptPDF(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	receipt, err := h.services.Fees.GetReceipt(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	verifyURL := h.cfg.App.URL + "/receipts/verify/" + receipt.QRToken
	if h.cfg.App.URL == "" {
		verifyURL = "/receipts/verify/" + receipt.QRToken
	}
	data, err := pdf.GenerateReceipt(receipt, verifyURL)
	if err != nil {
		return c.Status(500).SendString("PDF generation failed")
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=receipt-"+receipt.ReceiptNumber+".pdf")
	return c.Send(data)
}

func (h *Handler) ReceiptVerify(c fiber.Ctx) error {
	token := c.Params("token")
	receipt, err := h.services.Fees.VerifyReceipt(c.Context(), token)
	if err != nil {
		return h.render(c, fiber.StatusOK, web.ReceiptVerifyPage{Valid: false})
	}
	return h.render(c, fiber.StatusOK, web.ReceiptVerifyPage{Valid: true, Receipt: receipt})
}

func (h *Handler) ParentFeeView(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	if err := h.requireStudentAccess(c, id, model.PermFeeReportView); err != nil {
		return err
	}
	summary, err := h.services.Fees.ParentFeeSummary(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.ParentFeePage{Summary: summary})
}

func (h *Handler) parseFinanceFilter(c fiber.Ctx) dto.FinanceReportFilter {
	f := dto.FinanceReportFilter{Method: c.Query("method")}
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
	return f
}

func (h *Handler) ReportDailyCollection(c fiber.Ctx) error {
	f := h.parseFinanceFilter(c)
	payments, _ := h.services.Fees.ListPayments(c.Context(), f, 1, 500)
	return h.render(c, fiber.StatusOK, web.CollectionReportPage{Title: "Daily Collection", Payments: payments, Filter: f})
}

func (h *Handler) ReportMonthlyCollection(c fiber.Ctx) error {
	now := time.Now()
	f := dto.FinanceReportFilter{
		From: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()),
		To:   now,
	}
	payments, _ := h.services.Fees.ListPayments(c.Context(), f, 1, 500)
	return h.render(c, fiber.StatusOK, web.CollectionReportPage{Title: "Monthly Collection", Payments: payments, Filter: f})
}

func (h *Handler) ReportYearlyCollection(c fiber.Ctx) error {
	now := time.Now()
	f := dto.FinanceReportFilter{
		From: time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()),
		To:   now,
	}
	payments, _ := h.services.Fees.ListPayments(c.Context(), f, 1, 1000)
	return h.render(c, fiber.StatusOK, web.CollectionReportPage{Title: "Yearly Collection", Payments: payments, Filter: f})
}

func (h *Handler) ReportStudentLedger(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	now := time.Now()
	from := time.Date(now.Year(), now.Month()-6, 1, 0, 0, 0, 0, now.Location())
	entries, _ := h.services.Fees.StudentLedger(c.Context(), id, from, now)
	return h.render(c, fiber.StatusOK, web.StudentLedgerPage{Entries: entries, StudentID: id})
}

func (h *Handler) ReportPaymentHistory(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	f := dto.FinanceReportFilter{StudentID: id}
	payments, _ := h.services.Fees.ListPayments(c.Context(), f, 1, 100)
	return h.render(c, fiber.StatusOK, web.PaymentHistoryPage{Payments: payments, StudentID: id})
}

func (h *Handler) ReportDueStatement(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	filter := dto.BillSearchFilter{StudentID: id}
	data, _ := h.services.Fees.ListBills(c.Context(), filter)
	return h.render(c, fiber.StatusOK, web.DueStatementPage{Data: data, StudentID: id})
}

func (h *Handler) ReportIncomeSummary(c fiber.Ctx) error {
	stats, _ := h.services.Fees.DashboardStats(c.Context())
	return h.render(c, fiber.StatusOK, web.IncomeSummaryPage{Stats: stats})
}

func (h *Handler) ReportByMethod(c fiber.Ctx) error {
	f := h.parseFinanceFilter(c)
	if f.From.IsZero() {
		now := time.Now()
		f.From = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		f.To = now
	}
	stats, _ := h.services.Fees.DashboardStats(c.Context())
	return h.render(c, fiber.StatusOK, web.MethodReportPage{Stats: stats, Filter: f})
}

func (h *Handler) ReportByFeeType(c fiber.Ctx) error {
	f := h.parseFinanceFilter(c)
	if f.From.IsZero() {
		now := time.Now()
		f.From = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		f.To = now
	}
	items, _ := h.services.Fees.FeeTypeCollection(c.Context(), f.From, f.To)
	return h.render(c, fiber.StatusOK, web.FeeTypeReportPage{Items: items, Filter: f})
}

func (h *Handler) ExportCollectionCSV(c fiber.Ctx) error {
	f := h.parseFinanceFilter(c)
	payments, _ := h.services.Fees.ListPayments(c.Context(), f, 1, 5000)
	data, _ := export.CollectionCSV(payments)
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=collection.csv")
	return c.Send(data)
}

func (h *Handler) ExportCollectionExcel(c fiber.Ctx) error {
	f := h.parseFinanceFilter(c)
	payments, _ := h.services.Fees.ListPayments(c.Context(), f, 1, 5000)
	data, _ := export.CollectionExcel(payments)
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=collection.xlsx")
	return c.Send(data)
}

func (h *Handler) ExportBillsCSV(c fiber.Ctx) error {
	filter := dto.BillSearchFilter{PerPage: 5000}
	data, _ := h.services.Fees.ListBills(c.Context(), filter)
	csvData, _ := export.BillsCSV(data.Items)
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=bills.csv")
	return c.Send(csvData)
}

func (h *Handler) PaymentRefund(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	if err := h.services.Fees.RefundPayment(c.Context(), id, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/reports/fees/collection/daily")
}

func parsePaymentAllocations(c fiber.Ctx) []dto.PaymentAllocationInput {
	var allocs []dto.PaymentAllocationInput
	c.Request().PostArgs().VisitAll(func(key, value []byte) {
		k := string(key)
		if len(k) < 6 || k[:6] != "alloc_" {
			return
		}
		billID, err := uuid.Parse(k[6:])
		if err != nil {
			return
		}
		amt, err := strconv.ParseFloat(string(value), 64)
		if err != nil || amt <= 0 {
			return
		}
		allocs = append(allocs, dto.PaymentAllocationInput{BillID: billID, Amount: amt})
	})
	return allocs
}
