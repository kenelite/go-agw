package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kenelite/go-agw/internal/config"
	"github.com/kenelite/go-agw/internal/observability"
	"github.com/kenelite/go-agw/internal/plugin"
	"github.com/kenelite/go-agw/internal/scheduler"
	"github.com/kenelite/go-agw/internal/upstream"
)

func newTestRouter(t *testing.T, backend http.Handler) *Router {
	t.Helper()
	be := httptest.NewServer(backend)
	t.Cleanup(func() { be.Close() })

	ucfg := []config.UpstreamConfig{{Name: "echo", Targets: []string{be.URL}, Timeout: 2000}}
	logger := observability.NewLogger(nil)
	upm, err := upstream.NewManager(ucfg, logger)
	if err != nil {
		t.Fatalf("upstream manager: %v", err)
	}
	rr := scheduler.NewRoundRobin()
	pm := plugin.NewManager(logger)
	_ = pm.Init(config.PluginsConfig{Available: []config.PluginRef{{Name: "rewrite", Config: map[string]any{"add_headers": map[string]any{"X-Test": "1"}}}}})
	metrics := observability.NewMetrics()
	routes := []config.RouteConfig{{Path: "/", Methods: []string{"GET"}, UpstreamRef: "echo"}}
	r, err := NewRouter(routes, upm, rr, pm, metrics, logger)
	if err != nil {
		t.Fatalf("router: %v", err)
	}
	return r
}

func TestRouterProxy(t *testing.T) {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get("X-Test"); v != "" {
			w.Header().Set("X-Test", v)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r := newTestRouter(t, backend)
	req := httptest.NewRequest(http.MethodGet, "http://agw/hello", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
	if rec.Header().Get("X-Test") != "1" {
		t.Fatalf("expected header X-Test injected by rewrite plugin")
	}
}

func TestRouterRateLimit(t *testing.T) {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r := newTestRouter(t, backend)
	// enable rate limit for the first route
	r.routes[0].RateLimit.RequestsPerSecond = 1
	r.routes[0].RateLimit.Burst = 1

	req1 := httptest.NewRequest(http.MethodGet, "http://agw/", nil)
	req1.RemoteAddr = "1.2.3.4:12345"
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request unexpected code: %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "http://agw/", nil)
	req2.RemoteAddr = "1.2.3.4:12345"
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request expected 429 got %d", rec2.Code)
	}
}

func TestRewritePluginUpstreamOverride(t *testing.T) {
	// backend A returns body A, backend B returns body B
	backendA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("A"))
	}))
	defer backendA.Close()
	backendB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("B"))
	}))
	defer backendB.Close()

	logger := observability.NewLogger(nil)
	upm, err := upstream.NewManager([]config.UpstreamConfig{
		{Name: "a", Targets: []string{backendA.URL}, Timeout: 2000},
		{Name: "b", Targets: []string{backendB.URL}, Timeout: 2000},
	}, logger)
	if err != nil {
		t.Fatalf("upstream manager: %v", err)
	}
	rr := scheduler.NewRoundRobin()
	pm := plugin.NewManager(logger)
	// rewrite to force upstream b
	_ = pm.Init(config.PluginsConfig{Available: []config.PluginRef{{Name: "rewrite", Config: map[string]any{"set_upstream": "b"}}}})
	metrics := observability.NewMetrics()
	routes := []config.RouteConfig{{Path: "/", Methods: []string{"GET"}, UpstreamRef: "a"}}
	r, err := NewRouter(routes, upm, rr, pm, metrics, logger)
	if err != nil {
		t.Fatalf("router: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://agw/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected code: %d", rec.Code)
	}
	if rec.Body.String() != "B" {
		t.Fatalf("expected response from upstream b, got %q", rec.Body.String())
	}
}
