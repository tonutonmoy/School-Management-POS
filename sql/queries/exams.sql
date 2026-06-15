-- name: ListExams :many
SELECT * FROM exams WHERE deleted_at IS NULL ORDER BY start_date DESC;

-- name: UpsertStudentMark :one
INSERT INTO student_marks (exam_id, exam_subject_id, student_id, written_score, mcq_score, practical_score, total_score, entered_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (exam_subject_id, student_id) DO UPDATE SET
    written_score = EXCLUDED.written_score, mcq_score = EXCLUDED.mcq_score,
    practical_score = EXCLUDED.practical_score, total_score = EXCLUDED.total_score,
    entered_by = EXCLUDED.entered_by, updated_at = NOW()
RETURNING *;

-- name: CountExamResults :one
SELECT COUNT(*) FROM exam_results WHERE exam_id = $1 AND is_passed = $2;
