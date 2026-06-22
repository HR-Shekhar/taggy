-- +goose Up

CREATE TABLE skills (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE current_status AS ENUM (
    'DRAFT',
    'ACTIVE',
    'ARCHIVED'
);

CREATE TABLE roadmaps (
    id BIGSERIAL PRIMARY KEY,
    current_version_id BIGINT,
    skill_id BIGINT NOT NULL REFERENCES skills(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE roadmap_version (
    id BIGSERIAL PRIMARY KEY,
    roadmap_id BIGINT REFERENCES roadmaps(id),
    version_number INTEGER NOT NULL,
    status current_status NOT NULL,
    generated_by VARCHAR(20) NOT NULL,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE roadmaps
ADD CONSTRAINT fk_roadmap_current_version
FOREIGN KEY (current_version_id)
REFERENCES roadmap_version(id);

CREATE TABLE milestones (
    id BIGSERIAL PRIMARY KEY,
    roadmap_version_id BIGINT REFERENCES roadmap_version(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    estimated_hours INTEGER,
    order_index INTEGER NOT NULL,
    difficulty VARCHAR(20),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT estimated_hours_check CHECK (estimated_hours > 0),
    CONSTRAINT unique_milestone_order UNIQUE(roadmap_version_id, order_index)
);

CREATE TYPE status_value AS ENUM (
    'ACTIVE',
    'COMPLETED'
);

CREATE TABLE userskill (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    skill_id BIGINT NOT NULL REFERENCES skills(id) ON DELETE RESTRICT,
    roadmap_version_id BIGINT NOT NULL REFERENCES roadmap_version(id),
    status status_value NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ 
);

CREATE TYPE progress_status AS ENUM (
    'NOT_STARTED',
    'IN_PROGRESS',
    'COMPLETED',
    'POSTPONED'
);

CREATE TABLE user_milestone_progress (
    id BIGSERIAL PRIMARY KEY,
    user_skill_id BIGINT REFERENCES userskill(id) ON DELETE CASCADE,
    milestone_id BIGINT REFERENCES milestones(id) ON DELETE CASCADE,
    status progress_status NOT NULL,
    completed_at TIMESTAMPTZ,
    postponed_until TIMESTAMPTZ,
    CONSTRAINT unique_user_milestone_progress UNIQUE (user_skill_id, milestone_id)
);

CREATE TABLE study_session (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    skill_id BIGINT REFERENCES skills(id),
    duration_minutes INTEGER NOT NULL,
    notes TEXT,
    studied_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT duration_minutes_check CHECK (duration_minutes > 0)
);

CREATE TABLE streak (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    current_streak INTEGER NOT NULL DEFAULT 0,
    longest_streak INTEGER NOT NULL DEFAULT 0,
    last_activity_date DATE NULL,
    freeze_count INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT invalid_streak_check CHECK(current_streak >= 0),
    CONSTRAINT positive_value_check CHECK(longest_streak >= 0),
    CONSTRAINT invalid_freeze_check CHECK(freeze_count >= 0)
);

-- +goose Down
DROP TABLE IF EXISTS streak;
DROP TABLE IF EXISTS study_session;
DROP TABLE IF EXISTS user_milestone_progress;
DROP TABLE IF EXISTS userskill;
DROP TABLE IF EXISTS milestone;
DROP TABLE IF EXISTS roadmap_version;
DROP TABLE IF EXISTS roadmap;
DROP TABLE IF EXISTS skill;

DROP TYPE IF EXISTS current_status;
DROP TYPE IF EXISTS status_value;
DROP TYPE IF EXISTS progress_status;
