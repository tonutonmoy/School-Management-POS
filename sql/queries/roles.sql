-- name: CreateRole :one
INSERT INTO roles (name, slug, description, is_system)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1 AND deleted_at IS NULL;

-- name: GetRoleBySlug :one
SELECT * FROM roles WHERE slug = $1 AND deleted_at IS NULL;

-- name: UpdateRole :one
UPDATE roles
SET name = $2, slug = $3, description = $4, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND is_system = FALSE
RETURNING *;

-- name: SoftDeleteRole :exec
UPDATE roles SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND is_system = FALSE;

-- name: ListRoles :many
SELECT * FROM roles WHERE deleted_at IS NULL ORDER BY name;

-- name: AssignRolePermission :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveRolePermission :exec
DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2;

-- name: ClearRolePermissions :exec
DELETE FROM role_permissions WHERE role_id = $1;

-- name: GetRolePermissions :many
SELECT p.*
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
WHERE rp.role_id = $1
ORDER BY p.module, p.name;

-- name: GetRolePermissionSlugs :many
SELECT p.slug FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
WHERE rp.role_id = $1;
