package roadmap

// HTTP DTOs for roadmap catalog and version endpoints.

type roadmapSummaryResponse struct {
	SkillSlug       string                `json:"skill_slug"`
	SkillName       string                `json:"skill_name"`
	CurrentVersion  *versionSummaryResponse `json:"current_version"`
	Versions        []versionSummaryResponse `json:"versions"`
}

type versionSummaryResponse struct {
	VersionNumber  int32   `json:"version_number"`
	Status         string  `json:"status"`
	GeneratedBy    string  `json:"generated_by"`
	IsCurrent      bool    `json:"is_current"`
	MilestoneCount int64   `json:"milestone_count"`
	PublishedAt    *string `json:"published_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

type versionDetailResponse struct {
	SkillSlug     string              `json:"skill_slug"`
	SkillName     string              `json:"skill_name"`
	VersionNumber int32               `json:"version_number"`
	Status        string              `json:"status"`
	GeneratedBy   string              `json:"generated_by"`
	IsCurrent     bool                `json:"is_current"`
	PublishedAt   *string             `json:"published_at,omitempty"`
	CreatedAt     string              `json:"created_at"`
	Milestones    []milestoneResponse `json:"milestones"`
}

type milestoneResponse struct {
	Slug           string  `json:"slug"`
	Title          string  `json:"title"`
	Description    *string `json:"description,omitempty"`
	EstimatedHours *int32  `json:"estimated_hours,omitempty"`
	OrderIndex     int32   `json:"order_index"`
	Difficulty     *string `json:"difficulty,omitempty"`
	Chapter        *string `json:"chapter,omitempty"`
	Kind           string  `json:"kind"`
}
