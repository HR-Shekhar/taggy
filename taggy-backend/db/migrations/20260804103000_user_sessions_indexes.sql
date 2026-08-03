-- +goose Up
CREATE UNIQUE INDEX idx_user_sessions_refresh_token_hash
    ON user_sessions (refresh_token_hash);

CREATE INDEX idx_user_sessions_user_id
    ON user_sessions (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_user_sessions_user_id;
DROP INDEX IF EXISTS idx_user_sessions_refresh_token_hash;
