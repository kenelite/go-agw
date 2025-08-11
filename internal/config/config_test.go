package config

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoadYAML(t *testing.T) {
    yaml := `
server:
  http_addr: ":18080"
  admin_addr: ":19000"
upstreams:
- name: u
  targets: ["http://example.com"]
routes:
- path: "/"
  methods: ["GET"]
  upstream: "u"
`
    dir := t.TempDir()
    p := filepath.Join(dir, "cfg.yaml")
    if err := os.WriteFile(p, []byte(yaml), 0o644); err != nil {
        t.Fatalf("write temp yaml: %v", err)
    }
    c, err := Load(p)
    if err != nil {
        t.Fatalf("load yaml: %v", err)
    }
    if c.Server.HTTPAddr != ":18080" || c.Server.AdminAddr != ":19000" {
        t.Fatalf("server addrs unexpected: %+v", c.Server)
    }
    if len(c.Upstreams) != 1 || c.Upstreams[0].Name != "u" {
        t.Fatalf("upstreams unexpected: %+v", c.Upstreams)
    }
    if len(c.Routes) != 1 || c.Routes[0].UpstreamRef != "u" {
        t.Fatalf("routes unexpected: %+v", c.Routes)
    }
}

