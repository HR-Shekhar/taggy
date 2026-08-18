package jwt

import "errors"

var (
	ErrExpiredToken         = errors.New("token expired")
	ErrInvalidToken         = errors.New("invalid token")
	ErrInvalidSignature     = errors.New("invalid token signature")
	ErrInvalidIssuer        = errors.New("invalid token issuer")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
)
