-- +goose Up
DELETE FROM audio_room WHERE status = 'ENDED';

-- +goose Down
SELECT 1;
