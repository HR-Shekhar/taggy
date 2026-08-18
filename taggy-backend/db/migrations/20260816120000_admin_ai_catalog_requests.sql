-- +goose Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TYPE catalog_request_status AS ENUM (
    'PENDING',
    'APPROVED',
    'REJECTED',
    'CANCELLED'
);

CREATE TABLE skill_creation_request (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    requester_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug_candidate VARCHAR(255) NOT NULL,
    description TEXT,
    status catalog_request_status NOT NULL DEFAULT 'PENDING',
    similar_skills JSONB NOT NULL DEFAULT '[]'::jsonb,
    draft_milestones JSONB NOT NULL DEFAULT '[]'::jsonb,
    admin_note TEXT,
    reviewed_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_skill_id BIGINT NULL REFERENCES skills(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT skill_creation_request_name_len CHECK (char_length(trim(name)) BETWEEN 3 AND 255)
);

CREATE UNIQUE INDEX skill_creation_request_one_pending_per_requester_name
    ON skill_creation_request (requester_id, lower(name))
    WHERE status = 'PENDING';

CREATE INDEX skill_creation_request_status_created_idx
    ON skill_creation_request (status, created_at DESC);

CREATE TABLE roadmap_edit_request (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    skill_id BIGINT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    requester_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rationale TEXT,
    status catalog_request_status NOT NULL DEFAULT 'PENDING',
    base_version_number INTEGER NOT NULL,
    draft_milestones JSONB NOT NULL DEFAULT '[]'::jsonb,
    admin_note TEXT,
    reviewed_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_version_id BIGINT NULL REFERENCES roadmap_version(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX roadmap_edit_request_one_pending_per_requester_skill
    ON roadmap_edit_request (requester_id, skill_id)
    WHERE status = 'PENDING';

CREATE INDEX roadmap_edit_request_status_created_idx
    ON roadmap_edit_request (status, created_at DESC);

CREATE INDEX skills_name_trgm_idx ON skills USING gin (name gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS skills_name_trgm_idx;
DROP INDEX IF EXISTS roadmap_edit_request_status_created_idx;
DROP INDEX IF EXISTS roadmap_edit_request_one_pending_per_requester_skill;
DROP TABLE IF EXISTS roadmap_edit_request;
DROP INDEX IF EXISTS skill_creation_request_status_created_idx;
DROP INDEX IF EXISTS skill_creation_request_one_pending_per_requester_name;
DROP TABLE IF EXISTS skill_creation_request;
DROP TYPE IF EXISTS catalog_request_status;
ALTER TABLE users DROP COLUMN IF EXISTS is_admin;
DROP EXTENSION IF EXISTS pg_trgm;
