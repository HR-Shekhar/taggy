-- =========================================
-- GET
-- =========================================

-- name: GetSessionByRefreshHash :one
SELECT *
FROM user_sessions
WHERE refresh_token_hash = $1;


-- =========================================
-- CREATE
-- =========================================

-- name: CreateSession :one
INSERT INTO user_sessions (
    user_id,
    refresh_token_hash,
    user_agent,
    ip_address,
    expires_at
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;


-- =========================================
-- UPDATE
-- =========================================

-- name: RotateRefreshToken :one
UPDATE user_sessions
SET
    refresh_token_hash = $2,
    expires_at = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- =========================================
-- DELETE
-- =========================================

-- name: DeleteSession :exec
DELETE FROM user_sessions
WHERE refresh_token_hash = $1;


-- name: DeleteAllSessions :exec
DELETE FROM user_sessions
WHERE user_id = $1;
