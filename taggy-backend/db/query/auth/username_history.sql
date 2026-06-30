-- name: CreateUsernameHistory :one
INSERT INTO username_history (
    user_id,
    username
)
VALUES (
    $1,
    $2
)
RETURNING *;