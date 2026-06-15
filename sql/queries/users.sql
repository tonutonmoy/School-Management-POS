-- name: GetUserByEmail :one
SELECT
    u.id, u.email, u.password_hash, u.first_name, u.last_name, u.phone,
    u.role_id, u.is_active, u.last_login_at, u.created_at, u.updated_at, u.deleted_at,
    r.name AS role_name,
    r.slug AS role_slug
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE LOWER(u.email) = LOWER($1) AND u.deleted_at IS NULL;

-- name: GetUserByID :one
SELECT
    u.id, u.email, u.password_hash, u.first_name, u.last_name, u.phone,
    u.role_id, u.is_active, u.last_login_at, u.created_at, u.updated_at, u.deleted_at,
    r.name AS role_name,
    r.slug AS role_slug
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.id = $1 AND u.deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, first_name, last_name, phone, role_id, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET email = $2,
    first_name = $3,
    last_name = $4,
    phone = $5,
    role_id = $6,
    is_active = $7,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteUser :exec
UPDATE users SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SetUserActive :one
UPDATE users SET is_active = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserLastLogin :exec
UPDATE users SET last_login_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: ListUsers :many
SELECT
    u.id, u.email, u.password_hash, u.first_name, u.last_name, u.phone,
    u.role_id, u.is_active, u.last_login_at, u.created_at, u.updated_at, u.deleted_at,
    r.name AS role_name,
    r.slug AS role_slug
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.deleted_at IS NULL
ORDER BY u.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*)::bigint AS count FROM users WHERE deleted_at IS NULL;

-- name: CountUsersByRoleSlug :one
SELECT COUNT(*)::bigint AS count
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.deleted_at IS NULL AND r.slug = $1;

-- name: GetUserPermissions :many
SELECT p.slug
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
JOIN users u ON u.role_id = rp.role_id
WHERE u.id = $1;
