-- +goose Up
CREATE TABLE fee_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_fee_types_slug_active ON fee_types (slug) WHERE deleted_at IS NULL;

CREATE TABLE fee_structures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fee_type_id UUID NOT NULL REFERENCES fee_types(id),
    session_id UUID NOT NULL REFERENCES academic_sessions(id),
    class_id UUID NOT NULL REFERENCES classes(id),
    section_id UUID REFERENCES sections(id) ON DELETE SET NULL,
    amount NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
    due_day INT NOT NULL DEFAULT 10 CHECK (due_day BETWEEN 1 AND 28),
    frequency VARCHAR(20) NOT NULL DEFAULT 'monthly',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_fee_frequency CHECK (frequency IN ('one_time', 'monthly', 'quarterly', 'half_yearly', 'yearly'))
);

CREATE INDEX idx_fee_structures_session_class ON fee_structures (session_id, class_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_fee_structures_fee_type ON fee_structures (fee_type_id) WHERE deleted_at IS NULL;

CREATE TABLE student_discounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    session_id UUID NOT NULL REFERENCES academic_sessions(id),
    discount_type VARCHAR(20) NOT NULL,
    discount_value NUMERIC(12,2) NOT NULL CHECK (discount_value >= 0),
    reason VARCHAR(50) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_discount_type CHECK (discount_type IN ('fixed', 'percentage')),
    CONSTRAINT chk_discount_reason CHECK (reason IN ('scholarship', 'waiver', 'sibling', 'special'))
);

CREATE INDEX idx_student_discounts_student ON student_discounts (student_id, session_id) WHERE deleted_at IS NULL AND is_active = true;

CREATE TABLE finance_sequences (
    entity_type VARCHAR(20) NOT NULL,
    year INT NOT NULL,
    last_number INT NOT NULL DEFAULT 0,
    PRIMARY KEY (entity_type, year)
);

CREATE TABLE student_bills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(30) NOT NULL,
    student_id UUID NOT NULL REFERENCES students(id),
    session_id UUID NOT NULL REFERENCES academic_sessions(id),
    class_id UUID NOT NULL REFERENCES classes(id),
    section_id UUID NOT NULL REFERENCES sections(id),
    bill_period VARCHAR(20) NOT NULL,
    due_date DATE NOT NULL,
    subtotal NUMERIC(12,2) NOT NULL DEFAULT 0,
    discount_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    paid_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_bill_status CHECK (status IN ('pending', 'partial', 'paid', 'overdue', 'cancelled'))
);

CREATE UNIQUE INDEX idx_student_bills_invoice ON student_bills (invoice_number) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_student_bills_period ON student_bills (student_id, bill_period) WHERE deleted_at IS NULL AND status != 'cancelled';
CREATE INDEX idx_student_bills_student ON student_bills (student_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_student_bills_status ON student_bills (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_student_bills_due ON student_bills (due_date) WHERE deleted_at IS NULL;

CREATE TABLE bill_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bill_id UUID NOT NULL REFERENCES student_bills(id) ON DELETE CASCADE,
    fee_type_id UUID NOT NULL REFERENCES fee_types(id),
    fee_structure_id UUID REFERENCES fee_structures(id) ON DELETE SET NULL,
    description VARCHAR(200) NOT NULL,
    amount NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bill_items_bill ON bill_items (bill_id);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_number VARCHAR(30) NOT NULL,
    student_id UUID NOT NULL REFERENCES students(id),
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    payment_method VARCHAR(20) NOT NULL,
    collected_by UUID REFERENCES users(id) ON DELETE SET NULL,
    collection_date DATE NOT NULL DEFAULT CURRENT_DATE,
    remarks TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_payment_method CHECK (payment_method IN ('cash', 'bank', 'bkash', 'nagad', 'rocket', 'card')),
    CONSTRAINT chk_payment_status CHECK (status IN ('completed', 'refunded'))
);

CREATE UNIQUE INDEX idx_payments_number ON payments (payment_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_payments_student ON payments (student_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_payments_date ON payments (collection_date) WHERE deleted_at IS NULL;

CREATE TABLE payment_allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    bill_id UUID NOT NULL REFERENCES student_bills(id),
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payment_allocations_payment ON payment_allocations (payment_id);
CREATE INDEX idx_payment_allocations_bill ON payment_allocations (bill_id);

CREATE TABLE receipts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    receipt_number VARCHAR(30) NOT NULL,
    payment_id UUID NOT NULL REFERENCES payments(id),
    student_id UUID NOT NULL REFERENCES students(id),
    total_amount NUMERIC(12,2) NOT NULL,
    qr_token VARCHAR(64) NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    issued_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_receipts_number ON receipts (receipt_number);
CREATE UNIQUE INDEX idx_receipts_qr ON receipts (qr_token);
CREATE INDEX idx_receipts_payment ON receipts (payment_id);
CREATE INDEX idx_receipts_student ON receipts (student_id);

INSERT INTO fee_types (name, slug, description) VALUES
    ('Admission Fee', 'admission-fee', 'One-time admission charge'),
    ('Monthly Tuition Fee', 'monthly-tuition', 'Monthly tuition'),
    ('Exam Fee', 'exam-fee', 'Examination fee'),
    ('Registration Fee', 'registration-fee', 'Annual registration'),
    ('Transport Fee', 'transport-fee', 'School transport'),
    ('Library Fee', 'library-fee', 'Library access'),
    ('Hostel Fee', 'hostel-fee', 'Hostel accommodation'),
    ('Development Fee', 'development-fee', 'Infrastructure development'),
    ('Fine', 'fine', 'Penalty / fine');

-- +goose Down
DROP TABLE IF EXISTS receipts;
DROP TABLE IF EXISTS payment_allocations;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS bill_items;
DROP TABLE IF EXISTS student_bills;
DROP TABLE IF EXISTS finance_sequences;
DROP TABLE IF EXISTS student_discounts;
DROP TABLE IF EXISTS fee_structures;
DROP TABLE IF EXISTS fee_types;
