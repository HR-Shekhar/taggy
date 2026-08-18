-- +goose Up
-- Public API uses slugs instead of numeric IDs for milestones.

ALTER TABLE milestones ADD COLUMN slug VARCHAR(255);

UPDATE milestones
SET slug = lower(trim(both '-' from regexp_replace(title, '[^a-zA-Z0-9]+', '-', 'g')))
WHERE slug IS NULL;

ALTER TABLE milestones ALTER COLUMN slug SET NOT NULL;

CREATE UNIQUE INDEX milestones_roadmap_version_slug_unique
    ON milestones (roadmap_version_id, slug);

-- +goose Down
DROP INDEX IF EXISTS milestones_roadmap_version_slug_unique;
ALTER TABLE milestones DROP COLUMN IF EXISTS slug;
