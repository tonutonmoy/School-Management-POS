package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountingRepository interface {
	WithTx(ctx context.Context, fn func(pgx.Tx) error) error

	// Accounts
	CreateAccount(ctx context.Context, p CreateAccountParams) (*AccountRecord, error)
	UpdateAccount(ctx context.Context, id uuid.UUID, p CreateAccountParams) (*AccountRecord, error)
	DisableAccount(ctx context.Context, id uuid.UUID) error
	GetAccount(ctx context.Context, id uuid.UUID) (*AccountRecord, error)
	GetAccountByCode(ctx context.Context, code string) (*AccountRecord, error)
	ListAccounts(ctx context.Context, accountType, query string) ([]AccountRecord, error)

	// Journal
	NextJournalNumber(ctx context.Context, tx pgx.Tx, year int) (string, error)
	JournalExistsBySource(ctx context.Context, tx pgx.Tx, sourceType string, sourceID uuid.UUID) (bool, error)
	CreateJournalEntry(ctx context.Context, tx pgx.Tx, p CreateJournalParams) (*JournalEntryRecord, error)
	ListJournalEntries(ctx context.Context, f JournalSearchParams) ([]JournalEntryRecord, error)
	CountJournalEntries(ctx context.Context, f JournalSearchParams) (int64, error)
	GetJournalEntry(ctx context.Context, id uuid.UUID) (*JournalEntryRecord, error)
	ListJournalLines(ctx context.Context, entryID uuid.UUID) ([]JournalLineRecord, error)

	// Fee integration
	PostFeePaymentJournal(ctx context.Context, tx pgx.Tx, p FeePaymentJournalParams) error
	PostFeeRefundJournal(ctx context.Context, tx pgx.Tx, paymentID uuid.UUID, actorID uuid.UUID) error

	// Ledger & balances
	AccountBalance(ctx context.Context, accountID uuid.UUID, asOf time.Time) (float64, error)
	ListLedgerEntries(ctx context.Context, accountID uuid.UUID, from, to time.Time) ([]LedgerLineRecord, error)
	ListCashBook(ctx context.Context, from, to time.Time) ([]LedgerLineRecord, error)
	ListBankBook(ctx context.Context, from, to time.Time) ([]LedgerLineRecord, error)

	// Reports
	TrialBalance(ctx context.Context, asOf time.Time) ([]BalanceRowRecord, error)
	AccountActivity(ctx context.Context, accountType string, from, to time.Time) ([]BalanceRowRecord, error)

	// Expenses
	ListExpenseCategories(ctx context.Context) ([]ExpenseCategoryRecord, error)
	CreateExpense(ctx context.Context, p CreateExpenseParams) (*ExpenseRecord, error)
	UpdateExpense(ctx context.Context, id uuid.UUID, p CreateExpenseParams) (*ExpenseRecord, error)
	GetExpense(ctx context.Context, id uuid.UUID) (*ExpenseRecord, error)
	ListExpenses(ctx context.Context, f ExpenseSearchParams) ([]ExpenseRecord, error)
	CountExpenses(ctx context.Context, f ExpenseSearchParams) (int64, error)
	ApproveExpense(ctx context.Context, id, approverID uuid.UUID) (*ExpenseRecord, error)
	ApproveExpenseTx(ctx context.Context, tx pgx.Tx, id, approverID uuid.UUID) error
	SetExpenseJournal(ctx context.Context, id, journalID uuid.UUID) error
	SetExpenseJournalTx(ctx context.Context, tx pgx.Tx, id, journalID uuid.UUID) error
	SetExpenseAttachment(ctx context.Context, id uuid.UUID, url string) error

	// Income
	CreateIncomeEntry(ctx context.Context, p CreateIncomeParams) (*IncomeRecord, error)
	CreateIncomeEntryTx(ctx context.Context, tx pgx.Tx, p CreateIncomeParams) (uuid.UUID, error)
	GetIncomeEntry(ctx context.Context, id uuid.UUID) (*IncomeRecord, error)
	ListIncomeEntries(ctx context.Context, from, to time.Time) ([]IncomeRecord, error)
	SetIncomeJournal(ctx context.Context, id, journalID uuid.UUID) error

	// Periods
	ListFinancialPeriods(ctx context.Context) ([]FinancialPeriodRecord, error)
	GetOpenPeriod(ctx context.Context, date time.Time) (*FinancialPeriodRecord, error)
	CreateFinancialPeriod(ctx context.Context, name string, start, end time.Time) (*FinancialPeriodRecord, error)
	CloseFinancialPeriod(ctx context.Context, id, actorID uuid.UUID) error
	IsPeriodLocked(ctx context.Context, date time.Time) (bool, error)

	// Dashboard
	SumAccountDebitsCredits(ctx context.Context, accountCode string, from, to time.Time) (debits, credits float64, err error)
	MonthlyTrend(ctx context.Context, accountType string, months int) ([]TrendRecord, error)
}

type CreateAccountParams struct {
	Code, Name, AccountType, Description string
	ParentID                             *uuid.UUID
}

type AccountRecord struct {
	ID, ParentID       uuid.UUID
	ParentName         string
	Code, Name         string
	AccountType        string
	Description        string
	IsSystem, IsActive bool
}

type CreateJournalParams struct {
	EntryNumber, Description, SourceType, Status string
	EntryDate                                    time.Time
	SourceID, PeriodID, CreatedBy                *uuid.UUID
	Lines                                        []JournalLineParams
}

type JournalLineParams struct {
	AccountID             uuid.UUID
	Debit, Credit         float64
	Description           string
	LineOrder             int
}

