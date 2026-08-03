package password

import (
	"golang.org/x/crypto/bcrypt"
)

// Service handles password hashing and verification.
//
// We NEVER store plain-text passwords in the database.
// On register: Hash() produces a one-way bcrypt hash stored in user_identity.
// On login: Verify() compares the submitted password against that stored hash.
//
// bcrypt automatically embeds a random "salt" inside the hash string,
// so two users with the same password get different hashes.
type Service struct {
	config Config
}

func New(config Config) *Service {
	return &Service{
		config: config,
	}
}

// Hash converts a plain-text password into a bcrypt hash safe to store in PostgreSQL.
func (s *Service) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		s.config.Cost,
	)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// Verify checks whether password matches the stored hash.
// Returns nil on success, ErrInvalidPassword on mismatch.
//
// We return a domain error (not bcrypt's raw error) so callers can treat
// "wrong password" as an expected auth failure without leaking internals.
func (s *Service) Verify(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return ErrInvalidPassword
	}

	return nil
}
