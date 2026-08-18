-- Wipe all skills, communities, pods, roadmaps, and related data.
-- Keeps users / auth. Run:
--   docker exec -i taggy psql -U postgres -d taggy -v ON_ERROR_STOP=1 -f - < scripts/wipe_skills.sql
-- Or paste into: docker exec -it taggy psql -U postgres -d taggy

BEGIN;

DELETE FROM audio_room_participant
WHERE audio_room_id IN (
  SELECT ar.id FROM audio_room ar
  WHERE ar.community_channel_id IN (
    SELECT cc.id FROM community_channel cc
    JOIN community c ON c.id = cc.community_id
  )
  OR ar.pod_id IN (SELECT id FROM pods)
);

DELETE FROM audio_room
WHERE community_channel_id IN (
  SELECT cc.id FROM community_channel cc
  JOIN community c ON c.id = cc.community_id
)
OR pod_id IN (SELECT id FROM pods);

DELETE FROM message
WHERE community_channel_id IN (
  SELECT cc.id FROM community_channel cc
  JOIN community c ON c.id = cc.community_id
)
OR pod_id IN (SELECT id FROM pods);

DELETE FROM pod_membership WHERE pod_id IN (SELECT id FROM pods);
DELETE FROM pods;

DELETE FROM community_channel WHERE community_id IN (SELECT id FROM community);
DELETE FROM community;

DELETE FROM user_milestone_progress
WHERE milestone_id IN (SELECT id FROM milestones)
   OR user_skill_id IN (SELECT id FROM userskill);

DELETE FROM study_session WHERE skill_id IS NOT NULL;
DELETE FROM userskill;

DELETE FROM proposal_vote
WHERE proposal_id IN (SELECT id FROM milestone_proposal);
DELETE FROM milestone_proposal;

DELETE FROM roadmap_edit_request;
UPDATE skill_creation_request SET created_skill_id = NULL WHERE created_skill_id IS NOT NULL;
DELETE FROM skill_creation_request;

UPDATE roadmaps SET current_version_id = NULL;
DELETE FROM milestones;
DELETE FROM roadmap_version;
DELETE FROM roadmaps;

DELETE FROM skills;

COMMIT;

SELECT
  (SELECT count(*) FROM skills) AS skills,
  (SELECT count(*) FROM community) AS communities,
  (SELECT count(*) FROM pods) AS pods,
  (SELECT count(*) FROM roadmaps) AS roadmaps,
  (SELECT count(*) FROM milestones) AS milestones,
  (SELECT count(*) FROM users) AS users_kept;