type JournalEntryRecord struct {
	ID          uuid.UUID
	EntryNumber string
	EntryDate   time.Time
	Description string
	SourceType  string
	Status      string
	TotalDebit  float64
	TotalCredit float64
	CreatedAt   time.Time
}

type JournalLineRecord struct {
	ID, AccountID uuid.UUID
	AccountCode, AccountName string
	Debit, Credit float64
	Description   string
}

type JournalSearchParams struct {
	From, To time.Time
	SourceType, Query string
	Limit, Offset int32
}

type LedgerLineRecord struct {
	EntryDate   time.Time
	EntryNumber string
	Description string
	Debit, Credit float64
	SourceType  string
	AccountType string
}

type BalanceRowRecord struct {
	AccountID   uuid.UUID
	AccountCode, AccountName, AccountType string
	Debit, Credit, Balance float64
}

type FeePaymentJournalParams struct {
	PaymentID      uuid.UUID
	Amount         float64
	PaymentMethod  string
	CollectionDate time.Time
	Allocations    []FeeAllocationJournal
	ActorID        uuid.UUID
}

type FeeAllocationJournal struct {
	BillID uuid.UUID
	Amount float64
}

type ExpenseCategoryRecord struct {
	ID       uuid.UUID
	Name, Slug string
	AccountID uuid.UUID
	IsActive bool
}

type CreateExpenseParams struct {
	CategoryID, ExpenseAccountID, PayFromAccountID uuid.UUID
	Amount                                         float64
	ExpenseDate                                    time.Time
	Description, PaymentMethod                     string
	Status                                         string
	CreatedBy                                      uuid.UUID
}

type ExpenseRecord struct {
	ID, CategoryID uuid.UUID
	CategoryName   string
	ExpenseAccountID, PayFromAccountID uuid.UUID
	Amount         float64
	ExpenseDate    time.Time
	Description, PaymentMethod, Status, AttachmentURL string
	CreatedByName, ApprovedByName string
	ApprovedAt *time.Time
	JournalEntryID *uuid.UUID
}

type ExpenseSearchParams struct {
	Status string
	From, To time.Time
	Limit, Offset int32
}

type CreateIncomeParams struct {
	IncomeAccountID, ReceiveToAccountID uuid.UUID
	Amount                              float64
	IncomeDate                          time.Time
	Source, Description                 string
	CreatedBy                           uuid.UUID
}

type IncomeRecord struct {
	ID uuid.UUID
	IncomeAccountID, ReceiveToAccountID uuid.UUID
	IncomeAccountName, ReceiveAccountName string
	Amount float64
	IncomeDate time.Time
	Source, Description string
}

type FinancialPeriodRecord struct {
	ID        uuid.UUID
	Name      string
	StartDate, EndDate time.Time
	Status    string
	IsLocked  bool
	ClosedAt  *time.Time
}

type TrendRecord struct {
	Label  string
	Amount float64
}

type accountingRepository struct{ pool *pgxpool.Pool }

func NewAccountingRepository(pool *pgxpool.Pool) AccountingRepository {
	return &accountingRepository{pool: pool}
}

func (r *accountingRepository) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *accountingRepository) CreateAccount(ctx context.Context, p CreateAccountParams) (*AccountRecord, error) {
	var id uuid.UUID
	var parentID *uuid.UUID
	if p.ParentID != nil && *p.ParentID != uuid.Nil {
		parentID = p.ParentID
	}
	err := r.pool.QueryRow(ctx, `
INSERT INTO accounts (code, name, account_type, parent_id, description)
VALUES ($1,$2,$3,$4,$5) RETURNING id`, p.Code, p.Name, p.AccountType, parentID, p.Description).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetAccount(ctx, id)
}

func (r *accountingRepository) UpdateAccount(ctx context.Context, id uuid.UUID, p CreateAccountParams) (*AccountRecord, error) {
	var parentID *uuid.UUID
	if p.ParentID != nil && *p.ParentID != uuid.Nil {
		parentID = p.ParentID
	}
	_, err := r.pool.Exec(ctx, `
UPDATE accounts SET code=$2, name=$3, account_type=$4, parent_id=$5, description=$6, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL AND is_system=false`, id, p.Code, p.Name, p.AccountType, parentID, p.Description)
	if err != nil {
		return nil, err
	}
	return r.GetAccount(ctx, id)
}

