-- =========================================
-- PENDING REGISTRATIONS (pre-OTP)
-- =========================================

-- name: UpsertPendingRegistration :one
INSERT INTO pending_registrations (
    email,
    username,
    name,
    password_hash,
    otp_hash,
    expires_at
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (email) DO UPDATE
SET
    username = EXCLUDED.username,
    name = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash,
    otp_hash = EXCLUDED.otp_hash,
    expires_at = EXCLUDED.expires_at,
    updated_at = NOW()
RETURNING *;

-- name: GetPendingRegistrationByEmail :one
SELECT *
FROM pending_registrations
WHERE email = $1;

-- name: GetActivePendingRegistrationByEmail :one
SELECT *
FROM pending_registrations
WHERE email = $1
  AND expires_at > NOW();

-- name: GetActivePendingRegistrationByUsername :one
SELECT *
FROM pending_registrations
WHERE username = $1
  AND expires_at > NOW();

-- name: UpdatePendingRegistrationOTP :one
UPDATE pending_registrations
SET
    otp_hash = $2,
    expires_at = $3,
    updated_at = NOW()
WHERE email = $1
RETURNING *;

-- name: DeletePendingRegistrationByID :exec
DELETE FROM pending_registrations
WHERE id = $1;

-- name: DeleteExpiredPendingRegistrations :exec
DELETE FROM pending_registrations
WHERE expires_at <= NOW();
