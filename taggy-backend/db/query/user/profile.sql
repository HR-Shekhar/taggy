-- =========================================
-- USERNAME
-- =========================================

-- name: UsernameHistoryExists :one
SELECT EXISTS (
    SELECT 1
    FROM username_history
    WHERE username = $1
);


-- name: UpdateUsername :one
UPDATE users
SET
    username = $2,
    updated_at = NOW()
WHERE id = $1
    AND is_deleted = FALSE
RETURNING *;
