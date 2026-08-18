package progress

import "errors"

var (
	ErrNotEnrolledInSkill = errors.New("not enrolled in this skill")
	ErrNoStreak             = errors.New("streak not found")
)
