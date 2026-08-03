-- =========================================
-- GET
-- =========================================

-- name: GetIdentityByID :one
SELECT *
FROM user_identity
WHERE id = $1;


-- name: GetIdentityByUserID :many
SELECT *
FROM user_identity
WHERE user_id = $1;


-- name: GetIdentityByProvider :one
SELECT *
FROM user_identity
WHERE user_id = $1
  AND provider = $2;


-- name: GetIdentityByProviderUserID :one
SELECT *
FROM user_identity
WHERE provider = $1
  AND provider_user_id = $2;


-- name: GetIdentityByEmail :one
SELECT ui.*
FROM user_identity ui
INNER JOIN users u ON u.id = ui.user_id
WHERE u.email = $1
  AND ui.provider = 'local'
  AND u.is_deleted = FALSE;


-- =========================================
-- CREATE
-- =========================================

-- name: CreateIdentity :one
INSERT INTO user_identity (
    user_id,
    provider,
    provider_user_id,
    password_hash
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;


-- =========================================
-- UPDATE
-- =========================================

-- name: UpdatePasswordHash :one
UPDATE user_identity
SET
    password_hash = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: UpdateProviderUserID :one
UPDATE user_identity
SET
    provider_user_id = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- =========================================
-- DELETE
-- =========================================

-- Intentionally omitted.
-- user_identity records are never physically deleted.