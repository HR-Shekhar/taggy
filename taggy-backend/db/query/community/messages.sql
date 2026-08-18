-- =========================================
-- MESSAGES
-- =========================================

-- name: CreateChannelMessage :one
INSERT INTO message (
    author_id,
    community_channel_id,
    pod_id,
    content,
    reply_to_message_id
)
VALUES ($1, $2, NULL, $3, $4)
RETURNING id, author_id, community_channel_id, pod_id, content, edited_at, created_at, reply_to_message_id;


-- name: CreatePodMessage :one
INSERT INTO message (
    author_id,
    community_channel_id,
    pod_id,
    content,
    reply_to_message_id
)
VALUES ($1, NULL, $2, $3, $4)
RETURNING id, author_id, community_channel_id, pod_id, content, edited_at, created_at, reply_to_message_id;


-- name: ListChannelMessages :many
SELECT
    m.id,
    m.author_id,
    m.community_channel_id,
    m.pod_id,
    m.content,
    m.edited_at,
    m.created_at,
    m.reply_to_message_id,
    u.username AS author_username,
    u.name AS author_name,
    parent.content AS reply_to_content,
    parent_author.username AS reply_to_author_username
FROM message m
INNER JOIN users u ON u.id = m.author_id
LEFT JOIN message parent ON parent.id = m.reply_to_message_id
LEFT JOIN users parent_author ON parent_author.id = parent.author_id
WHERE m.community_channel_id = sqlc.arg(channel_id)
  AND (sqlc.arg(before_id)::bigint = 0 OR m.id < sqlc.arg(before_id))
ORDER BY m.id DESC
LIMIT sqlc.arg(message_limit);


-- name: ListPodMessages :many
SELECT
    m.id,
    m.author_id,
    m.community_channel_id,
    m.pod_id,
    m.content,
    m.edited_at,
    m.created_at,
    m.reply_to_message_id,
    u.username AS author_username,
    u.name AS author_name,
    parent.content AS reply_to_content,
    parent_author.username AS reply_to_author_username
FROM message m
INNER JOIN users u ON u.id = m.author_id
LEFT JOIN message parent ON parent.id = m.reply_to_message_id
LEFT JOIN users parent_author ON parent_author.id = parent.author_id
WHERE m.pod_id = sqlc.arg(pod_id)
  AND (sqlc.arg(before_id)::bigint = 0 OR m.id < sqlc.arg(before_id))
ORDER BY m.id DESC
LIMIT sqlc.arg(message_limit);


-- name: GetMessageByID :one
SELECT
    m.id,
    m.author_id,
    m.community_channel_id,
    m.pod_id,
    m.content,
    m.edited_at,
    m.created_at,
    m.reply_to_message_id,
    u.username AS author_username,
    u.name AS author_name,
    s.slug AS skill_slug,
    cc.slug AS channel_slug,
    p.slug AS pod_slug,
    parent.content AS reply_to_content,
    parent_author.username AS reply_to_author_username
FROM message m
INNER JOIN users u ON u.id = m.author_id
LEFT JOIN community_channel cc ON cc.id = m.community_channel_id
LEFT JOIN community c ON c.id = cc.community_id
LEFT JOIN skills s ON s.id = c.skill_id
LEFT JOIN pods p ON p.id = m.pod_id
LEFT JOIN message parent ON parent.id = m.reply_to_message_id
LEFT JOIN users parent_author ON parent_author.id = parent.author_id
WHERE m.id = $1;


-- name: UpdateMessageContent :one
UPDATE message
SET
    content = $2,
    edited_at = NOW()
WHERE id = $1
  AND author_id = $3
RETURNING id, author_id, community_channel_id, pod_id, content, edited_at, created_at, reply_to_message_id;


-- name: DeleteMessage :execrows
DELETE FROM message
WHERE id = $1
  AND author_id = $2;
