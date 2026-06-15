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

func (h *Handler) registerAccountingRoutes(auth fiber.Router, mw *middleware.Middleware) {
	acct := auth.Group("/accounting", mw.RequirePermission(model.PermAccountingView))
	acct.Get("/dashboard", h.AccountingDashboard)
	acct.Get("/accounts", h.AccountList)
	acct.Get("/accounts/new", mw.RequirePermission(model.PermAccountingManage), h.AccountCreatePage)
	acct.Post("/accounts", mw.CSRFProtect(), mw.RequirePermission(model.PermAccountingManage), h.AccountCreate)
	acct.Get("/accounts/:id/edit", mw.RequirePermission(model.PermAccountingManage), h.AccountEditPage)
	acct.Post("/accounts/:id", mw.CSRFProtect(), mw.RequirePermission(model.PermAccountingManage), h.AccountUpdate)
	acct.Post("/accounts/:id/disable", mw.CSRFProtect(), mw.RequirePermission(model.PermAccountingManage), h.AccountDisable)

	acct.Get("/journal", h.JournalList)
	acct.Get("/journal/new", mw.RequirePermission(model.PermAccountingManage), h.JournalCreatePage)
	acct.Post("/journal", mw.CSRFProtect(), mw.RequirePermission(model.PermAccountingManage), h.JournalCreate)
	acct.Get("/journal/:id", h.JournalDetail)

	acct.Get("/ledger", h.LedgerPage)
	acct.Get("/ledger/export.csv", h.LedgerExportCSV)
	acct.Get("/cash-book", h.CashBookPage)
	acct.Get("/bank-book", h.BankBookPage)

	exp := auth.Group("/expenses")
	exp.Get("/", mw.RequirePermission(model.PermAccountingView), h.ExpenseList)
	exp.Get("/new", mw.RequirePermission(model.PermExpenseCreate), h.ExpenseCreatePage)
	exp.Post("/", mw.CSRFProtect(), mw.RequirePermission(model.PermExpenseCreate), h.ExpenseCreate)
	exp.Post("/:id/approve", mw.CSRFProtect(), mw.RequirePermission(model.PermExpenseApprove), h.ExpenseApprove)
	exp.Post("/:id/attachment", mw.CSRFProtect(), mw.RequirePermission(model.PermExpenseCreate), h.ExpenseAttachment)

	auth.Get("/income", mw.RequirePermission(model.PermAccountingManage), h.IncomeList)
	auth.Get("/income/new", mw.RequirePermission(model.PermAccountingManage), h.IncomeCreatePage)
	auth.Post("/income", mw.CSRFProtect(), mw.RequirePermission(model.PermAccountingManage), h.IncomeCreate)

	reports := auth.Group("/reports/accounting", mw.RequirePermission(model.PermFinanceReportView))
	reports.Get("/trial-balance", h.TrialBalanceReport)
	reports.Get("/trial-balance/export.csv", h.TrialBalanceCSV)
	reports.Get("/trial-balance/export.xlsx", h.TrialBalanceExcel)
	reports.Get("/trial-balance/export.pdf", h.TrialBalancePDF)
	reports.Get("/income-statement", h.IncomeStatementReport)
	reports.Get("/income-statement/export.pdf", h.IncomeStatementPDF)
	reports.Get("/balance-sheet", h.BalanceSheetReport)
	reports.Get("/cash-flow", h.CashFlowReport)
	reports.Get("/ledger", h.LedgerReportPage)

	periods := auth.Group("/accounting/periods", mw.RequirePermission(model.PermAccountingManage))
	periods.Get("/", h.PeriodList)
	periods.Get("/new", h.PeriodCreatePage)
	periods.Post("/", mw.CSRFProtect(), h.PeriodCreate)
	periods.Post("/:id/close", mw.CSRFProtect(), h.PeriodClose)
}

func (h *Handler) accountingFormData(c fiber.Ctx) *web.AccountingFormData {
	accounts, _ := h.services.Accounting.ListAccounts(c.Context(), "", "")
	categories, _ := h.services.Accounting.ListExpenseCategories(c.Context())
	return &web.AccountingFormData{Accounts: accounts, Categories: categories}
}

