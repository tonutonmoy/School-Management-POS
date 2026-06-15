-- name: UpsertStudentAttendance :one
INSERT INTO student_attendance (student_id, session_id, class_id, section_id, attendance_date, status, marked_by, remarks)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (student_id, attendance_date) WHERE deleted_at IS NULL
DO UPDATE SET status = EXCLUDED.status, remarks = EXCLUDED.remarks, marked_by = EXCLUDED.marked_by, updated_at = NOW()
RETURNING *;

-- name: ListStudentAttendanceByDate :many
SELECT sa.*, s.first_name, s.last_name, s.admission_number, s.roll_number
FROM student_attendance sa
JOIN students s ON s.id = sa.student_id
WHERE sa.deleted_at IS NULL AND sa.attendance_date = $1
ORDER BY s.first_name, s.last_name;

-- name: CountStudentAttendanceByStatus :one
SELECT COUNT(*) FROM student_attendance
WHERE deleted_at IS NULL AND attendance_date = $1 AND status = $2;

-- name: UpsertTeacherAttendance :one
INSERT INTO teacher_attendance (teacher_id, attendance_date, status, marked_by, remarks)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (teacher_id, attendance_date) WHERE deleted_at IS NULL
DO UPDATE SET status = EXCLUDED.status, remarks = EXCLUDED.remarks, marked_by = EXCLUDED.marked_by, updated_at = NOW()
RETURNING *;

-- name: UpsertStaffAttendance :one
INSERT INTO staff_attendance (staff_id, attendance_date, status, marked_by, remarks)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (staff_id, attendance_date) WHERE deleted_at IS NULL
DO UPDATE SET status = EXCLUDED.status, remarks = EXCLUDED.remarks, marked_by = EXCLUDED.marked_by, updated_at = NOW()
RETURNING *;

-- name: CreateLeaveRequest :one
INSERT INTO leave_requests (entity_type, teacher_id, staff_id, leave_type, start_date, end_date, reason, applied_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: StudentAttendanceSummary :one
SELECT
    COUNT(*) FILTER (WHERE status = 'present') AS present_days,
    COUNT(*) FILTER (WHERE status = 'absent') AS absent_days,
    COUNT(*) FILTER (WHERE status = 'late') AS late_days,
    COUNT(*) FILTER (WHERE status = 'leave') AS leave_days
FROM student_attendance
WHERE student_id = $1 AND deleted_at IS NULL;
