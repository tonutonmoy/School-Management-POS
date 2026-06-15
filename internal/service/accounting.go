package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/repository"
)

type AccountingService struct {
	repos *repository.Repositories
	audit *AuditService
}

func NewAccountingService(repos *repository.Repositories, audit *AuditService) *AccountingService {
	return &AccountingService{repos: repos, audit: audit}
}

// --- Accounts ---

func (s *AccountingService) CreateAccount(ctx context.Context, req dto.AccountRequest, actorID uuid.UUID, ip string) (*dto.AccountResponse, error) {
	var parentID *uuid.UUID
	if req.ParentID != uuid.Nil {
		parentID = &req.ParentID
	}
	rec, err := s.repos.Accounting.CreateAccount(ctx, repository.CreateAccountParams{
		Code: req.Code, Name: req.Name, AccountType: req.AccountType,
		ParentID: parentID, Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	resp := mapAccount(rec, 0)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityAccount, &rec.ID, ip, nil)
	return &resp, nil
}

func (s *AccountingService) UpdateAccount(ctx context.Context, id uuid.UUID, req dto.AccountRequest, actorID uuid.UUID, ip string) (*dto.AccountResponse, error) {
	var parentID *uuid.UUID
	if req.ParentID != uuid.Nil {
		parentID = &req.ParentID
	}
	rec, err := s.repos.Accounting.UpdateAccount(ctx, id, repository.CreateAccountParams{
		Code: req.Code, Name: req.Name, AccountType: req.AccountType,
		ParentID: parentID, Description: req.Description,
	})
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	bal, _ := s.repos.Accounting.AccountBalance(ctx, id, time.Now())
	resp := mapAccount(rec, bal)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityAccount, &id, ip, nil)
	return &resp, nil
}

func (s *AccountingService) DisableAccount(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.Accounting.DisableAccount(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityAccount, &id, ip, map[string]any{"disabled": true})
	return nil
}

func (s *AccountingService) ListAccounts(ctx context.Context, accountType, query string) ([]dto.AccountResponse, error) {
	recs, err := s.repos.Accounting.ListAccounts(ctx, accountType, query)
	if err != nil {
		return nil, err
	}
	items := make([]dto.AccountResponse, 0, len(recs))
	for _, r := range recs {
		bal, _ := s.repos.Accounting.AccountBalance(ctx, r.ID, time.Now())
		items = append(items, mapAccount(&r, bal))
	}
	return items, nil
}

func (s *AccountingService) GetAccount(ctx context.Context, id uuid.UUID) (*dto.AccountResponse, error) {
	rec, err := s.repos.Accounting.GetAccount(ctx, id)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	bal, _ := s.repos.Accounting.AccountBalance(ctx, id, time.Now())
	resp := mapAccount(rec, bal)
	return &resp, nil
}

// --- Journal Entries ---

