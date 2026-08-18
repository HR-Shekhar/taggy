package user

// Service-layer input for profile updates.
// nil pointer = "do not change this field".
// non-nil pointer = "set to this value" (including empty string to clear bio/url).
type UpdateProfileInput struct {
	Name              *string
	Bio               *string
	ProfilePictureURL *string
	Username          *string
}
