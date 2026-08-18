-- +goose Up
CREATE INDEX idx_study_session_user_studied_at
    ON study_session (user_id, studied_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_study_session_user_studied_at;
