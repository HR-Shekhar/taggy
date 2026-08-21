package auth

// HTTP request/response DTOs for auth endpoints.
// These are the JSON shapes clients send and receive.
// Handlers validate these, then map to service-layer types in types.go.

type registerRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=30"`
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Password string `json:"password" validate:"required,min=8,strong_password"`
}

type userResponse struct {
	PublicID      string `json:"public_id"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	Name          string `json:"name"`
	EmailVerified bool   `json:"email_verified"`
	Subscription  string `json:"subscription"`
	IsAdmin       bool   `json:"is_admin"`
}

// registerResponse includes dev_otp in development so local testing needs no SMTP.
type registerResponse struct {
	PublicID      string `json:"public_id,omitempty"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	Name          string `json:"name"`
	EmailVerified bool   `json:"email_verified"`
	Subscription  string `json:"subscription"`
	DevOTP        string `json:"dev_otp,omitempty"`
}

type resendVerificationResponse struct {
	DevOTP string `json:"dev_otp,omitempty"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type verifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	Otp   string `json:"otp" validate:"required,len=6,numeric"`
}

type resendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// tokenResponse is returned after login or refresh.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username"`
	IsAdmin      bool   `json:"is_admin"`
	Subscription string `json:"subscription"`
}

type adminMeResponse struct {
	PublicID     string `json:"public_id"`
	Username     string `json:"username"`
	IsAdmin      bool   `json:"is_admin"`
	Subscription string `json:"subscription"`
}

type googleAuthURLResponse struct {
	URL string `json:"url"`
}

type completeGoogleRegistrationRequest struct {
	RegistrationToken string `json:"registration_token" validate:"required"`
	Username          string `json:"username" validate:"required,min=3,max=30"`
	Name              string `json:"name" validate:"omitempty,min=1,max=100"`
}

type pendingGoogleRegistrationResponse struct {
	RegistrationToken string `json:"registration_token"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	Picture           string `json:"picture,omitempty"`
}
