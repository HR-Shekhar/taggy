package progress

import "time"

// LogStudySessionInput is the service-layer input for recording study time.
type LogStudySessionInput struct {
	SkillSlug        string
	DurationMinutes  int32
	Notes            *string
	StudiedAt        time.Time
}

// StreakWrite holds streak values to persist after a study session is logged.
type StreakWrite struct {
	CurrentStreak    int32
	LongestStreak    int32
	LastActivityDate time.Time
	IsNew            bool
}
