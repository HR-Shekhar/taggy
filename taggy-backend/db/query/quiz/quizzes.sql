-- name: CreatePodQuiz :one
INSERT INTO pod_quiz (
    pod_id,
    user_id,
    skill_id,
    status,
    topic_count,
    completed_topic_titles
) VALUES (
    $1, $2, $3, 'IN_PROGRESS', $4, $5
)
RETURNING *;

-- name: GetPodQuizByPublicID :one
SELECT * FROM pod_quiz
WHERE public_id = $1;

-- name: GetInProgressPodQuiz :one
SELECT * FROM pod_quiz
WHERE user_id = $1
  AND pod_id = $2
  AND status = 'IN_PROGRESS';

-- name: ListMyPodQuizzes :many
SELECT * FROM pod_quiz
WHERE user_id = $1
  AND pod_id = $2
ORDER BY created_at DESC
LIMIT $3;

-- name: CompletePodQuiz :one
UPDATE pod_quiz
SET
    status = 'COMPLETED',
    correct_count = $2,
    score = $3,
    completed_at = NOW()
WHERE id = $1
  AND status = 'IN_PROGRESS'
RETURNING *;

-- name: AbandonPodQuiz :one
UPDATE pod_quiz
SET status = 'ABANDONED'
WHERE id = $1
  AND status = 'IN_PROGRESS'
RETURNING *;

-- name: CreatePodQuizQuestion :one
INSERT INTO pod_quiz_question (
    quiz_id,
    order_index,
    difficulty,
    prompt,
    options,
    correct_indices,
    topic_title,
    weight
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ListPodQuizQuestions :many
SELECT * FROM pod_quiz_question
WHERE quiz_id = $1
ORDER BY order_index;

-- name: GetPodQuizQuestionByOrder :one
SELECT * FROM pod_quiz_question
WHERE quiz_id = $1
  AND order_index = $2;

-- name: UpsertPodQuizAnswerStart :one
INSERT INTO pod_quiz_answer (quiz_id, question_id, started_at)
VALUES ($1, $2, NOW())
ON CONFLICT (quiz_id, question_id) DO UPDATE
SET started_at = COALESCE(pod_quiz_answer.started_at, EXCLUDED.started_at)
RETURNING *;

-- name: GetPodQuizAnswer :one
SELECT * FROM pod_quiz_answer
WHERE quiz_id = $1
  AND question_id = $2;

-- name: SavePodQuizAnswer :one
UPDATE pod_quiz_answer
SET
    selected_indices = $3,
    is_correct = $4,
    answered_at = NOW(),
    timed_out = $5
WHERE quiz_id = $1
  AND question_id = $2
  AND answered_at IS NULL
RETURNING *;

-- name: CountCorrectPodQuizAnswers :one
SELECT COUNT(*)::bigint
FROM pod_quiz_answer
WHERE quiz_id = $1
  AND is_correct = TRUE
  AND answered_at IS NOT NULL;

-- name: ListCompletedTopicTitlesForUserSkill :many
SELECT m.title
FROM userskill us
INNER JOIN user_milestone_progress ump ON ump.user_skill_id = us.id
INNER JOIN milestones m ON m.id = ump.milestone_id
WHERE us.user_id = $1
  AND us.skill_id = $2
  AND ump.status = 'COMPLETED'
  AND m.kind = 'TOPIC'
  AND m.roadmap_version_id = us.roadmap_version_id
ORDER BY m.order_index;

-- name: ListPodLeaderboard :many
WITH best AS (
    SELECT
        pq.user_id,
        MAX(pq.score)::int AS best_score,
        MAX(pq.topic_count)::int AS topic_count
    FROM pod_quiz pq
    WHERE pq.pod_id = $1
      AND pq.status = 'COMPLETED'
    GROUP BY pq.user_id
),
members AS (
    SELECT pm.user_id, u.username, u.public_id, u.name
    FROM pod_membership pm
    INNER JOIN users u ON u.id = pm.user_id
    WHERE pm.pod_id = $1
      AND pm.status = 'ACCEPTED'
)
SELECT
    m.username,
    m.public_id,
    m.name,
    COALESCE(b.best_score, 0)::int AS best_score,
    COALESCE(b.topic_count, 0)::int AS topic_count
FROM members m
LEFT JOIN best b ON b.user_id = m.user_id
ORDER BY best_score DESC, m.username ASC;

-- name: ListCommunityPodLeaderboard :many
WITH best AS (
    SELECT
        pq.pod_id,
        pq.user_id,
        MAX(pq.score)::int AS best_score
    FROM pod_quiz pq
    INNER JOIN pods p ON p.id = pq.pod_id
    WHERE p.skill_id = $1
      AND pq.status = 'COMPLETED'
    GROUP BY pq.pod_id, pq.user_id
),
pod_totals AS (
    SELECT
        p.id AS pod_id,
        p.slug AS pod_slug,
        p.name AS pod_name,
        COALESCE(SUM(b.best_score), 0)::int AS total_score,
        COUNT(DISTINCT pm.user_id)::int AS member_count
    FROM pods p
    INNER JOIN pod_membership pm ON pm.pod_id = p.id AND pm.status = 'ACCEPTED'
    LEFT JOIN best b ON b.pod_id = p.id AND b.user_id = pm.user_id
    WHERE p.skill_id = $1
    GROUP BY p.id, p.slug, p.name
)
SELECT pod_id, pod_slug, pod_name, total_score, member_count
FROM pod_totals
ORDER BY total_score DESC, pod_name ASC;
