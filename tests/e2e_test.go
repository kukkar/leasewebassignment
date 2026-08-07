package tests

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sahil/leasewebassignment/internal/server"
	"github.com/sahil/leasewebassignment/internal/service"
	"github.com/sahil/leasewebassignment/internal/store"
)

func TestE2E_ListServers(t *testing.T) {
	repo := store.NewRepository(store.RepositoryConfig{UploadDir: ""})
	svc := service.NewServerService(repo)
	srv := server.NewServer(server.Config{Service: svc})
	ts := httptest.NewServer(srv)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v1/servers?storage_min=500GB&storage_max=1TB")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// TestE2E_UnknownRoute_Returns404 locks in that a path matching nothing
// registered gets a real 404 in the standard JSON error shape - not the
// catch-all redirect-to-UI behavior this used to have (see routes.go).
func TestE2E_UnknownRoute_Returns404(t *testing.T) {
	repo := store.NewRepository(store.RepositoryConfig{UploadDir: t.TempDir()})
	svc := service.NewServerService(repo)
	srv := server.NewServer(server.Config{Service: svc})
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// A stale/typo'd path missing the /v1 prefix - exactly the case that
	// used to silently 302 into the UI instead of failing clearly.
	resp, err := http.Get(ts.URL + "/servers")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// TestE2E_RootPath_StillRedirectsToUI locks in that the one exception to the
// 404 catch-all - the literal root path - still redirects, since that's the
// documented, intended entry point into the UI.
func TestE2E_RootPath_StillRedirectsToUI(t *testing.T) {
	repo := store.NewRepository(store.RepositoryConfig{UploadDir: t.TempDir()})
	svc := service.NewServerService(repo)
	srv := server.NewServer(server.Config{Service: svc})
	ts := httptest.NewServer(srv)
	defer ts.Close()

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/ui/" {
		t.Fatalf("expected redirect to /ui/, got %q", loc)
	}
}

// TestE2E_WrongMethod_Returns405 locks in the method-specific routing: a
// method that doesn't match any registered handler for a known path gets a
// 405, not silent processing by the wrong handler.
func TestE2E_WrongMethod_Returns405(t *testing.T) {
	repo := store.NewRepository(store.RepositoryConfig{UploadDir: t.TempDir()})
	svc := service.NewServerService(repo)
	srv := server.NewServer(server.Config{Service: svc})
	ts := httptest.NewServer(srv)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodDelete, ts.URL+"/v1/servers", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

// TestE2E_OversizedUpload_Returns413 locks in the request-size cap on
// POST /v1/admin/upload (middleware.NewMaxBytesMiddleware): a body larger
// than the configured limit must be rejected with 413, in the standard
// error JSON shape - not accepted and left to consume unbounded server
// resources, and not a generic 400 that doesn't say what was actually wrong.
func TestE2E_OversizedUpload_Returns413(t *testing.T) {
	const authKey = "test-upload-key"
	repo := store.NewRepository(store.RepositoryConfig{UploadDir: t.TempDir()})
	svc := service.NewServerService(repo)
	srv := server.NewServer(server.Config{Service: svc, AuthKey: authKey})
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// server.go's route registration caps uploads at 20 MiB - comfortably
	// exceed that with a single oversized field, no real CSV content needed.
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("file", "servers.csv")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(strings.Repeat("x", 21<<20))); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/v1/admin/upload", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+authKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", resp.StatusCode)
	}

	var respBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if respBody.Error.Code != "request_too_large" {
		t.Fatalf("expected error code request_too_large, got %q", respBody.Error.Code)
	}
}

// TestE2E_SwaggerDocs_Available locks in that the Swagger UI and the spec it
// serves are both reachable - this is the documentation surface added on
// top of docs/api.md, so a regression here would silently break API
// consumers' ability to discover the contract interactively.
func TestE2E_SwaggerDocs_Available(t *testing.T) {
	repo := store.NewRepository(store.RepositoryConfig{UploadDir: t.TempDir()})
	svc := service.NewServerService(repo)
	srv := server.NewServer(server.Config{Service: svc})
	ts := httptest.NewServer(srv)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/docs/")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/docs/: expected 200, got %d", resp.StatusCode)
	}

	specResp, err := http.Get(ts.URL + "/openapi.yaml")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = specResp.Body.Close() }()
	if specResp.StatusCode != http.StatusOK {
		t.Fatalf("/openapi.yaml: expected 200, got %d", specResp.StatusCode)
	}
}

// TestE2E_Healthz_AlwaysOK locks in that /healthz is a pure liveness check -
// it must report the process as healthy even when no data has ever loaded,
// unlike /readyz below.
func TestE2E_Healthz_AlwaysOK(t *testing.T) {
	repo := store.NewRepository(store.RepositoryConfig{UploadDir: t.TempDir()})
	svc := service.NewServerService(repo)
	srv := server.NewServer(server.Config{Service: svc})
	ts := httptest.NewServer(srv)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// TestE2E_Readyz_ReflectsDataState locks in the graceful-boot design: a
// server with no data loaded reports 503 on /readyz, and a real
// POST /v1/admin/upload - the documented recovery path - flips it to 200.
// This exercises the actual HTTP surface end to end rather than poking the
// internal ready flag directly.
func TestE2E_Readyz_ReflectsDataState(t *testing.T) {
	const authKey = "test-upload-key"
	repo := store.NewRepository(store.RepositoryConfig{UploadDir: t.TempDir()})
	svc := service.NewServerService(repo)
	srv := server.NewServer(server.Config{Service: svc, AuthKey: authKey})
	ts := httptest.NewServer(srv)
	defer ts.Close()

	assertStatus := func(t *testing.T, path string, want int) {
		t.Helper()
		resp, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatalf("%s: request failed: %v", path, err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != want {
			t.Fatalf("%s: expected %d, got %d", path, want, resp.StatusCode)
		}
	}

	assertStatus(t, "/readyz", http.StatusServiceUnavailable)

	csvData := "Model,RAM,HDD,Location,Price\nDell R210,16GB,2x2TBSATA2,AmsterdamAMS-01,49.99\n"
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("file", "servers.csv")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(csvData)); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/v1/admin/upload", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+authKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload request failed: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected upload to return 204, got %d", resp.StatusCode)
	}

	assertStatus(t, "/readyz", http.StatusOK)
}
