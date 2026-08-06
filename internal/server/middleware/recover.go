package middleware

import (
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"

	"github.com/sahil/leasewebassignment/internal/api"
)

// NewRecoverMiddleware converts a panic in any handler into the API's
// standard structured 500 response instead of net/http's default behavior
// (an ungraceful stack trace on stderr and a closed connection with no
// body). Mount this outermost in the chain so it can catch panics from
// every other middleware and the handler itself.
func NewRecoverMiddleware(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					if logger != nil {
						logger.Errorw("panic recovered",
							"panic", rec,
							"path", r.URL.Path,
							"request_id", RequestIDFromContext(r.Context()),
							"stack", string(debug.Stack()),
						)
					}
					api.WriteError(w, api.InternalError("internal server error", "unexpected error"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