func (h *Handler) parseAccountingDates(c fiber.Ctx) (from, to time.Time) {
	to = time.Now()
	from = time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, to.Location())
	if f := c.Query("from"); f != "" {
		if t, err := parseDate(f); err == nil {
			from = t
		}
	}
	if t := c.Query("to"); t != "" {
		if d, err := parseDate(t); err == nil {
			to = d
		}
	}
	return from, to
}

func (h *Handler) AccountingDashboard(c fiber.Ctx) error {
	stats, _ := h.services.Accounting.DashboardStats(c.Context())
	return h.render(c, fiber.StatusOK, web.AccountingDashboardPage{Stats: stats})
}

func (h *Handler) AccountList(c fiber.Ctx) error {
	accounts, _ := h.services.Accounting.ListAccounts(c.Context(), c.Query("type"), c.Query("q"))
	return h.render(c, fiber.StatusOK, web.AccountListPage{Accounts: accounts, FilterType: c.Query("type")})
}

func (h *Handler) AccountCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.AccountFormPage{Title: "Create Account", FormData: h.accountingFormData(c)})
}

func (h *Handler) AccountCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req := parseAccountRequest(c)
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/accounting/accounts/new")
	}
	if _, err := h.services.Accounting.CreateAccount(c.Context(), req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/accounting/accounts/new")
	}
	return c.Redirect().To("/accounting/accounts")
}

func (h *Handler) AccountEditPage(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	acct, err := h.services.Accounting.GetAccount(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.AccountFormPage{Title: "Edit Account", Account: acct, FormData: h.accountingFormData(c)})
}

func (h *Handler) AccountUpdate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	req := parseAccountRequest(c)
	if _, err := h.services.Accounting.UpdateAccount(c.Context(), id, req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/accounting/accounts")
}

func (h *Handler) AccountDisable(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	_ = h.services.Accounting.DisableAccount(c.Context(), id, user.ID, c.IP())
	return c.Redirect().To("/accounting/accounts")
}

func (h *Handler) JournalList(c fiber.Ctx) error {
	from, to := h.parseAccountingDates(c)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	data, _ := h.services.Accounting.ListJournalEntries(c.Context(), dto.AccountingSearchFilter{From: from, To: to, Query: c.Query("q"), Page: page, PerPage: 20})
	return h.render(c, fiber.StatusOK, web.JournalListPage{Data: data, From: from, To: to})
}

func (h *Handler) JournalCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.JournalFormPage{FormData: h.accountingFormData(c)})
}

func (h *Handler) JournalCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	entryDate, err := parseDate(c.FormValue("entry_date"))
	if err != nil {
		h.flash(c, "Invalid date", true)
		return c.Redirect().To("/accounting/journal/new")
	}
	req := dto.JournalEntryRequest{
		EntryDate: entryDate, Description: c.FormValue("description"),
		Lines: parseJournalLines(c),
	}
	if _, err := h.services.Accounting.CreateJournalEntry(c.Context(), req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/accounting/journal/new")
	}
	return c.Redirect().To("/accounting/journal")
}

func (h *Handler) JournalDetail(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	entry, err := h.services.Accounting.GetJournalEntry(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.JournalDetailPage{Entry: entry})
}

func (h *Handler) LedgerPage(c fiber.Ctx) error {
	from, to := h.parseAccountingDates(c)
	accountID, _ := uuid.Parse(c.Query("account_id"))
	var report *dto.LedgerReport
	if accountID != uuid.Nil {
		report, _ = h.services.Accounting.LedgerReport(c.Context(), accountID, from, to)
	}
	return h.render(c, fiber.StatusOK, web.LedgerPage{Report: report, FormData: h.accountingFormData(c), From: from, To: to, AccountID: accountID})
}

func (h *Handler) LedgerExportCSV(c fiber.Ctx) error {
	from, to := h.parseAccountingDates(c)
	accountID, _ := uuid.Parse(c.Query("account_id"))
	report, err := h.services.Accounting.LedgerReport(c.Context(), accountID, from, to)
	if err != nil || report == nil {
		return c.Status(400).SendString("Select account")
	}
	data, _ := export.LedgerCSV(report)
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", `attachment; filename="ledger.csv"`)
	return c.Send(data)
}

