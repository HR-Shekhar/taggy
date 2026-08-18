package auth

import (
	"regexp"
	"strings"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9._]{3,30}$`)

func validateUsername(username string) error {
	username = strings.TrimSpace(username)
	if !usernamePattern.MatchString(username) {
		return ErrInvalidUsername
	}
	return nil
}
