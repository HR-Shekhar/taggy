package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Service signs and verifies short-lived access tokens.
//
// Access tokens are self-contained JWTs: middleware can verify them without
// hitting the database. They expire quickly (default 15m) so a stolen token
// has a short window of abuse.
type Service struct {
	config Config
}

func New(config Config) *Service {
	return &Service{
		config: config,
	}
}

// Generate builds a signed JWT for an authenticated session.
//
// Claims we embed:
//   - sub: user's public UUID (who is logged in)
//   - sid: session's public UUID (which device/session issued this token)
//   - iss, iat, exp: standard JWT metadata (issuer, issued-at, expiry)
func (s *Service) Generate(userID, sessionID uuid.UUID) (string, error) {
	claims := Claims{
		SessionID: sessionID,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    s.config.Issuer,
			IssuedAt:  jwtv5.NewNumericDate(time.Now()),
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(s.config.TTL)),
		},
	}

	token := jwtv5.NewWithClaims(s.config.SigningMethod, claims)

	return token.SignedString(s.config.SecretKey)
}

// Verify parses and validates a JWT string from the Authorization header.
//
// Checks performed:
//   - signature matches our secret (token wasn't tampered with)
//   - issuer matches config
//   - token hasn't expired
//
// No database call here — that's intentional for speed on every request.
func (s *Service) Verify(tokenString string) (*Claims, error) {
	token, err := jwtv5.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwtv5.Token) (any, error) {
			// Reject tokens signed with an unexpected algorithm (security hardening).
			if token.Method.Alg() != s.config.SigningMethod.Alg() {
				return nil, ErrInvalidSigningMethod
			}

			return s.config.SecretKey, nil
		},
		jwtv5.WithIssuer(s.config.Issuer),
	)

	if err != nil {
		switch {
		case errors.Is(err, jwtv5.ErrTokenExpired):
			return nil, ErrExpiredToken

		case errors.Is(err, jwtv5.ErrTokenSignatureInvalid):
			return nil, ErrInvalidSignature

		case errors.Is(err, jwtv5.ErrTokenInvalidIssuer):
			return nil, ErrInvalidIssuer

		default:
			// Any other parse failure is a malformed/invalid token, not an expiry.
			return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
		}
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
