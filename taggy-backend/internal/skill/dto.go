package skill

// HTTP request/response DTOs for skill and milestone endpoints.

type skillResponse struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
}

type communityResponse struct {
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type skillDetailResponse struct {
	Skill     skillResponse     `json:"skill"`
	Community communityResponse `json:"community"`
}

type userSkillResponse struct {
	SkillSlug             string  `json:"skill_slug"`
	SkillName             string  `json:"skill_name"`
	Status                string  `json:"status"`
	StartedAt             string  `json:"started_at"`
	RoadmapVersionNumber  int32   `json:"roadmap_version_number"`
	RoadmapVersionStatus  string  `json:"roadmap_version_status"`
	MilestoneCount        int64   `json:"milestone_count"`
	CompletedCount        int64   `json:"completed_count"`
	CompletionPercent     float64 `json:"completion_percent"`
}

type switchRoadmapVersionRequest struct {
	VersionNumber int32 `json:"version_number" validate:"required,min=1"`
}

type joinSkillResponse struct {
	UserSkill userSkillResponse `json:"user_skill"`
	Community communityResponse `json:"community"`
}

type milestoneResponse struct {
	Slug           string  `json:"slug"`
	Title          string  `json:"title"`
	Description    *string `json:"description"`
	EstimatedHours *int32  `json:"estimated_hours"`
	OrderIndex     int32   `json:"order_index"`
	Difficulty     *string `json:"difficulty"`
	Chapter        *string `json:"chapter,omitempty"`
	Kind           string  `json:"kind"`
	Status         string  `json:"status"`
	CompletedAt    *string `json:"completed_at"`
	PostponedUntil *string `json:"postponed_until"`
}

type updateMilestoneRequest struct {
	Action         string  `json:"action" validate:"required,oneof=COMPLETE POSTPONE"`
	PostponedUntil *string `json:"postponed_until" validate:"required_if=Action POSTPONE"`
}
