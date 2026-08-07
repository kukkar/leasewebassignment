package server

import (
	"net/http"

	"github.com/sahil/leasewebassignment/internal/api"
	"github.com/sahil/leasewebassignment/internal/platform/httperr"
	"github.com/sahil/leasewebassignment/internal/server/handlers"
	"github.com/sahil/leasewebassignment/internal/server/middleware"
)

// maxUploadBytes caps the total POST /v1/admin/upload request body. It's
// deliberately larger than handlers.MaxUploadMemory (the in-memory portion
// multipart parsing uses before spilling to a temp file): a real catalog
// upload spilling a little to a short-lived temp file is fine, an unbounded
// body filling disk before any size check ever runs is not.
const maxUploadBytes = 20 << 20 // 20 MiB

func (s *Server) routes() http.Handler {
	if s.once != nil {
		return s.once
	}

	mux := http.NewServeMux()

	// v1 API. Versioned so the request/response contract can evolve later
	// without breaking existing clients - see docs/api.md.
	registerRoute(mux, http.MethodGet, "/v1/servers", handlers.Adapt(s.logger, s.handler.GetServers))
	registerRoute(mux, http.MethodPost, "/v1/admin/upload",
		middleware.NewMaxBytesMiddleware(maxUploadBytes)(
			s.authMiddleware.Middleware(handlers.Adapt(s.logger, s.handler.Upload)),
		),
	)

	// Ops endpoints are deliberately unversioned - they're infrastructure
	// contracts (what an orchestrator/load balancer polls), not part of the
	// API surface clients integrate against. See docs/api.md.
	registerRoute(mux, http.MethodGet, "/healthz", http.HandlerFunc(handleHealthz))
	registerRoute(mux, http.MethodGet, "/readyz", http.HandlerFunc(s.handleReadyz))

	// Serve the filter UI under /ui/
	fileServer := http.FileServer(http.Dir("web"))
	mux.Handle("GET /ui/", http.StripPrefix("/ui/", fileServer))

	// Interactive API docs (Swagger UI) under /docs/, backed by the
	// hand-written spec at docs/openapi.yaml - see swaggerui.go.
	mux.Handle("GET /docs/", http.StripPrefix("/docs/", newSwaggerUIHandler()))
	registerRoute(mux, http.MethodGet, "/openapi.yaml", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/openapi.yaml")
	}))

	// "/" is Go's catch-all pattern - it matches any request that didn't
	// match a more specific pattern above, not just the literal root path.
	// It used to unconditionally redirect everywhere to /ui/, which meant
	// any typo'd or stale API path (e.g. missing the /v1 prefix) silently
	// "succeeded" with a 302 into an HTML page instead of failing with a
	// clear 404 - actively unhelpful for exactly the kind of mistake it was
	// hiding. Only the exact root path redirects now; everything else gets
	// a real 404 in the same JSON shape every other error uses.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/ui/", http.StatusFound)
			return
		}
		api.WriteError(w, httperr.NotFound("not found", "no route matches "+r.Method+" "+r.URL.Path))
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

// registerRoute binds handler to method+path, plus an explicit 405 fallback
// for every other method at that same path. This isn't redundant with Go's
// built-in cross-method 405 detection: that only triggers when *multiple*
// method-specific patterns share a path (e.g. both "GET /x" and "POST /x"
// registered) - a route with exactly one method registered has nothing to
// conflict with, so a mismatched method falls through to the next-broadest
// pattern (here, the catch-all above) and would silently become a plain 404
// instead of the more precise 405 a client can actually act on.
func registerRoute(mux *http.ServeMux, method, path string, handler http.Handler) {
	mux.Handle(method+" "+path, handler)
	mux.Handle(path, methodNotAllowed(method))
}

func methodNotAllowed(allowed string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", allowed)
		api.WriteError(w, httperr.MethodNotAllowed(
			"method not allowed",
			r.Method+" is not supported on this path, use "+allowed,
		))
	})
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
		api.WriteError(w, httperr.ServiceUnavailable(
			"server has no data loaded yet",
			"upload a catalog via POST /v1/admin/upload to recover",
		))
		return
	}
	w.WriteHeader(http.StatusOK)
}
