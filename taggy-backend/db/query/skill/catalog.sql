-- name: ListSimilarSkills :many
SELECT
    id,
    name,
    slug,
    description,
    is_active,
    created_at,
    updated_at,
    similarity(name, sqlc.arg(query)::text) AS score
FROM skills
WHERE is_active = TRUE
  AND (
    similarity(name, sqlc.arg(query)::text) >= sqlc.arg(min_score)::float4
    OR name ILIKE '%' || sqlc.arg(query)::text || '%'
  )
ORDER BY score DESC, name ASC
LIMIT sqlc.arg(result_limit);


-- name: CreateSkill :one
INSERT INTO skills (name, slug, description, is_active)
VALUES ($1, $2, $3, TRUE)
RETURNING *;


-- name: CreateRoadmap :one
INSERT INTO roadmaps (skill_id)
VALUES ($1)
RETURNING *;


-- name: CreateRoadmapVersion :one
INSERT INTO roadmap_version (
    roadmap_id,
    version_number,
    status,
    generated_by,
    published_at
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;


-- name: SetRoadmapCurrentVersion :exec
UPDATE roadmaps
SET current_version_id = $2,
    updated_at = NOW()
WHERE id = $1;


-- name: ArchiveActiveRoadmapVersions :exec
UPDATE roadmap_version
SET status = 'ARCHIVED'
WHERE roadmap_id = $1
  AND status = 'ACTIVE';


-- name: CreateMilestone :one
INSERT INTO milestones (
    roadmap_version_id,
    title,
    description,
    estimated_hours,
    order_index,
    difficulty,
    slug,
    chapter,
    kind
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;


-- name: CreateCommunity :one
INSERT INTO community (skill_id, name, description)
VALUES ($1, $2, $3)
RETURNING *;


-- name: CreateCommunityChannel :one
INSERT INTO community_channel (community_id, name, slug, description)
VALUES ($1, $2, $3, $4)
RETURNING *;


-- name: GetMaxRoadmapVersionNumber :one
SELECT COALESCE(MAX(version_number), 0)::int AS max_version
FROM roadmap_version
WHERE roadmap_id = $1;


-- name: GetActiveCatalogRoadmapVersionBySkillID :one
SELECT
    rv.id,
    rv.roadmap_id,
    rv.version_number,
    rv.status,
    rv.generated_by,
    rv.published_at,
    rv.created_at
FROM roadmap_version rv
INNER JOIN roadmaps r ON r.id = rv.roadmap_id
WHERE r.skill_id = $1
  AND rv.status = 'ACTIVE'
LIMIT 1;


-- name: ListUserIDsEnrolledInSkill :many
SELECT user_id
FROM userskill
WHERE skill_id = $1
  AND status = 'ACTIVE';
