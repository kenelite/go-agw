package listener

import (
	"context"
	"net/http"
	"time"

	"github.com/kenelite/go-agw/internal/observability"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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
			Addr: addr,
			// Enable h2c so that gRPC over HTTP/2 without TLS is accepted.
			Handler:           h2c.NewHandler(handler, &http2.Server{}),
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
