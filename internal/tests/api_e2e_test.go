package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sahil/leasewebassignment/internal/model"
	"github.com/sahil/leasewebassignment/internal/server"
	"github.com/sahil/leasewebassignment/internal/service"
	"github.com/sahil/leasewebassignment/internal/store"
	"github.com/sahil/leasewebassignment/internal/testutil"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	repo := store.NewRepository(store.RepositoryConfig{UploadDir: t.TempDir()})
	if err := repo.ReplaceServers(context.Background(), testutil.SampleServers); err != nil {
		t.Fatalf("seed repo: %v", err)
	}
	svc := service.NewServerService(repo)
	srv := server.NewServer(server.Config{Service: svc})
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)
	return ts
}

type serversResponse struct {
	Data []model.Server `json:"data"`
	Meta struct {
		Total  int `json:"total"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"meta"`
}

func getServersPage(t *testing.T, ts *httptest.Server, query string) serversResponse {
	t.Helper()
	resp, err := http.Get(ts.URL + "/v1/servers" + query)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d for query %q", resp.StatusCode, query)
	}
	var body serversResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}

func getServers(t *testing.T, ts *httptest.Server, query string) []model.Server {
	t.Helper()
	return getServersPage(t, ts, query).Data
}

// TestAPI_E2E_Filters locks in the acceptance criteria from the assignment's
// filter spec sheet: results narrow correctly for each supported filter, and
// filters compose with AND semantics when combined.
func TestAPI_E2E_Filters(t *testing.T) {
	ts := newTestServer(t)

	cases := []struct {
		name       string
		query      string
		wantModels []string
	}{
		{"no filters returns full catalog", "", []string{
			"Dell R210", "HP DL180", "RH2288", "Dell R730XD", "HP DL180G6",
			"Dell R930", "HP DL120G7", "Dell R210-II", "HP DL380eG8", "Supermicro SC846",
		}},
		{"storage range 500GB-1TB", "?storage_min=500GB&storage_max=1TB", []string{"Dell R210-II"}},
		{"ram 16GB", "?ram=16GB", []string{"Dell R210", "Dell R210-II"}},
		{"ram checkbox multi-select is OR, not AND", "?ram=16GB&ram=32GB", []string{
			"Dell R210", "HP DL180", "HP DL180G6", "Dell R210-II", "Supermicro SC846",
		}},
		{"location Washington", "?location=Washington", []string{"Dell R730XD", "Supermicro SC846"}},
		{"model substring case-insensitive", "?model=dell", []string{
			"Dell R210", "Dell R730XD", "Dell R930", "Dell R210-II",
		}},
		{"combined ram + location (AND semantics)", "?ram=32GB&location=Washington", []string{"Supermicro SC846"}},
		{"disk_type SATA matches SATA2-labeled drives", "?disk_type=SATA", []string{
			"Dell R210", "HP DL180", "Dell R730XD", "HP DL120G7", "Dell R210-II", "HP DL380eG8", "Supermicro SC846",
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := getServers(t, ts, tc.query)
			gotModels := make([]string, 0, len(got))
			for _, s := range got {
				gotModels = append(gotModels, s.Model)
			}
			if len(gotModels) != len(tc.wantModels) {
				t.Fatalf("query %q: expected %d results %v, got %d %v", tc.query, len(tc.wantModels), tc.wantModels, len(gotModels), gotModels)
			}
			want := map[string]bool{}
			for _, m := range tc.wantModels {
				want[m] = true
			}
			for _, m := range gotModels {
				if !want[m] {
					t.Fatalf("query %q: unexpected model %q in results %v", tc.query, m, gotModels)
				}
			}
		})
	}
}

func TestAPI_E2E_InvalidFilterReturns400(t *testing.T) {
	ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/v1/servers?ram=not-a-real-ram-size")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid ram value, got %d", resp.StatusCode)
	}
}

// TestAPI_E2E_InvalidFilters_AggregatesAllErrors locks in that a request with
// multiple invalid params reports every problem in one response, not just
// the first one encountered.
func TestAPI_E2E_InvalidFilters_AggregatesAllErrors(t *testing.T) {
	ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/v1/servers?ram=bogus-ram&disk_type=bogus-disk")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	var body struct {
		Error struct {
			Details string `json:"details"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !strings.Contains(body.Error.Details, "ram") || !strings.Contains(body.Error.Details, "disk_type") {
		t.Fatalf("expected both ram and disk_type errors in details, got %q", body.Error.Details)
	}
}

// TestAPI_E2E_Pagination locks in the limit/offset acceptance criterion:
// meta.total reflects the full filtered match count even when only one
// page of it is returned in data.
func TestAPI_E2E_Pagination(t *testing.T) {
	ts := newTestServer(t)

	page := getServersPage(t, ts, "?limit=3&offset=0")
	if len(page.Data) != 3 {
		t.Fatalf("expected 3 results on first page, got %d", len(page.Data))
	}
	if page.Meta.Total != 10 {
		t.Fatalf("expected meta.total=10 (full catalog), got %d", page.Meta.Total)
	}
	if page.Meta.Limit != 3 || page.Meta.Offset != 0 {
		t.Fatalf("expected meta limit=3 offset=0, got limit=%d offset=%d", page.Meta.Limit, page.Meta.Offset)
	}

	next := getServersPage(t, ts, "?limit=3&offset=3")
	if len(next.Data) != 3 {
		t.Fatalf("expected 3 results on second page, got %d", len(next.Data))
	}
	if next.Data[0].Model == page.Data[0].Model {
		t.Fatalf("expected second page to differ from first, both started with %q", page.Data[0].Model)
	}

	last := getServersPage(t, ts, "?limit=50&offset=9")
	if len(last.Data) != 1 {
		t.Fatalf("expected 1 remaining result past offset 9, got %d", len(last.Data))
	}
}
