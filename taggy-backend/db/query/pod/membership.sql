-- =========================================
-- POD MEMBERSHIP
-- =========================================

-- name: CreatePodMembership :one
INSERT INTO pod_membership (
    pod_id,
    user_id,
    status,
    role,
    joined_at
)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, pod_id, user_id, status, joined_at, created_at, role;


-- name: GetPodMembershipByPodAndUser :one
SELECT id, pod_id, user_id, status, joined_at, created_at, role
FROM pod_membership
WHERE pod_id = $1
  AND user_id = $2;


-- name: GetAcceptedPodMembershipByUserID :one
SELECT id, pod_id, user_id, status, joined_at, created_at, role
FROM pod_membership
WHERE user_id = $1
  AND status = 'ACCEPTED'
LIMIT 1;


-- name: UpdatePodMembershipStatus :one
UPDATE pod_membership
SET
    status = $2,
    joined_at = $3
WHERE id = $1
RETURNING id, pod_id, user_id, status, joined_at, created_at, role;


-- name: UpdatePodMembershipRole :one
UPDATE pod_membership
SET role = $2
WHERE id = $1
RETURNING id, pod_id, user_id, status, joined_at, created_at, role;


-- name: UpdatePodMembershipStatusAndRole :one
UPDATE pod_membership
SET
    status = $2,
    role = $3,
    joined_at = $4
WHERE id = $1
RETURNING id, pod_id, user_id, status, joined_at, created_at, role;


-- name: ListPodMembershipsByUserID :many
SELECT
    pm.id,
    pm.pod_id,
    pm.user_id,
    pm.status,
    pm.joined_at,
    pm.created_at,
    pm.role,
    p.slug AS pod_slug,
    p.name AS pod_name,
    s.slug AS skill_slug,
    s.name AS skill_name
FROM pod_membership pm
INNER JOIN pods p ON p.id = pm.pod_id
INNER JOIN skills s ON s.id = p.skill_id
WHERE pm.user_id = $1
ORDER BY pm.created_at DESC;


-- name: ListAcceptedMembersByPodID :many
SELECT
    pm.id,
    pm.pod_id,
    pm.user_id,
    pm.status,
    pm.joined_at,
    pm.created_at,
    pm.role,
    u.username,
    u.name AS user_name
FROM pod_membership pm
INNER JOIN users u ON u.id = pm.user_id
WHERE pm.pod_id = $1
  AND pm.status = 'ACCEPTED'
ORDER BY
    CASE pm.role
        WHEN 'OWNER' THEN 0
        WHEN 'ADMIN' THEN 1
        ELSE 2
    END,
    pm.joined_at ASC NULLS LAST,
    pm.created_at ASC;


-- name: ListPendingMembersByPodID :many
SELECT
    pm.id,
    pm.pod_id,
    pm.user_id,
    pm.status,
    pm.joined_at,
    pm.created_at,
    pm.role,
    u.username,
    u.name AS user_name
FROM pod_membership pm
INNER JOIN users u ON u.id = pm.user_id
WHERE pm.pod_id = $1
  AND pm.status = 'PENDING'
ORDER BY pm.created_at ASC;


-- name: ListAcceptedAdminsByPodID :many
SELECT id, pod_id, user_id, status, joined_at, created_at, role
FROM pod_membership
WHERE pod_id = $1
  AND status = 'ACCEPTED'
  AND role = 'ADMIN'
ORDER BY joined_at ASC NULLS LAST, created_at ASC;


-- name: ListAcceptedMembersExcludingUser :many
SELECT id, pod_id, user_id, status, joined_at, created_at, role
FROM pod_membership
WHERE pod_id = $1
  AND status = 'ACCEPTED'
  AND user_id <> $2
ORDER BY joined_at ASC NULLS LAST, created_at ASC;
