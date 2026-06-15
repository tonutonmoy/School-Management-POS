-- name: ListPermissions :many
SELECT * FROM permissions ORDER BY module, name;

-- name: GetPermissionByID :one
SELECT * FROM permissions WHERE id = $1;

-- name: GetPermissionBySlug :one
SELECT * FROM permissions WHERE slug = $1;

-- name: CreatePermission :one
INSERT INTO permissions (name, slug, description, module)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdatePermission :one
UPDATE permissions
SET name = $2, slug = $3, description = $4, module = $5, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = $1;
