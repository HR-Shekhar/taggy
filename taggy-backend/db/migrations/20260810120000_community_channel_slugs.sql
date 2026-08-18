-- +goose Up
ALTER TABLE community_channel
    ADD COLUMN slug VARCHAR(60);

-- Backfill existing rows from name (lowercase, spaces → hyphens).
UPDATE community_channel
SET slug = LOWER(REGEXP_REPLACE(TRIM(name), '[^a-zA-Z0-9]+', '-', 'g'))
WHERE slug IS NULL;

-- Collapse leading/trailing hyphens that the regex may produce.
UPDATE community_channel
SET slug = TRIM(BOTH '-' FROM slug)
WHERE slug IS NOT NULL;

-- Ensure uniqueness if collisions remain by appending id.
UPDATE community_channel c
SET slug = c.slug || '-' || c.id::text
WHERE EXISTS (
    SELECT 1
    FROM community_channel o
    WHERE o.community_id = c.community_id
      AND o.slug = c.slug
      AND o.id < c.id
);

ALTER TABLE community_channel
    ALTER COLUMN slug SET NOT NULL;

ALTER TABLE community_channel
    ADD CONSTRAINT community_channel_slug_unique UNIQUE (community_id, slug);

-- Seed default channels for the web-development community (idempotent).
INSERT INTO community_channel (community_id, name, slug, description)
SELECT c.id, v.name, v.slug, v.description
FROM community c
INNER JOIN skills s ON s.id = c.skill_id
CROSS JOIN (
    VALUES
        ('General', 'general', 'General discussion for web development learners.'),
        ('Resources', 'resources', 'Share tutorials, docs, and learning resources.'),
        ('Projects', 'projects', 'Show your builds and get feedback.')
) AS v(name, slug, description)
WHERE s.slug = 'web-development'
ON CONFLICT (community_id, slug) DO NOTHING;

-- +goose Down
ALTER TABLE community_channel DROP CONSTRAINT IF EXISTS community_channel_slug_unique;
ALTER TABLE community_channel DROP COLUMN IF EXISTS slug;
