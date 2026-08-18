package middleware

import (
	"sync"
	"time"

	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimit limits requests per client IP using a token bucket.
// `max` is both the burst size and the approximate number of allowed
// requests per `window`.
func IPRateLimit(max int, window time.Duration) echo.MiddlewareFunc {
	if max < 1 {
		max = 1
	}
	if window < time.Second {
		window = time.Second
	}

	var (
		mu       sync.Mutex
		visitors = make(map[string]*visitor)
		limit    = rate.Every(window / time.Duration(max))
		burst    = max
	)

	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			cutoff := time.Now().Add(-2 * window)
			for ip, v := range visitors {
				if v.lastSeen.Before(cutoff) {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			if ip == "" {
				ip = c.Request().RemoteAddr
			}

			mu.Lock()
			v, ok := visitors[ip]
			if !ok {
				v = &visitor{limiter: rate.NewLimiter(limit, burst)}
				visitors[ip] = v
			}
			v.lastSeen = time.Now()
			allowed := v.limiter.Allow()
			mu.Unlock()

			if !allowed {
				return apperrors.ErrTooManyRequests
			}
			return next(c)
		}
	}
}
