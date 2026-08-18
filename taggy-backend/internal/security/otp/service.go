package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const digitCount = 6

// Code is a one-time email verification code and its stored hash.
type Code struct {
	PlainText string
	Hash      string
}

// Service generates short numeric OTPs and hashes them for database storage.
type Service struct {
	hasher Hasher
}

// Hasher digests OTP plaintext for storage (inject token.Service.Hash in production wiring).
type Hasher func(plainText string) string

func New(hasher Hasher) *Service {
	return &Service{hasher: hasher}
}

// Generate creates a cryptographically random 6-digit OTP.
func (s *Service) Generate() (Code, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return Code{}, fmt.Errorf("generate otp: %w", err)
	}

	plain := fmt.Sprintf("%0*d", digitCount, n.Int64())

	return Code{
		PlainText: plain,
		Hash:      s.hasher(plain),
	}, nil
}

// Hash digests a client-submitted OTP for lookup.
func (s *Service) Hash(plainText string) string {
	return s.hasher(plainText)
}
