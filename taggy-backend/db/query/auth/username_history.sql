-- =========================================
-- GET
-- =========================================

-- name: GetUsernameHistoryByUserID :many
SELECT *
FROM username_history
WHERE user_id = $1
ORDER BY changed_at DESC;


-- name: GetLatestUsernameHistory :one
SELECT *
FROM username_history
WHERE user_id = $1
ORDER BY changed_at DESC
LIMIT 1;


-- =========================================
-- CREATE
-- =========================================

-- name: CreateUsernameHistory :one
INSERT INTO username_history (
    user_id,
    old_username,
    new_username,
    changed_at
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;