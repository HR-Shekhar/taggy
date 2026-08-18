package email

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/resend/resend-go/v3"
	"github.com/rs/zerolog"
)

const defaultFrom = "Taggy <onboarding@resend.dev>"

// Sender delivers transactional email.
type Sender interface {
	SendVerificationOTP(ctx context.Context, to string, otp string) error
}

// DevLogger prints OTP emails to the application log (local development only).
type DevLogger struct {
	log zerolog.Logger
}

func NewDevLogger(log zerolog.Logger) *DevLogger {
	return &DevLogger{log: log}
}

func (d *DevLogger) SendVerificationOTP(ctx context.Context, to string, otp string) error {
	d.log.Info().
		Str("to", to).
		Str("otp", otp).
		Msg("DEV EMAIL: verification OTP (set RESEND_API_KEY to send real mail)")
	return nil
}

// Resend sends email via the official Resend Go SDK.
type Resend struct {
	client *resend.Client
	from   string
	log    zerolog.Logger
}

type ResendConfig struct {
	APIKey string
	From   string
}

func NewResend(cfg ResendConfig, log zerolog.Logger) (*Resend, error) {
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("resend api key is required")
	}

	from := normalizeFrom(cfg.From)
	if looksLikeConsumerMailbox(from) {
		log.Warn().
			Str("from", from).
			Msg("RESEND_EMAIL_FROM looks like a personal inbox; Resend requires a verified domain (or Taggy <onboarding@resend.dev> for testing)")
	}

	log.Info().Str("from", from).Msg("email provider: Resend")
	return &Resend{
		client: resend.NewClient(apiKey),
		from:   from,
		log:    log,
	}, nil
}

func normalizeFrom(raw string) string {
	from := strings.TrimSpace(raw)
	if from == "" {
		return defaultFrom
	}
	// Already "Name <email@domain>"
	if strings.Contains(from, "<") && strings.Contains(from, ">") {
		return from
	}
	// Bare address → add product display name
	if strings.Contains(from, "@") {
		return "Taggy <" + from + ">"
	}
	return from
}

func looksLikeConsumerMailbox(from string) bool {
	lower := strings.ToLower(from)
	for _, domain := range []string{
		"@gmail.com", "@googlemail.com", "@yahoo.com", "@outlook.com",
		"@hotmail.com", "@icloud.com", "@live.com", "@me.com",
	} {
		if strings.Contains(lower, domain) {
			return true
		}
	}
	return false
}

func (r *Resend) SendVerificationOTP(ctx context.Context, to string, otp string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("recipient email is required")
	}
	otp = strings.TrimSpace(otp)
	if otp == "" {
		return fmt.Errorf("otp is required")
	}

	subject := "Your Taggy verification code"
	text := fmt.Sprintf(
		"Your Taggy email verification code is %s.\n\nThis code expires in 10 minutes. If you did not sign up for Taggy, you can ignore this email.",
		otp,
	)
	safeOTP := html.EscapeString(otp)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="margin:0;padding:0;background:#f4f6f4;font-family:Georgia,serif;">
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background:#f4f6f4;padding:32px 16px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" style="max-width:480px;background:#ffffff;border-radius:12px;padding:32px 28px;">
          <tr><td style="font-size:22px;font-weight:700;color:#2f4a3a;">Taggy</td></tr>
          <tr><td style="padding-top:16px;font-size:16px;line-height:1.5;color:#334155;">Your email verification code is:</td></tr>
          <tr><td style="padding-top:20px;font-size:32px;font-weight:700;letter-spacing:8px;color:#2f4a3a;font-family:ui-monospace,monospace;">%s</td></tr>
          <tr><td style="padding-top:20px;font-size:14px;line-height:1.5;color:#64748b;">This code expires in 10 minutes. If you did not sign up for Taggy, ignore this email.</td></tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, safeOTP)

	params := &resend.SendEmailRequest{
		From:    r.from,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
		Text:    text,
	}

	sent, err := r.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		r.log.Error().
			Str("to", to).
			Str("from", r.from).
			Err(err).
			Msg("resend API rejected email")
		return fmt.Errorf("resend send failed: %w", err)
	}

	r.log.Info().
		Str("to", to).
		Str("email_id", sent.Id).
		Msg("verification OTP email sent via Resend")
	return nil
}
