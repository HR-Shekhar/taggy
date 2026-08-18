package user

// HTTP request/response DTOs for user profile endpoints.

// profileResponse is GitHub-style: public fields always; private fields only when viewing yourself.
type profileResponse struct {
	Username          string  `json:"username"`
	Name              string  `json:"name"`
	Bio               *string `json:"bio"`
	ProfilePictureURL *string `json:"profile_picture_url"`
	PublicID          *string `json:"public_id,omitempty"`
	Email             *string `json:"email,omitempty"`
	EmailVerified     *bool   `json:"email_verified,omitempty"`
	Subscription      *string `json:"subscription,omitempty"`
	IsAdmin           *bool   `json:"is_admin,omitempty"`
}

// updateProfileRequest is the JSON body for PATCH /users/{username}.
type updateProfileRequest struct {
	Name              *string `json:"name" validate:"omitempty,min=1,max=100"`
	Bio               *string `json:"bio" validate:"omitempty,max=500"`
	ProfilePictureURL *string `json:"profile_picture_url" validate:"omitempty,url"`
	Username          *string `json:"username" validate:"omitempty,min=3,max=30"`
}
