-- name: GetPaymentGatewayBySlug :one
SELECT id, name, slug, is_active, is_sandbox, api_key, api_secret, merchant_id, store_id,
    callback_url, success_url, fail_url, created_at, updated_at
FROM payment_gateways
WHERE slug = $1 AND deleted_at IS NULL;

-- name: ListPaymentGateways :many
SELECT id, name, slug, is_active, is_sandbox, merchant_id, store_id,
    callback_url, success_url, fail_url, updated_at
FROM payment_gateways
WHERE deleted_at IS NULL
ORDER BY name;

-- name: GetGatewayTransactionByRef :one
SELECT id, gateway_id, transaction_ref, idempotency_key, gateway_ref, payment_type, reference_id,
    student_id, parent_id, admission_application_id, bill_ids, amount, currency, status,
    payment_id, signature_verified, error_message, completed_at, created_at
FROM gateway_transactions
WHERE transaction_ref = $1;

-- name: GetGatewayTransactionByIdempotency :one
SELECT id, transaction_ref, status, payment_id
FROM gateway_transactions
WHERE idempotency_key = $1;

-- name: CountGatewayTransactions :one
SELECT COUNT(*) FROM gateway_transactions gt
JOIN payment_gateways pg ON pg.id = gt.gateway_id
WHERE ($1::text = '' OR gt.status = $1)
AND ($2::text = '' OR pg.slug = $2)
AND ($3::text = '' OR gt.payment_type = $3)
AND ($4::timestamptz IS NULL OR gt.created_at >= $4)
AND ($5::timestamptz IS NULL OR gt.created_at <= $5);

-- name: PaymentDashboardStats :one
SELECT
    COALESCE(SUM(CASE WHEN status = 'completed' AND completed_at::date = CURRENT_DATE THEN amount ELSE 0 END), 0)::float8 AS today_collection,
    COALESCE(SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END), 0)::float8 AS gateway_collection,
    COUNT(*) FILTER (WHERE status = 'failed')::bigint AS failed_payments,
    COUNT(*) FILTER (WHERE status IN ('pending', 'processing'))::bigint AS pending_payments,
    COUNT(*) FILTER (WHERE created_at::date = CURRENT_DATE)::bigint AS today_transactions
FROM gateway_transactions;

-- name: GatewayCollectionReport :many
SELECT pg.slug, pg.name, COUNT(*)::bigint AS count, COALESCE(SUM(gt.amount), 0)::float8 AS total_amount
FROM gateway_transactions gt
JOIN payment_gateways pg ON pg.id = gt.gateway_id
WHERE gt.status = 'completed'
AND ($1::timestamptz IS NULL OR gt.completed_at >= $1)
AND ($2::timestamptz IS NULL OR gt.completed_at <= $2)
GROUP BY pg.slug, pg.name
ORDER BY total_amount DESC;
