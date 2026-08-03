package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims is the payload embedded inside every access token.
//
// SessionID (sid) ties the JWT to a row in user_sessions so we can
// support logout-all and refresh rotation later.
//
// RegisteredClaims carries standard fields — notably Subject (sub) = user public_id.
type Claims struct {
	SessionID uuid.UUID `json:"sid"`
	jwt.RegisteredClaims
}
