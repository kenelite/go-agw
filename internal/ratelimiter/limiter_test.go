package ratelimiter

import (
    "testing"
    "time"
)

func TestLimiterAllow(t *testing.T) {
    l := New()
    if !l.Allow("k", 1, 1) {
        t.Fatal("first request should pass")
    }
    if l.Allow("k", 1, 1) {
        t.Fatal("second immediate request should be limited")
    }
    time.Sleep(1100 * time.Millisecond)
    if !l.Allow("k", 1, 1) {
        t.Fatal("should allow after refill")
    }
}

