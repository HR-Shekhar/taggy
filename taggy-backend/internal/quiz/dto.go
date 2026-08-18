package quiz

type answerRequest struct {
	SelectedIndices []int `json:"selected_indices" validate:"required,dive,min=0,max=4"`
}

type questionResponse struct {
	OrderIndex int      `json:"order_index"`
	Difficulty int      `json:"difficulty"`
	Prompt     string   `json:"prompt"`
	Options    []string `json:"options"`
	TopicTitle string   `json:"topic_title"`
	Answered   bool     `json:"answered"`
	IsCorrect  *bool    `json:"is_correct,omitempty"`
	TimedOut   *bool    `json:"timed_out,omitempty"`
	// CorrectIndices only after the question is locked.
	CorrectIndices []int `json:"correct_indices,omitempty"`
	SecondsLeft    *int  `json:"seconds_left,omitempty"`
}

type quizResponse struct {
	ID          string             `json:"id"`
	Status      string             `json:"status"`
	TopicCount  int                `json:"topic_count"`
	CorrectCount int               `json:"correct_count"`
	Score       int                `json:"score"`
	Topics      []string           `json:"topics"`
	Questions   []questionResponse `json:"questions"`
	CreatedAt   string             `json:"created_at"`
	CompletedAt *string            `json:"completed_at,omitempty"`
}

type answerResponse struct {
	IsCorrect      bool  `json:"is_correct"`
	TimedOut       bool  `json:"timed_out"`
	CorrectIndices []int `json:"correct_indices"`
}

type startResponse struct {
	OrderIndex  int `json:"order_index"`
	SecondsLeft int `json:"seconds_left"`
}

type podLeaderboardEntry struct {
	Username   string `json:"username"`
	PublicID   string `json:"public_id"`
	Name       string `json:"name"`
	BestScore  int    `json:"best_score"`
	TopicCount int    `json:"topic_count"`
	Rank       int    `json:"rank"`
}

type communityLeaderboardEntry struct {
	PodSlug     string `json:"pod_slug"`
	PodName     string `json:"pod_name"`
	TotalScore  int    `json:"total_score"`
	MemberCount int    `json:"member_count"`
	Rank        int    `json:"rank"`
}
