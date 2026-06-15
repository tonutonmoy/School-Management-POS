-- name: GetAdmissionByNumber :one
SELECT id, application_number, status FROM admission_applications
WHERE application_number = $1 AND deleted_at IS NULL;

-- name: CountAdmissionsByStatus :one
SELECT COUNT(*) FROM admission_applications WHERE status = $1 AND deleted_at IS NULL;
