-- +goose Up
CREATE TABLE system_backups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    backup_type VARCHAR(20) NOT NULL DEFAULT 'manual',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    checksum VARCHAR(64),
    verified BOOLEAN NOT NULL DEFAULT false,
    error_message TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_backup_type CHECK (backup_type IN ('manual', 'scheduled')),
    CONSTRAINT chk_backup_status CHECK (status IN ('pending', 'completed', 'failed'))
);

CREATE INDEX idx_system_backups_created ON system_backups (created_at DESC);
CREATE INDEX idx_system_backups_status ON system_backups (status);

CREATE TABLE system_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category VARCHAR(50) NOT NULL,
    setting_key VARCHAR(100) NOT NULL,
    value JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_system_settings UNIQUE (category, setting_key)
);

CREATE INDEX idx_system_settings_category ON system_settings (category);

CREATE TABLE licenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_key VARCHAR(100) NOT NULL UNIQUE,
    school_name VARCHAR(200) NOT NULL,
    school_code VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    activated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    registered_email VARCHAR(200),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_license_status CHECK (status IN ('active', 'expired', 'revoked', 'pending'))
);

CREATE INDEX idx_licenses_status ON licenses (status);
CREATE INDEX idx_licenses_expires ON licenses (expires_at) WHERE expires_at IS NOT NULL;

CREATE TABLE login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(200) NOT NULL,
    ip_address VARCHAR(45),
    success BOOLEAN NOT NULL DEFAULT false,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_login_attempts_email ON login_attempts (email, created_at DESC);
CREATE INDEX idx_login_attempts_ip ON login_attempts (ip_address, created_at DESC);
CREATE INDEX idx_login_attempts_failed ON login_attempts (created_at DESC) WHERE success = false;

CREATE TABLE email_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    subject VARCHAR(200) NOT NULL,
    body_html TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO system_settings (category, setting_key, value) VALUES
    ('general', 'installed', '{"value": true}'),
    ('general', 'backup_schedule', '{"enabled": false, "cron": "0 2 * * *", "retention_days": 30}'),
    ('security', 'password_policy', '{"min_length": 8, "require_uppercase": true, "require_number": true, "require_special": false, "max_failed_attempts": 5}'),
    ('branding', 'white_label', '{"app_name": "School Management System", "logo_url": "", "favicon_url": "", "primary_color": "#4f46e5", "email_footer": ""}'),
    ('smtp', 'config', '{"host": "", "port": 587, "username": "", "password": "", "from": "", "enabled": false}'),
    ('sms', 'config', '{"provider": "log", "api_key": "", "sender_id": "", "enabled": false}');

INSERT INTO email_templates (slug, name, subject, body_html) VALUES
    ('payment_receipt', 'Payment Receipt', 'Payment Received - {{school_name}}', '<p>Dear Parent,</p><p>We received your payment of {{amount}}.</p><p>Receipt: {{receipt_number}}</p><p>{{footer}}</p>'),
    ('welcome', 'Welcome Email', 'Welcome to {{school_name}}', '<p>Welcome to {{school_name}} parent portal.</p><p>{{footer}}</p>'),
    ('password_reset', 'Password Reset', 'Reset Your Password', '<p>Click the link to reset your password: {{reset_url}}</p><p>{{footer}}</p>');

-- +goose Down
DROP TABLE IF EXISTS email_templates;
DROP TABLE IF EXISTS login_attempts;
DROP TABLE IF EXISTS licenses;
DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS system_backups;
