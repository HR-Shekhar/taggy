-- =========================================
-- CREATE
-- =========================================

-- name: CreateEmailVerificationOTP :one
INSERT INTO email_verification_otp (
    user_id,
    otp_hash,
    expires_at
)
VALUES ($1, $2, $3)
RETURNING id, user_id, otp_hash, expires_at, consumed_at, created_at;


-- =========================================
-- GET
-- =========================================

-- name: GetActiveEmailVerificationOTPByUserID :one
SELECT id, user_id, otp_hash, expires_at, consumed_at, created_at
FROM email_verification_otp
WHERE user_id = $1
  AND consumed_at IS NULL
  AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;


-- =========================================
-- UPDATE
-- =========================================

-- name: ConsumeEmailVerificationOTP :one
UPDATE email_verification_otp
SET consumed_at = NOW()
WHERE id = $1
  AND consumed_at IS NULL
RETURNING id, user_id, otp_hash, expires_at, consumed_at, created_at;


-- name: InvalidateActiveEmailVerificationOTPs :exec
UPDATE email_verification_otp
SET consumed_at = NOW()
WHERE user_id = $1
  AND consumed_at IS NULL;
