-- +goose Up
CREATE TABLE skill (
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

CREATE TABLE roadmap_version (
    id BIGSERIAL PRIMARY KEY,
    roadmap_id BIGINT references roadmap(id),
    version_number INTEGER NOT NULL,
    status current_status NOT NULL,
    generated_by VARCHAR(20) NOT NULL,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE roadmap (
    id BIGSERIAL PRIMARY KEY,
    skill_id BIGINT NOT NULL REFERENCES skill(id),
    current_version_id BIGINT REFERENCES roadmap_version(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE milestone (
    id BIGSERIAL PRIMARY KEY
    roadmap_version_id BIGINT REFERENCES roadmap_version(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    estimated_hours INTEGER,
    order_index INTEGER NOT NULL,
    difficulty VARCHAR(20),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT estimated_hours_check CHECK (estimated_hours > 0),
    CONSTRAINT duplication_check UNIQUE(roadmap_version_id, order_index)
);

-- +goose Down
DROP TABLE skill;
