-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE subscription_tier AS ENUM (
    'FREE',
    'PREMIUM'
);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    public_id get_random_uuid(),
    email CITEXT NOT NULL UNIQUE,
    username CITEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    profile_picture_url TEXT DEFAULT NULL,
    bio TEXT DEFAULT NULL,
    subscription subscription_tier NOT NULL DEFAULT 'FREE',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email
ON users(email);


-- +goose Down
DROP EXTENSION IF EXISTS pgcrypto;

DROP EXTENSION IF EXISTS citext;

DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS subscription_tier;  