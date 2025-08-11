package ratelimiter

import (
    "net"
    "sync"
    "time"
)

// Simple token bucket per key (e.g., client IP or route key)
type TokenBucket struct {
    capacity int
    tokens   float64
    rate     float64 // tokens per second
    last     time.Time
}

type Limiter struct {
    mu    sync.Mutex
    store map[string]*TokenBucket
}

func New() *Limiter { return &Limiter{store: map[string]*TokenBucket{}} }

func (l *Limiter) Allow(key string, rps, burst int) bool {
    if rps <= 0 { return true }
    now := time.Now()
    l.mu.Lock()
    defer l.mu.Unlock()
    b, ok := l.store[key]
    if !ok {
        b = &TokenBucket{capacity: max(1, burst), tokens: float64(burst), rate: float64(rps), last: now}
        l.store[key] = b
    }
    // refill
    elapsed := now.Sub(b.last).Seconds()
    b.tokens += elapsed * b.rate
    if b.tokens > float64(b.capacity) { b.tokens = float64(b.capacity) }
    b.last = now
    if b.tokens >= 1 {
        b.tokens -= 1
        return true
    }
    return false
}

func ClientIP(hostport string) string {
    host, _, err := net.SplitHostPort(hostport)
    if err != nil { return hostport }
    return host
}

func max(a, b int) int { if a > b { return a }; return b }

