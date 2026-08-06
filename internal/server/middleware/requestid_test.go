package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	var gotFromContext string
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFromContext = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := NewRequestIDMiddleware()(base)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	headerID := rec.Header().Get(RequestIDHeader)
	if headerID == "" {
		t.Fatal("expected non-empty request ID header")
	}
	if gotFromContext != headerID {
		t.Fatalf("context request ID %q does not match header %q", gotFromContext, headerID)
	}
}

func TestRequestIDFromContext_Empty(t *testing.T) {
	if got := RequestIDFromContext(context.Background()); got != "" {
		t.Fatalf("expected empty request ID for context without one, got %q", got)
	}
}
