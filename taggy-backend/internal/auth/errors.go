package auth

import "errors"

// Service-layer auth errors. Handlers and the global error handler map these
// to HTTP status codes (e.g. ErrInvalidCredentials → 401).
var (
	// Returned when email or password is wrong. Same message for both cases
	// so attackers cannot tell whether an email is registered.
	ErrInvalidCredentials = errors.New("invalid email or password")

	// Returned when the refresh token does not match any session.
	ErrInvalidRefreshToken = errors.New("invalid refresh token")

	// Returned when the session row exists but expires_at is in the past.
	ErrSessionExpired = errors.New("session expired")

	// Returned when a local account has not completed email verification.
	ErrEmailNotVerified = errors.New("email not verified")

	// Returned when the submitted OTP is wrong.
	ErrInvalidOTP = errors.New("invalid verification code")

	// Returned when the OTP row is missing or past expires_at.
	ErrOTPExpired = errors.New("verification code expired")

	// Returned when email is already verified.
	ErrEmailAlreadyVerified = errors.New("email already verified")

	// Returned when register email is already taken.
	ErrEmailInUse = errors.New("email already in use")

	// Returned when register username is already taken.
	ErrUsernameInUse = errors.New("username already in use")

	// Returned when Google OAuth is not configured on this server.
	ErrOAuthNotConfigured = errors.New("oauth not configured")

	// Returned when OAuth state parameter fails validation (CSRF protection).
	ErrInvalidOAuthState = errors.New("invalid oauth state")

	// Returned when Google returns an unusable profile.
	ErrOAuthAccountInvalid = errors.New("oauth account invalid")

	// Returned when username format is invalid (3-30 chars: letters, digits, . _).
	ErrInvalidUsername = errors.New("username format is invalid")

	// Returned when Google signup completion is missing a username.
	ErrOAuthUsernameRequired = errors.New("username is required")

	// Returned when the pending Google registration token is missing, forged, or expired.
	ErrInvalidRegistrationToken = errors.New("invalid or expired google registration token")

	// Returned when Resend (or the configured mailer) fails to deliver the OTP email.
	ErrEmailDeliveryFailed = errors.New("failed to send verification email")
)
