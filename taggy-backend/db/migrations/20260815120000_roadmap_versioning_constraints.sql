-- +goose Up
-- Align roadmap schema with db_design: one roadmap per skill, unique version numbers,
-- one enrollment per user+skill, at most one ACTIVE version per roadmap.

CREATE UNIQUE INDEX IF NOT EXISTS roadmaps_skill_id_unique
    ON roadmaps (skill_id);

CREATE UNIQUE INDEX IF NOT EXISTS roadmap_version_roadmap_id_version_number_unique
    ON roadmap_version (roadmap_id, version_number);

-- Learners stay on a pinned version; only one row per user+skill.
CREATE UNIQUE INDEX IF NOT EXISTS userskill_user_id_skill_id_unique
    ON userskill (user_id, skill_id);

-- Official "current" catalog version: at most one ACTIVE per roadmap.
CREATE UNIQUE INDEX IF NOT EXISTS roadmap_version_one_active_per_roadmap
    ON roadmap_version (roadmap_id)
    WHERE status = 'ACTIVE';

-- +goose Down
DROP INDEX IF EXISTS roadmap_version_one_active_per_roadmap;
DROP INDEX IF EXISTS userskill_user_id_skill_id_unique;
DROP INDEX IF EXISTS roadmap_version_roadmap_id_version_number_unique;
DROP INDEX IF EXISTS roadmaps_skill_id_unique;
