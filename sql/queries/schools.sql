-- name: GetSchool :one
SELECT * FROM schools ORDER BY created_at ASC LIMIT 1;

-- name: UpsertSchool :one
INSERT INTO schools (name, logo_url, address, email, phone, website)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT DO NOTHING
RETURNING *;

-- name: UpdateSchool :one
UPDATE schools
SET name = $2,
    logo_url = COALESCE($3, logo_url),
    address = $4,
    email = $5,
    phone = $6,
    website = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CreateSchool :one
INSERT INTO schools (name, logo_url, address, email, phone, website)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
