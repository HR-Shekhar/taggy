-- =========================================
-- GET
-- =========================================

-- name: ListMilestonesByRoadmapVersionID :many
SELECT id, roadmap_version_id, title, slug, description, estimated_hours, order_index, difficulty, chapter, kind, created_at, updated_at
FROM milestones
WHERE roadmap_version_id = $1
ORDER BY order_index;


-- name: GetMilestoneByID :one
SELECT id, roadmap_version_id, title, slug, description, estimated_hours, order_index, difficulty, chapter, kind, created_at, updated_at
FROM milestones
WHERE id = $1;


-- name: GetMilestoneBySlugAndRoadmapVersion :one
SELECT id, roadmap_version_id, title, slug, description, estimated_hours, order_index, difficulty, chapter, kind, created_at, updated_at
FROM milestones
WHERE roadmap_version_id = $1
  AND slug = $2;
