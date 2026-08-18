package audio

type CreateRoomInput struct {
	Title           string
	Description     *string
	MaxParticipants *int32
}

type JoinRoomResult struct {
	RoomID          string
	LiveKitURL      string
	LiveKitRoomName string
	Token           string
	Role            string
}
