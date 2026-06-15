-- +goose Up
CREATE TABLE parents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    phone VARCHAR(30),
    address TEXT,
    occupation VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_parents_user ON parents (user_id) WHERE deleted_at IS NULL;

CREATE TABLE parent_students (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID NOT NULL REFERENCES parents(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    relationship VARCHAR(30) NOT NULL DEFAULT 'guardian',
    is_primary BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_parent_relationship CHECK (relationship IN ('father', 'mother', 'guardian', 'other')),
    CONSTRAINT uq_parent_student UNIQUE (parent_id, student_id)
);

CREATE INDEX idx_parent_students_parent ON parent_students (parent_id);
CREATE INDEX idx_parent_students_student ON parent_students (student_id);

CREATE TABLE notices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    notice_type VARCHAR(30) NOT NULL DEFAULT 'general',
    target_audience VARCHAR(30) NOT NULL DEFAULT 'all_parents',
    publish_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    is_published BOOLEAN NOT NULL DEFAULT true,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_notice_type CHECK (notice_type IN ('general', 'exam', 'holiday', 'fee', 'urgent')),
    CONSTRAINT chk_notice_audience CHECK (target_audience IN ('all_parents', 'all_users', 'specific_class'))
);

CREATE INDEX idx_notices_type ON notices (notice_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_notices_published ON notices (is_published, publish_at) WHERE deleted_at IS NULL;

CREATE TABLE notice_reads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notice_id UUID NOT NULL REFERENCES notices(id) ON DELETE CASCADE,
    parent_id UUID NOT NULL REFERENCES parents(id) ON DELETE CASCADE,
    read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_notice_read UNIQUE (notice_id, parent_id)
);

CREATE INDEX idx_notice_reads_parent ON notice_reads (parent_id);

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID REFERENCES parents(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    category VARCHAR(50) NOT NULL,
    reference_type VARCHAR(50),
    reference_id UUID,
    is_read BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_parent ON notifications (parent_id, is_read, created_at DESC);
CREATE INDEX idx_notifications_user ON notifications (user_id, is_read, created_at DESC);

CREATE TABLE sms_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_phone VARCHAR(30) NOT NULL,
    message TEXT NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    provider VARCHAR(50) NOT NULL DEFAULT 'log',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reference_type VARCHAR(50),
    reference_id UUID,
    parent_id UUID REFERENCES parents(id) ON DELETE SET NULL,
    error_message TEXT,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_sms_status CHECK (status IN ('pending', 'sent', 'failed', 'queued'))
);

CREATE INDEX idx_sms_logs_status ON sms_logs (status, created_at DESC);
CREATE INDEX idx_sms_logs_event ON sms_logs (event_type);

CREATE TABLE email_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_email VARCHAR(255) NOT NULL,
    subject VARCHAR(300) NOT NULL,
    body TEXT NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    provider VARCHAR(50) NOT NULL DEFAULT 'log',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reference_type VARCHAR(50),
    reference_id UUID,
    parent_id UUID REFERENCES parents(id) ON DELETE SET NULL,
    error_message TEXT,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_email_status CHECK (status IN ('pending', 'sent', 'failed', 'queued'))
);

CREATE INDEX idx_email_logs_status ON email_logs (status, created_at DESC);
CREATE INDEX idx_email_logs_event ON email_logs (event_type);

CREATE TABLE notification_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel VARCHAR(20) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_queue_channel CHECK (channel IN ('sms', 'email', 'push')),
    CONSTRAINT chk_queue_status CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
);

CREATE INDEX idx_notification_queue_pending ON notification_queue (status, scheduled_at) WHERE status = 'pending';

-- +goose Down
DROP TABLE IF EXISTS notification_queue;
DROP TABLE IF EXISTS email_logs;
DROP TABLE IF EXISTS sms_logs;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS notice_reads;
DROP TABLE IF EXISTS notices;
DROP TABLE IF EXISTS parent_students;
DROP TABLE IF EXISTS parents;
