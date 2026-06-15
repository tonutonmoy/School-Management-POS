-- name: CreateDepartment :one
INSERT INTO departments (name, slug, description) VALUES ($1, $2, $3) RETURNING *;

-- name: ListDepartments :many
SELECT * FROM departments WHERE deleted_at IS NULL ORDER BY name;

-- name: GetDepartmentByID :one
SELECT * FROM departments WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateClass :one
INSERT INTO classes (name, code, description, sort_order) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateClass :one
UPDATE classes SET name=$2, code=$3, description=$4, sort_order=$5, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL RETURNING *;

-- name: SoftDeleteClass :exec
UPDATE classes SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL;

-- name: GetClassByID :one
SELECT * FROM classes WHERE id = $1 AND deleted_at IS NULL;

-- name: ListClasses :many
SELECT * FROM classes WHERE deleted_at IS NULL ORDER BY sort_order, name;

-- name: CreateSection :one
INSERT INTO sections (class_id, name, capacity) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateSection :one
UPDATE sections SET class_id=$2, name=$3, capacity=$4, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL RETURNING *;

-- name: SoftDeleteSection :exec
UPDATE sections SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL;

-- name: GetSectionByID :one
SELECT * FROM sections WHERE id = $1 AND deleted_at IS NULL;

-- name: ListSectionsByClass :many
SELECT * FROM sections WHERE class_id = $1 AND deleted_at IS NULL ORDER BY name;

-- name: ListSections :many
SELECT s.*, c.name AS class_name FROM sections s
JOIN classes c ON c.id = s.class_id
WHERE s.deleted_at IS NULL ORDER BY c.sort_order, s.name;

-- name: CreateSubject :one
INSERT INTO subjects (name, code, description) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateSubject :one
UPDATE subjects SET name=$2, code=$3, description=$4, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL RETURNING *;

-- name: SoftDeleteSubject :exec
UPDATE subjects SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL;

-- name: GetSubjectByID :one
SELECT * FROM subjects WHERE id = $1 AND deleted_at IS NULL;

-- name: ListSubjects :many
SELECT * FROM subjects WHERE deleted_at IS NULL ORDER BY name;

-- name: AssignSubjectToClass :exec
INSERT INTO class_subjects (class_id, subject_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemoveClassSubject :exec
DELETE FROM class_subjects WHERE class_id = $1 AND subject_id = $2;

-- name: ClearClassSubjects :exec
DELETE FROM class_subjects WHERE class_id = $1;

-- name: ListSubjectsByClass :many
SELECT s.* FROM subjects s
JOIN class_subjects cs ON cs.subject_id = s.id
WHERE cs.class_id = $1 AND s.deleted_at IS NULL ORDER BY s.name;
