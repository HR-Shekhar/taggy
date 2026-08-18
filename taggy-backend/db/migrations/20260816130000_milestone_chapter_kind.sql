-- +goose Up
ALTER TABLE milestones
    ADD COLUMN IF NOT EXISTS chapter TEXT,
    ADD COLUMN IF NOT EXISTS kind VARCHAR(20) NOT NULL DEFAULT 'TOPIC';

ALTER TABLE milestones
    DROP CONSTRAINT IF EXISTS milestones_kind_check;

ALTER TABLE milestones
    ADD CONSTRAINT milestones_kind_check
    CHECK (kind IN ('CHAPTER', 'TOPIC'));

COMMENT ON COLUMN milestones.chapter IS 'Curriculum chapter/section title for grouping topic milestones';
COMMENT ON COLUMN milestones.kind IS 'CHAPTER = section overview; TOPIC = concrete learning unit';

-- +goose Down
ALTER TABLE milestones DROP CONSTRAINT IF EXISTS milestones_kind_check;
ALTER TABLE milestones DROP COLUMN IF EXISTS kind;
ALTER TABLE milestones DROP COLUMN IF EXISTS chapter;
