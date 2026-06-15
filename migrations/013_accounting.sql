-- +goose Up
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) NOT NULL,
    name VARCHAR(150) NOT NULL,
    account_type VARCHAR(20) NOT NULL,
    parent_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_account_type CHECK (account_type IN ('assets', 'liabilities', 'equity', 'income', 'expenses'))
);

CREATE UNIQUE INDEX idx_accounts_code_active ON accounts (code) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_type ON accounts (account_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_parent ON accounts (parent_id) WHERE deleted_at IS NULL;

CREATE TABLE accounting_sequences (
    entity_type VARCHAR(20) NOT NULL,
    year INT NOT NULL,
    last_number INT NOT NULL DEFAULT 0,
    PRIMARY KEY (entity_type, year)
);

CREATE TABLE financial_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    is_locked BOOLEAN NOT NULL DEFAULT false,
    closed_at TIMESTAMPTZ,
    closed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_period_status CHECK (status IN ('open', 'closed')),
    CONSTRAINT chk_period_dates CHECK (end_date >= start_date)
);

CREATE INDEX idx_financial_periods_dates ON financial_periods (start_date, end_date);

CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_number VARCHAR(30) NOT NULL,
    entry_date DATE NOT NULL,
    description TEXT NOT NULL,
    source_type VARCHAR(30) NOT NULL DEFAULT 'manual',
    source_id UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'posted',
    period_id UUID REFERENCES financial_periods(id) ON DELETE SET NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    posted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_journal_status CHECK (status IN ('draft', 'posted', 'reversed')),
    CONSTRAINT chk_journal_source CHECK (source_type IN ('manual', 'fee_payment', 'fee_refund', 'expense', 'income'))
);

CREATE UNIQUE INDEX idx_journal_entry_number ON journal_entries (entry_number) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_journal_source_unique ON journal_entries (source_type, source_id) WHERE source_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_journal_entry_date ON journal_entries (entry_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_journal_period ON journal_entries (period_id) WHERE deleted_at IS NULL;

CREATE TABLE journal_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id),
    debit NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (debit >= 0),
    credit NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (credit >= 0),
    description TEXT,
    line_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_journal_line_amount CHECK (debit > 0 OR credit > 0),
    CONSTRAINT chk_journal_line_not_both CHECK (NOT (debit > 0 AND credit > 0))
);

CREATE INDEX idx_journal_lines_entry ON journal_lines (journal_entry_id);
CREATE INDEX idx_journal_lines_account ON journal_lines (account_id);

CREATE TABLE expense_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_expense_categories_slug ON expense_categories (slug) WHERE deleted_at IS NULL;

CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES expense_categories(id),
    expense_account_id UUID NOT NULL REFERENCES accounts(id),
    pay_from_account_id UUID NOT NULL REFERENCES accounts(id),
    amount NUMERIC(14,2) NOT NULL CHECK (amount > 0),
    expense_date DATE NOT NULL,
    description TEXT NOT NULL,
    payment_method VARCHAR(20) NOT NULL DEFAULT 'cash',
    status VARCHAR(20) NOT NULL DEFAULT 'pending_approval',
    attachment_url TEXT,
    journal_entry_id UUID REFERENCES journal_entries(id) ON DELETE SET NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_expense_status CHECK (status IN ('draft', 'pending_approval', 'approved', 'paid', 'rejected')),
    CONSTRAINT chk_expense_method CHECK (payment_method IN ('cash', 'bank'))
);

CREATE INDEX idx_expenses_date ON expenses (expense_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_expenses_status ON expenses (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_expenses_category ON expenses (category_id) WHERE deleted_at IS NULL;

CREATE TABLE income_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    income_account_id UUID NOT NULL REFERENCES accounts(id),
    receive_to_account_id UUID NOT NULL REFERENCES accounts(id),
    amount NUMERIC(14,2) NOT NULL CHECK (amount > 0),
    income_date DATE NOT NULL,
    source VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    journal_entry_id UUID REFERENCES journal_entries(id) ON DELETE SET NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_income_source CHECK (source IN ('donation', 'event', 'admission_form', 'misc'))
);

CREATE INDEX idx_income_entries_date ON income_entries (income_date) WHERE deleted_at IS NULL;

-- Default chart of accounts
INSERT INTO accounts (code, name, account_type, is_system) VALUES
    ('1000', 'Cash', 'assets', true),
    ('1010', 'Bank', 'assets', true),
    ('1100', 'Accounts Receivable', 'assets', true),
    ('3000', 'Retained Earnings', 'equity', true),
    ('4000', 'Tuition Income', 'income', true),
    ('4010', 'Exam Fee Income', 'income', true),
    ('4020', 'Transport Income', 'income', true),
    ('4030', 'Library Income', 'income', true),
    ('5000', 'Salary Expense', 'expenses', true),
    ('5010', 'Utility Expense', 'expenses', true),
    ('5090', 'Misc Expense', 'expenses', true);

-- Expense categories linked to accounts
INSERT INTO expense_categories (name, slug, account_id)
SELECT 'Salary', 'salary', id FROM accounts WHERE code = '5000';
INSERT INTO expense_categories (name, slug, account_id)
SELECT 'Utilities', 'utilities', id FROM accounts WHERE code = '5010';
INSERT INTO expense_categories (name, slug, account_id)
SELECT 'Maintenance', 'maintenance', id FROM accounts WHERE code = '5090';
INSERT INTO expense_categories (name, slug, account_id)
SELECT 'Office Supplies', 'office_supplies', id FROM accounts WHERE code = '5090';
INSERT INTO expense_categories (name, slug, account_id)
SELECT 'Marketing', 'marketing', id FROM accounts WHERE code = '5090';
INSERT INTO expense_categories (name, slug, account_id)
SELECT 'Miscellaneous', 'miscellaneous', id FROM accounts WHERE code = '5090';

-- Default open financial period (current year)
INSERT INTO financial_periods (name, start_date, end_date, status)
VALUES (
    'FY ' || EXTRACT(YEAR FROM CURRENT_DATE)::TEXT,
    DATE_TRUNC('year', CURRENT_DATE)::DATE,
    (DATE_TRUNC('year', CURRENT_DATE) + INTERVAL '1 year' - INTERVAL '1 day')::DATE,
    'open'
);

-- +goose Down
DROP TABLE IF EXISTS income_entries;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS expense_categories;
DROP TABLE IF EXISTS journal_lines;
DROP TABLE IF EXISTS journal_entries;
DROP TABLE IF EXISTS financial_periods;
DROP TABLE IF EXISTS accounting_sequences;
DROP TABLE IF EXISTS accounts;
