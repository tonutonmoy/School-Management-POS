-- name: NextAdmissionNumber :one
INSERT INTO admission_sequences (year, last_number) VALUES ($1, 1)
ON CONFLICT (year) DO UPDATE SET last_number = admission_sequences.last_number + 1
RETURNING last_number;

-- name: CreateStudent :one
INSERT INTO students (
    admission_number, roll_number, first_name, last_name, date_of_birth, gender,
    blood_group, religion, nationality, photo_url, phone, email, address,
    session_id, class_id, section_id, department_id, admission_date, status
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
RETURNING *;

-- name: UpdateStudent :one
UPDATE students SET
    roll_number=$2, first_name=$3, last_name=$4, date_of_birth=$5, gender=$6,
    blood_group=$7, religion=$8, nationality=$9,
    photo_url=CASE WHEN $10='' THEN photo_url ELSE $10 END,
    phone=$11, email=$12, address=$13,
    session_id=$14, class_id=$15, section_id=$16, department_id=$17,
    admission_date=$18, status=$19, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL RETURNING *;

-- name: SoftDeleteStudent :exec
UPDATE students SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL;

-- name: GetStudentByID :one
SELECT st.*,
    sess.name AS session_name, c.name AS class_name, sec.name AS section_name,
    d.name AS department_name
FROM students st
JOIN academic_sessions sess ON sess.id = st.session_id
JOIN classes c ON c.id = st.class_id
JOIN sections sec ON sec.id = st.section_id
LEFT JOIN departments d ON d.id = st.department_id
WHERE st.id = $1 AND st.deleted_at IS NULL;

-- name: UpsertStudentParents :one
INSERT INTO student_parents (
    student_id, father_name, father_phone, father_occupation,
    mother_name, mother_phone, mother_occupation, guardian_name, guardian_phone
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (student_id) DO UPDATE SET
    father_name=$2, father_phone=$3, father_occupation=$4,
    mother_name=$5, mother_phone=$6, mother_occupation=$7,
    guardian_name=$8, guardian_phone=$9, updated_at=NOW()
RETURNING *;

-- name: GetStudentParents :one
SELECT * FROM student_parents WHERE student_id = $1;

-- name: CreateStudentDocument :one
INSERT INTO student_documents (student_id, doc_type, file_name, file_url)
VALUES ($1,$2,$3,$4) RETURNING *;

-- name: ListStudentDocuments :many
SELECT * FROM student_documents WHERE student_id=$1 AND deleted_at IS NULL ORDER BY created_at;

-- name: SoftDeleteStudentDocument :exec
UPDATE student_documents SET deleted_at=NOW() WHERE id=$1;

-- name: CreateStudentPromotion :one
INSERT INTO student_promotions (
    student_id, promotion_type, from_session_id, to_session_id,
    from_class_id, to_class_id, from_section_id, to_section_id,
    promotion_date, notes, created_by
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING *;

-- name: ListStudentPromotions :many
SELECT * FROM student_promotions WHERE student_id=$1 ORDER BY created_at DESC;

-- name: SearchStudents :many
SELECT st.*,
    sess.name AS session_name, c.name AS class_name, sec.name AS section_name,
    d.name AS department_name
FROM students st
JOIN academic_sessions sess ON sess.id = st.session_id
JOIN classes c ON c.id = st.class_id
JOIN sections sec ON sec.id = st.section_id
LEFT JOIN departments d ON d.id = st.department_id
WHERE st.deleted_at IS NULL
AND ($1::text = '' OR st.admission_number ILIKE '%' || $1 || '%')
AND ($2::text = '' OR st.roll_number ILIKE '%' || $2 || '%')
AND ($3::text = '' OR st.first_name ILIKE '%' || $3 || '%' OR st.last_name ILIKE '%' || $3 || '%')
AND ($4::uuid IS NULL OR st.class_id = $4)
AND ($5::uuid IS NULL OR st.section_id = $5)
AND ($6::uuid IS NULL OR st.session_id = $6)
ORDER BY st.created_at DESC
LIMIT $7 OFFSET $8;

-- name: CountStudents :one
SELECT COUNT(*)::bigint FROM students WHERE deleted_at IS NULL;

-- name: CountActiveStudents :one
SELECT COUNT(*)::bigint FROM students WHERE deleted_at IS NULL AND status = 'active';

-- name: CountNewAdmissionsThisMonth :one
SELECT COUNT(*)::bigint FROM students
WHERE deleted_at IS NULL
AND admission_date >= date_trunc('month', CURRENT_DATE)::date;

-- name: CountStudentsByClass :many
SELECT c.id, c.name, COUNT(st.id)::bigint AS student_count
FROM classes c
LEFT JOIN students st ON st.class_id = c.id AND st.deleted_at IS NULL AND st.status = 'active'
WHERE c.deleted_at IS NULL
GROUP BY c.id, c.name
ORDER BY c.sort_order, c.name;

-- name: CountSearchStudents :one
SELECT COUNT(*)::bigint FROM students st
WHERE st.deleted_at IS NULL
AND ($1::text = '' OR st.admission_number ILIKE '%' || $1 || '%')
AND ($2::text = '' OR st.roll_number ILIKE '%' || $2 || '%')
AND ($3::text = '' OR st.first_name ILIKE '%' || $3 || '%' OR st.last_name ILIKE '%' || $3 || '%')
AND ($4::uuid IS NULL OR st.class_id = $4)
AND ($5::uuid IS NULL OR st.section_id = $5)
AND ($6::uuid IS NULL OR st.session_id = $6);

-- name: ListStudentsForReport :many
SELECT st.*,
    sess.name AS session_name, c.name AS class_name, sec.name AS section_name,
    d.name AS department_name
FROM students st
JOIN academic_sessions sess ON sess.id = st.session_id
JOIN classes c ON c.id = st.class_id
JOIN sections sec ON sec.id = st.section_id
LEFT JOIN departments d ON d.id = st.department_id
WHERE st.deleted_at IS NULL
AND ($1::uuid IS NULL OR st.class_id = $1)
AND ($2::uuid IS NULL OR st.session_id = $2)
AND ($3::text = '' OR st.status = $3)
ORDER BY c.sort_order, sec.name, st.roll_number, st.last_name;

-- name: ListAdmissionsReport :many
SELECT st.*,
    sess.name AS session_name, c.name AS class_name, sec.name AS section_name
FROM students st
JOIN academic_sessions sess ON sess.id = st.session_id
JOIN classes c ON c.id = st.class_id
JOIN sections sec ON sec.id = st.section_id
WHERE st.deleted_at IS NULL
AND st.admission_date >= $1 AND st.admission_date <= $2
ORDER BY st.admission_date DESC;
