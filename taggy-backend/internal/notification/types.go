package notification

type CreateInput struct {
	UserID     int64
	Type       string
	EntityType *string
	EntityID   *int64
	Title      string
	Body       string
}
