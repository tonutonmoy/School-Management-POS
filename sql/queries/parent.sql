-- name: GetParentByUserID :one
SELECT id, user_id, phone, address, occupation, created_at, updated_at
FROM parents WHERE user_id = $1 AND deleted_at IS NULL;

-- name: GetParentByID :one
SELECT id, user_id, phone, address, occupation, created_at, updated_at
FROM parents WHERE id = $1 AND deleted_at IS NULL;

-- name: ListParentStudents :many
SELECT ps.id, ps.parent_id, ps.student_id, ps.relationship, ps.is_primary, ps.created_at,
       s.first_name, s.last_name, s.admission_number, s.roll_number, c.name AS class_name
FROM parent_students ps
JOIN students s ON s.id = ps.student_id AND s.deleted_at IS NULL
LEFT JOIN classes c ON c.id = s.class_id
WHERE ps.parent_id = $1
ORDER BY ps.is_primary DESC, s.first_name;

-- name: ParentHasStudent :one
SELECT EXISTS(
    SELECT 1 FROM parent_students ps
    JOIN parents p ON p.id = ps.parent_id AND p.deleted_at IS NULL
    WHERE p.user_id = $1 AND ps.student_id = $2
) AS has_access;

-- name: ListParents :many
SELECT p.id, p.user_id, p.phone, p.address, p.occupation, p.created_at, p.updated_at,
       u.email, u.first_name, u.last_name, u.is_active,
       (SELECT COUNT(*) FROM parent_students ps WHERE ps.parent_id = p.id) AS child_count
FROM parents p
JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
WHERE p.deleted_at IS NULL
ORDER BY u.last_name, u.first_name
LIMIT $1 OFFSET $2;

-- name: CountParents :one
SELECT COUNT(*) FROM parents WHERE deleted_at IS NULL;
