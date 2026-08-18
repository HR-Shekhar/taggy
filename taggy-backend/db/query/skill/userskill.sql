-- =========================================
-- GET
-- =========================================

-- name: CountActiveUserSkillsByUserID :one
SELECT COUNT(*)::bigint
FROM userskill
WHERE user_id = $1
  AND status = 'ACTIVE';


-- name: GetUserSkillByUserAndSkill :one
SELECT id, user_id, skill_id, roadmap_version_id, status, started_at, completed_at
FROM userskill
WHERE user_id = $1
  AND skill_id = $2;


-- name: GetUserSkillByUserAndSkillSlug :one
SELECT us.id, us.user_id, us.skill_id, us.roadmap_version_id, us.status, us.started_at, us.completed_at
FROM userskill us
INNER JOIN skills s ON s.id = us.skill_id
WHERE us.user_id = $1
  AND s.slug = $2;


-- name: ListUserSkillsByUserID :many
SELECT us.id, us.user_id, us.skill_id, us.roadmap_version_id, us.status, us.started_at, us.completed_at,
       s.name AS skill_name, s.slug AS skill_slug,
       rv.version_number AS roadmap_version_number,
       rv.status AS roadmap_version_status,
       (
           SELECT COUNT(*)::bigint
           FROM user_milestone_progress ump
           WHERE ump.user_skill_id = us.id
       ) AS milestone_count,
       (
           SELECT COUNT(*)::bigint
           FROM user_milestone_progress ump
           WHERE ump.user_skill_id = us.id
             AND ump.status = 'COMPLETED'
       ) AS completed_count
FROM userskill us
INNER JOIN skills s ON s.id = us.skill_id
INNER JOIN roadmap_version rv ON rv.id = us.roadmap_version_id
WHERE us.user_id = $1
ORDER BY us.started_at DESC;


-- name: GetUserSkillByID :one
SELECT id, user_id, skill_id, roadmap_version_id, status, started_at, completed_at
FROM userskill
WHERE id = $1
  AND user_id = $2;


-- =========================================
-- CREATE
-- =========================================

-- name: CreateUserSkill :one
INSERT INTO userskill (
    user_id,
    skill_id,
    roadmap_version_id,
    status
)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, skill_id, roadmap_version_id, status, started_at, completed_at;


-- =========================================
-- UPDATE
-- =========================================

-- name: UpdateUserSkillRoadmapVersion :one
UPDATE userskill
SET roadmap_version_id = $2
WHERE id = $1
RETURNING id, user_id, skill_id, roadmap_version_id, status, started_at, completed_at;
