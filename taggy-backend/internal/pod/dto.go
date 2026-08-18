package pod

type createPodRequest struct {
	Slug        string  `json:"slug" validate:"required,min=3,max=60"`
	Name        string  `json:"name" validate:"required,min=3,max=255"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
}

type setMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=OWNER ADMIN MEMBER owner admin member"`
}

type podResponse struct {
	ID            int64   `json:"id"`
	Slug          string  `json:"slug"`
	Name          string  `json:"name"`
	Description   *string `json:"description,omitempty"`
	SkillSlug     string  `json:"skill_slug"`
	SkillName     string  `json:"skill_name"`
	OwnerUsername string  `json:"owner_username"`
	MaxMembers    int32   `json:"max_members"`
	AcceptedCount int64   `json:"accepted_count"`
}

type membershipResponse struct {
	PodSlug   string  `json:"pod_slug"`
	PodName   string  `json:"pod_name"`
	SkillSlug string  `json:"skill_slug"`
	SkillName string  `json:"skill_name"`
	Status    string  `json:"status"`
	Role      string  `json:"role"`
	JoinedAt  *string `json:"joined_at,omitempty"`
}

type memberResponse struct {
	Username string  `json:"username"`
	Name     string  `json:"name"`
	Role     string  `json:"role"`
	Status   string  `json:"status,omitempty"`
	JoinedAt *string `json:"joined_at,omitempty"`
}

type podDetailResponse struct {
	Pod          podResponse      `json:"pod"`
	Members      []memberResponse `json:"members"`
	JoinRequests []memberResponse `json:"join_requests"`
}
