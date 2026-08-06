package handlers

import (
	"encoding/json"
	"hash/fnv"
	"net/http"
	"strconv"

	"github.com/sahil/leasewebassignment/internal/api"
	"github.com/sahil/leasewebassignment/internal/platform/httperr"
	"go.uber.org/zap"
)

type HandlerResult struct {
	Status  int
	Body    any
	Headers map[string]string
	// Cacheable marks a response as safe for conditional-request caching.
	// When true, Adapt computes a content-hash ETag and honors an incoming
	// If-None-Match with a 304. Only set this for GET responses whose body
	// is fully determined by server-side state the client can't mutate.
	Cacheable bool
}

type HandlerFunc func(*http.Request) (*HandlerResult, error)

// Adapt bridges a HandlerFunc (a handler that returns its result/error
// rather than writing directly to the ResponseWriter) to http.HandlerFunc.
// It's the single place responses are serialized, which is what makes ETag
// support and 5xx logging possible without every handler repeating them.
func Adapt(logger *zap.SugaredLogger, fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := fn(r)
		if err != nil {
			status, apiErr := api.MapError(err)
			// 4xx is normal traffic (bad client input) and already visible
			// in the access log's status field; 5xx means the server did
			// something wrong and needs the full error captured here.
			if status >= http.StatusInternalServerError && logger != nil {
				logger.Errorw("handler error", "path", r.URL.Path, "status", status, "error", err)
			}
			httperr.WriteJSON(w, status, httperr.ErrorResponse{Error: apiErr})
			return
		}
		if result == nil {
			return
		}
		for key, value := range result.Headers {
			w.Header().Set(key, value)
		}
		if result.Status == 0 {
			result.Status = http.StatusOK
		}
		if result.Body == nil {
			w.WriteHeader(result.Status)
			return
		}
		writeJSON(w, r, result.Status, result.Body, result.Cacheable)
	}
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, body any, cacheable bool) {
	data, err := json.Marshal(body)
	if err != nil {
		http.Error(w, `{"error":{"code":"internal_error","message":"failed to encode response"}}`, http.StatusInternalServerError)
		return
	}
	if cacheable {
		etag := computeETag(data)
		w.Header().Set("ETag", etag)
		// "no-cache" is a common misnomer - it means "cache it, but always
		// revalidate with the server first", which is exactly right for a
		// resource that changes on upload but should never be served stale.
		w.Header().Set("Cache-Control", "no-cache")
		if match := r.Header.Get("If-None-Match"); match != "" && match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

func computeETag(data []byte) string {
	h := fnv.New64a()
	_, _ = h.Write(data)
	return `"` + strconv.FormatUint(h.Sum64(), 16) + `"`
}
