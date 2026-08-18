-- +goose Up
CREATE TYPE pod_quiz_status AS ENUM (
    'IN_PROGRESS',
    'COMPLETED',
    'ABANDONED'
);

CREATE TABLE pod_quiz (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    pod_id BIGINT NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    skill_id BIGINT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    status pod_quiz_status NOT NULL DEFAULT 'IN_PROGRESS',
    topic_count INTEGER NOT NULL DEFAULT 0 CHECK (topic_count >= 0),
    correct_count INTEGER NOT NULL DEFAULT 0 CHECK (correct_count >= 0),
    score INTEGER NOT NULL DEFAULT 0 CHECK (score >= 0),
    completed_topic_titles JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX pod_quiz_one_in_progress_per_user_pod
    ON pod_quiz (user_id, pod_id)
    WHERE status = 'IN_PROGRESS';

CREATE INDEX pod_quiz_pod_user_score_idx
    ON pod_quiz (pod_id, user_id, score DESC)
    WHERE status = 'COMPLETED';

CREATE INDEX pod_quiz_pod_status_idx ON pod_quiz (pod_id, status);

CREATE TABLE pod_quiz_question (
    id BIGSERIAL PRIMARY KEY,
    quiz_id BIGINT NOT NULL REFERENCES pod_quiz(id) ON DELETE CASCADE,
    order_index INTEGER NOT NULL CHECK (order_index BETWEEN 1 AND 10),
    difficulty INTEGER NOT NULL CHECK (difficulty BETWEEN 1 AND 10),
    prompt TEXT NOT NULL,
    options JSONB NOT NULL,
    correct_indices JSONB NOT NULL,
    topic_title TEXT NOT NULL,
    weight INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT pod_quiz_question_quiz_order UNIQUE (quiz_id, order_index)
);

CREATE TABLE pod_quiz_answer (
    id BIGSERIAL PRIMARY KEY,
    quiz_id BIGINT NOT NULL REFERENCES pod_quiz(id) ON DELETE CASCADE,
    question_id BIGINT NOT NULL REFERENCES pod_quiz_question(id) ON DELETE CASCADE,
    selected_indices JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    started_at TIMESTAMPTZ,
    answered_at TIMESTAMPTZ,
    timed_out BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT pod_quiz_answer_unique UNIQUE (quiz_id, question_id)
);

CREATE INDEX pod_quiz_answer_quiz_idx ON pod_quiz_answer (quiz_id);

-- +goose Down
DROP TABLE IF EXISTS pod_quiz_answer;
DROP TABLE IF EXISTS pod_quiz_question;
DROP TABLE IF EXISTS pod_quiz;
DROP TYPE IF EXISTS pod_quiz_status;
