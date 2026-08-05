package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sahil/leasewebassignment/internal/config"
	internalServer "github.com/sahil/leasewebassignment/internal/server"
	"github.com/sahil/leasewebassignment/internal/service"
	"github.com/sahil/leasewebassignment/internal/shutdown"
	"github.com/sahil/leasewebassignment/internal/store"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	repo := store.NewRepository(store.RepositoryConfig{UploadDir: cfg.App.UploadDir})
	svc := service.NewServerService(repo)

	if err := svc.LoadServerData(context.Background(), cfg.App.DataFile); err != nil {
		log.Fatalf("failed to load startup data: %v", err)
	}

	runServer(cfg, svc)
}

func runServer(cfg *config.Config, svc service.Service) {
	httpServer := internalServer.NewServer(svc, cfg.App.JWTSigningKey)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           httpServer,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on %s", addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := shutdown.WaitForSignals()
	<-stop

	if err := shutdown.GracefulShutdown(context.Background(), 5*time.Second, srv.Shutdown); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
