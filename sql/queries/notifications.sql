-- name: ListNotices :many
SELECT id, title, body, notice_type, target_audience, publish_at, expires_at, is_published, created_by, created_at, updated_at
FROM notices WHERE deleted_at IS NULL
ORDER BY publish_at DESC
LIMIT $1 OFFSET $2;

-- name: CountSMSLogsToday :one
SELECT COUNT(*) FROM sms_logs WHERE created_at >= CURRENT_DATE;

-- name: CountEmailLogsToday :one
SELECT COUNT(*) FROM email_logs WHERE created_at >= CURRENT_DATE;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications WHERE parent_id = $1 AND is_read = false;

-- name: SMSDeliveryRate :one
SELECT
    CASE WHEN COUNT(*) = 0 THEN 0
    ELSE ROUND(100.0 * COUNT(*) FILTER (WHERE status = 'sent') / COUNT(*), 2)
    END AS rate
FROM sms_logs WHERE created_at >= CURRENT_DATE - INTERVAL '30 days';
