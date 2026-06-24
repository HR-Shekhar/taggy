-- +goose Up
CREATE TYPE report_target_type AS ENUM (
	'USER',
	'PROPOSAL',
    'POD',
    'MESSAGE',
	'AUDIO_ROOM',
	'COMMUNITY_CHANNEL'
);

CREATE TYPE report_status AS ENUM (
	'OPEN',
	'UNDER_REVIEW',
	'RESOLVED',
	'DISMISSED'
);

CREATE TABLE report (
	id BIGSERIAL PRIMARY KEY,
	reporter_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
	target_type report_target_type NOT NULL,
	target_id BIGINT NOT NULL,
	reason TEXT NOT NULL,
	status report_status NOT NULL DEFAULT 'OPEN',
	resolved_at TIMESTAMPTZ NULL,
	resolved_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT report_resolution_state CHECK (
		(resolved_by IS NULL AND resolved_at IS NULL)
		OR (resolved_by IS NOT NULL)
	),
    CONSTRAINT valid_report_resolution CHECK (
        (resolved_by IS NULL AND resolved_at IS NULL)
        OR
        (resolved_by IS NOT NULL AND resolved_at IS NOT NULL)
    )
);

-- +goose Down
DROP TABLE IF EXISTS report;
DROP TYPE IF EXISTS report_status;
DROP TYPE IF EXISTS report_target_type;
