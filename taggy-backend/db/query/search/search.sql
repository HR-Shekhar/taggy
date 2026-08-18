-- =========================================
-- SEARCH
-- =========================================

-- name: SearchSkills :many
SELECT id, name, slug, description, is_active, created_at, updated_at
FROM skills
WHERE is_active = TRUE
  AND (
    name ILIKE '%' || sqlc.arg(query)::text || '%'
    OR slug ILIKE '%' || sqlc.arg(query)::text || '%'
  )
ORDER BY name
LIMIT sqlc.arg(lim);


-- name: SearchUsers :many
SELECT public_id, username, name, profile_picture_url, bio
FROM users
WHERE is_deleted = FALSE
  AND username ILIKE '%' || sqlc.arg(query)::text || '%'
ORDER BY username
LIMIT sqlc.arg(lim);


-- name: SearchCommunities :many
SELECT
    c.id,
    c.skill_id,
    c.name,
    c.description,
    c.created_at,
    c.updated_at,
    s.slug AS skill_slug,
    s.name AS skill_name
FROM community c
INNER JOIN skills s ON s.id = c.skill_id
WHERE s.is_active = TRUE
  AND (
    c.name ILIKE '%' || sqlc.arg(query)::text || '%'
    OR s.name ILIKE '%' || sqlc.arg(query)::text || '%'
    OR s.slug ILIKE '%' || sqlc.arg(query)::text || '%'
  )
ORDER BY c.name
LIMIT sqlc.arg(lim);
