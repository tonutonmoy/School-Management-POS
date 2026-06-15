-- +goose Up
-- SQLC schema mirror for payment gateways (applied via migrations/019_payment_gateways.sql)

CREATE TABLE IF NOT EXISTS payment_gateways (
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
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS gateway_transactions (
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
    amount NUMERIC(12,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'BDT',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_id UUID REFERENCES payments(id) ON DELETE SET NULL,
    signature_verified BOOLEAN NOT NULL DEFAULT false,
    gateway_response JSONB,
    error_message TEXT,
    ip_address VARCHAR(45),
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    gateway_transaction_id UUID REFERENCES gateway_transactions(id) ON DELETE SET NULL,
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    amount NUMERIC(12,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'requested',
    reason TEXT NOT NULL,
    requested_by UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
