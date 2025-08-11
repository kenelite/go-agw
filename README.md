# go-agw

A lightweight API Gateway (AGW) in Go with plugin, rate limiting, scheduling and observability.

## Features
- Minimal HTTP reverse proxy with routing by path and method
- Pluggable request lifecycle hooks (before/after dispatch)
- Round-robin load balancing across upstream targets
- Simple token-bucket rate limiting per client IP and route (extensible)
- Basic observability: counters endpoint and healthz, config dump

## Getting Started
1. Install Go 1.21+
2. Build and run:

```bash
go mod tidy
go run ./cmd/go-agw --config ./deploy/config.yaml
```

Admin: `http://localhost:9000/healthz`, `http://localhost:9000/metrics`, `http://localhost:9000/config`

## Config
See `deploy/config.yaml` for an example.
