// Service-layer input/output types for the auth module.
//
// These are NOT HTTP request/response DTOs. Handlers define their own
// JSON-tagged structs and map to/from these types.
package auth

import (
	"net/netip"
)

// RegisterInput is the service-layer input for creating a new local account.
// Validation (email format, password strength, etc.) happens in the handler
// before values are passed here.
type RegisterInput struct {
	Email    string
	Username string
	Name     string
	Password string
}

// PendingSignup is returned after register. No users row exists until OTP succeeds.
type PendingSignup struct {
	Email    string
	Username string
	Name     string
}

// LoginInput is the service-layer input for email/password authentication.
// UserAgent and IPAddress are optional metadata stored on the session row.
type LoginInput struct {
	Email     string
	Password  string
	UserAgent *string
	IPAddress *netip.Addr
}

// TokenPair is the service-layer output returned after login or refresh.
// Handlers map this to the JSON response { "access_token", "refresh_token" }.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	Username     string
	IsAdmin      bool
	Subscription string
}
