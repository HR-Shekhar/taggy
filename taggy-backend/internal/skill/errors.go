package skill

import "errors"

var (
	ErrSkillNotFound            = errors.New("skill not found")
	ErrCommunityNotFound        = errors.New("community not found")
	ErrRoadmapNotFound          = errors.New("active roadmap not found for skill")
	ErrAlreadyEnrolled          = errors.New("already enrolled in this skill")
	ErrActiveSkillLimit         = errors.New("free users can only have one active skill")
	ErrUserSkillNotFound        = errors.New("skill enrollment not found")
	ErrMilestoneNotFound        = errors.New("milestone not found")
	ErrMilestoneOutOfOrder      = errors.New("complete previous milestones first")
	ErrSubtopicsIncomplete      = errors.New("complete all subtopics before finishing this topic")
	ErrMilestoneAlreadyComplete = errors.New("milestone already completed")
	ErrInvalidMilestoneAction   = errors.New("invalid milestone action")
	ErrVersionNotFound          = errors.New("roadmap version not found")
	ErrVersionNotSelectable     = errors.New("draft roadmap versions cannot be selected")
	ErrAlreadyOnVersion         = errors.New("already on this roadmap version")
)
