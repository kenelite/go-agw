package router

import (
    "net/http"
)

// integrate rate limit checks into routing
func (r *Router) maybeAllow(w http.ResponseWriter, req *http.Request, rtIdx int) bool {
    rl := r.routes[rtIdx].RateLimit
    if rl.RequestsPerSecond == 0 { return true }
    if r._rlmw == nil { r._rlmw = newRateLimitMiddleware() }
    return r._rlmw.allow(w, req, rl.RequestsPerSecond, rl.Burst)
}

// internal state
func (r *Router) ensureRateLimit(w http.ResponseWriter, req *http.Request, idx int) bool {
    return r.maybeAllow(w, req, idx)
}

// call at match point
func (r *Router) preflight(w http.ResponseWriter, req *http.Request, idx int) bool {
    return r.ensureRateLimit(w, req, idx)
}

// augment Router with field for middleware
func init() {}

