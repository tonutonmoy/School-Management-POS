-- name: CreateAcademicSession :one
INSERT INTO academic_sessions (name, start_date, end_date, is_active)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAcademicSessionByID :one
SELECT * FROM academic_sessions WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateAcademicSession :one
UPDATE academic_sessions
SET name = $2, start_date = $3, end_date = $4, is_active = $5, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteAcademicSession :exec
UPDATE academic_sessions SET deleted_at = NOW(), updated_at = NOW(), is_active = FALSE
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListAcademicSessions :many
SELECT * FROM academic_sessions WHERE deleted_at IS NULL ORDER BY start_date DESC;

-- name: GetActiveAcademicSession :one
SELECT * FROM academic_sessions
WHERE is_active = TRUE AND deleted_at IS NULL
LIMIT 1;

-- name: DeactivateAllAcademicSessions :exec
UPDATE academic_sessions SET is_active = FALSE, updated_at = NOW()
WHERE deleted_at IS NULL AND is_active = TRUE;

-- name: SetAcademicSessionActive :one
UPDATE academic_sessions SET is_active = TRUE, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;
