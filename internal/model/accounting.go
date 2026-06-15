package model

const (
	EntityAccount         = "account"
	EntityJournalEntry    = "journal_entry"
	EntityExpense         = "expense"
	EntityIncomeEntry     = "income_entry"
	EntityFinancialPeriod = "financial_period"
)

const (
	PermAccountingView   = "accounting.view"
	PermAccountingManage = "accounting.manage"
	PermExpenseCreate    = "expense.create"
	PermExpenseApprove   = "expense.approve"
	PermFinanceReportView = "finance.report.view"
)

const (
	AccountTypeAssets      = "assets"
	AccountTypeLiabilities = "liabilities"
	AccountTypeEquity      = "equity"
	AccountTypeIncome      = "income"
	AccountTypeExpenses    = "expenses"
)

const (
	JournalSourceManual      = "manual"
	JournalSourceFeePayment  = "fee_payment"
	JournalSourceFeeRefund   = "fee_refund"
	JournalSourceExpense     = "expense"
	JournalSourceIncome      = "income"
)

const (
	JournalStatusDraft    = "draft"
	JournalStatusPosted   = "posted"
	JournalStatusReversed = "reversed"
)

const (
	ExpenseStatusDraft            = "draft"
	ExpenseStatusPendingApproval  = "pending_approval"
	ExpenseStatusApproved         = "approved"
	ExpenseStatusPaid             = "paid"
	ExpenseStatusRejected         = "rejected"
)

const (
	IncomeSourceDonation      = "donation"
	IncomeSourceEvent         = "event"
	IncomeSourceAdmissionForm = "admission_form"
	IncomeSourceMisc          = "misc"
)

const (
	PeriodStatusOpen   = "open"
	PeriodStatusClosed = "closed"
)

const (
	SeqJournal = "journal"
)

const (
	AccountCodeCash            = "1000"
	AccountCodeBank            = "1010"
	AccountCodeReceivable      = "1100"
	AccountCodeTuitionIncome   = "4000"
	AccountCodeExamFeeIncome   = "4010"
	AccountCodeTransportIncome = "4020"
	AccountCodeLibraryIncome   = "4030"
	AccountCodeSalaryExpense   = "5000"
	AccountCodeUtilityExpense  = "5010"
	AccountCodeMiscExpense     = "5090"
)
