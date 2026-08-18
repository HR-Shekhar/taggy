package progress

// HTTP request/response DTOs for study sessions and streak endpoints.

type logStudySessionRequest struct {
	SkillSlug        string  `json:"skill_slug" validate:"required"`
	DurationMinutes  int32   `json:"duration_minutes" validate:"required,min=1"`
	Notes            *string `json:"notes" validate:"omitempty,max=500"`
	StudiedAt        *string `json:"studied_at" validate:"omitempty"`
}

type studySessionResponse struct {
	SkillSlug       string  `json:"skill_slug"`
	DurationMinutes int32   `json:"duration_minutes"`
	Notes           *string `json:"notes"`
	StudiedAt       string  `json:"studied_at"`
	CreatedAt       string  `json:"created_at"`
}

type streakResponse struct {
	CurrentStreak    int32   `json:"current_streak"`
	LongestStreak    int32   `json:"longest_streak"`
	LastActivityDate *string `json:"last_activity_date"`
	FreezeCount      int32   `json:"freeze_count"`
}

type progressSummaryResponse struct {
	TotalMinutes   int64 `json:"total_minutes"`
	WeeklyMinutes  int64 `json:"weekly_minutes"`
	MonthlyMinutes int64 `json:"monthly_minutes"`
	CurrentStreak  int32 `json:"current_streak"`
	LongestStreak  int32 `json:"longest_streak"`
}

type logStudySessionResponse struct {
	Session studySessionResponse `json:"session"`
	Streak  streakResponse       `json:"streak"`
}
