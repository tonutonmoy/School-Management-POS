-- name: CreateAuditLog :one
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, ip_address, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListRecentAuditLogs :many
SELECT a.*, u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name
FROM audit_logs a
LEFT JOIN users u ON u.id = a.user_id
ORDER BY a.created_at DESC
LIMIT $1;

-- name: ListAuditLogs :many
SELECT a.*, u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name
FROM audit_logs a
LEFT JOIN users u ON u.id = a.user_id
ORDER BY a.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAuditLogs :one
SELECT COUNT(*)::bigint AS count FROM audit_logs;

-- name: CountDashboardActivitiesByActions :many
SELECT action, COUNT(*)::bigint AS count
FROM audit_logs
WHERE action = ANY($1::text[])
GROUP BY action;
