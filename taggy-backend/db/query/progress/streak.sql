-- =========================================
-- GET
-- =========================================

-- name: GetStreakByUserID :one
SELECT id, user_id, current_streak, longest_streak, last_activity_date, freeze_count, updated_at
FROM streak
WHERE user_id = $1;


-- name: GetStreakByUserIDForUpdate :one
SELECT id, user_id, current_streak, longest_streak, last_activity_date, freeze_count, updated_at
FROM streak
WHERE user_id = $1
FOR UPDATE;


-- =========================================
-- CREATE
-- =========================================

-- name: CreateStreak :one
INSERT INTO streak (
    user_id,
    current_streak,
    longest_streak,
    last_activity_date
)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, current_streak, longest_streak, last_activity_date, freeze_count, updated_at;


-- =========================================
-- UPDATE
-- =========================================

-- name: UpdateStreak :one
UPDATE streak
SET
    current_streak = $2,
    longest_streak = $3,
    last_activity_date = $4,
    updated_at = NOW()
WHERE user_id = $1
RETURNING id, user_id, current_streak, longest_streak, last_activity_date, freeze_count, updated_at;
