-- +goose Up
CREATE TABLE community (
    id BIGSERIAL PRIMARY KEY,
    skill_id BIGINT UNIQUE NOT NULL REFERENCES skills(id),
    name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE community_channel (
    id BIGSERIAL PRIMARY KEY,
    community_id BIGINT NOT NULL REFERENCES community(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT duplicate_channel_name UNIQUE (community_id, name)
);

-- +goose Down
DROP TABLE IF EXISTS community_channel;
DROP TABLE IF EXISTS community;