func (s *AccountingService) CreateJournalEntry(ctx context.Context, req dto.JournalEntryRequest, actorID uuid.UUID, ip string) (*dto.JournalEntryResponse, error) {
	if locked, _ := s.repos.Accounting.IsPeriodLocked(ctx, req.EntryDate); locked {
		return nil, fmt.Errorf("%w: period is locked", ErrValidation)
	}
	var totalDebit, totalCredit float64
	for _, l := range req.Lines {
		totalDebit += l.Debit
		totalCredit += l.Credit
	}
	if math.Abs(totalDebit-totalCredit) > 0.01 || totalDebit == 0 {
		return nil, fmt.Errorf("%w: debits must equal credits", ErrValidation)
	}
	var rec *repository.JournalEntryRecord
	err := s.repos.Accounting.WithTx(ctx, func(tx pgx.Tx) error {
		year := req.EntryDate.Year()
		entryNum, err := s.repos.Accounting.NextJournalNumber(ctx, tx, year)
		if err != nil {
			return err
		}
		period, _ := s.repos.Accounting.GetOpenPeriod(ctx, req.EntryDate)
		var periodID *uuid.UUID
		if period != nil {
			periodID = &period.ID
		}
		lines := make([]repository.JournalLineParams, 0, len(req.Lines))
		for i, l := range req.Lines {
			lines = append(lines, repository.JournalLineParams{
				AccountID: l.AccountID, Debit: l.Debit, Credit: l.Credit,
				Description: l.Description, LineOrder: i + 1,
			})
		}
		rec, err = s.repos.Accounting.CreateJournalEntry(ctx, tx, repository.CreateJournalParams{
			EntryNumber: entryNum, EntryDate: req.EntryDate, Description: req.Description,
			SourceType: model.JournalSourceManual, Status: model.JournalStatusPosted,
			PeriodID: periodID, CreatedBy: &actorID, Lines: lines,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	resp := mapJournal(rec, nil)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityJournalEntry, &rec.ID, ip, nil)
	return &resp, nil
}

func (s *AccountingService) ListJournalEntries(ctx context.Context, f dto.AccountingSearchFilter) (*dto.PaginatedJournalEntries, error) {
	page, perPage := f.Page, f.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	params := repository.JournalSearchParams{
		From: f.From, To: f.To, Query: f.Query,
		Limit: int32(perPage), Offset: int32((page - 1) * perPage),
	}
	total, _ := s.repos.Accounting.CountJournalEntries(ctx, params)
	recs, err := s.repos.Accounting.ListJournalEntries(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.JournalEntryResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapJournal(&r, nil))
	}
	return &dto.PaginatedJournalEntries{
		Items: items, Total: total, Page: page, PerPage: perPage,
		TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
	}, nil
}

func (s *AccountingService) GetJournalEntry(ctx context.Context, id uuid.UUID) (*dto.JournalEntryResponse, error) {
	rec, err := s.repos.Accounting.GetJournalEntry(ctx, id)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	lines, _ := s.repos.Accounting.ListJournalLines(ctx, id)
	resp := mapJournal(rec, lines)
	return &resp, nil
}

// --- Ledger ---

func (s *AccountingService) LedgerReport(ctx context.Context, accountID uuid.UUID, from, to time.Time) (*dto.LedgerReport, error) {
	acct, err := s.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	openBal, _ := s.repos.Accounting.AccountBalance(ctx, accountID, from.Add(-24*time.Hour))
	lines, err := s.repos.Accounting.ListLedgerEntries(ctx, accountID, from, to)
	if err != nil {
		return nil, err
	}
	rec, _ := s.repos.Accounting.GetAccount(ctx, accountID)
	acctType := ""
	if rec != nil {
		acctType = rec.AccountType
	}
	running := openBal
	entries := make([]dto.LedgerEntry, 0, len(lines))
	for _, l := range lines {
		if acctType == "" {
			acctType = l.AccountType
		}
		running += balanceDelta(acctType, l.Debit, l.Credit)
		entries = append(entries, dto.LedgerEntry{
			EntryDate: l.EntryDate, EntryNumber: l.EntryNumber, Description: l.Description,
			Debit: l.Debit, Credit: l.Credit, Balance: running, SourceType: l.SourceType,
		})
	}
	return &dto.LedgerReport{Account: *acct, Entries: entries, OpenBalance: openBal, CloseBalance: running}, nil
}

func balanceDelta(acctType string, debit, credit float64) float64 {
	switch acctType {
	case model.AccountTypeAssets, model.AccountTypeExpenses:
		return debit - credit
	default:
		return credit - debit
	}
}

func (s *AccountingService) CashBook(ctx context.Context, from, to time.Time) ([]dto.CashBookEntry, error) {
	return s.bookEntries(ctx, from, to, true)
}

func (s *AccountingService) BankBook(ctx context.Context, from, to time.Time) ([]dto.BankBookEntry, error) {
	lines, err := s.repos.Accounting.ListBankBook(ctx, from, to)
	if err != nil {
		return nil, err
	}
	cash, _ := s.repos.Accounting.GetAccountByCode(ctx, model.AccountCodeBank)
	openBal := 0.0
	if cash != nil {
		openBal, _ = s.repos.Accounting.AccountBalance(ctx, cash.ID, from.Add(-24*time.Hour))
	}
	running := openBal
	items := make([]dto.BankBookEntry, 0, len(lines))
	for _, l := range lines {
		running += l.Debit - l.Credit
		items = append(items, dto.BankBookEntry{
			EntryDate: l.EntryDate, EntryNumber: l.EntryNumber, Description: l.Description,
			Deposit: l.Debit, Withdrawal: l.Credit, Balance: running,
		})
	}
	return items, nil
}

func (s *AccountingService) bookEntries(ctx context.Context, from, to time.Time, cash bool) ([]dto.CashBookEntry, error) {
	var lines []repository.LedgerLineRecord
	var err error
	if cash {
		lines, err = s.repos.Accounting.ListCashBook(ctx, from, to)
	} else {
		lines, err = s.repos.Accounting.ListBankBook(ctx, from, to)
	}
	if err != nil {
		return nil, err
	}
	code := model.AccountCodeCash
	if !cash {
		code = model.AccountCodeBank
	}
	acct, _ := s.repos.Accounting.GetAccountByCode(ctx, code)
	openBal := 0.0
	if acct != nil {
		openBal, _ = s.repos.Accounting.AccountBalance(ctx, acct.ID, from.Add(-24*time.Hour))
	}
	running := openBal
	items := make([]dto.CashBookEntry, 0, len(lines))
	for _, l := range lines {
		running += l.Debit - l.Credit
		items = append(items, dto.CashBookEntry{
			EntryDate: l.EntryDate, EntryNumber: l.EntryNumber, Description: l.Description,
			CashIn: l.Debit, CashOut: l.Credit, Balance: running,
		})
	}
	return items, nil
}

// --- Expenses ---

func (s *AccountingService) CreateExpense(ctx context.Context, req dto.ExpenseRequest, actorID uuid.UUID, ip string) (*dto.ExpenseResponse, error) {
	cat, cats := s.getCategory(ctx, req.CategoryID)
	if cat == nil {
		return nil, ErrNotFound
	}
	_ = cats
	expAcct := cat.AccountID
	if expAcct == uuid.Nil {
		def, _ := s.repos.Accounting.GetAccountByCode(ctx, model.AccountCodeMiscExpense)
		if def != nil {
			expAcct = def.ID
		}
	}
	rec, err := s.repos.Accounting.CreateExpense(ctx, repository.CreateExpenseParams{
		CategoryID: req.CategoryID, ExpenseAccountID: expAcct,
		PayFromAccountID: req.PayFromAccountID, Amount: req.Amount,
		ExpenseDate: req.ExpenseDate, Description: req.Description,
		PaymentMethod: req.PaymentMethod, Status: model.ExpenseStatusPendingApproval,
		CreatedBy: actorID,
	})
	if err != nil {
		return nil, err
	}
	resp := mapExpense(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityExpense, &rec.ID, ip, nil)
	return &resp, nil
}

func (s *AccountingService) getCategory(ctx context.Context, id uuid.UUID) (*repository.ExpenseCategoryRecord, []repository.ExpenseCategoryRecord) {
	cats, _ := s.repos.Accounting.ListExpenseCategories(ctx)
	for _, c := range cats {
		if c.ID == id {
			return &c, cats
		}
	}
	return nil, cats
}

func (s *AccountingService) ApproveExpense(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) (*dto.ExpenseResponse, error) {
	exp, err := s.repos.Accounting.GetExpense(ctx, id)
	if err != nil || exp == nil {
		return nil, ErrNotFound
	}
	if exp.Status != model.ExpenseStatusPendingApproval {
		return nil, fmt.Errorf("%w: expense not pending approval", ErrValidation)
	}
	if locked, _ := s.repos.Accounting.IsPeriodLocked(ctx, exp.ExpenseDate); locked {
		return nil, fmt.Errorf("%w: period is locked", ErrValidation)
	}
	var journalID uuid.UUID
	err = s.repos.Accounting.WithTx(ctx, func(tx pgx.Tx) error {
		if err := s.repos.Accounting.ApproveExpenseTx(ctx, tx, id, actorID); err != nil {
			return err
		}
		year := exp.ExpenseDate.Year()
		entryNum, err := s.repos.Accounting.NextJournalNumber(ctx, tx, year)
		if err != nil {
			return err
		}
		srcID := id
		je, err := s.repos.Accounting.CreateJournalEntry(ctx, tx, repository.CreateJournalParams{
			EntryNumber: entryNum, EntryDate: exp.ExpenseDate,
			Description: fmt.Sprintf("Expense: %s", exp.Description),
			SourceType: model.JournalSourceExpense, SourceID: &srcID,
			Status: model.JournalStatusPosted, CreatedBy: &actorID,
			Lines: []repository.JournalLineParams{
				{AccountID: exp.ExpenseAccountID, Debit: exp.Amount, Description: exp.Description},
				{AccountID: exp.PayFromAccountID, Credit: exp.Amount, Description: "Payment"},
			},
		})
		if err != nil {
			return err
		}
		journalID = je.ID
		return s.repos.Accounting.SetExpenseJournalTx(ctx, tx, id, journalID)
	})
	if err != nil {
		return nil, err
	}
	rec, _ := s.repos.Accounting.GetExpense(ctx, id)
	resp := mapExpense(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityExpense, &id, ip, map[string]any{"approved": true, "journal_id": journalID})
	return &resp, nil
}

func (s *AccountingService) ListExpenses(ctx context.Context, status string, from, to time.Time, page, perPage int) (*dto.PaginatedExpenses, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	params := repository.ExpenseSearchParams{Status: status, From: from, To: to, Limit: int32(perPage), Offset: int32((page - 1) * perPage)}
	total, _ := s.repos.Accounting.CountExpenses(ctx, params)
	recs, err := s.repos.Accounting.ListExpenses(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.ExpenseResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapExpense(&r))
	}
	return &dto.PaginatedExpenses{
		Items: items, Total: total, Page: page, PerPage: perPage,
		TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
	}, nil
}

func (s *AccountingService) ListExpenseCategories(ctx context.Context) ([]dto.ExpenseCategoryResponse, error) {
	recs, err := s.repos.Accounting.ListExpenseCategories(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.ExpenseCategoryResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, dto.ExpenseCategoryResponse{ID: r.ID, Name: r.Name, Slug: r.Slug, IsActive: r.IsActive})
	}
	return items, nil
}

