-- +goose Up
CREATE TABLE skill (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)

CREATE TABLE roadmap_version (
        
)

CREATE TABLE roadmap (
    id BIGSERIAL PRIMARY KEY DEFAULT,
    skill_id BIGINT REFERENCES skill(id),
    current_version_id BIGINT NULL REFERENCES roadmap_version(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)

-- +goose Down
DROP TABLE skill;
