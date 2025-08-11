package controlplane

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/kenelite/go-agw/internal/config"
    "github.com/kenelite/go-agw/internal/observability"
)

func TestHealthz(t *testing.T) {
    mux := http.NewServeMux()
    metrics := observability.NewMetrics()
    cfg := &config.Config{}
    logger := observability.NewLogger(nil)
    RegisterAdminHandlers(mux, metrics, cfg, logger)

    req := httptest.NewRequest(http.MethodGet, "http://admin/healthz", nil)
    rec := httptest.NewRecorder()
    mux.ServeHTTP(rec, req)
    if rec.Code != http.StatusOK {
        t.Fatalf("unexpected status: %d", rec.Code)
    }
    if rec.Body.String() != "ok" {
        t.Fatalf("unexpected body: %q", rec.Body.String())
    }
}

