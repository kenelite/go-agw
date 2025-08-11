package upstream

import (
    "errors"
    "net/http"
    "net/url"
    "sync"
    "time"

    "github.com/kenelite/go-agw/internal/config"
    "github.com/kenelite/go-agw/internal/observability"
)

type Target struct {
    URL *url.URL
}

type Upstream struct {
    Name    string
    Targets []Target
    Client  *http.Client
}

type Manager struct {
    mu        sync.RWMutex
    upstreams map[string]*Upstream
    logger    *observability.Logger
}

func NewManager(cfgs []config.UpstreamConfig, logger *observability.Logger) (*Manager, error) {
    m := &Manager{upstreams: make(map[string]*Upstream), logger: logger}
    for _, uc := range cfgs {
        if uc.Name == "" || len(uc.Targets) == 0 {
            return nil, errors.New("upstream name and targets required")
        }
        ups := &Upstream{Name: uc.Name, Client: &http.Client{Timeout: time.Duration(uc.Timeout) * time.Millisecond}}
        for _, t := range uc.Targets {
            u, err := url.Parse(t)
            if err != nil { return nil, err }
            ups.Targets = append(ups.Targets, Target{URL: u})
        }
        m.upstreams[uc.Name] = ups
    }
    return m, nil
}

func (m *Manager) Get(name string) (*Upstream, bool) {
    m.mu.RLock(); defer m.mu.RUnlock()
    u, ok := m.upstreams[name]
    return u, ok
}

