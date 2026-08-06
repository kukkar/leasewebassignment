package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestLoggingMiddleware_PassesThroughStatusAndBody(t *testing.T) {
	logger := zap.NewNop().Sugar()
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("hi"))
	})

	handler := NewLoggingMiddleware(logger)(base)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/servers", nil))

	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected status %d, got %d", http.StatusTeapot, rec.Code)
	}
	if rec.Body.String() != "hi" {
		t.Fatalf("expected body %q, got %q", "hi", rec.Body.String())
	}
}

func TestLoggingMiddleware_NilLoggerIsNoop(t *testing.T) {
	called := false
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := NewLoggingMiddleware(nil)(base)
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !called {
		t.Fatal("expected base handler to still run with a nil logger")
	}
}
