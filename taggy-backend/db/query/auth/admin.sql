-- name: SetUsersAdminByUsernames :exec
UPDATE users
SET role = 'ADMIN',
    updated_at = NOW()
WHERE username = ANY(sqlc.arg(usernames)::citext[])
  AND is_deleted = FALSE;


-- name: GetUserIsAdminByPublicID :one
SELECT role = 'ADMIN' AS is_admin
FROM users
WHERE public_id = $1
  AND is_deleted = FALSE;
