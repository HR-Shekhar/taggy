-- =========================================
-- COMMUNITY CHANNELS
-- =========================================

-- name: ListChannelsByCommunityID :many
SELECT id, community_id, name, slug, description, created_at, updated_at
FROM community_channel
WHERE community_id = $1
ORDER BY created_at ASC, id ASC;


-- name: GetChannelByCommunityIDAndSlug :one
SELECT id, community_id, name, slug, description, created_at, updated_at
FROM community_channel
WHERE community_id = $1
  AND slug = $2;


-- name: GetChannelBySkillSlugAndChannelSlug :one
SELECT
    cc.id,
    cc.community_id,
    cc.name,
    cc.slug,
    cc.description,
    cc.created_at,
    cc.updated_at,
    c.skill_id,
    s.slug AS skill_slug
FROM community_channel cc
INNER JOIN community c ON c.id = cc.community_id
INNER JOIN skills s ON s.id = c.skill_id
WHERE s.slug = sqlc.arg(skill_slug)
  AND cc.slug = sqlc.arg(channel_slug);
