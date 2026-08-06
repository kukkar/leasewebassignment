// Package handlers implements the HTTP-facing request/response translation
// for each route: parsing and validating query params (filter.go), mapping
// domain types to wire DTOs (contracts.go), and adapting the
// (*HandlerResult, error)-returning handler style to http.HandlerFunc,
// centralizing error serialization and response caching (adapters.go).
package handlers

import (
	"sync/atomic"

	"github.com/sahil/leasewebassignment/internal/service"
	"go.uber.org/zap"
)

// defaultAllowedRAM and defaultAllowedDiskTypes mirror the filter values
// from the assignment's spec sheet, used whenever config doesn't override them.
var (
	defaultAllowedRAM       = []string{"2GB", "4GB", "8GB", "12GB", "16GB", "24GB", "32GB", "48GB", "64GB", "96GB"}
	defaultAllowedDiskTypes = []string{"SAS", "SATA", "SSD"}
)

func withDefaults(allowedRAM, allowedDiskTypes []string) ([]string, []string) {
	if len(allowedRAM) == 0 {
		allowedRAM = defaultAllowedRAM
	}
	if len(allowedDiskTypes) == 0 {
		allowedDiskTypes = defaultAllowedDiskTypes
	}
	return allowedRAM, allowedDiskTypes
}

type Handler struct {
	Service          service.Service
	Logger           *zap.SugaredLogger
	AllowedRAM       []string
	AllowedDiskTypes []string
	// Ready is flipped to true the first time server data is successfully
	// loaded - at startup or via a later admin upload - and read by
	// GET /readyz. Never nil: NewHandler defaults it if the caller omits one.
	Ready *atomic.Bool
}

// Config holds everything NewHandler needs. A struct instead of positional
// parameters so a new field doesn't silently shift the meaning of existing
// call sites - see internal/server.Config for the same reasoning.
type Config struct {
	Service          service.Service
	Logger           *zap.SugaredLogger
	AllowedRAM       []string
	AllowedDiskTypes []string
	Ready            *atomic.Bool
}

func NewHandler(cfg Config) *Handler {
	allowedRAM, allowedDiskTypes := withDefaults(cfg.AllowedRAM, cfg.AllowedDiskTypes)
	ready := cfg.Ready
	if ready == nil {
		ready = &atomic.Bool{}
	}
	return &Handler{
		Service:          cfg.Service,
		Logger:           cfg.Logger,
		AllowedRAM:       allowedRAM,
		AllowedDiskTypes: allowedDiskTypes,
		Ready:            ready,
	}
}
