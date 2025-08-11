package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kenelite/go-agw/internal/config"
	"github.com/kenelite/go-agw/internal/controlplane"
	"github.com/kenelite/go-agw/internal/listener"
	"github.com/kenelite/go-agw/internal/observability"
	"github.com/kenelite/go-agw/internal/plugin"
	"github.com/kenelite/go-agw/internal/router"
	"github.com/kenelite/go-agw/internal/scheduler"
	"github.com/kenelite/go-agw/internal/upstream"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", os.Getenv("GO_AGW_CONFIG"), "Path to config file (yaml)")
	flag.Parse()

	if configPath == "" {
		configPath = "./deploy/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := observability.NewLogger(cfg)
	defer func() { _ = logger.Sync() }()

	logger.Infow("starting go-agw", "http_addr", cfg.Server.HTTPAddr, "admin_addr", cfg.Server.AdminAddr)

	// Observability setup (metrics/tracing minimal)
	metrics := observability.NewMetrics()

	// Upstream and scheduler
	upstreamMgr, err := upstream.NewManager(cfg.Upstreams, logger)
	if err != nil {
		logger.Fatalw("failed to init upstream manager", "err", err)
	}

	sched := scheduler.NewRoundRobin()

	// Plugins
	pluginMgr := plugin.NewManager(logger)
	if err := pluginMgr.Init(cfg.Plugins); err != nil {
		logger.Fatalw("failed to initialize plugins", "err", err)
	}

	// Router
	rtr, err := router.NewRouter(cfg.Routes, upstreamMgr, sched, pluginMgr, metrics, logger)
	if err != nil {
		logger.Fatalw("failed to init router", "err", err)
	}

	// Data plane server
	dataSrv := listener.NewServer(cfg.Server.HTTPAddr, rtr, logger)

	// Admin plane server
	adminMux := http.NewServeMux()
	controlplane.RegisterAdminHandlers(adminMux, metrics, cfg, logger)
	adminSrv := &http.Server{Addr: cfg.Server.AdminAddr, Handler: adminMux, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		logger.Infow("admin server listening", "addr", cfg.Server.AdminAddr)
		if err := adminSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalw("admin server error", "err", err)
		}
	}()

	go func() {
		if err := dataSrv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Fatalw("data server error", "err", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = dataSrv.Shutdown(ctx)
	_ = adminSrv.Shutdown(ctx)
}
