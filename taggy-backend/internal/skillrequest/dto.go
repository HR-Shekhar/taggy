package skillrequest

type createRequestBody struct {
	Name        string `json:"name" validate:"required,min=3,max=255"`
	Description string `json:"description" validate:"omitempty,max=4000"`
	Force       bool   `json:"force"`
}

type rejectRequestBody struct {
	AdminNote string `json:"admin_note" validate:"omitempty,max=2000"`
}

type similarResponse struct {
	Query   string         `json:"query"`
	Similar []SimilarSkill `json:"similar"`
}

type createResponse struct {
	RequiresConfirm bool           `json:"requires_confirm,omitempty"`
	Similar         []SimilarSkill `json:"similar,omitempty"`
	Request         *RequestView   `json:"request,omitempty"`
	Message         string         `json:"message,omitempty"`
}
