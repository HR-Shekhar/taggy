package search

type searchResponse struct {
	Query       string             `json:"query"`
	Skills      []skillHitResponse `json:"skills,omitempty"`
	Users       []userHitResponse  `json:"users,omitempty"`
	Communities []communityHitResponse `json:"communities,omitempty"`
}

type skillHitResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
}

type userHitResponse struct {
	PublicID          string  `json:"public_id"`
	Username          string  `json:"username"`
	Name              string  `json:"name"`
	ProfilePictureURL *string `json:"profile_picture_url"`
	Bio               *string `json:"bio"`
}

type communityHitResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	SkillSlug   string  `json:"skill_slug"`
	SkillName   string  `json:"skill_name"`
}
