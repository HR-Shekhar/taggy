-- =========================================
-- AUDIO ROOM PARTICIPANTS
-- =========================================

-- name: UpsertAudioRoomParticipant :one
INSERT INTO audio_room_participant (
    audio_room_id,
    user_id,
    joined_at,
    left_at,
    duration_seconds,
    role
)
VALUES ($1, $2, $3, NULL, NULL, $4)
ON CONFLICT (audio_room_id, user_id) DO UPDATE
SET
    joined_at = EXCLUDED.joined_at,
    left_at = NULL,
    duration_seconds = NULL,
    role = CASE
        WHEN audio_room_participant.role = 'HOST' THEN audio_room_participant.role
        ELSE EXCLUDED.role
    END
RETURNING id, audio_room_id, user_id, joined_at, left_at, duration_seconds, role;


-- name: LeaveAudioRoomParticipant :one
UPDATE audio_room_participant
SET
    left_at = NOW(),
    duration_seconds = GREATEST(
        0,
        EXTRACT(EPOCH FROM (NOW() - joined_at))::integer
    )
WHERE audio_room_id = $1
  AND user_id = $2
  AND left_at IS NULL
RETURNING id, audio_room_id, user_id, joined_at, left_at, duration_seconds, role;


-- name: LeaveAllActiveAudioParticipants :execrows
UPDATE audio_room_participant
SET
    left_at = NOW(),
    duration_seconds = GREATEST(
        0,
        EXTRACT(EPOCH FROM (NOW() - joined_at))::integer
    )
WHERE left_at IS NULL;


-- name: GetAudioRoomParticipant :one
SELECT id, audio_room_id, user_id, joined_at, left_at, duration_seconds, role
FROM audio_room_participant
WHERE audio_room_id = $1
  AND user_id = $2;


-- name: ListActiveAudioRoomParticipants :many
SELECT
    arp.id,
    arp.audio_room_id,
    arp.user_id,
    arp.joined_at,
    arp.left_at,
    arp.duration_seconds,
    arp.role,
    u.username,
    u.name AS user_name
FROM audio_room_participant arp
INNER JOIN users u ON u.id = arp.user_id
WHERE arp.audio_room_id = $1
  AND arp.left_at IS NULL
ORDER BY
    CASE arp.role
        WHEN 'HOST' THEN 0
        WHEN 'SPEAKER' THEN 1
        ELSE 2
    END,
    arp.joined_at ASC;
