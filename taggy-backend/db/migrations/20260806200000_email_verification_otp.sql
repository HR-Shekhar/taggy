-- +goose Up
-- OTP codes for email verification. Store hash only (same idea as refresh tokens).

CREATE TABLE email_verification_otp (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    otp_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX email_verification_otp_user_active_idx
    ON email_verification_otp (user_id, expires_at DESC)
    WHERE consumed_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS email_verification_otp;
