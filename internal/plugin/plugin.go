package plugin

import (
	"context"
	"net/http"

	"github.com/kenelite/go-agw/internal/config"
	"github.com/kenelite/go-agw/internal/observability"
)

// RequestContext provides per-request context passed to plugins.
type RequestContext struct {
	Context context.Context
	Writer  http.ResponseWriter
	Request *http.Request
	// Plugin shared storage could be added later
	Response *Response
    Logger   *observability.Logger
    Metrics  *observability.Metrics
    // Resolved upstream info
    UpstreamName   string
    UpstreamTarget string
}

// Plugin defines request lifecycle hooks.
type Plugin interface {
	Name() string
	Init(cfg map[string]any) error
	// BeforeDispatch can modify request or short-circuit by writing response and returning handled=true.
	BeforeDispatch(*RequestContext) (handled bool, err error)
	// AfterDispatch observes/adjusts response. Errors are logged but not returned to client.
	AfterDispatch(*RequestContext)
}

// Response represents a mutable response for plugins in AfterDispatch.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	Trailer    http.Header
}

// Manager wires configured plugins into the request flow.
type Manager struct {
	plugins []Plugin
	logger  *observability.Logger
}

func NewManager(logger *observability.Logger) *Manager { return &Manager{logger: logger} }

func (m *Manager) Init(cfg config.PluginsConfig) error {
	m.plugins = []Plugin{}
	for _, pref := range cfg.Available {
		ctor := getConstructor(pref.Name)
		if ctor == nil {
			if m.logger != nil {
				m.logger.Warnw("unknown plugin", "name", pref.Name)
			}
			continue
		}
		p := ctor()
		if err := p.Init(pref.Config); err != nil {
			if m.logger != nil {
				m.logger.Errorw("plugin init failed", "name", pref.Name, "err", err)
			}
			continue
		}
		m.plugins = append(m.plugins, p)
		if m.logger != nil {
			m.logger.Infow("plugin loaded", "name", p.Name())
		}
	}
	return nil
}

func (m *Manager) Chain() []Plugin { return m.plugins }
