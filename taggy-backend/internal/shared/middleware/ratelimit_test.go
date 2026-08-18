package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/labstack/echo/v4"
)

func TestIPRateLimit(t *testing.T) {
	e := echo.New()
	limit := IPRateLimit(2, time.Minute)

	handler := limit(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	call := func() error {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		return handler(c)
	}

	if err := call(); err != nil {
		t.Fatalf("1st call: %v", err)
	}
	if err := call(); err != nil {
		t.Fatalf("2nd call: %v", err)
	}
	if err := call(); err != apperrors.ErrTooManyRequests {
		t.Fatalf("3rd call want ErrTooManyRequests, got %v", err)
	}
}
