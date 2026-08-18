package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

const pendingGoogleRegistrationTTL = 30 * time.Minute

// GoogleUser is the subset of Google profile data we use for signup/login.
type GoogleUser struct {
	ID            string
	Email         string
	Name          string
	Picture       string
	EmailVerified bool
}

// PendingGoogleRegistration is returned when Google succeeds but no Taggy user exists yet.
// Nothing is written to the DB until CompleteGoogleRegistration is called.
type PendingGoogleRegistration struct {
	RegistrationToken string
	Email             string
	Name              string
	Picture           string
}

// GoogleOAuthResult is either a finished session or a pending profile step.
type GoogleOAuthResult struct {
	Tokens  *TokenPair
	Pending *PendingGoogleRegistration
}

// GoogleOAuth wraps the Google OAuth2 authorization-code flow.
type GoogleOAuth struct {
	config      oauth2.Config
	stateSecret []byte
}

func NewGoogleOAuth(clientID, clientSecret, redirectURL string, stateSecret []byte) *GoogleOAuth {
	return &GoogleOAuth{
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		stateSecret: stateSecret,
	}
}

func (g *GoogleOAuth) Enabled() bool {
	return g.config.ClientID != "" && g.config.ClientSecret != ""
}

func (g *GoogleOAuth) AuthURL() (string, error) {
	state, err := g.signState(time.Now())
	if err != nil {
		return "", err
	}
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

func (g *GoogleOAuth) ExchangeUser(ctx context.Context, code, state string) (GoogleUser, error) {
	if err := g.verifyState(state); err != nil {
		return GoogleUser{}, err
	}

	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return GoogleUser{}, fmt.Errorf("exchange google code: %w", err)
	}

	client := g.config.Client(ctx, token)
	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		return GoogleUser{}, fmt.Errorf("fetch google userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return GoogleUser{}, fmt.Errorf("google userinfo status %d: %s", resp.StatusCode, string(body))
	}

	var payload struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		VerifiedEmail bool   `json:"verified_email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return GoogleUser{}, fmt.Errorf("decode google userinfo: %w", err)
	}

	if payload.ID == "" || payload.Email == "" {
		return GoogleUser{}, fmt.Errorf("google userinfo missing id or email")
	}

	return GoogleUser{
		ID:            payload.ID,
		Email:         payload.Email,
		Name:          payload.Name,
		Picture:       payload.Picture,
		EmailVerified: payload.VerifiedEmail,
	}, nil
}

type oauthStatePayload struct {
	Nonce string    `json:"n"`
	Exp   time.Time `json:"exp"`
}

func (g *GoogleOAuth) signState(now time.Time) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	payload := oauthStatePayload{
		Nonce: base64.RawURLEncoding.EncodeToString(nonce),
		Exp:   now.Add(10 * time.Minute),
	}

	return g.signPayload(payload)
}

func (g *GoogleOAuth) verifyState(state string) error {
	var payload oauthStatePayload
	if err := g.verifyPayload(state, &payload); err != nil {
		return err
	}
	if time.Now().After(payload.Exp) {
		return ErrInvalidOAuthState
	}
	return nil
}

type pendingGooglePayload struct {
	GoogleID string    `json:"gid"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Picture  string    `json:"picture,omitempty"`
	Exp      time.Time `json:"exp"`
}

func (g *GoogleOAuth) IssuePendingRegistration(user GoogleUser) (string, error) {
	payload := pendingGooglePayload{
		GoogleID: user.ID,
		Email:    user.Email,
		Name:     strings.TrimSpace(user.Name),
		Picture:  strings.TrimSpace(user.Picture),
		Exp:      time.Now().Add(pendingGoogleRegistrationTTL),
	}
	return g.signPayload(payload)
}

func (g *GoogleOAuth) ParsePendingRegistration(token string) (GoogleUser, error) {
	var payload pendingGooglePayload
	if err := g.verifyPayload(token, &payload); err != nil {
		return GoogleUser{}, ErrInvalidRegistrationToken
	}
	if time.Now().After(payload.Exp) {
		return GoogleUser{}, ErrInvalidRegistrationToken
	}
	if payload.GoogleID == "" || payload.Email == "" {
		return GoogleUser{}, ErrInvalidRegistrationToken
	}
	return GoogleUser{
		ID:            payload.GoogleID,
		Email:         payload.Email,
		Name:          payload.Name,
		Picture:       payload.Picture,
		EmailVerified: true,
	}, nil
}

func (g *GoogleOAuth) signPayload(payload any) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, g.stateSecret)
	mac.Write(raw)
	sig := mac.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(raw) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (g *GoogleOAuth) verifyPayload(token string, dest any) error {
	parts := splitOnce(token, ".")
	if len(parts) != 2 {
		return ErrInvalidOAuthState
	}

	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ErrInvalidOAuthState
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ErrInvalidOAuthState
	}

	mac := hmac.New(sha256.New, g.stateSecret)
	mac.Write(raw)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return ErrInvalidOAuthState
	}

	if err := json.Unmarshal(raw, dest); err != nil {
		return ErrInvalidOAuthState
	}
	return nil
}

func splitOnce(s, sep string) []string {
	idx := strings.Index(s, sep)
	if idx < 0 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+len(sep):]}
}
