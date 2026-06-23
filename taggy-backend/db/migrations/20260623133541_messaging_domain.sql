-- +goose Up
CREATE TABLE message (
    id BIGSERIAL PRIMARY KEY,
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,   
    community_channel_id BIGINT REFERENCES community_channel(id) ON DELETE CASCADE,
    pod_id BIGINT REFERENCES pods(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    edited_at TIMESTAMPTZ,             
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT message_one_location 
    CHECK (
        (
            community_channel_id IS NOT NULL
            AND pod_id IS NULL
        )
        OR 
        (
            community_channel_id IS NULL
            AND pod_id IS NOT NULL
        )
    ) 
);  
-- +goose Down
DROP TABLE IF EXISTS message;