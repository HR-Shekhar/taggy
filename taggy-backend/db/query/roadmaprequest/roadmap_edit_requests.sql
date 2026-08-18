-- name: CreateRoadmapEditRequest :one
INSERT INTO roadmap_edit_request (
    skill_id,
    requester_id,
    rationale,
    status,
    base_version_number,
    draft_milestones
)
VALUES ($1, $2, $3, 'PENDING', $4, $5)
RETURNING *;


-- name: GetRoadmapEditRequestByPublicID :one
SELECT
    r.*,
    s.slug AS skill_slug,
    s.name AS skill_name
FROM roadmap_edit_request r
INNER JOIN skills s ON s.id = r.skill_id
WHERE r.public_id = $1;


-- name: ListRoadmapEditRequestsByRequester :many
SELECT
    r.*,
    s.slug AS skill_slug,
    s.name AS skill_name
FROM roadmap_edit_request r
INNER JOIN skills s ON s.id = r.skill_id
WHERE r.requester_id = $1
ORDER BY r.created_at DESC
LIMIT sqlc.arg(result_limit);


-- name: ListPendingRoadmapEditRequests :many
SELECT
    r.*,
    s.slug AS skill_slug,
    s.name AS skill_name
FROM roadmap_edit_request r
INNER JOIN skills s ON s.id = r.skill_id
WHERE r.status = 'PENDING'
ORDER BY r.created_at ASC
LIMIT sqlc.arg(result_limit);


-- name: CancelRoadmapEditRequest :one
UPDATE roadmap_edit_request
SET status = 'CANCELLED',
    updated_at = NOW()
WHERE public_id = $1
  AND requester_id = $2
  AND status = 'PENDING'
RETURNING *;


-- name: ApproveRoadmapEditRequest :one
UPDATE roadmap_edit_request
SET status = 'APPROVED',
    reviewed_by = $2,
    reviewed_at = NOW(),
    created_version_id = $3,
    admin_note = $4,
    updated_at = NOW()
WHERE id = $1
  AND status = 'PENDING'
RETURNING *;


-- name: RejectRoadmapEditRequest :one
UPDATE roadmap_edit_request
SET status = 'REJECTED',
    reviewed_by = $2,
    reviewed_at = NOW(),
    admin_note = $3,
    updated_at = NOW()
WHERE id = $1
  AND status = 'PENDING'
RETURNING *;


-- name: GetPendingRoadmapEditByRequesterAndSkill :one
SELECT *
FROM roadmap_edit_request
WHERE requester_id = $1
  AND skill_id = $2
  AND status = 'PENDING'
LIMIT 1;
