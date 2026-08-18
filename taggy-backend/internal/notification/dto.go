package notification

type notificationResponse struct {
	ID         int64   `json:"id"`
	Type       string  `json:"type"`
	EntityType *string `json:"entity_type,omitempty"`
	EntityID   *int64  `json:"entity_id,omitempty"`
	Title      string  `json:"title"`
	Body       string  `json:"body"`
	IsRead     bool    `json:"is_read"`
	ReadAt     *string `json:"read_at,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

type listNotificationsResponse struct {
	UnreadCount   int64                  `json:"unread_count"`
	Notifications []notificationResponse `json:"notifications"`
}
