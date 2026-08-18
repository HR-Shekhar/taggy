package roadmaprequest

import "github.com/HR-Shekhar/taggy-backend/internal/infrastructure/openrouter"

type CreateInput struct {
	Rationale string
}

type MilestoneDraft = openrouter.MilestoneDraft

type RequestView struct {
	ID                string           `json:"id"`
	SkillSlug         string           `json:"skill_slug"`
	SkillName         string           `json:"skill_name"`
	Rationale         *string          `json:"rationale,omitempty"`
	Status            string           `json:"status"`
	BaseVersionNumber int32            `json:"base_version_number"`
	DraftMilestones   []MilestoneDraft `json:"draft_milestones"`
	AdminNote         *string          `json:"admin_note,omitempty"`
	CreatedVersionID  *int64           `json:"created_version_id,omitempty"`
	CreatedAt         string           `json:"created_at"`
	UpdatedAt         string           `json:"updated_at"`
}

type createRequestBody struct {
	Rationale string `json:"rationale" validate:"omitempty,max=4000"`
}

type rejectRequestBody struct {
	AdminNote string `json:"admin_note" validate:"omitempty,max=2000"`
}
