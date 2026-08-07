package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/sahil/leasewebassignment/internal/config"
	applog "github.com/sahil/leasewebassignment/internal/platform/log"
	"github.com/sahil/leasewebassignment/internal/platform/shutdown"
	internalServer "github.com/sahil/leasewebassignment/internal/server"
	"github.com/sahil/leasewebassignment/internal/service"
	"github.com/sahil/leasewebassignment/internal/store"
	"go.uber.org/zap"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	// No structured logger exists yet, so bootstrap failures use the stdlib logger.
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := applog.NewLogger(cfg.App.Logging.Level)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync() //nolint:errcheck

	logEffectiveConfig(logger, cfg)

	repo := store.NewRepository(store.RepositoryConfig{UploadDir: cfg.App.UploadDir})
	svc := service.NewLoggingService(service.NewServerServiceWithParser(repo, service.NewCSVParser()), logger)

	// Startup data load failure is deliberately non-fatal: the server has
	// its own recovery path (POST /v1/admin/upload), so refusing to boot
	// over a bad/missing data file would take the whole service down for a
	// problem it can already self-heal from at runtime. ready starts false
	// and GET /readyz reports that honestly until a load succeeds - here or
	// via a later admin upload.
	ready := &atomic.Bool{}
	if err := svc.LoadServerData(context.Background(), cfg.App.DataFile); err != nil {
		logger.Errorw("failed to load startup data - starting anyway with an empty catalog",
			"data_file", cfg.App.DataFile,
			"error", err,
			"recovery", "POST /v1/admin/upload",
		)
	} else {
		ready.Store(true)
	}

	runServer(cfg, svc, logger, ready)
}

// logEffectiveConfig logs what was actually loaded so "what config is this
// instance running with" is answerable from logs alone, without needing
// shell access to re-read the file. Secrets are never logged - only whether
// one was set, since even a redacted-looking value can leak length/shape.
func logEffectiveConfig(logger *zap.SugaredLogger, cfg *config.Config) {
	logger.Infow("effective config",
		"server_host", cfg.Server.Host,
		"server_port", cfg.Server.Port,
		"server_timeout_seconds", cfg.Server.Timeout,
		"data_file", cfg.App.DataFile,
		"upload_dir", cfg.App.UploadDir,
		"admin_token_set", cfg.App.AdminToken != "",
		"logging_level", cfg.App.Logging.Level,
		"allowed_ram", cfg.App.AllowedRAM,
		"allowed_disk_types", cfg.App.AllowedDiskTypes,
	)
}

func runServer(cfg *config.Config, svc service.Service, logger *zap.SugaredLogger, ready *atomic.Bool) {
	httpServer := internalServer.NewServer(internalServer.Config{
		Service:          svc,
		Logger:           logger,
		AuthKey:          cfg.App.AdminToken,
		AllowedRAM:       cfg.App.AllowedRAM,
		AllowedDiskTypes: cfg.App.AllowedDiskTypes,
		Ready:            ready,
	})
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	requestTimeout := time.Duration(cfg.Server.Timeout) * time.Second
	srv := &http.Server{
		Addr:              addr,
		Handler:           httpServer,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       requestTimeout,
		WriteTimeout:      requestTimeout,
		IdleTimeout:       60 * time.Second,
	}

	logger.Infof("listening on %s", addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server error: %v", err)
		}
	}()

	stop := shutdown.WaitForSignals()
	<-stop

	if err := shutdown.GracefulShutdown(context.Background(), 5*time.Second, srv.Shutdown); err != nil {
		logger.Errorf("graceful shutdown failed: %v", err)
	}
	logger.Info("server stopped")
}
