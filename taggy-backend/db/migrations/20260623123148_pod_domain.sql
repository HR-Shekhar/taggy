-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE pods (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    slug CITEXT UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    owner_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    skill_id BIGINT NOT NULL REFERENCES skills(id),
    max_members INTEGER NOT NULL DEFAULT 7,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_number_of_members CHECK (max_members > 0 AND max_members <= 7)
);

CREATE TYPE user_pod_membership_status AS ENUM (
    'PENDING',
    'ACCEPTED',
    'REJECTED',
    'LEFT',
    'REMOVED'
);

CREATE TABLE pod_membership (
    id BIGSERIAL PRIMARY KEY,
    pod_id BIGINT NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL  REFERENCES users(id),
    status user_pod_membership_status NOT NULL,
    joined_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_joining_same_pod UNIQUE (pod_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS pods;
DROP TABLE IF EXISTS pod_membership;

DROP EXTENSION IF EXISTS pgcrypto;
DROP EXTENSION IF EXISTS citext;
