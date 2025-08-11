package router

import (
	"io"
	"net/http"
	"strings"

	"github.com/kenelite/go-agw/internal/config"
	"github.com/kenelite/go-agw/internal/observability"
	"github.com/kenelite/go-agw/internal/plugin"
	"github.com/kenelite/go-agw/internal/scheduler"
	"github.com/kenelite/go-agw/internal/upstream"
)

type Router struct {
	routes   []config.RouteConfig
	upstream *upstream.Manager
	sched    scheduler.Scheduler
	plugins  *plugin.Manager
	metrics  *observability.Metrics
	logger   *observability.Logger
	_rlmw    *rateLimitMiddleware
}

func NewRouter(routes []config.RouteConfig, up *upstream.Manager, sch scheduler.Scheduler, pl *plugin.Manager, m *observability.Metrics, l *observability.Logger) (*Router, error) {
	return &Router{routes: routes, upstream: up, sched: sch, plugins: pl, metrics: m, logger: l}, nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.metrics.IncRequests()
	for i, rt := range r.routes {
		if !matchRoute(rt, req) {
			continue
		}
		if !r.preflight(w, req, i) {
			return
		}
		// plugins: before (plugins may mutate request and choose upstream)
		prc := &plugin.RequestContext{Context: req.Context(), Writer: w, Request: req}
		for _, p := range r.plugins.Chain() {
			handled, err := p.BeforeDispatch(prc)
			if err != nil {
				r.logger.Errorw("plugin before error", "plugin", p.Name(), "err", err)
			}
			if handled {
				return
			}
		}
		// choose upstream after plugins
		upstreamName := rt.UpstreamRef
		if name, ok := plugin.UpstreamOverrideFrom(prc.Request.Context()); ok && name != "" {
			upstreamName = name
		}
		ups, ok := r.upstream.Get(upstreamName)
		if !ok || len(ups.Targets) == 0 {
			http.Error(w, "upstream not found", http.StatusBadGateway)
			return
		}

		// pick target
		idx := r.sched.Next(len(ups.Targets))
		if idx < 0 {
			http.Error(w, "no backend", http.StatusServiceUnavailable)
			return
		}
		target := ups.Targets[idx]

		// proxy minimal
		outReq := prc.Request.Clone(prc.Request.Context())
		outReq.URL.Scheme = target.URL.Scheme
		outReq.URL.Host = target.URL.Host
		outReq.URL.Path = singleJoiningSlash(target.URL.Path, prc.Request.URL.Path)
		outReq.RequestURI = ""
		// sanitize and adjust headers
		outReq.Header = cloneHeader(prc.Request.Header)
		removeHopByHopHeaders(outReq.Header)
		if isGRPC(prc.Request) {
			// gRPC requires TE: trailers on HTTP/2; set to be safe for upstreams that expect it
			outReq.Header.Set("TE", "trailers")
		}

		// Ensure HTTP/2 when proxying gRPC if possible; http.Client will negotiate automatically over TLS.
		// For h2c upstreams, users should provide http:// targets; std client will still use HTTP/1.1.
		resp, err := ups.Client.Do(outReq)
		if err != nil {
			r.metrics.IncFailures()
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		// Prepare trailers for downstream if upstream provided any
		if len(resp.Trailer) > 0 {
			for k := range resp.Trailer {
				w.Header().Add("Trailer", k)
			}
		}
		// copy response headers except Trailer (we already declared it above)
		copyHeaderExcept(w.Header(), resp.Header, map[string]struct{}{"Trailer": {}})
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
		// set trailer values after body is written
		for k, vv := range resp.Trailer {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		// plugins: after
		for _, p := range r.plugins.Chain() {
			p.AfterDispatch(prc)
		}
		return
	}
	http.NotFound(w, req)
}

func matchRoute(rt config.RouteConfig, req *http.Request) bool {
	if rt.Path != "" && !strings.HasPrefix(req.URL.Path, rt.Path) {
		return false
	}
	if len(rt.Methods) > 0 {
		ok := false
		for _, m := range rt.Methods {
			if strings.EqualFold(m, req.Method) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func singleJoiningSlash(a, b string) string {
	as := strings.HasSuffix(a, "/")
	bs := strings.HasPrefix(b, "/")
	switch {
	case as && bs:
		return a + b[1:]
	case !as && !bs:
		return a + "/" + b
	default:
		return a + b
	}
}
