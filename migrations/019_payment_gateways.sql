-- +goose Up
CREATE TABLE payment_gateways (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,
    is_sandbox BOOLEAN NOT NULL DEFAULT true,
    api_key TEXT,
    api_secret TEXT,
    merchant_id VARCHAR(100),
    store_id VARCHAR(100),
    callback_url TEXT,
    success_url TEXT,
    fail_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_gateway_slug CHECK (slug IN ('bkash', 'nagad', 'sslcommerz'))
);

CREATE UNIQUE INDEX idx_payment_gateways_slug ON payment_gateways (slug) WHERE deleted_at IS NULL;

CREATE TABLE gateway_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    gateway_id UUID NOT NULL REFERENCES payment_gateways(id),
    transaction_ref VARCHAR(50) NOT NULL UNIQUE,
    idempotency_key VARCHAR(100) NOT NULL UNIQUE,
    gateway_ref VARCHAR(200),
    payment_type VARCHAR(30) NOT NULL,
    reference_id UUID NOT NULL,
    student_id UUID REFERENCES students(id) ON DELETE SET NULL,
    parent_id UUID REFERENCES parents(id) ON DELETE SET NULL,
    admission_application_id UUID REFERENCES admission_applications(id) ON DELETE SET NULL,
    bill_ids JSONB,
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'BDT',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_id UUID REFERENCES payments(id) ON DELETE SET NULL,
    signature_verified BOOLEAN NOT NULL DEFAULT false,
    gateway_response JSONB,
    error_message TEXT,
    ip_address VARCHAR(45),
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_gw_payment_type CHECK (payment_type IN ('admission_fee', 'student_fee')),
    CONSTRAINT chk_gw_status CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'refunded', 'cancelled'))
);

CREATE INDEX idx_gw_tx_status ON gateway_transactions (status, created_at DESC);
CREATE INDEX idx_gw_tx_gateway ON gateway_transactions (gateway_id, created_at DESC);
CREATE INDEX idx_gw_tx_student ON gateway_transactions (student_id) WHERE student_id IS NOT NULL;
CREATE INDEX idx_gw_tx_admission ON gateway_transactions (admission_application_id) WHERE admission_application_id IS NOT NULL;
CREATE INDEX idx_gw_tx_payment ON gateway_transactions (payment_id) WHERE payment_id IS NOT NULL;

CREATE TABLE payment_refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    gateway_transaction_id UUID REFERENCES gateway_transactions(id) ON DELETE SET NULL,
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'requested',
    reason TEXT NOT NULL,
    requested_by UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_refund_status CHECK (status IN ('requested', 'approved', 'processed', 'rejected'))
);

CREATE INDEX idx_payment_refunds_status ON payment_refunds (status, created_at DESC);

INSERT INTO payment_gateways (name, slug, is_active, is_sandbox, callback_url, success_url, fail_url) VALUES
    ('bKash', 'bkash', true, true, '/webhooks/payment/bkash', '/payments/success', '/payments/failed'),
    ('Nagad', 'nagad', true, true, '/webhooks/payment/nagad', '/payments/success', '/payments/failed'),
    ('SSLCommerz', 'sslcommerz', true, true, '/webhooks/payment/sslcommerz', '/payments/success', '/payments/failed');

-- +goose Down
DROP TABLE IF EXISTS payment_refunds;
DROP TABLE IF EXISTS gateway_transactions;
DROP TABLE IF EXISTS payment_gateways;