// --- Income ---

func (s *AccountingService) CreateIncome(ctx context.Context, req dto.IncomeEntryRequest, actorID uuid.UUID, ip string) (*dto.IncomeEntryResponse, error) {
	if locked, _ := s.repos.Accounting.IsPeriodLocked(ctx, req.IncomeDate); locked {
		return nil, fmt.Errorf("%w: period is locked", ErrValidation)
	}
	var incomeID uuid.UUID
	err := s.repos.Accounting.WithTx(ctx, func(tx pgx.Tx) error {
		var err error
		incomeID, err = s.repos.Accounting.CreateIncomeEntryTx(ctx, tx, repository.CreateIncomeParams{
			IncomeAccountID: req.IncomeAccountID, ReceiveToAccountID: req.ReceiveToAccountID,
			Amount: req.Amount, IncomeDate: req.IncomeDate, Source: req.Source,
			Description: req.Description, CreatedBy: actorID,
		})
		if err != nil {
			return err
		}
		year := req.IncomeDate.Year()
		entryNum, err := s.repos.Accounting.NextJournalNumber(ctx, tx, year)
		if err != nil {
			return err
		}
		srcID := incomeID
		je, err := s.repos.Accounting.CreateJournalEntry(ctx, tx, repository.CreateJournalParams{
			EntryNumber: entryNum, EntryDate: req.IncomeDate, Description: req.Description,
			SourceType: model.JournalSourceIncome, SourceID: &srcID,
			Status: model.JournalStatusPosted, CreatedBy: &actorID,
			Lines: []repository.JournalLineParams{
				{AccountID: req.ReceiveToAccountID, Debit: req.Amount, Description: req.Description},
				{AccountID: req.IncomeAccountID, Credit: req.Amount, Description: req.Description},
			},
		})
		if err != nil {
			return err
		}
		return s.repos.Accounting.SetIncomeJournal(ctx, incomeID, je.ID)
	})
	if err != nil {
		return nil, err
	}
	rec, err := s.repos.Accounting.GetIncomeEntry(ctx, incomeID)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	resp := mapIncome(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityIncomeEntry, &incomeID, ip, nil)
	return &resp, nil
}

