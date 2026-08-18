-- =========================================
-- CREATE
-- =========================================

-- name: CreateStudySession :one
INSERT INTO study_session (
    user_id,
    skill_id,
    duration_minutes,
    notes,
    studied_at
)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, skill_id, duration_minutes, notes, studied_at, created_at;


-- =========================================
-- GET
-- =========================================

-- name: ListStudySessionsByUserID :many
SELECT ss.id, ss.user_id, ss.skill_id, ss.duration_minutes, ss.notes, ss.studied_at, ss.created_at,
       s.slug AS skill_slug
FROM study_session ss
INNER JOIN skills s ON s.id = ss.skill_id
WHERE ss.user_id = $1
ORDER BY ss.studied_at DESC;


-- name: ListStudySessionsByUserAndSkillSlug :many
SELECT ss.id, ss.user_id, ss.skill_id, ss.duration_minutes, ss.notes, ss.studied_at, ss.created_at,
       s.slug AS skill_slug
FROM study_session ss
INNER JOIN skills s ON s.id = ss.skill_id
WHERE ss.user_id = $1
  AND s.slug = $2
ORDER BY ss.studied_at DESC;


-- name: SumStudyMinutesByUserID :one
SELECT COALESCE(SUM(duration_minutes), 0)::bigint
FROM study_session
WHERE user_id = $1;


-- name: SumStudyMinutesLast7Days :one
SELECT COALESCE(SUM(duration_minutes), 0)::bigint
FROM study_session
WHERE user_id = $1
  AND studied_at >= NOW() - INTERVAL '7 days';


-- name: SumStudyMinutesThisMonth :one
SELECT COALESCE(SUM(duration_minutes), 0)::bigint
FROM study_session
WHERE user_id = $1
  AND studied_at >= date_trunc('month', NOW());
