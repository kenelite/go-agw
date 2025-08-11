package router

import (
    "net/http"

    "github.com/kenelite/go-agw/internal/ratelimiter"
)

type rateLimitMiddleware struct { limiter *ratelimiter.Limiter }

func newRateLimitMiddleware() *rateLimitMiddleware { return &rateLimitMiddleware{limiter: ratelimiter.New()} }

func (m *rateLimitMiddleware) allow(w http.ResponseWriter, r *http.Request, rps, burst int) bool {
    key := ratelimiter.ClientIP(r.RemoteAddr) + "|" + r.URL.Path
    if !m.limiter.Allow(key, rps, burst) {
        http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
        return false
    }
    return true
}

