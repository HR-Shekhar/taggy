package skillrequest

import "github.com/HR-Shekhar/taggy-backend/internal/infrastructure/openrouter"

type CreateInput struct {
	Name        string
	Description string
	Force       bool
}

type RejectInput struct {
	AdminNote string
}

type SimilarSkill struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	Score       float32 `json:"score"`
}

type MilestoneDraft = openrouter.MilestoneDraft

type CreateResult struct {
	Request         RequestView
	Similar         []SimilarSkill
	RequiresConfirm bool
}

type RequestView struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	SlugCandidate   string           `json:"slug_candidate"`
	Description     *string          `json:"description,omitempty"`
	Status          string           `json:"status"`
	SimilarSkills   []SimilarSkill   `json:"similar_skills"`
	DraftMilestones []MilestoneDraft `json:"draft_milestones"`
	AdminNote       *string          `json:"admin_note,omitempty"`
	CreatedSkillID  *int64           `json:"created_skill_id,omitempty"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       string           `json:"updated_at"`
}