func (s *AccountingService) ListIncome(ctx context.Context, from, to time.Time) ([]dto.IncomeEntryResponse, error) {
	recs, err := s.repos.Accounting.ListIncomeEntries(ctx, from, to)
	if err != nil {
		return nil, err
	}
	items := make([]dto.IncomeEntryResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapIncome(&r))
	}
	return items, nil
}

// --- Reports ---

func (s *AccountingService) TrialBalance(ctx context.Context, asOf time.Time) (*dto.TrialBalanceReport, error) {
	rows, err := s.repos.Accounting.TrialBalance(ctx, asOf)
	if err != nil {
		return nil, err
	}
	report := &dto.TrialBalanceReport{AsOf: asOf}
	for _, r := range rows {
		report.Rows = append(report.Rows, dto.TrialBalanceRow{
			AccountCode: r.AccountCode, AccountName: r.AccountName, AccountType: r.AccountType,
			Debit: r.Debit, Credit: r.Credit,
		})
		report.TotalDebit += r.Debit
		report.TotalCredit += r.Credit
	}
	return report, nil
}

func (s *AccountingService) IncomeStatement(ctx context.Context, from, to time.Time) (*dto.IncomeStatementReport, error) {
	income, _ := s.repos.Accounting.AccountActivity(ctx, model.AccountTypeIncome, from, to)
	expense, _ := s.repos.Accounting.AccountActivity(ctx, model.AccountTypeExpenses, from, to)
	report := &dto.IncomeStatementReport{From: from, To: to}
	for _, r := range income {
		amt := r.Credit - r.Debit
		report.IncomeItems = append(report.IncomeItems, dto.TrialBalanceRow{
			AccountCode: r.AccountCode, AccountName: r.AccountName, Credit: amt,
		})
		report.TotalIncome += amt
	}
	for _, r := range expense {
		amt := r.Debit - r.Credit
		report.ExpenseItems = append(report.ExpenseItems, dto.TrialBalanceRow{
			AccountCode: r.AccountCode, AccountName: r.AccountName, Debit: amt,
		})
		report.TotalExpense += amt
	}
	report.NetProfit = report.TotalIncome - report.TotalExpense
	return report, nil
}

