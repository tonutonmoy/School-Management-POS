-- name: ListDepartments :many
SELECT d.*,
    (SELECT COUNT(*) FROM teachers t WHERE t.department_id = d.id AND t.deleted_at IS NULL) AS teacher_count,
    (SELECT COUNT(*) FROM staffs s WHERE s.department_id = d.id AND s.deleted_at IS NULL) AS staff_count
FROM departments d
WHERE d.deleted_at IS NULL
ORDER BY d.name;

-- name: GetDepartment :one
SELECT * FROM departments WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateDepartment :one
INSERT INTO departments (name, slug, description, dept_type)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateDepartment :one
UPDATE departments SET name = $2, slug = $3, description = $4, dept_type = $5, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteDepartment :exec
UPDATE departments SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL;

-- name: ListDesignations :many
SELECT * FROM designations WHERE deleted_at IS NULL ORDER BY name;

-- name: CreateTeacher :one
INSERT INTO teachers (employee_id, first_name, last_name, gender, joining_date, status, employment_type)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: SearchTeachers :many
SELECT t.* FROM teachers t
WHERE t.deleted_at IS NULL
ORDER BY t.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountTeachers :one
SELECT COUNT(*) FROM teachers WHERE deleted_at IS NULL;

-- name: CreateStaff :one
INSERT INTO staffs (employee_id, name, joining_date, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: SearchStaff :many
SELECT s.* FROM staffs s
WHERE s.deleted_at IS NULL
ORDER BY s.created_at DESC
LIMIT $1 OFFSET $2;

-- name: NextEmployeeSequence :one
INSERT INTO employee_sequences (entity_type, year, last_number)
VALUES ($1, $2, 1)
ON CONFLICT (entity_type, year) DO UPDATE SET last_number = employee_sequences.last_number + 1
RETURNING last_number;
