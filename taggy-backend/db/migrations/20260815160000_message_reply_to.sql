-- +goose Up
ALTER TABLE message
    ADD COLUMN reply_to_message_id BIGINT NULL REFERENCES message(id) ON DELETE SET NULL;

CREATE INDEX message_reply_to_message_id_idx ON message (reply_to_message_id)
    WHERE reply_to_message_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS message_reply_to_message_id_idx;
ALTER TABLE message DROP COLUMN IF EXISTS reply_to_message_id;