func (h *Handler) CashBookPage(c fiber.Ctx) error {
	from, to := h.parseAccountingDates(c)
	entries, _ := h.services.Accounting.CashBook(c.Context(), from, to)
	return h.render(c, fiber.StatusOK, web.CashBookPage{Entries: entries, From: from, To: to})
}

func (h *Handler) BankBookPage(c fiber.Ctx) error {
	from, to := h.parseAccountingDates(c)
	entries, _ := h.services.Accounting.BankBook(c.Context(), from, to)
	return h.render(c, fiber.StatusOK, web.BankBookPage{Entries: entries, From: from, To: to})
}

func (h *Handler) ExpenseList(c fiber.Ctx) error {
	from, to := h.parseAccountingDates(c)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	data, _ := h.services.Accounting.ListExpenses(c.Context(), c.Query("status"), from, to, page, 20)
	return h.render(c, fiber.StatusOK, web.ExpenseListPage{Data: data, Status: c.Query("status"), From: from, To: to})
}

func (h *Handler) ExpenseCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.ExpenseFormPage{FormData: h.accountingFormData(c)})
}

func (h *Handler) ExpenseCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	expDate, _ := parseDate(c.FormValue("expense_date"))
	req := dto.ExpenseRequest{
		CategoryID: parseFormUUID(c, "category_id"), Amount: parseFormFloat(c, "amount"),
		ExpenseDate: expDate, Description: c.FormValue("description"),
		PaymentMethod: c.FormValue("payment_method"), PayFromAccountID: parseFormUUID(c, "pay_from_account_id"),
	}
	if _, err := h.services.Accounting.CreateExpense(c.Context(), req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/expenses/new")
	}
	return c.Redirect().To("/expenses")
}

func (h *Handler) ExpenseApprove(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	if _, err := h.services.Accounting.ApproveExpense(c.Context(), id, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/expenses")
}

func (h *Handler) ExpenseAttachment(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	file, err := c.FormFile("attachment")
	if err != nil {
		return c.Redirect().To("/expenses")
	}
	f, _ := file.Open()
	defer f.Close()
	url, _ := h.storage.Upload(c.Context(), file.Filename, f, file.Header.Get("Content-Type"))
	_ = h.services.Accounting.SetExpenseAttachment(c.Context(), id, url)
	return c.Redirect().To("/expenses")
}

func (h *Handler) IncomeList(c fiber.Ctx) error {
	from, to := h.parseAccountingDates(c)
	entries, _ := h.services.Accounting.ListIncome(c.Context(), from, to)
	return h.render(c, fiber.StatusOK, web.IncomeListPage{Entries: entries, From: from, To: to})
}

func (h *Handler) IncomeCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.IncomeFormPage{FormData: h.accountingFormData(c)})
}

func (h *Handler) IncomeCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	incDate, _ := parseDate(c.FormValue("income_date"))
	req := dto.IncomeEntryRequest{
		IncomeAccountID: parseFormUUID(c, "income_account_id"),
		ReceiveToAccountID: parseFormUUID(c, "receive_to_account_id"),
		Amount: parseFormFloat(c, "amount"), IncomeDate: incDate,
		Source: c.FormValue("source"), Description: c.FormValue("description"),
	}
	if _, err := h.services.Accounting.CreateIncome(c.Context(), req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/income/new")
	}
	return c.Redirect().To("/income")
}

func (h *Handler) TrialBalanceReport(c fiber.Ctx) error {
	asOf := time.Now()
	if d := c.Query("as_of"); d != "" {
		if t, err := parseDate(d); err == nil {
			asOf = t
		}
	}
	report, _ := h.services.Accounting.TrialBalance(c.Context(), asOf)
	return h.render(c, fiber.StatusOK, web.TrialBalancePage{Report: report, AsOf: asOf})
}

func (h *Handler) TrialBalanceCSV(c fiber.Ctx) error {
	asOf := time.Now()
	if d := c.Query("as_of"); d != "" {
		if t, err := parseDate(d); err == nil {
			asOf = t
		}
	}
	report, _ := h.services.Accounting.TrialBalance(c.Context(), asOf)
	data, _ := export.TrialBalanceCSV(report)
	c.Set("Content-Type", "text/csv")
	return c.Send(data)
}

