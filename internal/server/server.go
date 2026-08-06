// Package server wires the HTTP server together: routes (routes.go), the
// middleware chain (see internal/server/middleware), and the handlers (see
// internal/server/handlers). NewServer/Config is the only exported surface -
// everything else is assembled internally so callers can't construct a
// Server in a partially-wired state.
package server

import (
	"net/http"
	"sync/atomic"

	handlers "github.com/sahil/leasewebassignment/internal/server/handlers"
	"github.com/sahil/leasewebassignment/internal/server/middleware"
	"github.com/sahil/leasewebassignment/internal/service"
	"go.uber.org/zap"
)

type Server struct {
	handler        *handlers.Handler
	authMiddleware *middleware.AuthMiddleware
	logger         *zap.SugaredLogger
	ready          *atomic.Bool
	once           http.Handler
}

// Config holds everything NewServer needs to wire up the HTTP server.
// Using a struct instead of positional parameters keeps call sites stable
// as fields are added or reordered.
type Config struct {
	Service          service.Service
	Logger           *zap.SugaredLogger
	AuthKey          string
	AllowedRAM       []string
	AllowedDiskTypes []string
	// Ready reflects whether server data has been successfully loaded at
	// least once - read by GET /readyz, flipped true by a successful admin
	// upload even if it was false at construction time. Pass the same
	// *atomic.Bool the caller sets after its own startup load so /readyz
	// reflects reality immediately; if nil, one is created starting false.
	Ready *atomic.Bool
}

func NewServer(cfg Config) *Server {
	ready := cfg.Ready
	if ready == nil {
		ready = &atomic.Bool{}
	}
	return &Server{
		handler: handlers.NewHandler(handlers.Config{
			Service:          cfg.Service,
			Logger:           cfg.Logger,
			AllowedRAM:       cfg.AllowedRAM,
			AllowedDiskTypes: cfg.AllowedDiskTypes,
			Ready:            ready,
		}),
		authMiddleware: middleware.NewAuthMiddleware(cfg.AuthKey),
		logger:         cfg.Logger,
		ready:          ready,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.routes().ServeHTTP(w, r)
}
