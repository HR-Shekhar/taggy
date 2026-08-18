-- =========================================
-- PODS
-- =========================================

-- name: CreatePod :one
INSERT INTO pods (
    slug,
    name,
    description,
    owner_id,
    skill_id,
    max_members
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, public_id, slug, name, description, owner_id, skill_id, max_members, created_at, updated_at;


-- name: GetPodBySlug :one
SELECT
    p.id,
    p.public_id,
    p.slug,
    p.name,
    p.description,
    p.owner_id,
    p.skill_id,
    p.max_members,
    p.created_at,
    p.updated_at,
    s.slug AS skill_slug,
    s.name AS skill_name,
    u.username AS owner_username
FROM pods p
INNER JOIN skills s ON s.id = p.skill_id
INNER JOIN users u ON u.id = p.owner_id
WHERE p.slug = $1;


-- name: GetPodByID :one
SELECT id, public_id, slug, name, description, owner_id, skill_id, max_members, created_at, updated_at
FROM pods
WHERE id = $1;


-- name: PodSlugExists :one
SELECT EXISTS (
    SELECT 1
    FROM pods
    WHERE slug = $1
);


-- name: ListPodsBySkillSlug :many
SELECT
    p.id,
    p.public_id,
    p.slug,
    p.name,
    p.description,
    p.owner_id,
    p.skill_id,
    p.max_members,
    p.created_at,
    p.updated_at,
    s.slug AS skill_slug,
    s.name AS skill_name,
    u.username AS owner_username,
    (
        SELECT COUNT(*)::bigint
        FROM pod_membership pm
        WHERE pm.pod_id = p.id
          AND pm.status = 'ACCEPTED'
    ) AS accepted_count
FROM pods p
INNER JOIN skills s ON s.id = p.skill_id
INNER JOIN users u ON u.id = p.owner_id
WHERE s.slug = $1
ORDER BY p.created_at DESC;


-- name: CountAcceptedPodMembers :one
SELECT COUNT(*)::bigint
FROM pod_membership
WHERE pod_id = $1
  AND status = 'ACCEPTED';


-- name: UpdatePodOwner :one
UPDATE pods
SET
    owner_id = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, public_id, slug, name, description, owner_id, skill_id, max_members, created_at, updated_at;


-- name: DeletePod :exec
DELETE FROM pods
WHERE id = $1;
