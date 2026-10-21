package handlers

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	count       int
	windowStart time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]visitor
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]visitor),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		now := time.Now()

		rl.mu.Lock()
		v := rl.visitors[ip]

		if v.windowStart.IsZero() || now.Sub(v.windowStart) >= rl.window {
			v = visitor{
				count:       0,
				windowStart: now,
			}
		}

		if v.count >= rl.limit {
			rl.visitors[ip] = v
			rl.mu.Unlock()

			w.Header().Set("Retry-After", "60")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		v.count++
		rl.visitors[ip] = v
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
