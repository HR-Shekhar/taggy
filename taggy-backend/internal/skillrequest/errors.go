package skillrequest

import "errors"

var (
	ErrInvalidName       = errors.New("skill name is invalid")
	ErrInvalidDescription = errors.New("skill description is invalid")
	ErrSimilarFound      = errors.New("similar skills found; set force=true to continue")
	ErrNearDuplicate     = errors.New("a very similar skill roadmap already exists")
	ErrDuplicatePending  = errors.New("a pending request for this skill name already exists")
	ErrRequestNotFound   = errors.New("skill creation request not found")
	ErrNotPending        = errors.New("request is not pending")
	ErrAIUnavailable     = errors.New("roadmap generation is unavailable")
	ErrAIFailed          = errors.New("roadmap generation failed")
	ErrAIBusy            = errors.New("AI generation is busy; try again shortly")
	ErrSlugTaken         = errors.New("skill slug is already taken")
	ErrNotAdmin          = errors.New("admin access required")
)
