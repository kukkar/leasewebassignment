package tests

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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
