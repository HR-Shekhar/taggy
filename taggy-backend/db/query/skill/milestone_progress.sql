-- =========================================
-- GET
-- =========================================

-- name: ListMilestoneProgressByUserSkillID :many
SELECT ump.id, ump.user_skill_id, ump.milestone_id, ump.status, ump.completed_at, ump.postponed_until,
       m.title, m.slug, m.description, m.estimated_hours, m.order_index, m.difficulty, m.chapter, m.kind
FROM user_milestone_progress ump
INNER JOIN milestones m ON m.id = ump.milestone_id
WHERE ump.user_skill_id = $1
ORDER BY m.order_index;


-- name: GetMilestoneProgress :one
SELECT ump.id, ump.user_skill_id, ump.milestone_id, ump.status, ump.completed_at, ump.postponed_until
FROM user_milestone_progress ump
WHERE ump.user_skill_id = $1
  AND ump.milestone_id = $2;


-- name: GetMilestoneProgressBySlug :one
SELECT ump.id, ump.user_skill_id, ump.milestone_id, ump.status, ump.completed_at, ump.postponed_until,
       m.title, m.slug, m.description, m.estimated_hours, m.order_index, m.difficulty, m.chapter, m.kind
FROM user_milestone_progress ump
INNER JOIN milestones m ON m.id = ump.milestone_id
WHERE ump.user_skill_id = $1
  AND m.slug = $2;


-- name: CountIncompleteMilestonesBefore :one
SELECT COUNT(*)::bigint
FROM user_milestone_progress ump
INNER JOIN milestones m ON m.id = ump.milestone_id
WHERE ump.user_skill_id = $1
  AND m.order_index < $2
  AND ump.status != 'COMPLETED';

-- name: CountIncompleteTopicMilestonesBefore :one
-- Subtopics (TOPIC) may be completed without finishing their parent CHAPTER first.
SELECT COUNT(*)::bigint
FROM user_milestone_progress ump
INNER JOIN milestones m ON m.id = ump.milestone_id
WHERE ump.user_skill_id = $1
  AND m.order_index < $2
  AND m.kind = 'TOPIC'
  AND ump.status != 'COMPLETED';

-- name: CountIncompleteTopicsInChapter :one
-- Parent CHAPTER cannot complete until every TOPIC in the same chapter is done.
SELECT COUNT(*)::bigint
FROM user_milestone_progress ump
INNER JOIN milestones m ON m.id = ump.milestone_id
WHERE ump.user_skill_id = $1
  AND m.kind = 'TOPIC'
  AND m.chapter = $2
  AND ump.status != 'COMPLETED';

-- name: CountIncompleteChaptersBefore :one
SELECT COUNT(*)::bigint
FROM user_milestone_progress ump
INNER JOIN milestones m ON m.id = ump.milestone_id
WHERE ump.user_skill_id = $1
  AND m.order_index < $2
  AND m.kind = 'CHAPTER'
  AND ump.status != 'COMPLETED';


-- =========================================
-- CREATE
-- =========================================

-- name: CreateMilestoneProgress :one
INSERT INTO user_milestone_progress (
    user_skill_id,
    milestone_id,
    status
)
VALUES ($1, $2, $3)
RETURNING id, user_skill_id, milestone_id, status, completed_at, postponed_until;


-- =========================================
-- UPDATE
-- =========================================

-- name: CompleteMilestoneProgress :one
UPDATE user_milestone_progress
SET
    status = 'COMPLETED',
    completed_at = NOW(),
    postponed_until = NULL
WHERE id = $1
RETURNING id, user_skill_id, milestone_id, status, completed_at, postponed_until;


-- name: PostponeMilestoneProgress :one
UPDATE user_milestone_progress
SET
    status = 'POSTPONED',
    postponed_until = $2,
    completed_at = NULL
WHERE id = $1
RETURNING id, user_skill_id, milestone_id, status, completed_at, postponed_until;


-- =========================================
-- DELETE
-- =========================================

-- name: DeleteMilestoneProgressByUserSkillID :exec
DELETE FROM user_milestone_progress
WHERE user_skill_id = $1;


-- name: ListCompletedMilestoneSlugsByUserSkillID :many
SELECT m.slug
FROM user_milestone_progress ump
INNER JOIN milestones m ON m.id = ump.milestone_id
WHERE ump.user_skill_id = $1
  AND ump.status = 'COMPLETED';