func (s *AccountingService) BalanceSheet(ctx context.Context, asOf time.Time) (*dto.BalanceSheetReport, error) {
	rows, _ := s.repos.Accounting.TrialBalance(ctx, asOf)
	report := &dto.BalanceSheetReport{AsOf: asOf}
	for _, r := range rows {
		row := dto.TrialBalanceRow{AccountCode: r.AccountCode, AccountName: r.AccountName, AccountType: r.AccountType, Debit: r.Debit, Credit: r.Credit}
		switch r.AccountType {
		case model.AccountTypeAssets:
			report.Assets = append(report.Assets, row)
			report.TotalAssets += r.Balance
		case model.AccountTypeLiabilities:
			report.Liabilities = append(report.Liabilities, row)
			report.TotalLiab += r.Balance
		case model.AccountTypeEquity:
			report.Equity = append(report.Equity, row)
			report.TotalEquity += r.Balance
		case model.AccountTypeIncome:
			report.Equity = append(report.Equity, row)
			report.TotalEquity += r.Balance
		}
	}
	return report, nil
}

func (s *AccountingService) CashFlow(ctx context.Context, from, to time.Time) (*dto.CashFlowReport, error) {
	dIn, cOut, _ := s.repos.Accounting.SumAccountDebitsCredits(ctx, model.AccountCodeCash, from, to)
	bIn, bOut, _ := s.repos.Accounting.SumAccountDebitsCredits(ctx, model.AccountCodeBank, from, to)
	report := &dto.CashFlowReport{From: from, To: to}
	report.OperatingIn = dIn + bIn
	report.OperatingOut = cOut + bOut
	report.NetCashFlow = report.OperatingIn - report.OperatingOut
	return report, nil
}

// --- Periods ---

func (s *AccountingService) ListPeriods(ctx context.Context) ([]dto.FinancialPeriodResponse, error) {
	recs, err := s.repos.Accounting.ListFinancialPeriods(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.FinancialPeriodResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapPeriod(&r))
	}
	return items, nil
}

func (s *AccountingService) ClosePeriod(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.Accounting.CloseFinancialPeriod(ctx, id, actorID); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityFinancialPeriod, &id, ip, map[string]any{"closed": true})
	return nil
}

