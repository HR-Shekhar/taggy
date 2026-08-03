package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// 32 bytes = 256 bits of randomness. Same strength as a good AES key.
// Encoded as hex this becomes a 64-character string sent to the client once.
const refreshTokenByteLength = 32

// RefreshToken is a value object that pairs the client-visible secret with
// its database-safe hash. Callers must store Hash in PostgreSQL and return
// PlainText to the client exactly once (at login/refresh). Keeping both
// fields together reduces the risk of accidentally persisting PlainText.
type RefreshToken struct {
	PlainText string // sent to client, never stored in DB
	Hash      string // SHA-256 hex digest, stored in user_sessions.refresh_token_hash
}

// Service generates and hashes opaque refresh tokens.
//
// Refresh tokens are NOT JWTs — they are random secrets with no payload.
// The server looks them up in user_sessions to issue new access tokens.
type Service struct{}

func New() *Service {
	return &Service{}
}

// Generate creates a new cryptographically random refresh token.
//
// Flow at login:
//  1. Generate() produces PlainText + Hash
//  2. Hash goes into user_sessions
//  3. PlainText goes to the client (JSON response)
func (s *Service) Generate() (RefreshToken, error) {
	bytes := make([]byte, refreshTokenByteLength)
	if _, err := rand.Read(bytes); err != nil {
		return RefreshToken{}, fmt.Errorf("generate refresh token: %w", err)
	}

	plainText := hex.EncodeToString(bytes)

	return RefreshToken{
		PlainText: plainText,
		Hash:      s.Hash(plainText),
	}, nil
}

// Hash produces a deterministic SHA-256 hex digest of a refresh token.
//
// Used when the client sends PlainText back (refresh/logout):
// we hash it the same way and look up the row by refresh_token_hash.
//
// Why hash refresh tokens but use bcrypt for passwords?
//   - Passwords are low-entropy (users pick "password123") → need slow bcrypt
//   - Refresh tokens are high-entropy random strings → SHA-256 is enough
func (s *Service) Hash(plainText string) string {
	sum := sha256.Sum256([]byte(plainText))
	return hex.EncodeToString(sum[:])
}
