-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE subscription_tier AS ENUM (
    'FREE',
    'PREMIUM'
);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    email CITEXT NOT NULL UNIQUE,
    username CITEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    profile_picture_url TEXT,
    bio TEXT,
    subscription subscription_tier NOT NULL DEFAULT 'FREE',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_username_length_check
        CHECK (
            length(username) BETWEEN 3 AND 30
        ),
    CONSTRAINT deleted_user_check
        CHECK (
            (is_deleted = FALSE AND deleted_at IS NULL)
            OR
            (is_deleted = TRUE)
        )
);

CREATE TYPE provider_name AS ENUM (
    'local',
    'google'
);

CREATE TABLE user_identity (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider provider_name NOT NULL DEFAULT 'local',
    provider_user_id VARCHAR(255) NULL,
    password_hash TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT identity_provider_check
        CHECK (
            (provider = 'google' AND provider_user_id IS NOT NULL)
            OR
            (provider = 'local' AND password_hash IS NOT NULL)
        ),
    CONSTRAINT unique_user_provider UNIQUE (user_id, provider),
    CONSTRAINT unique_provider_identity UNIQUE (provider, provider_user_id)
);

CREATE TABLE username_history (
    id  BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    username CITEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT username_history_username_length_check
        CHECK (
            length(username) BETWEEN 3 AND 30
        )
);

-- +goose Down
DROP TABLE IF EXISTS username_history;
DROP TABLE IF EXISTS user_identity;
DROP TABLE IF EXISTS users;


DROP TYPE IF EXISTS subscription_tier;  
DROP TYPE IF EXISTS provider_name;  

DROP EXTENSION IF EXISTS pgcrypto;
DROP EXTENSION IF EXISTS citext;
