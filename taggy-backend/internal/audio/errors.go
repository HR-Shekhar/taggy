package audio

import "errors"

var (
	ErrRoomNotFound          = errors.New("audio room not found")
	ErrPodNotFound           = errors.New("pod not found")
	ErrChannelNotFound       = errors.New("channel not found")
	ErrNotEnrolledInSkill    = errors.New("must be enrolled in the skill to access community audio rooms")
	ErrNotAcceptedPodMember  = errors.New("must be an accepted pod member to access pod audio rooms")
	ErrNotRoomHost           = errors.New("only the room host can perform this action")
	ErrRoomNotActive         = errors.New("audio room is not active")
	ErrActiveRoomExists      = errors.New("an active audio room already exists")
	ErrRoomFull              = errors.New("audio room is full")
	ErrNotParticipant        = errors.New("not an active participant in this audio room")
	ErrInvalidRoomTitle      = errors.New("audio room title is invalid")
	ErrLiveKitNotConfigured  = errors.New("livekit not configured")
)
