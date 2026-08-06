package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddleware_CompressesWhenAccepted(t *testing.T) {
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hello":"world"}`))
	})

	handler := NewGzipMiddleware()(base)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected Content-Encoding: gzip, got %q", rec.Header().Get("Content-Encoding"))
	}
	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("expected valid gzip body: %v", err)
	}
	defer func() { _ = gr.Close() }()
	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to decompress body: %v", err)
	}
	if string(decompressed) != `{"hello":"world"}` {
		t.Fatalf("unexpected decompressed body: %q", decompressed)
	}
}

func TestGzipMiddleware_SkipsWhenNotAccepted(t *testing.T) {
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("plain"))
	})

	handler := NewGzipMiddleware()(base)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Fatal("expected no gzip encoding when client doesn't advertise support")
	}
	if rec.Body.String() != "plain" {
		t.Fatalf("expected uncompressed body, got %q", rec.Body.String())
	}
}
