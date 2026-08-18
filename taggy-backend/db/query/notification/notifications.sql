-- =========================================
-- NOTIFICATIONS
-- =========================================

-- name: CreateNotification :one
INSERT INTO notification (
    user_id,
    type,
    entity_type,
    entity_id,
    title,
    body
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, type, entity_type, entity_id, title, body, is_read, read_at, created_at;


-- name: ListNotificationsByUserID :many
SELECT id, user_id, type, entity_type, entity_id, title, body, is_read, read_at, created_at
FROM notification
WHERE user_id = sqlc.arg(user_id)
  AND (sqlc.arg(unread_only)::boolean = FALSE OR is_read = FALSE)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(notif_limit);


-- name: GetNotificationByIDAndUserID :one
SELECT id, user_id, type, entity_type, entity_id, title, body, is_read, read_at, created_at
FROM notification
WHERE id = $1
  AND user_id = $2;


-- name: MarkNotificationRead :one
UPDATE notification
SET
    is_read = TRUE,
    read_at = NOW()
WHERE id = $1
  AND user_id = $2
  AND is_read = FALSE
RETURNING id, user_id, type, entity_type, entity_id, title, body, is_read, read_at, created_at;


-- name: MarkAllNotificationsRead :execrows
UPDATE notification
SET
    is_read = TRUE,
    read_at = NOW()
WHERE user_id = $1
  AND is_read = FALSE;


-- name: CountUnreadNotifications :one
SELECT COUNT(*)::bigint
FROM notification
WHERE user_id = $1
  AND is_read = FALSE;


-- name: DeleteReadNotificationsByUserID :execrows
DELETE FROM notification
WHERE user_id = $1
  AND is_read = TRUE;
