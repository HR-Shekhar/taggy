-- =========================================
-- LIST / GET
-- =========================================

-- name: ListActiveSkills :many
SELECT id, name, slug, description, is_active, created_at, updated_at
FROM skills
WHERE is_active = TRUE
ORDER BY name;


-- name: GetSkillByID :one
SELECT id, name, slug, description, is_active, created_at, updated_at
FROM skills
WHERE id = $1;


-- name: GetSkillBySlug :one
SELECT id, name, slug, description, is_active, created_at, updated_at
FROM skills
WHERE slug = $1;


-- name: GetSkillByIDActive :one
SELECT id, name, slug, description, is_active, created_at, updated_at
FROM skills
WHERE id = $1 AND is_active = TRUE;
