-- =========================================
-- AUDIO ROOMS
-- =========================================

-- name: CreateAudioRoom :one
INSERT INTO audio_room (
    public_id,
    room_type,
    pod_id,
    community_channel_id,
    host_id,
    title,
    description,
    livekit_room_name,
    status,
    actual_start_time,
    max_participants
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING
    id,
    public_id,
    room_type,
    pod_id,
    community_channel_id,
    host_id,
    title,
    description,
    livekit_room_name,
    status,
    scheduled_start_time,
    actual_start_time,
    ended_at,
    max_participants,
    created_at,
    updated_at;


-- name: GetAudioRoomByPublicID :one
SELECT
    ar.id,
    ar.public_id,
    ar.room_type,
    ar.pod_id,
    ar.community_channel_id,
    ar.host_id,
    ar.title,
    ar.description,
    ar.livekit_room_name,
    ar.status,
    ar.scheduled_start_time,
    ar.actual_start_time,
    ar.ended_at,
    ar.max_participants,
    ar.created_at,
    ar.updated_at,
    u.username AS host_username,
    p.slug AS pod_slug,
    s.slug AS skill_slug,
    cc.slug AS channel_slug
FROM audio_room ar
INNER JOIN users u ON u.id = ar.host_id
LEFT JOIN pods p ON p.id = ar.pod_id
LEFT JOIN community_channel cc ON cc.id = ar.community_channel_id
LEFT JOIN community c ON c.id = cc.community_id
LEFT JOIN skills s ON s.id = c.skill_id
WHERE ar.public_id = $1;


-- name: ListAudioRoomsByPodID :many
SELECT
    ar.id,
    ar.public_id,
    ar.room_type,
    ar.pod_id,
    ar.community_channel_id,
    ar.host_id,
    ar.title,
    ar.description,
    ar.livekit_room_name,
    ar.status,
    ar.scheduled_start_time,
    ar.actual_start_time,
    ar.ended_at,
    ar.max_participants,
    ar.created_at,
    ar.updated_at,
    u.username AS host_username
FROM audio_room ar
INNER JOIN users u ON u.id = ar.host_id
WHERE ar.pod_id = sqlc.arg(pod_id)
  AND (sqlc.narg(status)::audio_room_status IS NULL OR ar.status = sqlc.narg(status))
ORDER BY ar.created_at DESC;


-- name: ListAudioRoomsByChannelID :many
SELECT
    ar.id,
    ar.public_id,
    ar.room_type,
    ar.pod_id,
    ar.community_channel_id,
    ar.host_id,
    ar.title,
    ar.description,
    ar.livekit_room_name,
    ar.status,
    ar.scheduled_start_time,
    ar.actual_start_time,
    ar.ended_at,
    ar.max_participants,
    ar.created_at,
    ar.updated_at,
    u.username AS host_username
FROM audio_room ar
INNER JOIN users u ON u.id = ar.host_id
WHERE ar.community_channel_id = sqlc.arg(channel_id)
  AND (sqlc.narg(status)::audio_room_status IS NULL OR ar.status = sqlc.narg(status))
ORDER BY ar.created_at DESC;


-- name: GetActiveAudioRoomByPodID :one
SELECT id, public_id, room_type, pod_id, community_channel_id, host_id, title, description,
       livekit_room_name, status, scheduled_start_time, actual_start_time, ended_at,
       max_participants, created_at, updated_at
FROM audio_room
WHERE pod_id = $1
  AND status = 'ACTIVE'
LIMIT 1;


-- name: GetActiveAudioRoomByChannelID :one
SELECT id, public_id, room_type, pod_id, community_channel_id, host_id, title, description,
       livekit_room_name, status, scheduled_start_time, actual_start_time, ended_at,
       max_participants, created_at, updated_at
FROM audio_room
WHERE community_channel_id = $1
  AND status = 'ACTIVE'
LIMIT 1;


-- name: EndAudioRoom :one
UPDATE audio_room
SET
    status = 'ENDED',
    ended_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND status = 'ACTIVE'
RETURNING
    id,
    public_id,
    room_type,
    pod_id,
    community_channel_id,
    host_id,
    title,
    description,
    livekit_room_name,
    status,
    scheduled_start_time,
    actual_start_time,
    ended_at,
    max_participants,
    created_at,
    updated_at;


-- name: EndAllActiveAudioRooms :execrows
UPDATE audio_room
SET
    status = 'ENDED',
    ended_at = NOW(),
    updated_at = NOW()
WHERE status = 'ACTIVE';


-- name: CountActiveParticipants :one
SELECT COUNT(*)::bigint
FROM audio_room_participant
WHERE audio_room_id = $1
  AND left_at IS NULL;


-- name: ListStaleEmptyActiveAudioRooms :many
SELECT
    ar.id,
    ar.public_id,
    ar.livekit_room_name
FROM audio_room ar
WHERE ar.status = 'ACTIVE'
  AND NOT EXISTS (
      SELECT 1
      FROM audio_room_participant arp
      WHERE arp.audio_room_id = ar.id
        AND arp.left_at IS NULL
  )
  AND COALESCE(
      (
          SELECT MAX(arp.left_at)
          FROM audio_room_participant arp
          WHERE arp.audio_room_id = ar.id
      ),
      ar.actual_start_time,
      ar.created_at
  ) < sqlc.arg(empty_before);


-- name: ListActiveAudioRooms :many
SELECT
    id,
    public_id,
    livekit_room_name
FROM audio_room
WHERE status = 'ACTIVE'
ORDER BY id;


-- name: DeleteAudioRoom :exec
DELETE FROM audio_room
WHERE id = $1;
