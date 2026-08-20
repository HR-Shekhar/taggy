package roadmaprequest

import "errors"

var (
	ErrNotEnrolled      = errors.New("must be enrolled in the skill")
	ErrSkillNotFound    = errors.New("skill not found")
	ErrDuplicatePending = errors.New("a pending roadmap edit request already exists")
	ErrRequestNotFound  = errors.New("roadmap edit request not found")
	ErrNotPending       = errors.New("request is not pending")
	ErrAIUnavailable    = errors.New("roadmap generation is unavailable")
	ErrAIFailed         = errors.New("roadmap generation failed")
	ErrAIBusy           = errors.New("AI generation is busy; try again shortly")
	ErrNoActiveVersion  = errors.New("skill has no active roadmap version")
	ErrNotAdmin         = errors.New("admin access required")
)
