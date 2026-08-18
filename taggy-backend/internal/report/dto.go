package report

type createReportRequest struct {
	TargetType string `json:"target_type" validate:"required"`
	TargetID   int64  `json:"target_id" validate:"required,gt=0"`
	Reason     string `json:"reason" validate:"required,min=3,max=2000"`
}

type reportResponse struct {
	ID         int64   `json:"id"`
	TargetType string  `json:"target_type"`
	TargetID   int64   `json:"target_id"`
	Reason     string  `json:"reason"`
	Status     string  `json:"status"`
	ResolvedAt *string `json:"resolved_at,omitempty"`
	CreatedAt  string  `json:"created_at"`
}
