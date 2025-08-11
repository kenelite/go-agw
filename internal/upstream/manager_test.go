package upstream

import (
    "testing"

    "github.com/kenelite/go-agw/internal/config"
)

func TestManagerBasic(t *testing.T) {
    m, err := NewManager([]config.UpstreamConfig{{Name: "u", Targets: []string{"http://example.com"}, Timeout: 1000}}, nil)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    u, ok := m.Get("u")
    if !ok {
        t.Fatal("expected upstream 'u' present")
    }
    if len(u.Targets) != 1 || u.Targets[0].URL.String() != "http://example.com" {
        t.Fatalf("unexpected targets: %+v", u.Targets)
    }
}

