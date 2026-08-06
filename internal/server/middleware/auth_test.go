package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cases := []struct {
		name       string
		authKey    string
		header     string
		wantStatus int
	}{
		{"valid bearer token", "secret", "Bearer secret", http.StatusOK},
		{"missing header", "secret", "", http.StatusUnauthorized},
		{"wrong token", "secret", "Bearer wrong", http.StatusUnauthorized},
		{"missing bearer prefix", "secret", "secret", http.StatusUnauthorized},
		{"empty configured key always fails closed", "", "Bearer anything", http.StatusUnauthorized},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mw := NewAuthMiddleware(tc.authKey)
			req := httptest.NewRequest(http.MethodPost, "/admin/upload", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()

			mw.Middleware(next).ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}
		})
	}
}
