package middleware

import "net/http"

// NewMaxBytesMiddleware caps the total request body at maxBytes, returning
// a clean error instead of letting an unbounded body exhaust server
// resources. This matters specifically for multipart uploads: parsing only
// bounds how much is held *in memory* before spilling the rest to a disk
// temp file - without a hard cap on the body itself, a large enough upload
// fills disk before that in-memory threshold ever comes into play.
func NewMaxBytesMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