func (s *AccountingService) CreatePeriod(ctx context.Context, req dto.FinancialPeriodRequest, actorID uuid.UUID, ip string) (*dto.FinancialPeriodResponse, error) {
	rec, err := s.repos.Accounting.CreateFinancialPeriod(ctx, req.Name, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	resp := mapPeriod(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityFinancialPeriod, &rec.ID, ip, nil)
	return &resp, nil
}

// --- Dashboard ---

func (s *AccountingService) DashboardStats(ctx context.Context) (*dto.AccountingDashboardStats, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	cashAcct, _ := s.repos.Accounting.GetAccountByCode(ctx, model.AccountCodeCash)
	bankAcct, _ := s.repos.Accounting.GetAccountByCode(ctx, model.AccountCodeBank)
	stats := &dto.AccountingDashboardStats{}
	if cashAcct != nil {
		stats.CashBalance, _ = s.repos.Accounting.AccountBalance(ctx, cashAcct.ID, now)
	}
	if bankAcct != nil {
		stats.BankBalance, _ = s.repos.Accounting.AccountBalance(ctx, bankAcct.ID, now)
	}
	incRows, _ := s.repos.Accounting.AccountActivity(ctx, model.AccountTypeIncome, monthStart, now)
	for _, r := range incRows {
		stats.MonthlyIncome += r.Credit - r.Debit
	}
	expRows, _ := s.repos.Accounting.AccountActivity(ctx, model.AccountTypeExpenses, monthStart, now)
	for _, r := range expRows {
		stats.MonthlyExpense += r.Debit - r.Credit
	}
	stats.NetProfit = stats.MonthlyIncome - stats.MonthlyExpense
	incTrend, _ := s.repos.Accounting.MonthlyTrend(ctx, model.AccountTypeIncome, 6)
	for _, t := range incTrend {
		stats.IncomeTrend = append(stats.IncomeTrend, dto.FinanceTrendPoint{Label: t.Label, Amount: t.Amount})
	}
	expTrend, _ := s.repos.Accounting.MonthlyTrend(ctx, model.AccountTypeExpenses, 6)
	for _, t := range expTrend {
		stats.ExpenseTrend = append(stats.ExpenseTrend, dto.FinanceTrendPoint{Label: t.Label, Amount: t.Amount})
	}
	return stats, nil
}

func (s *AccountingService) SetExpenseAttachment(ctx context.Context, id uuid.UUID, url string) error {
	return s.repos.Accounting.SetExpenseAttachment(ctx, id, url)
}

// --- Mappers ---

func mapAccount(r *repository.AccountRecord, bal float64) dto.AccountResponse {
	resp := dto.AccountResponse{
		ID: r.ID, Code: r.Code, Name: r.Name, AccountType: r.AccountType,
		Description: r.Description, IsSystem: r.IsSystem, IsActive: r.IsActive, Balance: bal,
	}
	if r.ParentID != uuid.Nil {
		resp.ParentID = &r.ParentID
		resp.ParentName = r.ParentName
	}
	return resp
}

func mapJournal(r *repository.JournalEntryRecord, lines []repository.JournalLineRecord) dto.JournalEntryResponse {
	resp := dto.JournalEntryResponse{
		ID: r.ID, EntryNumber: r.EntryNumber, EntryDate: r.EntryDate, Description: r.Description,
		SourceType: r.SourceType, Status: r.Status, TotalDebit: r.TotalDebit, TotalCredit: r.TotalCredit,
		CreatedAt: r.CreatedAt,
	}
	for _, l := range lines {
		resp.Lines = append(resp.Lines, dto.JournalLineResponse{
			ID: l.ID, AccountID: l.AccountID, AccountCode: l.AccountCode, AccountName: l.AccountName,
			Debit: l.Debit, Credit: l.Credit, Description: l.Description,
		})
	}
	return resp
}

func mapExpense(r *repository.ExpenseRecord) dto.ExpenseResponse {
	return dto.ExpenseResponse{
		ID: r.ID, CategoryID: r.CategoryID, CategoryName: r.CategoryName, Amount: r.Amount,
		ExpenseDate: r.ExpenseDate, Description: r.Description, PaymentMethod: r.PaymentMethod,
		Status: r.Status, AttachmentURL: r.AttachmentURL, CreatedByName: r.CreatedByName,
		ApprovedByName: r.ApprovedByName, ApprovedAt: r.ApprovedAt,
	}
}

func mapIncome(r *repository.IncomeRecord) dto.IncomeEntryResponse {
	return dto.IncomeEntryResponse{
		ID: r.ID, IncomeAccountID: r.IncomeAccountID, IncomeAccountName: r.IncomeAccountName,
		ReceiveToAccountID: r.ReceiveToAccountID, ReceiveAccountName: r.ReceiveAccountName,
		Amount: r.Amount, IncomeDate: r.IncomeDate, Source: r.Source, Description: r.Description,
	}
}

func mapPeriod(r *repository.FinancialPeriodRecord) dto.FinancialPeriodResponse {
	return dto.FinancialPeriodResponse{
		ID: r.ID, Name: r.Name, StartDate: r.StartDate, EndDate: r.EndDate,
		Status: r.Status, IsLocked: r.IsLocked, ClosedAt: r.ClosedAt,
	}
}
