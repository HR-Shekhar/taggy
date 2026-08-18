package community

type SendMessageInput struct {
	Content          string
	ReplyToMessageID *int64
}

type ListMessagesInput struct {
	BeforeID int64
	Limit    int32
}
