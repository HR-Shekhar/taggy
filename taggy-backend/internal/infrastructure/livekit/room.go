package livekit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// httpBaseURL converts a LiveKit WebSocket URL (wss:// / ws://) to HTTPS/HTTP for Twirp.
func (c *TokenClient) httpBaseURL() string {
	u := strings.TrimSpace(c.cfg.URL)
	u = strings.TrimRight(u, "/")
	switch {
	case strings.HasPrefix(u, "wss://"):
		return "https://" + strings.TrimPrefix(u, "wss://")
	case strings.HasPrefix(u, "ws://"):
		return "http://" + strings.TrimPrefix(u, "ws://")
	default:
		return u
	}
}

func (c *TokenClient) mintServerToken(validFor time.Duration) (string, error) {
	if !c.Configured() {
		return "", ErrNotConfigured
	}
	if validFor <= 0 {
		validFor = 5 * time.Minute
	}
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iss": c.cfg.APIKey,
		"sub": c.cfg.APIKey,
		"nbf": now.Unix(),
		"exp": now.Add(validFor).Unix(),
		"video": map[string]any{
			"roomCreate": true,
			"roomList":   true,
			"roomAdmin":  true,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.cfg.APISecret))
}

// DeleteRoom forcibly disconnects all participants and removes the room on LiveKit.
// A missing room is treated as success (already gone).
func (c *TokenClient) DeleteRoom(ctx context.Context, roomName string) error {
	if !c.Configured() {
		return ErrNotConfigured
	}
	roomName = strings.TrimSpace(roomName)
	if roomName == "" {
		return fmt.Errorf("room name is required")
	}

	token, err := c.mintServerToken(5 * time.Minute)
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]string{"room": roomName})
	if err != nil {
		return err
	}

	endpoint := c.httpBaseURL() + "/twirp/livekit.RoomService/DeleteRoom"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	msg := strings.TrimSpace(string(respBody))
	// Twirp not_found / room does not exist → treat as already deleted.
	lower := strings.ToLower(msg)
	if resp.StatusCode == http.StatusNotFound ||
		strings.Contains(lower, "not_found") ||
		strings.Contains(lower, "does not exist") ||
		strings.Contains(lower, "room not found") {
		return nil
	}

	if msg == "" {
		msg = resp.Status
	}
	return fmt.Errorf("livekit DeleteRoom %s: %s", roomName, msg)
}
