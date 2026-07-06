package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter per API key.
type RateLimiter struct {
	mu      sync.Mutex
	limits  map[string]*clientLimit
	maxReqs int
	window  time.Duration
}

type clientLimit struct {
	count   int
	resetAt time.Time
}

// NewRateLimiter creates a new rate limiter.
// maxReqs: maximum number of requests allowed within the window.
// window: time window for rate limiting (e.g., 1 minute).
func NewRateLimiter(maxReqs int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limits:  make(map[string]*clientLimit),
		maxReqs: maxReqs,
		window:  window,
	}
}

// Middleware returns an HTTP middleware that enforces rate limiting.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Health check endpoints are not rate-limited
		if r.URL.Path == "/health" || r.URL.Path == "/ready" {
			next.ServeHTTP(w, r)
			return
		}

		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			http.Error(w, `{"error":{"code":"missing_api_key","message":"header X-API-Key é obrigatório"}}`, http.StatusUnauthorized)
			return
		}

		rl.mu.Lock()
		now := time.Now()
		lim, exists := rl.limits[apiKey]
		if !exists || now.After(lim.resetAt) {
			lim = &clientLimit{count: 0, resetAt: now.Add(rl.window)}
			rl.limits[apiKey] = lim
		}

		if lim.count >= rl.maxReqs {
			rl.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"code":"rate_limit_exceeded","message":"limite de requisições excedido"}}`))
			return
		}

		lim.count++
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}