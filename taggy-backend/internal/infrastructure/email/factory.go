package email

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

// NewSender picks Resend when configured; otherwise DevLogger in development only.
// Production without Resend fails fast.
func NewSender(environment string, resend ResendConfig, log zerolog.Logger) (Sender, bool, error) {
	env := strings.ToLower(strings.TrimSpace(environment))
	hasResend := strings.TrimSpace(resend.APIKey) != ""

	if hasResend {
		client, err := NewResend(resend, log)
		if err != nil {
			return nil, false, err
		}
		// Prefer real email when key is set; do not expose OTP in API responses.
		return client, false, nil
	}

	if env == "development" || env == "dev" || env == "local" {
		return NewDevLogger(log), true, nil
	}

	return nil, false, fmt.Errorf("email not configured: set RESEND_API_KEY and RESEND_EMAIL_FROM for non-development environments")
}
