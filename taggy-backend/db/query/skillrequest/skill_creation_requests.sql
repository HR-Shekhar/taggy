-- name: CreateSkillCreationRequest :one
INSERT INTO skill_creation_request (
    requester_id,
    name,
    slug_candidate,
    description,
    status,
    similar_skills,
    draft_milestones
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;


-- name: GetSkillCreationRequestByPublicID :one
SELECT *
FROM skill_creation_request
WHERE public_id = $1;


-- name: GetSkillCreationRequestByID :one
SELECT *
FROM skill_creation_request
WHERE id = $1;


-- name: ListSkillCreationRequestsByRequester :many
SELECT *
FROM skill_creation_request
WHERE requester_id = $1
ORDER BY created_at DESC
LIMIT sqlc.arg(result_limit);


-- name: ListPendingSkillCreationRequests :many
SELECT *
FROM skill_creation_request
WHERE status = 'PENDING'
ORDER BY created_at ASC
LIMIT sqlc.arg(result_limit);


-- name: ListGeneratingSkillCreationRequests :many
SELECT id
FROM skill_creation_request
WHERE status = 'GENERATING'
ORDER BY created_at ASC;


-- name: CompleteSkillCreationDraft :one
UPDATE skill_creation_request
SET draft_milestones = $2,
    status = 'PENDING',
    updated_at = NOW()
WHERE id = $1
  AND status = 'GENERATING'
RETURNING *;


-- name: FailSkillCreationRequest :one
UPDATE skill_creation_request
SET status = 'FAILED',
    admin_note = $2,
    updated_at = NOW()
WHERE id = $1
  AND status = 'GENERATING'
RETURNING *;


-- name: CancelSkillCreationRequest :one
UPDATE skill_creation_request
SET status = 'CANCELLED',
    updated_at = NOW()
WHERE public_id = $1
  AND requester_id = $2
  AND status IN ('PENDING', 'GENERATING')
RETURNING *;


-- name: ApproveSkillCreationRequest :one
UPDATE skill_creation_request
SET status = 'APPROVED',
    reviewed_by = $2,
    reviewed_at = NOW(),
    created_skill_id = $3,
    admin_note = $4,
    updated_at = NOW()
WHERE id = $1
  AND status = 'PENDING'
RETURNING *;


-- name: RejectSkillCreationRequest :one
UPDATE skill_creation_request
SET status = 'REJECTED',
    reviewed_by = $2,
    reviewed_at = NOW(),
    admin_note = $3,
    updated_at = NOW()
WHERE id = $1
  AND status = 'PENDING'
RETURNING *;


-- name: AutoRejectGeneratingSkillCreationRequest :one
UPDATE skill_creation_request
SET status = 'REJECTED',
    reviewed_by = NULL,
    reviewed_at = NOW(),
    admin_note = $2,
    updated_at = NOW()
WHERE id = $1
  AND status = 'GENERATING'
RETURNING *;


-- name: GetPendingSkillCreationByRequesterAndName :one
SELECT *
FROM skill_creation_request
WHERE requester_id = $1
  AND lower(name) = lower($2)
  AND status IN ('PENDING', 'GENERATING')
LIMIT 1;
