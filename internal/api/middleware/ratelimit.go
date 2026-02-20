package middleware

import (
	"net/http"
	"sync"
	"time"
)

type rateLimitEntry struct {
	mu          sync.Mutex
	count       int
	windowStart time.Time
}

func RateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	var store sync.Map

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := UserIDFromContext(r.Context())
			if key == "" {
				key = r.RemoteAddr
			}

			val, _ := store.LoadOrStore(key, &rateLimitEntry{windowStart: time.Now()})
			entry := val.(*rateLimitEntry)

			entry.mu.Lock()
			if time.Since(entry.windowStart) >= time.Minute {
				entry.count = 0
				entry.windowStart = time.Now()
			}
			entry.count++
			count := entry.count
			entry.mu.Unlock()

			if count > requestsPerMinute {
				writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded, try again in a minute"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
