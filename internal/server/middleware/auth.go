package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/sahil/leasewebassignment/internal/api"
	"github.com/sahil/leasewebassignment/internal/platform/httperr"
)

const bearerPrefix = "Bearer "

type AuthMiddleware struct {
	authKey string
}

func NewAuthMiddleware(authKey string) *AuthMiddleware {
	return &AuthMiddleware{authKey: authKey}
}

// Middleware requires an "Authorization: Bearer <token>" header matching the
// configured auth key. An empty configured key always fails closed - there's
// no way to "disable" auth by leaving the config value blank.
func (a *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(header, bearerPrefix)
		if !ok || a.authKey == "" || !constantTimeEquals(token, a.authKey) {
			api.WriteError(w, httperr.Unauthorized("authorization required", "authorization header must be 'Bearer <token>' with a valid token"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func constantTimeEquals(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
