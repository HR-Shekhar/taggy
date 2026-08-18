package pod

import "errors"

var (
	ErrPodNotFound          = errors.New("pod not found")
	ErrNotEnrolledInSkill   = errors.New("must be enrolled in the skill to create or join a pod")
	ErrAlreadyInActivePod   = errors.New("already in an active pod")
	ErrAlreadyMember        = errors.New("already a member of this pod")
	ErrAlreadyPending       = errors.New("join request already pending")
	ErrPodFull              = errors.New("pod is full")
	ErrNotPodOwner          = errors.New("only the pod owner can perform this action")
	ErrMembershipNotFound   = errors.New("membership not found")
	ErrMembershipNotPending = errors.New("membership is not pending")
	ErrCannotRemoveOwner    = errors.New("cannot remove the pod owner")
	ErrCannotChangeOwnRole  = errors.New("cannot change your own role")
	ErrNotAcceptedMember    = errors.New("not an accepted member of this pod")
	ErrInvalidPodName       = errors.New("pod name is invalid")
	ErrInvalidPodSlug       = errors.New("pod slug is invalid")
	ErrPodSlugTaken         = errors.New("pod slug already in use")
	ErrInvalidMemberRole    = errors.New("invalid member role")
	ErrPodNotEmpty          = errors.New("pod is not empty; transfer ownership or remove members first")
)
