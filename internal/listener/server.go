package listener

import (
    "context"
    "net/http"
    "time"

    "github.com/kenelite/go-agw/internal/observability"
)

type Server struct {
    addr   string
    srv    *http.Server
    logger *observability.Logger
}

func NewServer(addr string, handler http.Handler, logger *observability.Logger) *Server {
    return &Server{
        addr: addr,
        srv: &http.Server{
            Addr:              addr,
            Handler:           handler,
            ReadHeaderTimeout: 5 * time.Second,
        },
        logger: logger,
    }
}

func (s *Server) Start() error {
    s.logger.Infow("listening", "addr", s.addr)
    return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error { return s.srv.Shutdown(ctx) }

