-- +goose Up
CREATE TYPE pod_member_role AS ENUM (
    'OWNER',
    'ADMIN',
    'MEMBER'
);

ALTER TABLE pod_membership
    ADD COLUMN role pod_member_role NOT NULL DEFAULT 'MEMBER';

-- Existing accepted owners become OWNER; everyone else stays MEMBER.
UPDATE pod_membership pm
SET role = 'OWNER'
FROM pods p
WHERE pm.pod_id = p.id
  AND pm.user_id = p.owner_id
  AND pm.status = 'ACCEPTED';

-- +goose Down
ALTER TABLE pod_membership DROP COLUMN IF EXISTS role;
DROP TYPE IF EXISTS pod_member_role;
