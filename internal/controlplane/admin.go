package controlplane

import (
	"encoding/json"
	"net/http"

	"github.com/kenelite/go-agw/internal/config"
	"github.com/kenelite/go-agw/internal/observability"
)

func RegisterAdminHandlers(mux *http.ServeMux, metrics *observability.Metrics, cfg *config.Config, logger *observability.Logger) {
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	mux.Handle("/metrics", metrics.Handler())
	mux.Handle("/config", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cfg)
	}))
	_ = logger // avoid unused; in future audit endpoints will use it
}