func (r *accountingRepository) DisableAccount(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE accounts SET is_active=false, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL AND is_system=false`, id)
	return err
}

func scanAccount(row pgx.Row) (*AccountRecord, error) {
	var rec AccountRecord
	var parentID *uuid.UUID
	var parentName *string
	err := row.Scan(&rec.ID, &rec.Code, &rec.Name, &rec.AccountType, &parentID, &parentName,
		&rec.Description, &rec.IsSystem, &rec.IsActive)
	rec.ParentID = uuid.Nil
	if parentID != nil {
		rec.ParentID = *parentID
	}
	if parentName != nil {
		rec.ParentName = *parentName
	}
	return &rec, err
}

func (r *accountingRepository) accountSelect() string {
	return `
SELECT a.id, a.code, a.name, a.account_type, a.parent_id, p.name, a.description, a.is_system, a.is_active
FROM accounts a LEFT JOIN accounts p ON p.id = a.parent_id
WHERE a.deleted_at IS NULL`
}

func (r *accountingRepository) GetAccount(ctx context.Context, id uuid.UUID) (*AccountRecord, error) {
	rec, err := scanAccount(r.pool.QueryRow(ctx, r.accountSelect()+` AND a.id=$1`, id))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *accountingRepository) GetAccountByCode(ctx context.Context, code string) (*AccountRecord, error) {
	rec, err := scanAccount(r.pool.QueryRow(ctx, r.accountSelect()+` AND a.code=$1`, code))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *accountingRepository) ListAccounts(ctx context.Context, accountType, query string) ([]AccountRecord, error) {
	q := r.accountSelect()
	args := []any{}
	n := 1
	if accountType != "" {
		q += fmt.Sprintf(" AND a.account_type=$%d", n)
		args = append(args, accountType)
		n++
	}
	if query != "" {
		q += fmt.Sprintf(" AND (a.name ILIKE $%d OR a.code ILIKE $%d)", n, n)
		args = append(args, "%"+query+"%")
		n++
	}
	q += " ORDER BY a.code"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AccountRecord
	for rows.Next() {
		rec, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func (r *accountingRepository) NextJournalNumber(ctx context.Context, tx pgx.Tx, year int) (string, error) {
	var num int
	err := tx.QueryRow(ctx, `
INSERT INTO accounting_sequences (entity_type, year, last_number) VALUES ('journal', $1, 1)
ON CONFLICT (entity_type, year) DO UPDATE SET last_number = accounting_sequences.last_number + 1
RETURNING last_number`, year).Scan(&num)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("JE-%d-%05d", year, num), nil
}

func (r *accountingRepository) JournalExistsBySource(ctx context.Context, tx pgx.Tx, sourceType string, sourceID uuid.UUID) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
SELECT EXISTS(SELECT 1 FROM journal_entries WHERE source_type=$1 AND source_id=$2 AND deleted_at IS NULL)`,
		sourceType, sourceID).Scan(&exists)
	return exists, err
}

