package server

import (
	"net/http"

	"github.com/sahil/leasewebassignment/internal/api"
	"github.com/sahil/leasewebassignment/internal/server/handlers"
	"github.com/sahil/leasewebassignment/internal/server/middleware"
)

func (s *Server) routes() http.Handler {
	if s.once != nil {
		return s.once
	}

	mux := http.NewServeMux()

	// v1 API. Versioned so the request/response contract can evolve later
	// without breaking existing clients - see docs/api.md.
	mux.Handle("/v1/servers", handlers.Adapt(s.logger, s.handler.GetServers))
	mux.Handle("/v1/admin/upload", s.authMiddleware.Middleware(handlers.Adapt(s.logger, s.handler.Upload)))

	// Ops endpoints are deliberately unversioned - they're infrastructure
	// contracts (what an orchestrator/load balancer polls), not part of the
	// API surface clients integrate against. See docs/api.md.
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/readyz", s.handleReadyz)

	// Serve the filter UI under /ui/
	fileServer := http.FileServer(http.Dir("web"))
	mux.Handle("/ui/", http.StripPrefix("/ui/", fileServer))
	// redirect root to UI
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/", http.StatusFound)
	})

	// Outermost first: recover must see panics from every other middleware;
	// gzip innermost so it only ever touches the response body, never status.
	s.once = middleware.Chain(mux,
		middleware.NewRecoverMiddleware(s.logger),
		middleware.NewRequestIDMiddleware(),
		middleware.NewLoggingMiddleware(s.logger),
		middleware.NewGzipMiddleware(),
	)
	return s.once
}

// handleHealthz is a pure liveness check - it always returns 200 once the
// process is serving HTTP at all, regardless of whether data has loaded.
func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// handleReadyz is a readiness check - it returns 503 until server data has
// been successfully loaded at least once (at startup or via a later admin
// upload). Unlike handleHealthz, this can legitimately fail: the server is
// allowed to boot without data (see cmd/main.go), so an orchestrator using
// this to gate traffic won't route requests to an instance with nothing to serve.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	if !s.ready.Load() {
		api.WriteError(w, api.ServiceUnavailable(
			"server has no data loaded yet",
			"upload a catalog via POST /v1/admin/upload to recover",
		))
		return
	}
	w.WriteHeader(http.StatusOK)
}
