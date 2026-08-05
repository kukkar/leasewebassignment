package handlers

import (
	"net/http"

	"github.com/sahil/leasewebassignment/internal/api"
)

type HandlerResult struct {
	Status  int
	Body    any
	Headers map[string]string
}

type HandlerFunc func(*http.Request) (*HandlerResult, error)

func Adapt(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := fn(r)
		if err != nil {
			api.WriteError(w, err)
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
		api.WriteJSON(w, result.Status, result.Body)
	}
}
