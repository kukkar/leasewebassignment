package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

// NewGzipMiddleware compresses response bodies when the client advertises
// support for it. Mount it closest to the handler (innermost) so outer
// middlewares like logging see the real, uncompressed status code - gzip
// only wraps the response body, never WriteHeader.
func NewGzipMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Add("Vary", "Accept-Encoding")
			gz := gzip.NewWriter(w)
			defer func() { _ = gz.Close() }()
			next.ServeHTTP(gzipResponseWriter{ResponseWriter: w, writer: gz}, r)
		})
	}
}
