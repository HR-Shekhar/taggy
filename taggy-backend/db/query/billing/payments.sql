-- name: CreatePayment :one
INSERT INTO payments (
    user_id,
    razorpay_order_id,
    amount,
    currency,
    status,
    receipt
) VALUES (
    $1, $2, $3, $4, 'CREATED', $5
)
RETURNING *;

-- name: GetPaymentByOrderID :one
SELECT * FROM payments
WHERE razorpay_order_id = $1;

-- name: GetPaymentByPaymentID :one
SELECT * FROM payments
WHERE razorpay_payment_id = $1;

-- name: MarkPaymentPaid :one
UPDATE payments
SET
    status = 'PAID',
    razorpay_payment_id = $2,
    updated_at = NOW()
WHERE razorpay_order_id = $1
  AND status = 'CREATED'
RETURNING *;

-- name: MarkPaymentPaidIdempotent :one
UPDATE payments
SET
    status = 'PAID',
    razorpay_payment_id = COALESCE(razorpay_payment_id, $2),
    updated_at = NOW()
WHERE razorpay_order_id = $1
  AND (status = 'CREATED' OR (status = 'PAID' AND (razorpay_payment_id IS NULL OR razorpay_payment_id = $2)))
RETURNING *;
