package dto

import (
	"time"

	"github.com/google/uuid"
)

type AccountRequest struct {
	Code        string    `form:"code" validate:"required,min=2,max=20"`
	Name        string    `form:"name" validate:"required,min=2,max=150"`
	AccountType string    `form:"account_type" validate:"required,oneof=assets liabilities equity income expenses"`
	ParentID    uuid.UUID `form:"parent_id" validate:"omitempty"`
	Description string    `form:"description" validate:"omitempty,max=500"`
}

type AccountResponse struct {
	ID          uuid.UUID  `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	AccountType string     `json:"account_type"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	ParentName  string     `json:"parent_name,omitempty"`
	Description string     `json:"description,omitempty"`
	IsSystem    bool       `json:"is_system"`
	IsActive    bool       `json:"is_active"`
	Balance     float64    `json:"balance,omitempty"`
	Children    []AccountResponse `json:"children,omitempty"`
}

type JournalLineRequest struct {
	AccountID   uuid.UUID `form:"account_id" validate:"required"`
	Debit       float64   `form:"debit" validate:"omitempty,gte=0"`
	Credit      float64   `form:"credit" validate:"omitempty,gte=0"`
	Description string    `form:"description" validate:"omitempty,max=200"`
}

type JournalEntryRequest struct {
	EntryDate   time.Time `form:"entry_date" validate:"required"`
	Description string    `form:"description" validate:"required,min=3,max=500"`
	Lines       []JournalLineRequest
}

type JournalLineResponse struct {
	ID          uuid.UUID `json:"id"`
	AccountID   uuid.UUID `json:"account_id"`
	AccountCode string    `json:"account_code"`
	AccountName string    `json:"account_name"`
	Debit       float64   `json:"debit"`
	Credit      float64   `json:"credit"`
	Description string    `json:"description,omitempty"`
}

type JournalEntryResponse struct {
	ID          uuid.UUID             `json:"id"`
	EntryNumber string                `json:"entry_number"`
	EntryDate   time.Time             `json:"entry_date"`
	Description string                `json:"description"`
	SourceType  string                `json:"source_type"`
	Status      string                `json:"status"`
	TotalDebit  float64               `json:"total_debit"`
	TotalCredit float64               `json:"total_credit"`
	Lines       []JournalLineResponse `json:"lines"`
	CreatedAt   time.Time             `json:"created_at"`
}

type LedgerEntry struct {
	EntryDate   time.Time
	EntryNumber string
	Description string
	Debit       float64
	Credit      float64
	Balance     float64
	SourceType  string
}

type LedgerReport struct {
	Account     AccountResponse
	Entries     []LedgerEntry
	OpenBalance float64
	CloseBalance float64
}

type CashBookEntry struct {
	EntryDate   time.Time
	EntryNumber string
	Description string
	CashIn      float64
	CashOut     float64
	Balance     float64
}

type BankBookEntry struct {
	EntryDate   time.Time
	EntryNumber string
	Description string
	Deposit     float64
	Withdrawal  float64
	Balance     float64
}

type ExpenseRequest struct {
	CategoryID      uuid.UUID `form:"category_id" validate:"required"`
	Amount          float64   `form:"amount" validate:"required,gt=0"`
	ExpenseDate     time.Time `form:"expense_date" validate:"required"`
	Description     string    `form:"description" validate:"required,min=3,max=500"`
	PaymentMethod   string    `form:"payment_method" validate:"required,oneof=cash bank"`
	PayFromAccountID uuid.UUID `form:"pay_from_account_id" validate:"required"`
}

type ExpenseResponse struct {
	ID               uuid.UUID  `json:"id"`
	CategoryID       uuid.UUID  `json:"category_id"`
	CategoryName     string     `json:"category_name"`
	Amount           float64    `json:"amount"`
	ExpenseDate      time.Time  `json:"expense_date"`
	Description      string     `json:"description"`
	PaymentMethod    string     `json:"payment_method"`
	Status           string     `json:"status"`
	AttachmentURL    string     `json:"attachment_url,omitempty"`
	CreatedByName    string     `json:"created_by_name,omitempty"`
	ApprovedByName   string     `json:"approved_by_name,omitempty"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`
}

type ExpenseCategoryResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Slug     string    `json:"slug"`
	IsActive bool      `json:"is_active"`
}

type IncomeEntryRequest struct {
	IncomeAccountID    uuid.UUID `form:"income_account_id" validate:"required"`
	ReceiveToAccountID uuid.UUID `form:"receive_to_account_id" validate:"required"`
	Amount             float64   `form:"amount" validate:"required,gt=0"`
	IncomeDate         time.Time `form:"income_date" validate:"required"`
	Source             string    `form:"source" validate:"required,oneof=donation event admission_form misc"`
	Description        string    `form:"description" validate:"required,min=3,max=500"`
}

type IncomeEntryResponse struct {
	ID                 uuid.UUID `json:"id"`
	IncomeAccountID    uuid.UUID `json:"income_account_id"`
	IncomeAccountName  string    `json:"income_account_name"`
	ReceiveToAccountID uuid.UUID `json:"receive_to_account_id"`
	ReceiveAccountName string    `json:"receive_account_name"`
	Amount             float64   `json:"amount"`
	IncomeDate         time.Time `json:"income_date"`
	Source             string    `json:"source"`
	Description        string    `json:"description"`
}

type TrialBalanceRow struct {
	AccountCode string
	AccountName string
	AccountType string
	Debit       float64
	Credit      float64
}

type TrialBalanceReport struct {
	Rows        []TrialBalanceRow
	TotalDebit  float64
	TotalCredit float64
	AsOf        time.Time
}

type IncomeStatementReport struct {
	IncomeItems  []TrialBalanceRow
	ExpenseItems []TrialBalanceRow
	TotalIncome  float64
	TotalExpense float64
	NetProfit    float64
	From, To     time.Time
}

type BalanceSheetReport struct {
	Assets      []TrialBalanceRow
	Liabilities []TrialBalanceRow
	Equity      []TrialBalanceRow
	TotalAssets float64
	TotalLiab   float64
	TotalEquity float64
	AsOf        time.Time
}

type CashFlowReport struct {
	OperatingIn  float64
	OperatingOut float64
	NetCashFlow  float64
	From, To     time.Time
}

type FinancialPeriodRequest struct {
	Name      string    `form:"name" validate:"required,min=2,max=100"`
	StartDate time.Time `form:"start_date" validate:"required"`
	EndDate   time.Time `form:"end_date" validate:"required"`
}

type FinancialPeriodResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	StartDate time.Time  `json:"start_date"`
	EndDate   time.Time  `json:"end_date"`
	Status    string     `json:"status"`
	IsLocked  bool       `json:"is_locked"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
}

type AccountingDashboardStats struct {
	CashBalance    float64
	BankBalance    float64
	MonthlyIncome  float64
	MonthlyExpense float64
	NetProfit      float64
	IncomeTrend    []FinanceTrendPoint
	ExpenseTrend   []FinanceTrendPoint
}

type AccountingSearchFilter struct {
	AccountType string
	Query       string
	From, To    time.Time
	AccountID   uuid.UUID
	Page        int
	PerPage     int
}

type PaginatedJournalEntries struct {
	Items      []JournalEntryResponse
	Total      int64
	Page       int
	PerPage    int
	TotalPages int
}

type PaginatedExpenses struct {
	Items      []ExpenseResponse
	Total      int64
	Page       int
	PerPage    int
	TotalPages int
}