func (h *Handler) TrialBalanceExcel(c fiber.Ctx) error {
	asOf := time.Now()
	report, _ := h.services.Accounting.TrialBalance(c.Context(), asOf)
	data, _ := export.TrialBalanceExcel(report)
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	return c.Send(data)
}

func (h *Handler) TrialBalancePDF(c fiber.Ctx) error {
	asOf := time.Now()
	report, _ := h.services.Accounting.TrialBalance(c.Context(), asOf)
	data, _ := pdf.GenerateTrialBalance(report)
	c.Set("Content-Type", "application/pdf")
	return c.Send(data)
}

func (h *Handler) IncomeStatementReport(c fiber.Ctx) error {
	from, to := h.parseAccountingDates(c)
	report, _ := h.services.Accounting.IncomeStatement(c.Context(), from, to)
	return h.render(c, fiber.StatusOK, web.IncomeStatementPage{Report: report, From: from, To: to})
}

func (h *Handler) IncomeStatementPDF(c fiber.Ctx) error {
	from, to := h.parseAccountingDates(c)
	report, _ := h.services.Accounting.IncomeStatement(c.Context(), from, to)
	data, _ := pdf.GenerateIncomeStatement(report)
	c.Set("Content-Type", "application/pdf")
	return c.Send(data)
}

func (h *Handler) BalanceSheetReport(c fiber.Ctx) error {
	asOf := time.Now()
	if d := c.Query("as_of"); d != "" {
		if t, err := parseDate(d); err == nil {
			asOf = t
		}
	}
	report, _ := h.services.Accounting.BalanceSheet(c.Context(), asOf)
	return h.render(c, fiber.StatusOK, web.BalanceSheetPage{Report: report, AsOf: asOf})
}

func (h *Handler) CashFlowReport(c fiber.Ctx) error {
	from, to := h.parseAccountingDates(c)
	report, _ := h.services.Accounting.CashFlow(c.Context(), from, to)
	return h.render(c, fiber.StatusOK, web.CashFlowPage{Report: report, From: from, To: to})
}

func (h *Handler) LedgerReportPage(c fiber.Ctx) error {
	return h.LedgerPage(c)
}

func (h *Handler) PeriodList(c fiber.Ctx) error {
	periods, _ := h.services.Accounting.ListPeriods(c.Context())
	return h.render(c, fiber.StatusOK, web.PeriodListPage{Periods: periods})
}

func (h *Handler) PeriodCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.PeriodFormPage{})
}

func (h *Handler) PeriodCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	start, _ := parseDate(c.FormValue("start_date"))
	end, _ := parseDate(c.FormValue("end_date"))
	req := dto.FinancialPeriodRequest{Name: c.FormValue("name"), StartDate: start, EndDate: end}
	if _, err := h.services.Accounting.CreatePeriod(c.Context(), req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/accounting/periods")
}

func (h *Handler) PeriodClose(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	id, _ := parseUUIDParam(c, "id")
	if err := h.services.Accounting.ClosePeriod(c.Context(), id, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/accounting/periods")
}

func parseAccountRequest(c fiber.Ctx) dto.AccountRequest {
	return dto.AccountRequest{
		Code: c.FormValue("code"), Name: c.FormValue("name"),
		AccountType: c.FormValue("account_type"), ParentID: parseFormUUID(c, "parent_id"),
		Description: c.FormValue("description"),
	}
}

func parseJournalLines(c fiber.Ctx) []dto.JournalLineRequest {
	var lines []dto.JournalLineRequest
	for i := 0; i < 20; i++ {
		prefix := fmt.Sprintf("line_%d_", i)
		acctStr := c.FormValue(prefix + "account_id")
		if acctStr == "" {
			continue
		}
		acctID, err := uuid.Parse(acctStr)
		if err != nil {
			continue
		}
		debit, _ := strconv.ParseFloat(c.FormValue(prefix+"debit"), 64)
		credit, _ := strconv.ParseFloat(c.FormValue(prefix+"credit"), 64)
		if debit == 0 && credit == 0 {
			continue
		}
		lines = append(lines, dto.JournalLineRequest{
			AccountID: acctID, Debit: debit, Credit: credit,
			Description: c.FormValue(prefix + "description"),
		})
	}
	return lines
}
