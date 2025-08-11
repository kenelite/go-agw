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

// Manager wires configured plugins into the request flow.
type Manager struct {
    plugins []Plugin
    logger  *observability.Logger
}

func NewManager(logger *observability.Logger) *Manager { return &Manager{logger: logger} }

func (m *Manager) Init(cfg config.PluginsConfig) error {
    // For now we have no built-ins registered automatically.
    // In future, this could dynamically load via Go plugins or compile-time registry.
    m.plugins = []Plugin{}
    return nil
}

func (m *Manager) Chain() []Plugin { return m.plugins }

