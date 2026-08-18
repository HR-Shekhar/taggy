package livekit

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrNotConfigured = errors.New("livekit not configured")

type Config struct {
	URL       string
	APIKey    string
	APISecret string
}

func (c Config) Configured() bool {
	return strings.TrimSpace(c.URL) != "" &&
		strings.TrimSpace(c.APIKey) != "" &&
		strings.TrimSpace(c.APISecret) != ""
}

type TokenClient struct {
	cfg Config
}

func NewTokenClient(cfg Config) *TokenClient {
	return &TokenClient{cfg: cfg}
}

func (c *TokenClient) Configured() bool {
	return c.cfg.Configured()
}

func (c *TokenClient) URL() string {
	return c.cfg.URL
}

type MintParams struct {
	Identity   string
	Name       string
	RoomName   string
	CanPublish bool
	ValidFor   time.Duration
}

// MintJoinToken creates a LiveKit access token (JWT) with video room grants.
func (c *TokenClient) MintJoinToken(p MintParams) (string, error) {
	if !c.Configured() {
		return "", ErrNotConfigured
	}
	if strings.TrimSpace(p.Identity) == "" || strings.TrimSpace(p.RoomName) == "" {
		return "", errors.New("identity and room name are required")
	}

	validFor := p.ValidFor
	if validFor <= 0 {
		validFor = time.Hour
	}

	now := time.Now().UTC()
	canPublish := p.CanPublish
	canSubscribe := true
	canPublishData := true
	roomJoin := true

	claims := jwt.MapClaims{
		"iss":  c.cfg.APIKey,
		"sub":  p.Identity,
		"name": p.Name,
		"nbf":  now.Unix(),
		"exp":  now.Add(validFor).Unix(),
		"video": map[string]any{
			"roomJoin":       roomJoin,
			"room":           p.RoomName,
			"canPublish":     canPublish,
			"canSubscribe":   canSubscribe,
			"canPublishData": canPublishData,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.cfg.APISecret))
}
