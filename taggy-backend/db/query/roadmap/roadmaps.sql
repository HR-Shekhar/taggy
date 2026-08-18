-- =========================================
-- GET
-- =========================================

-- name: GetRoadmapBySkillID :one
SELECT id, current_version_id, skill_id, created_at, updated_at
FROM roadmaps
WHERE skill_id = $1;


-- name: GetRoadmapBySkillSlug :one
SELECT r.id, r.current_version_id, r.skill_id, r.created_at, r.updated_at,
       s.slug AS skill_slug, s.name AS skill_name
FROM roadmaps r
INNER JOIN skills s ON s.id = r.skill_id
WHERE s.slug = $1;


-- name: GetRoadmapVersionByID :one
SELECT id, roadmap_id, version_number, status, generated_by, published_at, created_at
FROM roadmap_version
WHERE id = $1;


-- name: GetActiveRoadmapVersionBySkillID :one
SELECT rv.id, rv.roadmap_id, rv.version_number, rv.status, rv.generated_by, rv.published_at, rv.created_at
FROM roadmap_version rv
INNER JOIN roadmaps r ON r.id = rv.roadmap_id
WHERE r.skill_id = $1
  AND rv.status = 'ACTIVE'
LIMIT 1;


-- name: GetCurrentRoadmapVersionBySkillID :one
-- Prefer roadmaps.current_version_id; fall back to ACTIVE status.
SELECT rv.id, rv.roadmap_id, rv.version_number, rv.status, rv.generated_by, rv.published_at, rv.created_at
FROM roadmaps r
INNER JOIN roadmap_version rv ON rv.id = COALESCE(
    r.current_version_id,
    (
        SELECT rv2.id
        FROM roadmap_version rv2
        WHERE rv2.roadmap_id = r.id
          AND rv2.status = 'ACTIVE'
        ORDER BY rv2.version_number DESC
        LIMIT 1
    )
)
WHERE r.skill_id = $1;


-- name: ListRoadmapVersionsBySkillSlug :many
SELECT rv.id, rv.roadmap_id, rv.version_number, rv.status, rv.generated_by, rv.published_at, rv.created_at,
       COALESCE(r.current_version_id = rv.id, false)::boolean AS is_current,
       (
           SELECT COUNT(*)::bigint
           FROM milestones m
           WHERE m.roadmap_version_id = rv.id
       ) AS milestone_count
FROM roadmap_version rv
INNER JOIN roadmaps r ON r.id = rv.roadmap_id
INNER JOIN skills s ON s.id = r.skill_id
WHERE s.slug = $1
ORDER BY rv.version_number DESC;


-- name: GetRoadmapVersionBySkillSlugAndNumber :one
SELECT rv.id, rv.roadmap_id, rv.version_number, rv.status, rv.generated_by, rv.published_at, rv.created_at,
       COALESCE(r.current_version_id = rv.id, false)::boolean AS is_current,
       s.slug AS skill_slug, s.name AS skill_name
FROM roadmap_version rv
INNER JOIN roadmaps r ON r.id = rv.roadmap_id
INNER JOIN skills s ON s.id = r.skill_id
WHERE s.slug = $1
  AND rv.version_number = $2;
