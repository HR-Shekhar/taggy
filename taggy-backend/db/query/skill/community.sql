-- =========================================
-- GET
-- =========================================

-- name: GetCommunityBySkillID :one
SELECT c.id, c.skill_id, c.name, c.description, c.created_at, c.updated_at,
       s.slug AS skill_slug
FROM community c
INNER JOIN skills s ON s.id = c.skill_id
WHERE c.skill_id = $1;


-- name: GetCommunityBySkillSlug :one
SELECT c.id, c.skill_id, c.name, c.description, c.created_at, c.updated_at,
       s.slug AS skill_slug
FROM community c
INNER JOIN skills s ON s.id = c.skill_id
WHERE s.slug = $1;
