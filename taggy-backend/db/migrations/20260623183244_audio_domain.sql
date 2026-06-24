-- +goose Up
CREATE TYPE audio_room_type AS ENUM (
	'POD',
	'COMMUNITY_CHANNEL'
);

CREATE TYPE audio_room_status AS ENUM (
	'SCHEDULED',
	'ACTIVE',
	'ENDED',
	'CANCELLED'
);

CREATE TYPE audio_room_participant_role AS ENUM (
	'HOST',
	'SPEAKER',
	'LISTENER'
);

CREATE TABLE audio_room (
	id BIGSERIAL PRIMARY KEY,
	public_id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
	room_type audio_room_type NOT NULL,
	pod_id BIGINT NULL REFERENCES pods(id) ON DELETE RESTRICT,
	community_channel_id BIGINT NULL REFERENCES community_channel(id) ON DELETE RESTRICT,
	host_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
	title VARCHAR(255) NOT NULL,
	description TEXT NULL,
	livekit_room_name VARCHAR(255) UNIQUE NOT NULL,
	status audio_room_status NOT NULL,
	scheduled_start_time TIMESTAMPTZ NULL,
	actual_start_time TIMESTAMPTZ NULL,
	ended_at TIMESTAMPTZ NULL,
	max_participants INTEGER NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT audio_room_target_check CHECK (
		(
			room_type = 'POD'
			AND pod_id IS NOT NULL
			AND community_channel_id IS NULL
		)
		OR
		(
			room_type = 'COMMUNITY_CHANNEL'
			AND community_channel_id IS NOT NULL
			AND pod_id IS NULL
		)
	),
    CONSTRAINT minimum_participant_value_check CHECK (
        max_participants IS NULL
        OR max_participants > 0
    )
);

CREATE TABLE audio_room_participant (
	id BIGSERIAL PRIMARY KEY,
	audio_room_id BIGINT NOT NULL REFERENCES audio_room(id) ON DELETE CASCADE,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	joined_at TIMESTAMPTZ NOT NULL,
	left_at TIMESTAMPTZ NULL,
	duration_seconds INTEGER NULL,
	role audio_room_participant_role NOT NULL,
	CONSTRAINT audio_room_participant_duration_check CHECK (
		duration_seconds IS NULL
		OR duration_seconds >= 0
	),
	CONSTRAINT audio_room_participant_unique UNIQUE (audio_room_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS audio_room_participant;
DROP TABLE IF EXISTS audio_room;
DROP TYPE IF EXISTS audio_room_participant_role;
DROP TYPE IF EXISTS audio_room_status;
DROP TYPE IF EXISTS audio_room_type;