package community

import "errors"

var (
	ErrCommunityNotFound     = errors.New("community not found")
	ErrChannelNotFound       = errors.New("channel not found")
	ErrMessageNotFound       = errors.New("message not found")
	ErrPodNotFound           = errors.New("pod not found")
	ErrNotEnrolledInSkill    = errors.New("must be enrolled in the skill to access community chat")
	ErrNotAcceptedPodMember  = errors.New("must be an accepted pod member to access pod chat")
	ErrNotMessageAuthor      = errors.New("only the message author can perform this action")
	ErrInvalidMessageContent = errors.New("message content is invalid")
	ErrInvalidReplyTarget    = errors.New("reply target is invalid")
	ErrInvalidChatRoom       = errors.New("chat room is invalid")
)
