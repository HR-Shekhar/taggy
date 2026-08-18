package jwt

import (
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"time"
)

type Config struct {
	SecretKey     []byte
	Issuer        string
	TTL           time.Duration
	SigningMethod jwtv5.SigningMethod
}
