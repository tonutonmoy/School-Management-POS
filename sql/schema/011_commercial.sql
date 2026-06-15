-- SQLC schema mirror (applied via migrations/021_commercial_deployment.sql)

CREATE TABLE IF NOT EXISTS system_backups (
    id UUID PRIMARY KEY,
    file_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    backup_type VARCHAR(20) NOT NULL DEFAULT 'manual',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    checksum VARCHAR(64),
    verified BOOLEAN NOT NULL DEFAULT false,
    error_message TEXT,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS system_settings (
    id UUID PRIMARY KEY,
    category VARCHAR(50) NOT NULL,
    setting_key VARCHAR(100) NOT NULL,
    value JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS licenses (
    id UUID PRIMARY KEY,
    license_key VARCHAR(100) NOT NULL UNIQUE,
    school_name VARCHAR(200) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS login_attempts (
    id UUID PRIMARY KEY,
    email VARCHAR(200) NOT NULL,
    ip_address VARCHAR(45),
    success BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- name: ListSystemBackups :many
SELECT id, file_name, file_size, backup_type, status, checksum, verified, created_at
FROM system_backups ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CountPendingQueue :one
SELECT COUNT(*) FROM notification_queue WHERE status = 'pending';

-- name: CountFailedLoginsSince :one
SELECT COUNT(*) FROM login_attempts WHERE success = false AND created_at >= $1;
