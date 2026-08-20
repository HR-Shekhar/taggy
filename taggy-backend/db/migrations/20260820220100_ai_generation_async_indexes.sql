-- +goose Up
-- Open catalog requests include both AI generation and admin review.
DROP INDEX IF EXISTS skill_creation_request_one_pending_per_requester_name;
CREATE UNIQUE INDEX skill_creation_request_one_open_per_requester_name
    ON skill_creation_request (requester_id, lower(name))
    WHERE status IN ('PENDING', 'GENERATING');

DROP INDEX IF EXISTS roadmap_edit_request_one_pending_per_requester_skill;
CREATE UNIQUE INDEX roadmap_edit_request_one_open_per_requester_skill
    ON roadmap_edit_request (requester_id, skill_id)
    WHERE status IN ('PENDING', 'GENERATING');

DROP INDEX IF EXISTS pod_quiz_one_in_progress_per_user_pod;
CREATE UNIQUE INDEX pod_quiz_one_active_per_user_pod
    ON pod_quiz (user_id, pod_id)
    WHERE status IN ('IN_PROGRESS', 'GENERATING');

-- +goose Down
DROP INDEX IF EXISTS pod_quiz_one_active_per_user_pod;
CREATE UNIQUE INDEX pod_quiz_one_in_progress_per_user_pod
    ON pod_quiz (user_id, pod_id)
    WHERE status = 'IN_PROGRESS';

DROP INDEX IF EXISTS roadmap_edit_request_one_open_per_requester_skill;
CREATE UNIQUE INDEX roadmap_edit_request_one_pending_per_requester_skill
    ON roadmap_edit_request (requester_id, skill_id)
    WHERE status = 'PENDING';

DROP INDEX IF EXISTS skill_creation_request_one_open_per_requester_name;
CREATE UNIQUE INDEX skill_creation_request_one_pending_per_requester_name
    ON skill_creation_request (requester_id, lower(name))
    WHERE status = 'PENDING';
