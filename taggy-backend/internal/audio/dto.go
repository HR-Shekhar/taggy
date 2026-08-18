package audio

type createRoomRequest struct {
	Title           string  `json:"title" validate:"required,min=3,max=255"`
	Description     *string `json:"description" validate:"omitempty,max=2000"`
	MaxParticipants *int32  `json:"max_participants" validate:"omitempty,min=1,max=100"`
}

type roomResponse struct {
	ID              string  `json:"id"`
	EntityID        int64   `json:"entity_id"`
	RoomType        string  `json:"room_type"`
	Title           string  `json:"title"`
	Description     *string `json:"description,omitempty"`
	Status          string  `json:"status"`
	HostUsername    string  `json:"host_username"`
	LiveKitRoomName string  `json:"livekit_room_name"`
	PodSlug         *string `json:"pod_slug,omitempty"`
	SkillSlug       *string `json:"skill_slug,omitempty"`
	ChannelSlug     *string `json:"channel_slug,omitempty"`
	MaxParticipants *int32  `json:"max_participants,omitempty"`
	ActualStartTime *string `json:"actual_start_time,omitempty"`
	EndedAt         *string `json:"ended_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

type participantResponse struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}

type roomDetailResponse struct {
	Room         roomResponse          `json:"room"`
	Participants []participantResponse `json:"participants"`
}

type joinResponse struct {
	RoomID          string `json:"room_id"`
	LiveKitURL      string `json:"livekit_url"`
	LiveKitRoomName string `json:"livekit_room_name"`
	Token           string `json:"token"`
	Role            string `json:"role"`
}
