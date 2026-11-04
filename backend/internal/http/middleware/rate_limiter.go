package middleware

import (
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// rateEntry tracks request attempts for a given key.
type rateEntry struct {
	count      int
	resetAt    time.Time
	blockUntil time.Time
}

// RateLimiter provides in-memory rate limiting.
type RateLimiter struct {
	mu     sync.Mutex
	entries map[string]*rateEntry
	maxReq int
	window time.Duration
}

// NewRateLimiter creates a new in-memory rate limiter.
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		entries: make(map[string]*rateEntry),
		maxReq:  maxRequests,
		window:  window,
	}
}

// RateLimiterMiddleware returns an Echo middleware that enforces rate limiting.
func (rl *RateLimiter) RateLimiterMiddleware(keyFunc func(c echo.Context) string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Determine the client key
			var key string
			if keyFunc != nil {
				key = keyFunc(c)
			} else {
				key = c.RealIP()
			}

			if key == "" {
				key = "unknown"
			}

			rl.mu.Lock()

			// Get or create the entry
			entry, exists := rl.entries[key]
			now := time.Now()

			if !exists || now.After(entry.resetAt) {
				// New window
				entry = &rateEntry{
					count:      0,
					resetAt:    now.Add(rl.window),
					blockUntil: time.Time{},
				}
				rl.entries[key] = entry
			}

			// Check if the client is blocked
			if !entry.blockUntil.IsZero() && now.Before(entry.blockUntil) {
				rl.mu.Unlock()
				retryAfter := int(time.Until(entry.blockUntil).Seconds()) + 1
				if retryAfter < 1 {
					retryAfter = 1
				}
				c.Response().Header().Set("Retry-After", string(rune(retryAfter)))
				return echo.NewHTTPError(429, "too many requests, please try again later")
			}

			entry.count++

			// Check if limit exceeded
			if entry.count > rl.maxReq {
				// Block for a longer period
				entry.blockUntil = now.Add(5 * time.Minute)
				rl.mu.Unlock()
				c.Response().Header().Set("Retry-After", "300")
				return echo.NewHTTPError(429, "too many requests, please try again later")
			}

			rl.mu.Unlock()

			// Store rate limit info in context for headers
			c.Set("rate_limit_remaining", rl.maxReq-entry.count)
			c.Set("rate_limit_reset", entry.resetAt.Unix())

			return next(c)
		}
	}
}

// CleanupOldEntries removes entries that have expired and are not blocked.
func (rl *RateLimiter) CleanupOldEntries() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, entry := range rl.entries {
		if !entry.blockUntil.IsZero() {
			if now.Before(entry.blockUntil) {
				continue
			}
			delete(rl.entries, key)
			continue
		}
		if now.After(entry.resetAt) {
			delete(rl.entries, key)
		}
	}
}

// RateLimitMiddleware is a helper function that creates a rate limiter middleware.
func RateLimitMiddleware(maxRequests int, window time.Duration) echo.MiddlewareFunc {
	rl := NewRateLimiter(maxRequests, window)
	return rl.RateLimiterMiddleware(nil)
}
