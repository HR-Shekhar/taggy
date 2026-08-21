-- +goose Up
-- Hold signup details until email OTP succeeds. No users row until then.
CREATE TABLE pending_registrations (
    id BIGSERIAL PRIMARY KEY,
    email CITEXT NOT NULL,
    username CITEXT NOT NULL,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    otp_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pending_registrations_username_length_check
        CHECK (length(username) BETWEEN 3 AND 30)
);

CREATE UNIQUE INDEX pending_registrations_email_key ON pending_registrations (email);
CREATE UNIQUE INDEX pending_registrations_username_key ON pending_registrations (username);

-- +goose Down
DROP TABLE IF EXISTS pending_registrations;
