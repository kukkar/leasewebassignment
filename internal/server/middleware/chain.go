// Package middleware holds the HTTP middlewares applied around routes:
// panic recovery, request ID assignment, access logging, gzip compression,
// and bearer-token auth (auth.go is applied per-route; the rest are applied
// globally - see internal/server/routes.go).
package middleware

import "net/http"

// Chain composes middlewares around a base handler. The first middleware in
// the list is outermost - it sees the request before all the others and the
// response after all the others. Chain(base, a, b, c) executes a -> b -> c
// -> base on the way in, and base -> c -> b -> a on the way out.
func Chain(base http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	h := base
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
