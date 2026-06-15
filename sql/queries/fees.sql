-- name: ListFeeTypes :many
SELECT * FROM fee_types WHERE deleted_at IS NULL ORDER BY name;

-- name: CreateStudentBill :one
INSERT INTO student_bills (invoice_number, student_id, session_id, class_id, section_id, bill_period, due_date, subtotal, discount_amount, total_amount, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: NextFinanceSequence :one
INSERT INTO finance_sequences (entity_type, year, last_number) VALUES ($1, $2, 1)
ON CONFLICT (entity_type, year) DO UPDATE SET last_number = finance_sequences.last_number + 1
RETURNING last_number;

-- name: SumOutstandingDues :one
SELECT COALESCE(SUM(total_amount - paid_amount), 0) FROM student_bills
WHERE deleted_at IS NULL AND status IN ('pending', 'partial', 'overdue');