func (r *accountingRepository) CreateJournalEntry(ctx context.Context, tx pgx.Tx, p CreateJournalParams) (*JournalEntryRecord, error) {
	var totalDebit, totalCredit float64
	for _, l := range p.Lines {
		totalDebit += l.Debit
		totalCredit += l.Credit
	}
	if totalDebit != totalCredit || totalDebit == 0 {
		return nil, fmt.Errorf("journal entry must balance: debit %.2f credit %.2f", totalDebit, totalCredit)
	}
	var id uuid.UUID
	err := tx.QueryRow(ctx, `
INSERT INTO journal_entries (entry_number, entry_date, description, source_type, source_id, status, period_id, created_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		p.EntryNumber, p.EntryDate, p.Description, p.SourceType, p.SourceID, p.Status, p.PeriodID, p.CreatedBy).Scan(&id)
	if err != nil {
		return nil, err
	}
	for i, l := range p.Lines {
		_, err := tx.Exec(ctx, `
INSERT INTO journal_lines (journal_entry_id, account_id, debit, credit, description, line_order)
VALUES ($1,$2,$3,$4,$5,$6)`, id, l.AccountID, l.Debit, l.Credit, l.Description, i+1)
		if err != nil {
			return nil, err
		}
	}
	return &JournalEntryRecord{
		ID: id, EntryNumber: p.EntryNumber, EntryDate: p.EntryDate, Description: p.Description,
		SourceType: p.SourceType, Status: p.Status, TotalDebit: totalDebit, TotalCredit: totalCredit,
	}, nil
}

func (r *accountingRepository) journalSelect() string {
	return `
SELECT je.id, je.entry_number, je.entry_date, je.description, je.source_type, je.status,
    COALESCE((SELECT SUM(debit) FROM journal_lines WHERE journal_entry_id=je.id),0),
    COALESCE((SELECT SUM(credit) FROM journal_lines WHERE journal_entry_id=je.id),0),
    je.created_at
FROM journal_entries je WHERE je.deleted_at IS NULL`
}

func scanJournal(row pgx.Row) (*JournalEntryRecord, error) {
	var rec JournalEntryRecord
	err := row.Scan(&rec.ID, &rec.EntryNumber, &rec.EntryDate, &rec.Description, &rec.SourceType,
		&rec.Status, &rec.TotalDebit, &rec.TotalCredit, &rec.CreatedAt)
	return &rec, err
}

func (r *accountingRepository) ListJournalEntries(ctx context.Context, f JournalSearchParams) ([]JournalEntryRecord, error) {
	q, args := r.journalSearchQuery(f, false)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []JournalEntryRecord
	for rows.Next() {
		rec, err := scanJournal(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func (r *accountingRepository) journalSearchQuery(f JournalSearchParams, count bool) (string, []any) {
	q := r.journalSelect()
	args := []any{}
	n := 1
	if !f.From.IsZero() {
		q += fmt.Sprintf(" AND je.entry_date >= $%d", n)
		args = append(args, f.From)
		n++
	}
	if !f.To.IsZero() {
		q += fmt.Sprintf(" AND je.entry_date <= $%d", n)
		args = append(args, f.To)
		n++
	}
	if f.SourceType != "" {
		q += fmt.Sprintf(" AND je.source_type=$%d", n)
		args = append(args, f.SourceType)
		n++
	}
	if f.Query != "" {
		q += fmt.Sprintf(" AND (je.description ILIKE $%d OR je.entry_number ILIKE $%d)", n, n)
		args = append(args, "%"+f.Query+"%")
		n++
	}
	if count {
		return "SELECT COUNT(*) FROM (" + q + ") sub", args
	}
	q += " ORDER BY je.entry_date DESC, je.entry_number DESC"
	if f.Limit > 0 {
		q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", n, n+1)
		args = append(args, f.Limit, f.Offset)
	}
	return q, args
}

func (r *accountingRepository) CountJournalEntries(ctx context.Context, f JournalSearchParams) (int64, error) {
	q, args := r.journalSearchQuery(f, true)
	var count int64
	return count, r.pool.QueryRow(ctx, q, args...).Scan(&count)
}

func (r *accountingRepository) GetJournalEntry(ctx context.Context, id uuid.UUID) (*JournalEntryRecord, error) {
	rec, err := scanJournal(r.pool.QueryRow(ctx, r.journalSelect()+` AND je.id=$1`, id))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *accountingRepository) ListJournalLines(ctx context.Context, entryID uuid.UUID) ([]JournalLineRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT jl.id, jl.account_id, a.code, a.name, jl.debit, jl.credit, COALESCE(jl.description,'')
FROM journal_lines jl JOIN accounts a ON a.id=jl.account_id
WHERE jl.journal_entry_id=$1 ORDER BY jl.line_order`, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []JournalLineRecord
	for rows.Next() {
		var rec JournalLineRecord
		if err := rows.Scan(&rec.ID, &rec.AccountID, &rec.AccountCode, &rec.AccountName, &rec.Debit, &rec.Credit, &rec.Description); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *accountingRepository) PostFeePaymentJournal(ctx context.Context, tx pgx.Tx, p FeePaymentJournalParams) error {
	exists, err := r.JournalExistsBySource(ctx, tx, "fee_payment", p.PaymentID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	cashAcct, err := r.getAccountByCodeTx(ctx, tx, "1000")
	if err != nil || cashAcct == nil {
		return fmt.Errorf("cash account not found")
	}
	bankAcct, err := r.getAccountByCodeTx(ctx, tx, "1010")
	if err != nil || bankAcct == nil {
		return fmt.Errorf("bank account not found")
	}
	debitAcct := cashAcct.ID
	if p.PaymentMethod == "bank" || p.PaymentMethod == "card" {
		debitAcct = bankAcct.ID
	}
	incomeCredits := map[uuid.UUID]float64{}
	for _, alloc := range p.Allocations {
		items, err := r.listBillItemsTx(ctx, tx, alloc.BillID)
		if err != nil {
			return err
		}
		var billTotal float64
		for _, it := range items {
			billTotal += it.Amount
		}
		if billTotal <= 0 {
			tuition, _ := r.getAccountByCodeTx(ctx, tx, "4000")
			if tuition != nil {
				incomeCredits[tuition.ID] += alloc.Amount
			}
			continue
		}
		for _, it := range items {
			share := alloc.Amount * (it.Amount / billTotal)
			acctCode := feeSlugToIncomeCode(it.FeeTypeSlug)
			incAcct, err := r.getAccountByCodeTx(ctx, tx, acctCode)
			if err != nil || incAcct == nil {
				incAcct, _ = r.getAccountByCodeTx(ctx, tx, "4000")
			}
			if incAcct != nil {
				incomeCredits[incAcct.ID] += share
			}
		}
	}
	var lines []JournalLineParams
	lines = append(lines, JournalLineParams{AccountID: debitAcct, Debit: p.Amount, Description: "Fee collection"})
	order := 1
	for acctID, amt := range incomeCredits {
		if amt > 0.001 {
			lines = append(lines, JournalLineParams{AccountID: acctID, Credit: round2(amt), Description: "Fee income", LineOrder: order})
			order++
		}
	}
	if len(lines) < 2 {
		tuition, _ := r.getAccountByCodeTx(ctx, tx, "4000")
		if tuition != nil {
			lines = []JournalLineParams{
				{AccountID: debitAcct, Debit: p.Amount, Description: "Fee collection"},
				{AccountID: tuition.ID, Credit: p.Amount, Description: "Tuition income"},
			}
		}
	}
	// Normalize credit lines to match debit total (rounding)
	var creditSum float64
	for i := range lines {
		if lines[i].Credit > 0 {
			creditSum += lines[i].Credit
		}
	}
	if diff := round2(p.Amount - creditSum); diff != 0 {
		for i := range lines {
			if lines[i].Credit > 0 {
				lines[i].Credit = round2(lines[i].Credit + diff)
				break
			}
		}
	}
	year := p.CollectionDate.Year()
	entryNum, err := r.NextJournalNumber(ctx, tx, year)
	if err != nil {
		return err
	}
	period, _ := r.getOpenPeriodTx(ctx, tx, p.CollectionDate)
	var periodID *uuid.UUID
	if period != nil {
		periodID = &period.ID
	}
	if locked, _ := r.isPeriodLockedTx(ctx, tx, p.CollectionDate); locked {
		return fmt.Errorf("financial period is locked for date %s", p.CollectionDate.Format("2006-01-02"))
	}
	srcID := p.PaymentID
	_, err = r.CreateJournalEntry(ctx, tx, CreateJournalParams{
		EntryNumber: entryNum,
		EntryDate:   p.CollectionDate,
		Description: fmt.Sprintf("Fee payment %s", p.PaymentID.String()[:8]),
		SourceType:  "fee_payment",
		SourceID:    &srcID,
		Status:      "posted",
		PeriodID:    periodID,
		CreatedBy:   &p.ActorID,
		Lines:       lines,
	})
	return err
}

func (r *accountingRepository) PostFeeRefundJournal(ctx context.Context, tx pgx.Tx, paymentID uuid.UUID, actorID uuid.UUID) error {
	exists, err := r.JournalExistsBySource(ctx, tx, "fee_refund", paymentID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	origExists, _ := r.JournalExistsBySource(ctx, tx, "fee_payment", paymentID)
	if !origExists {
		return nil
	}
	var amount float64
	var method string
	var collDate time.Time
	err = tx.QueryRow(ctx, `SELECT amount, payment_method, collection_date FROM payments WHERE id=$1`, paymentID).
		Scan(&amount, &method, &collDate)
	if err != nil {
		return err
	}
	cashAcct, _ := r.getAccountByCodeTx(ctx, tx, "1000")
	bankAcct, _ := r.getAccountByCodeTx(ctx, tx, "1010")
	tuition, _ := r.getAccountByCodeTx(ctx, tx, "4000")
	if cashAcct == nil || bankAcct == nil || tuition == nil {
		return fmt.Errorf("required accounts not found for refund journal")
	}
	creditAcct := cashAcct.ID
	if method == "bank" || method == "card" {
		creditAcct = bankAcct.ID
	}
	year := time.Now().Year()
	entryNum, err := r.NextJournalNumber(ctx, tx, year)
	if err != nil {
		return err
	}
	srcID := paymentID
	_, err = r.CreateJournalEntry(ctx, tx, CreateJournalParams{
		EntryNumber: entryNum,
		EntryDate:   time.Now().Truncate(24 * time.Hour),
		Description: fmt.Sprintf("Fee refund for payment %s", paymentID.String()[:8]),
		SourceType:  "fee_refund",
		SourceID:    &srcID,
		Status:      "posted",
		CreatedBy:   &actorID,
		Lines: []JournalLineParams{
			{AccountID: tuition.ID, Debit: amount, Description: "Reverse fee income"},
			{AccountID: creditAcct, Credit: amount, Description: "Refund payment"},
		},
	})
	return err
}

func feeSlugToIncomeCode(slug string) string {
	switch slug {
	case "exam", "exam_fee":
		return "4010"
	case "transport":
		return "4020"
	case "library":
		return "4030"
	default:
		return "4000"
	}
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

type billItemWithSlug struct {
	FeeTypeSlug string
	Amount      float64
}

func (r *accountingRepository) listBillItemsTx(ctx context.Context, tx pgx.Tx, billID uuid.UUID) ([]billItemWithSlug, error) {
	rows, err := tx.Query(ctx, `
SELECT ft.slug, bi.amount FROM bill_items bi JOIN fee_types ft ON ft.id=bi.fee_type_id WHERE bi.bill_id=$1`, billID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []billItemWithSlug
	for rows.Next() {
		var rec billItemWithSlug
		if err := rows.Scan(&rec.FeeTypeSlug, &rec.Amount); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *accountingRepository) getAccountByCodeTx(ctx context.Context, tx pgx.Tx, code string) (*AccountRecord, error) {
	rec, err := scanAccount(tx.QueryRow(ctx, r.accountSelect()+` AND a.code=$1`, code))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *accountingRepository) AccountBalance(ctx context.Context, accountID uuid.UUID, asOf time.Time) (float64, error) {
	var acctType string
	if err := r.pool.QueryRow(ctx, `SELECT account_type FROM accounts WHERE id=$1`, accountID).Scan(&acctType); err != nil {
		return 0, err
	}
	var debits, credits float64
	err := r.pool.QueryRow(ctx, `
SELECT COALESCE(SUM(jl.debit),0), COALESCE(SUM(jl.credit),0)
FROM journal_lines jl
JOIN journal_entries je ON je.id=jl.journal_entry_id
WHERE jl.account_id=$1 AND je.deleted_at IS NULL AND je.status='posted' AND je.entry_date <= $2`,
		accountID, asOf).Scan(&debits, &credits)
	if err != nil {
		return 0, err
	}
	return balanceForType(acctType, debits, credits), nil
}

func balanceForType(acctType string, debits, credits float64) float64 {
	switch acctType {
	case "assets", "expenses":
		return debits - credits
	default:
		return credits - debits
	}
}

func (r *accountingRepository) ListLedgerEntries(ctx context.Context, accountID uuid.UUID, from, to time.Time) ([]LedgerLineRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT je.entry_date, je.entry_number, COALESCE(jl.description, je.description), jl.debit, jl.credit, je.source_type, a.account_type
FROM journal_lines jl
JOIN journal_entries je ON je.id=jl.journal_entry_id
JOIN accounts a ON a.id=jl.account_id
WHERE jl.account_id=$1 AND je.deleted_at IS NULL AND je.status='posted'
AND je.entry_date BETWEEN $2 AND $3
ORDER BY je.entry_date, je.entry_number`, accountID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LedgerLineRecord
	for rows.Next() {
		var rec LedgerLineRecord
		if err := rows.Scan(&rec.EntryDate, &rec.EntryNumber, &rec.Description, &rec.Debit, &rec.Credit, &rec.SourceType, &rec.AccountType); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *accountingRepository) ListCashBook(ctx context.Context, from, to time.Time) ([]LedgerLineRecord, error) {
	cash, err := r.GetAccountByCode(ctx, "1000")
	if err != nil || cash == nil {
		return nil, err
	}
	return r.ListLedgerEntries(ctx, cash.ID, from, to)
}

func (r *accountingRepository) ListBankBook(ctx context.Context, from, to time.Time) ([]LedgerLineRecord, error) {
	bank, err := r.GetAccountByCode(ctx, "1010")
	if err != nil || bank == nil {
		return nil, err
	}
	return r.ListLedgerEntries(ctx, bank.ID, from, to)
}

func (r *accountingRepository) TrialBalance(ctx context.Context, asOf time.Time) ([]BalanceRowRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT a.id, a.code, a.name, a.account_type,
    COALESCE(SUM(jl.debit),0), COALESCE(SUM(jl.credit),0)
FROM accounts a
LEFT JOIN journal_lines jl ON jl.account_id=a.id
LEFT JOIN journal_entries je ON je.id=jl.journal_entry_id AND je.deleted_at IS NULL AND je.status='posted' AND je.entry_date <= $1
WHERE a.deleted_at IS NULL AND a.is_active=true
GROUP BY a.id, a.code, a.name, a.account_type
HAVING COALESCE(SUM(jl.debit),0) > 0 OR COALESCE(SUM(jl.credit),0) > 0
ORDER BY a.code`, asOf)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BalanceRowRecord
	for rows.Next() {
		var rec BalanceRowRecord
		if err := rows.Scan(&rec.AccountID, &rec.AccountCode, &rec.AccountName, &rec.AccountType, &rec.Debit, &rec.Credit); err != nil {
			return nil, err
		}
		rec.Balance = balanceForType(rec.AccountType, rec.Debit, rec.Credit)
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *accountingRepository) AccountActivity(ctx context.Context, accountType string, from, to time.Time) ([]BalanceRowRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT a.id, a.code, a.name, a.account_type,
    COALESCE(SUM(jl.debit),0), COALESCE(SUM(jl.credit),0)
FROM accounts a
JOIN journal_lines jl ON jl.account_id=a.id
JOIN journal_entries je ON je.id=jl.journal_entry_id
WHERE a.deleted_at IS NULL AND a.account_type=$1 AND je.deleted_at IS NULL AND je.status='posted'
AND je.entry_date BETWEEN $2 AND $3
GROUP BY a.id, a.code, a.name, a.account_type
ORDER BY a.code`, accountType, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BalanceRowRecord
	for rows.Next() {
		var rec BalanceRowRecord
		if err := rows.Scan(&rec.AccountID, &rec.AccountCode, &rec.AccountName, &rec.AccountType, &rec.Debit, &rec.Credit); err != nil {
			return nil, err
		}
		rec.Balance = balanceForType(rec.AccountType, rec.Debit, rec.Credit)
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *accountingRepository) ListExpenseCategories(ctx context.Context) ([]ExpenseCategoryRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, name, slug, COALESCE(account_id, '00000000-0000-0000-0000-000000000000'), is_active
FROM expense_categories WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExpenseCategoryRecord
	for rows.Next() {
		var rec ExpenseCategoryRecord
		if err := rows.Scan(&rec.ID, &rec.Name, &rec.Slug, &rec.AccountID, &rec.IsActive); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *accountingRepository) CreateExpense(ctx context.Context, p CreateExpenseParams) (*ExpenseRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO expenses (category_id, expense_account_id, pay_from_account_id, amount, expense_date, description, payment_method, status, created_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		p.CategoryID, p.ExpenseAccountID, p.PayFromAccountID, p.Amount, p.ExpenseDate,
		p.Description, p.PaymentMethod, p.Status, p.CreatedBy).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetExpense(ctx, id)
}

func (r *accountingRepository) expenseSelect() string {
	return `
SELECT e.id, e.category_id, ec.name, e.expense_account_id, e.pay_from_account_id,
    e.amount, e.expense_date, e.description, e.payment_method, e.status,
    COALESCE(e.attachment_url,''), COALESCE(u1.first_name||' '||u1.last_name,''),
    COALESCE(u2.first_name||' '||u2.last_name,''), e.approved_at, e.journal_entry_id
FROM expenses e
JOIN expense_categories ec ON ec.id=e.category_id
LEFT JOIN users u1 ON u1.id=e.created_by
LEFT JOIN users u2 ON u2.id=e.approved_by
WHERE e.deleted_at IS NULL`
}

func (r *accountingRepository) GetExpense(ctx context.Context, id uuid.UUID) (*ExpenseRecord, error) {
	var rec ExpenseRecord
	var journalID *uuid.UUID
	err := r.pool.QueryRow(ctx, r.expenseSelect()+` AND e.id=$1`, id).Scan(
		&rec.ID, &rec.CategoryID, &rec.CategoryName, &rec.ExpenseAccountID, &rec.PayFromAccountID,
		&rec.Amount, &rec.ExpenseDate, &rec.Description, &rec.PaymentMethod, &rec.Status,
		&rec.AttachmentURL, &rec.CreatedByName, &rec.ApprovedByName, &rec.ApprovedAt, &journalID)
	rec.JournalEntryID = journalID
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &rec, err
}

func (r *accountingRepository) UpdateExpense(ctx context.Context, id uuid.UUID, p CreateExpenseParams) (*ExpenseRecord, error) {
	_, err := r.pool.Exec(ctx, `
UPDATE expenses SET category_id=$2, expense_account_id=$3, pay_from_account_id=$4, amount=$5,
    expense_date=$6, description=$7, payment_method=$8, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL AND status IN ('draft','pending_approval')`,
		id, p.CategoryID, p.ExpenseAccountID, p.PayFromAccountID, p.Amount, p.ExpenseDate, p.Description, p.PaymentMethod)
	if err != nil {
		return nil, err
	}
	return r.GetExpense(ctx, id)
}

func (r *accountingRepository) ListExpenses(ctx context.Context, f ExpenseSearchParams) ([]ExpenseRecord, error) {
	q, args := r.expenseSearchQuery(f, false)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExpenseRecord
	for rows.Next() {
		var rec ExpenseRecord
		var journalID *uuid.UUID
		if err := rows.Scan(
			&rec.ID, &rec.CategoryID, &rec.CategoryName, &rec.ExpenseAccountID, &rec.PayFromAccountID,
			&rec.Amount, &rec.ExpenseDate, &rec.Description, &rec.PaymentMethod, &rec.Status,
			&rec.AttachmentURL, &rec.CreatedByName, &rec.ApprovedByName, &rec.ApprovedAt, &journalID); err != nil {
			return nil, err
		}
		rec.JournalEntryID = journalID
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *accountingRepository) expenseSearchQuery(f ExpenseSearchParams, count bool) (string, []any) {
	q := r.expenseSelect()
	args := []any{}
	n := 1
	if f.Status != "" {
		q += fmt.Sprintf(" AND e.status=$%d", n)
		args = append(args, f.Status)
		n++
	}
	if !f.From.IsZero() {
		q += fmt.Sprintf(" AND e.expense_date >= $%d", n)
		args = append(args, f.From)
		n++
	}
	if !f.To.IsZero() {
		q += fmt.Sprintf(" AND e.expense_date <= $%d", n)
		args = append(args, f.To)
		n++
	}
	if count {
		return "SELECT COUNT(*) FROM (" + q + ") sub", args
	}
	q += " ORDER BY e.expense_date DESC"
	if f.Limit > 0 {
		q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", n, n+1)
		args = append(args, f.Limit, f.Offset)
	}
	return q, args
}

func (r *accountingRepository) CountExpenses(ctx context.Context, f ExpenseSearchParams) (int64, error) {
	q, args := r.expenseSearchQuery(f, true)
	var count int64
	return count, r.pool.QueryRow(ctx, q, args...).Scan(&count)
}

func (r *accountingRepository) ApproveExpenseTx(ctx context.Context, tx pgx.Tx, id, approverID uuid.UUID) error {
	tag, err := tx.Exec(ctx, `
UPDATE expenses SET status='approved', approved_by=$2, approved_at=NOW(), updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL AND status='pending_approval'`, id, approverID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *accountingRepository) SetExpenseJournalTx(ctx context.Context, tx pgx.Tx, id, journalID uuid.UUID) error {
	_, err := tx.Exec(ctx, `UPDATE expenses SET journal_entry_id=$2, status='paid', updated_at=NOW() WHERE id=$1`, id, journalID)
	return err
}

func (r *accountingRepository) CreateIncomeEntryTx(ctx context.Context, tx pgx.Tx, p CreateIncomeParams) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `
INSERT INTO income_entries (income_account_id, receive_to_account_id, amount, income_date, source, description, created_by)
VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`, p.IncomeAccountID, p.ReceiveToAccountID, p.Amount, p.IncomeDate, p.Source, p.Description, p.CreatedBy).Scan(&id)
	return id, err
}

func (r *accountingRepository) ApproveExpense(ctx context.Context, id, approverID uuid.UUID) (*ExpenseRecord, error) {
	if err := r.WithTx(ctx, func(tx pgx.Tx) error {
		return r.ApproveExpenseTx(ctx, tx, id, approverID)
	}); err != nil {
		return nil, err
	}
	return r.GetExpense(ctx, id)
}

func (r *accountingRepository) SetExpenseJournal(ctx context.Context, id, journalID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE expenses SET journal_entry_id=$2, status='paid', updated_at=NOW() WHERE id=$1`, id, journalID)
	return err
}

func (r *accountingRepository) SetExpenseAttachment(ctx context.Context, id uuid.UUID, url string) error {
	_, err := r.pool.Exec(ctx, `UPDATE expenses SET attachment_url=$2, updated_at=NOW() WHERE id=$1`, id, url)
	return err
}

func (r *accountingRepository) GetIncomeEntry(ctx context.Context, id uuid.UUID) (*IncomeRecord, error) {
	return r.getIncome(ctx, id)
}

func (r *accountingRepository) CreateIncomeEntry(ctx context.Context, p CreateIncomeParams) (*IncomeRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO income_entries (income_account_id, receive_to_account_id, amount, income_date, source, description, created_by)
VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`, p.IncomeAccountID, p.ReceiveToAccountID, p.Amount, p.IncomeDate, p.Source, p.Description, p.CreatedBy).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.getIncome(ctx, id)
}

func (r *accountingRepository) getIncome(ctx context.Context, id uuid.UUID) (*IncomeRecord, error) {
	var rec IncomeRecord
	err := r.pool.QueryRow(ctx, `
SELECT ie.id, ie.income_account_id, ia.name, ie.receive_to_account_id, ra.name,
    ie.amount, ie.income_date, ie.source, ie.description
FROM income_entries ie
JOIN accounts ia ON ia.id=ie.income_account_id
JOIN accounts ra ON ra.id=ie.receive_to_account_id
WHERE ie.id=$1 AND ie.deleted_at IS NULL`, id).Scan(
		&rec.ID, &rec.IncomeAccountID, &rec.IncomeAccountName, &rec.ReceiveToAccountID, &rec.ReceiveAccountName,
		&rec.Amount, &rec.IncomeDate, &rec.Source, &rec.Description)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &rec, err
}

func (r *accountingRepository) ListIncomeEntries(ctx context.Context, from, to time.Time) ([]IncomeRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT ie.id, ie.income_account_id, ia.name, ie.receive_to_account_id, ra.name,
    ie.amount, ie.income_date, ie.source, ie.description
FROM income_entries ie
JOIN accounts ia ON ia.id=ie.income_account_id
JOIN accounts ra ON ra.id=ie.receive_to_account_id
WHERE ie.deleted_at IS NULL AND ie.income_date BETWEEN $1 AND $2
ORDER BY ie.income_date DESC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []IncomeRecord
	for rows.Next() {
		var rec IncomeRecord
		if err := rows.Scan(
			&rec.ID, &rec.IncomeAccountID, &rec.IncomeAccountName, &rec.ReceiveToAccountID, &rec.ReceiveAccountName,
			&rec.Amount, &rec.IncomeDate, &rec.Source, &rec.Description); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *accountingRepository) SetIncomeJournal(ctx context.Context, id, journalID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE income_entries SET journal_entry_id=$2, updated_at=NOW() WHERE id=$1`, id, journalID)
	return err
}

func (r *accountingRepository) ListFinancialPeriods(ctx context.Context) ([]FinancialPeriodRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, start_date, end_date, status, is_locked, closed_at FROM financial_periods ORDER BY start_date DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FinancialPeriodRecord
	for rows.Next() {
		var rec FinancialPeriodRecord
		if err := rows.Scan(&rec.ID, &rec.Name, &rec.StartDate, &rec.EndDate, &rec.Status, &rec.IsLocked, &rec.ClosedAt); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *accountingRepository) GetOpenPeriod(ctx context.Context, date time.Time) (*FinancialPeriodRecord, error) {
	var rec FinancialPeriodRecord
	err := r.pool.QueryRow(ctx, `
SELECT id, name, start_date, end_date, status, is_locked, closed_at FROM financial_periods
WHERE $1 BETWEEN start_date AND end_date AND status='open' LIMIT 1`, date).Scan(
		&rec.ID, &rec.Name, &rec.StartDate, &rec.EndDate, &rec.Status, &rec.IsLocked, &rec.ClosedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &rec, err
}

func (r *accountingRepository) getOpenPeriodTx(ctx context.Context, tx pgx.Tx, date time.Time) (*FinancialPeriodRecord, error) {
	var rec FinancialPeriodRecord
	err := tx.QueryRow(ctx, `
SELECT id, name, start_date, end_date, status, is_locked, closed_at FROM financial_periods
WHERE $1 BETWEEN start_date AND end_date AND status='open' LIMIT 1`, date).Scan(
		&rec.ID, &rec.Name, &rec.StartDate, &rec.EndDate, &rec.Status, &rec.IsLocked, &rec.ClosedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &rec, err
}

func (r *accountingRepository) CreateFinancialPeriod(ctx context.Context, name string, start, end time.Time) (*FinancialPeriodRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO financial_periods (name, start_date, end_date) VALUES ($1,$2,$3) RETURNING id`, name, start, end).Scan(&id)
	if err != nil {
		return nil, err
	}
	var rec FinancialPeriodRecord
	err = r.pool.QueryRow(ctx, `SELECT id, name, start_date, end_date, status, is_locked, closed_at FROM financial_periods WHERE id=$1`, id).
		Scan(&rec.ID, &rec.Name, &rec.StartDate, &rec.EndDate, &rec.Status, &rec.IsLocked, &rec.ClosedAt)
	return &rec, err
}

func (r *accountingRepository) CloseFinancialPeriod(ctx context.Context, id, actorID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
UPDATE financial_periods SET status='closed', is_locked=true, closed_at=NOW(), closed_by=$2, updated_at=NOW()
WHERE id=$1 AND status='open'`, id, actorID)
	return err
}

func (r *accountingRepository) IsPeriodLocked(ctx context.Context, date time.Time) (bool, error) {
	var locked bool
	err := r.pool.QueryRow(ctx, `
SELECT COALESCE(bool_or(is_locked), false) FROM financial_periods
WHERE $1 BETWEEN start_date AND end_date AND status='closed'`, date).Scan(&locked)
	return locked, err
}

func (r *accountingRepository) isPeriodLockedTx(ctx context.Context, tx pgx.Tx, date time.Time) (bool, error) {
	var locked bool
	err := tx.QueryRow(ctx, `
SELECT COALESCE(bool_or(is_locked), false) FROM financial_periods
WHERE $1 BETWEEN start_date AND end_date AND status='closed'`, date).Scan(&locked)
	return locked, err
}

func (r *accountingRepository) SumAccountDebitsCredits(ctx context.Context, accountCode string, from, to time.Time) (float64, float64, error) {
	var debits, credits float64
	err := r.pool.QueryRow(ctx, `
SELECT COALESCE(SUM(jl.debit),0), COALESCE(SUM(jl.credit),0)
FROM journal_lines jl
JOIN journal_entries je ON je.id=jl.journal_entry_id
JOIN accounts a ON a.id=jl.account_id
WHERE a.code=$1 AND je.deleted_at IS NULL AND je.status='posted'
AND je.entry_date BETWEEN $2 AND $3`, accountCode, from, to).Scan(&debits, &credits)
	return debits, credits, err
}

func (r *accountingRepository) MonthlyTrend(ctx context.Context, accountType string, months int) ([]TrendRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT TO_CHAR(je.entry_date, 'YYYY-MM') AS period,
    COALESCE(SUM(CASE WHEN a.account_type='income' THEN jl.credit - jl.debit ELSE jl.debit - jl.credit END), 0)
FROM journal_lines jl
JOIN journal_entries je ON je.id=jl.journal_entry_id
JOIN accounts a ON a.id=jl.account_id
WHERE a.account_type=$1 AND je.deleted_at IS NULL AND je.status='posted'
AND je.entry_date >= DATE_TRUNC('month', CURRENT_DATE) - ($2 || ' months')::INTERVAL
GROUP BY period ORDER BY period`, accountType, months)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TrendRecord
	for rows.Next() {
		var rec TrendRecord
		if err := rows.Scan(&rec.Label, &rec.Amount); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}
