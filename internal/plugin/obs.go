package plugin

import (
    "crypto/rand"
    "encoding/hex"
    "net/http"
    "time"
)

// ObservabilityPlugin adds request logging/audit, metrics tagging, and request IDs.
// Config:
//   request_id_header: "X-Request-ID"
//   correlation_id_header: "X-Correlation-ID"
//   log: true
//   metrics_labels: { "service": "checkout" }
//   baggage: { "user": "${user}" } // placeholder for future prop
type ObservabilityPlugin struct {
    requestIDHeader     string
    correlationIDHeader string
    enableLog           bool
    staticLabels        map[string]string
}

func (p *ObservabilityPlugin) Name() string { return "observability" }

func (p *ObservabilityPlugin) Init(cfg map[string]any) error {
    p.requestIDHeader = getStringOr(cfg, "request_id_header", "X-Request-ID")
    p.correlationIDHeader = getStringOr(cfg, "correlation_id_header", "X-Correlation-ID")
    if v, ok := cfg["log"].(bool); ok { p.enableLog = v } else { p.enableLog = true }
    if m, ok := cfg["metrics_labels"].(map[string]any); ok {
        p.staticLabels = map[string]string{}
        for k, vi := range m { if s, ok := vi.(string); ok { p.staticLabels[k] = s } }
    }
    return nil
}

func (p *ObservabilityPlugin) BeforeDispatch(ctx *RequestContext) (bool, error) {
    // Request ID / Correlation ID
    req := ctx.Request
    rid := headerOr(req.Header, p.requestIDHeader, randomID())
    cid := headerOr(req.Header, p.correlationIDHeader, rid)
    req.Header.Set(p.requestIDHeader, rid)
    req.Header.Set(p.correlationIDHeader, cid)
    // record start time in context
    ctx.Request = req.WithContext(withStartTime(req.Context(), time.Now()))
    return false, nil
}

func (p *ObservabilityPlugin) AfterDispatch(ctx *RequestContext) {
    if ctx.Logger == nil || ctx.Metrics == nil || ctx.Response == nil { return }
    dur := time.Since(startTimeFrom(ctx.Request.Context()))
    // basic structured log
    if p.enableLog {
        ctx.Logger.Infow("request",
            "method", ctx.Request.Method,
            "path", ctx.Request.URL.Path,
            "status", ctx.Response.StatusCode,
            "duration_ms", dur.Milliseconds(),
            "upstream", ctx.UpstreamName,
            "target", ctx.UpstreamTarget,
        )
    }
    // simple metrics: increment total and failures with static labels
    ctx.Metrics.IncRequests()
    if ctx.Response.StatusCode >= http.StatusBadRequest { ctx.Metrics.IncFailures() }
}

func headerOr(h http.Header, key, fallback string) string {
    if v := h.Get(key); v != "" { return v }
    return fallback
}

func randomID() string {
    var b [8]byte
    _, _ = rand.Read(b[:])
    return hex.EncodeToString(b[:])
}

