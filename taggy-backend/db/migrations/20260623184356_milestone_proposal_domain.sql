-- +goose Up
CREATE TYPE milestone_proposal_type AS ENUM (
	'ADD',
	'REMOVE',
	'EDIT',
	'REORDER',
	'MERGE',
	'SPLIT'
);

CREATE TYPE milestone_proposal_status AS ENUM (
	'PENDING',
	'APPROVED',
	'REJECTED',
	'IMPLEMENTED'
);

CREATE TYPE proposal_vote_type AS ENUM (
	'UPVOTE',
	'DOWNVOTE'
);

CREATE TABLE milestone_proposal (
	id BIGSERIAL PRIMARY KEY,
	roadmap_version_id BIGINT NOT NULL REFERENCES roadmap_version(id) ON DELETE RESTRICT,
	proposer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
	proposal_type milestone_proposal_type NOT NULL,
	title VARCHAR(255) NOT NULL,
	description TEXT NULL,
	status milestone_proposal_status NOT NULL DEFAULT 'PENDING',
	reviewed_at TIMESTAMPTZ NULL,
	reviewed_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT milestone_proposal_review_state CHECK (
		(reviewed_by IS NULL AND reviewed_at IS NULL)
		OR (reviewed_by IS NOT NULL)
	),
	CONSTRAINT milestone_proposal_review_required CHECK (
		status NOT IN ('APPROVED', 'REJECTED')
		OR reviewed_by IS NOT NULL
	),
    CONSTRAINT valid_review_check CHECK (
		(reviewed_by IS NULL AND reviewed_at IS NULL)
        OR
        (reviewed_by IS NOT NULL AND reviewed_at IS NOT NULL)
	)
);

CREATE TABLE proposal_vote (
	id BIGSERIAL PRIMARY KEY,
	proposal_id BIGINT NOT NULL REFERENCES milestone_proposal(id) ON DELETE CASCADE,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
	vote_type proposal_vote_type NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT proposal_vote_unique UNIQUE (proposal_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS proposal_vote;
DROP TABLE IF EXISTS milestone_proposal;
DROP TYPE IF EXISTS proposal_vote_type;
DROP TYPE IF EXISTS milestone_proposal_status;
DROP TYPE IF EXISTS milestone_proposal_type;
