package search

type Input struct {
	Query string
	Types []string
	Limit int32
}

type Result struct {
	Skills      []SkillHit
	Users       []UserHit
	Communities []CommunityHit
}

type SkillHit struct {
	ID          int64
	Name        string
	Slug        string
	Description *string
}

type UserHit struct {
	PublicID          string
	Username          string
	Name              string
	ProfilePictureURL *string
	Bio               *string
}

type CommunityHit struct {
	ID          int64
	Name        string
	Description *string
	SkillSlug   string
	SkillName   string
}
