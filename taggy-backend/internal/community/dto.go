package community

type sendMessageRequest struct {
	Content          string `json:"content" validate:"required,min=1,max=4000"`
	ReplyToMessageID *int64 `json:"reply_to_message_id" validate:"omitempty,gt=0"`
}

type editMessageRequest struct {
	Content string `json:"content" validate:"required,min=1,max=4000"`
}

type communityResponse struct {
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type channelResponse struct {
	ID          int64   `json:"id"`
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type messageResponse struct {
	ID                    int64   `json:"id"`
	AuthorUsername        string  `json:"author_username"`
	AuthorName            string  `json:"author_name"`
	Content               string  `json:"content"`
	EditedAt              *string `json:"edited_at,omitempty"`
	CreatedAt             string  `json:"created_at"`
	ReplyToMessageID      *int64  `json:"reply_to_message_id,omitempty"`
	ReplyToContent        *string `json:"reply_to_content,omitempty"`
	ReplyToAuthorUsername *string `json:"reply_to_author_username,omitempty"`
}
