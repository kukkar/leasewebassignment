package tests

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/sahil/leasewebassignment/internal/model"
	"github.com/sahil/leasewebassignment/internal/server"
	"github.com/sahil/leasewebassignment/internal/service"
	"github.com/sahil/leasewebassignment/internal/store"
)

// This file validates the filter API against the real production dataset
// (data/servers.csv, 486 rows), not the small hand-written fixture the rest
// of the suite uses (internal/testutil). Two different concerns:
//
//  1. Data validation: every row in the real file must parse into a Server
//     with sane derived values - a malformed RAM/HDD cell wouldn't fail to
//     load (see service.CSVParser.Parse, which only validates Price), it
//     would just silently stop matching disk_type/storage_min/storage_max
//     filters for that row. That failure mode is invisible unless something
//     specifically checks every row's derived fields, which is what
//     TestRealData_EveryRowParsesToValidFields does.
//
//  2. Filter correctness at real scale: the rest of the suite's expected
//     result sets are hand-typed against a 10-row fixture. That doesn't
//     prove the filter logic is right on 486 rows of messy real data - a
//     hand-typed expectation list for that many rows would be unreviewable
//     and untrustworthy. Instead, every test below computes its own
//     "expected" answer with an oracle implementation deliberately written
//     using different mechanics than the production code (index/split-based
//     here vs. regexp-based in internal/model) - see the oracle* functions.
//     If both independently arrive at the same answer for all 486 rows, the
//     production parsing/filtering logic is very unlikely to be wrong in a
//     way that would also be wrong in the oracle.

// oracleRAMFamily independently reproduces model.RAMFamily's contract
// ("<digits>GB" prefix, family only) using substring search instead of
// model's regexp, so a regex bug can't produce a matching wrong answer here.
func oracleRAMFamily(raw string) string {
	upper := strings.ToUpper(strings.TrimSpace(raw))
	idx := strings.Index(upper, "GB")
	if idx == -1 {
		return upper
	}
	return upper[:idx+2]
}

// oracleDiskType independently reproduces model.NormalizeDiskType's
// contract (collapse a raw HDD label to SAS/SATA/SSD) via substring
// containment instead of model's trailing-digit-stripping regex.
func oracleDiskType(hdd string) string {
	upper := strings.ToUpper(hdd)
	switch {
	case strings.Contains(upper, "SSD"):
		return "SSD"
	case strings.Contains(upper, "SATA"):
		return "SATA"
	case strings.Contains(upper, "SAS"):
		return "SAS"
	default:
		return ""
	}
}

// oracleTotalStorageGB independently reproduces model.ParseHDD's total
// capacity math ("<count>x<size><unit>...") via manual digit-scanning
// instead of model's regexp, catching a multiplication/unit-conversion bug
// that a shared implementation would hide from both sides of the check.
func oracleTotalStorageGB(hdd string) (int, error) {
	parts := strings.SplitN(hdd, "x", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("no 'x' separator in %q", hdd)
	}
	count, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("bad count in %q: %w", hdd, err)
	}
	rest := parts[1]
	i := 0
	for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
		i++
	}
	if i == 0 {
		return 0, fmt.Errorf("no size digits in %q", hdd)
	}
	size, err := strconv.Atoi(rest[:i])
	if err != nil {
		return 0, fmt.Errorf("bad size in %q: %w", hdd, err)
	}
	switch unit := strings.ToUpper(rest[i:]); {
	case strings.HasPrefix(unit, "TB"):
		return count * size * 1024, nil
	case strings.HasPrefix(unit, "GB"):
		return count * size, nil
	default:
		return 0, fmt.Errorf("unknown unit in %q", hdd)
	}
}

// loadRealServers parses the actual data/servers.csv through the same
// parser the server uses for startup load and POST /v1/admin/upload -
// TestMain (main_test.go) has already chdir'd the test process to the repo
// root, so this relative path matches how the app itself reads it.
func loadRealServers(t *testing.T) []model.Server {
	t.Helper()
	f, err := os.Open("data/servers.csv")
	if err != nil {
		t.Fatalf("open data/servers.csv: %v", err)
	}
	defer func() { _ = f.Close() }()

	servers, err := service.NewCSVParser().Parse(f)
	if err != nil {
		t.Fatalf("data/servers.csv failed to parse: %v", err)
	}
	if len(servers) == 0 {
		t.Fatal("data/servers.csv parsed to zero rows")
	}
	return servers
}

// newRealDataServer boots the full HTTP stack (routing, middleware, service,
// in-memory store) seeded with the real dataset - the same stack a real
// client of the API talks to, not a package-internal shortcut.
func newRealDataServer(t *testing.T, servers []model.Server) *httptest.Server {
	t.Helper()
	repo := store.NewRepository(store.RepositoryConfig{UploadDir: t.TempDir()})
	if err := repo.ReplaceServers(context.Background(), servers); err != nil {
		t.Fatalf("seed repo with real data: %v", err)
	}
	svc := service.NewServerService(repo)
	srv := server.NewServer(server.Config{Service: svc})
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)
	return ts
}

// fetchAllServers pages through the full result set for query (the server
// caps a single page at handlers.MaxLimit=200, well under the real
// catalog's 486 rows), so callers can diff an exact model set rather than
// trust a single page's worth of results.
func fetchAllServers(t *testing.T, ts *httptest.Server, query string) []model.Server {
	t.Helper()
	const pageSize = 200
	var all []model.Server
	offset := 0
	for {
		sep := "&"
		if !strings.Contains(query, "?") {
			sep = "?"
		}
		page := getServersPage(t, ts, fmt.Sprintf("%s%slimit=%d&offset=%d", query, sep, pageSize, offset))
		all = append(all, page.Data...)
		if len(page.Data) < pageSize {
			if len(all) != page.Meta.Total {
				t.Fatalf("query %q: paginated through %d results but meta.total=%d", query, len(all), page.Meta.Total)
			}
			return all
		}
		offset += pageSize
	}
}

func modelSet(servers []model.Server) map[string]int {
	set := map[string]int{}
	for _, s := range servers {
		set[s.Model]++
	}
	return set
}

// assertSameModels fails with a readable diff if got and want don't contain
// the exact same models with the exact same multiplicity (the real dataset
// has duplicate model names across locations, so a plain set isn't enough).
func assertSameModels(t *testing.T, label string, got, want []model.Server) {
	t.Helper()
	gotSet, wantSet := modelSet(got), modelSet(want)
	if len(got) != len(want) {
		t.Errorf("%s: expected %d results, got %d", label, len(want), len(got))
	}
	for m, wantCount := range wantSet {
		if gotSet[m] != wantCount {
			t.Errorf("%s: model %q expected %d occurrences, got %d", label, m, wantCount, gotSet[m])
		}
	}
	for m, gotCount := range gotSet {
		if wantSet[m] == 0 {
			t.Errorf("%s: unexpected model %q (%d occurrences) not in expected set", label, m, gotCount)
		}
	}
}

// TestRealData_EveryRowParsesToValidFields is the data-validation half: it
// doesn't touch the HTTP API at all, it just proves every one of the 486
// real rows produces sane derived values. A row that silently failed
// derived-field parsing (see newIndexedServer in internal/store) would
// never surface as a parse error - it would just quietly stop matching
// disk_type/storage filters, which is exactly the kind of bug this dataset
// is large and varied enough to hide from a hand-picked fixture.
func TestRealData_EveryRowParsesToValidFields(t *testing.T) {
	servers := loadRealServers(t)
	t.Logf("validating %d real rows", len(servers))

	for i, s := range servers {
		label := fmt.Sprintf("row %d (%s)", i+2, s.Model) // +2: header is row 1, data is 1-indexed from row 2
		if strings.TrimSpace(s.Model) == "" {
			t.Errorf("%s: empty Model", label)
		}
		if strings.TrimSpace(s.Location) == "" {
			t.Errorf("%s: empty Location", label)
		}
		if s.Price <= 0 {
			t.Errorf("%s: non-positive Price %v", label, s.Price)
		}

		fam := model.RAMFamily(s.RAM)
		if oracle := oracleRAMFamily(s.RAM); fam != oracle {
			t.Errorf("%s: RAM family mismatch: model.RAMFamily=%q oracle=%q (raw %q)", label, fam, oracle, s.RAM)
		}
		if !strings.HasSuffix(fam, "GB") || len(fam) < 3 {
			t.Errorf("%s: RAM family %q doesn't look like \"<digits>GB\" (raw %q)", label, fam, s.RAM)
		}

		totalGB, diskType, err := model.ParseHDD(s.HDD)
		if err != nil {
			t.Errorf("%s: HDD %q failed to parse: %v", label, s.HDD, err)
			continue
		}
		if oracleType := oracleDiskType(s.HDD); diskType != oracleType {
			t.Errorf("%s: disk type mismatch: model.ParseHDD=%q oracle=%q (raw %q)", label, diskType, oracleType, s.HDD)
		}
		if diskType != "SAS" && diskType != "SATA" && diskType != "SSD" {
			t.Errorf("%s: disk type %q is not one of SAS/SATA/SSD (raw %q)", label, diskType, s.HDD)
		}
		if oracleTotal, err := oracleTotalStorageGB(s.HDD); err != nil {
			t.Errorf("%s: oracle couldn't parse HDD %q: %v", label, s.HDD, err)
		} else if totalGB != oracleTotal {
			t.Errorf("%s: total storage mismatch: model.ParseHDD=%dGB oracle=%dGB (raw %q)", label, totalGB, oracleTotal, s.HDD)
		}
		if totalGB <= 0 {
			t.Errorf("%s: non-positive total storage %dGB (raw %q)", label, totalGB, s.HDD)
		}
	}
}

// TestRealData_FilterByRAM_MatchesOracle queries ram= for every RAM family
// actually present in the real data, plus every allowed-but-absent family
// (the assignment's spec sheet lists more RAM sizes than this particular
// dataset happens to contain - e.g. 12GB/24GB/48GB never appear), and
// checks meta.total against an independent per-row oracle count for each.
// A family with zero real matches must still return 200 with total=0, not
// an error - "valid filter value, no results" and "invalid filter value"
// are different outcomes and the API must not conflate them.
func TestRealData_FilterByRAM_MatchesOracle(t *testing.T) {
	servers := loadRealServers(t)
	ts := newRealDataServer(t, servers)

	present := map[string]bool{}
	for _, s := range servers {
		present[oracleRAMFamily(s.RAM)] = true
	}
	// Full allowed list per handlers.defaultAllowedRAM - includes families
	// with zero real matches, deliberately, to exercise that path too.
	allowed := []string{"2GB", "4GB", "8GB", "12GB", "16GB", "24GB", "32GB", "48GB", "64GB", "96GB"}

	for _, ram := range allowed {
		t.Run(ram, func(t *testing.T) {
			want := 0
			for _, s := range servers {
				if oracleRAMFamily(s.RAM) == ram {
					want++
				}
			}
			if present[ram] == (want == 0) {
				t.Fatalf("test setup inconsistency for %s: present=%v want=%d", ram, present[ram], want)
			}

			page := getServersPage(t, ts, "?ram="+strings.ReplaceAll(ram, " ", "%20")+"&limit=1")
			if page.Meta.Total != want {
				t.Errorf("ram=%s: expected meta.total=%d, got %d", ram, want, page.Meta.Total)
			}
		})
	}
}

// TestRealData_FilterByDiskType_MatchesOracle checks disk_type=SAS/SATA/SSD
// against the real data with a full exact-model-set comparison (not just a
// count), for the two disk types with a real, checkable number of results.
func TestRealData_FilterByDiskType_MatchesOracle(t *testing.T) {
	servers := loadRealServers(t)
	ts := newRealDataServer(t, servers)

	for _, diskType := range []string{"SAS", "SATA", "SSD"} {
		t.Run(diskType, func(t *testing.T) {
			var want []model.Server
			for _, s := range servers {
				if oracleDiskType(s.HDD) == diskType {
					want = append(want, s)
				}
			}
			got := fetchAllServers(t, ts, "?disk_type="+diskType)
			assertSameModels(t, "disk_type="+diskType, got, want)
		})
	}
}

// TestRealData_FilterByStorageRange_MatchesOracle picks range boundaries
// derived from the real data's own min/max/median total storage (not
// guessed round numbers), so the test is checking real inclusive-boundary
// behavior against real values rather than an arbitrary range that happens
// to land cleanly.
func TestRealData_FilterByStorageRange_MatchesOracle(t *testing.T) {
	servers := loadRealServers(t)
	ts := newRealDataServer(t, servers)

	totals := make([]int, len(servers))
	for i, s := range servers {
		total, err := oracleTotalStorageGB(s.HDD)
		if err != nil {
			t.Fatalf("oracle failed on row %d: %v", i, err)
		}
		totals[i] = total
	}
	sorted := append([]int(nil), totals...)
	sort.Ints(sorted)
	min, median, max := sorted[0], sorted[len(sorted)/2], sorted[len(sorted)-1]
	t.Logf("real data total storage: min=%dGB median=%dGB max=%dGB", min, median, max)

	toParam := func(gb int) string {
		if gb%1024 == 0 {
			return fmt.Sprintf("%dTB", gb/1024)
		}
		return fmt.Sprintf("%dGB", gb)
	}

	cases := []struct {
		name       string
		storageMin int
		hasMin     bool
		storageMax int
		hasMax     bool
	}{
		{name: "at or above median", storageMin: median, hasMin: true},
		{name: "at or below median", storageMax: median, hasMax: true},
		{name: "between min and median inclusive", storageMin: min, hasMin: true, storageMax: median, hasMax: true},
		{name: "everything (min..max inclusive)", storageMin: min, hasMin: true, storageMax: max, hasMax: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := 0
			for _, total := range totals {
				if tc.hasMin && total < tc.storageMin {
					continue
				}
				if tc.hasMax && total > tc.storageMax {
					continue
				}
				want++
			}

			query := ""
			if tc.hasMin {
				query += "storage_min=" + toParam(tc.storageMin)
			}
			if tc.hasMax {
				if query != "" {
					query += "&"
				}
				query += "storage_max=" + toParam(tc.storageMax)
			}
			page := getServersPage(t, ts, "?"+query+"&limit=1")
			if page.Meta.Total != want {
				t.Errorf("%s (%s): expected meta.total=%d, got %d", tc.name, query, want, page.Meta.Total)
			}
		})
	}

	if len(cases) == 0 || max == 0 {
		t.Fatal("test setup failure: no storage spread found in real data")
	}
}

// TestRealData_CombinedFilters_MatchesOracle checks AND-composition across
// three dimensions at once against the real data - the fixture-based suite
// already proves AND semantics generically, this proves it doesn't break
// down under the real data's actual value combinations (e.g. does every
// location actually have SATA drives in every RAM size, or does a
// real-world gap in the data expose an off-by-one in how filters combine).
func TestRealData_CombinedFilters_MatchesOracle(t *testing.T) {
	servers := loadRealServers(t)
	ts := newRealDataServer(t, servers)

	const ram, diskType, locationSubstr = "64GB", "SATA", "Amsterdam"
	var want []model.Server
	for _, s := range servers {
		if oracleRAMFamily(s.RAM) != ram {
			continue
		}
		if oracleDiskType(s.HDD) != diskType {
			continue
		}
		if !strings.Contains(strings.ToLower(s.Location), strings.ToLower(locationSubstr)) {
			continue
		}
		want = append(want, s)
	}
	if len(want) == 0 {
		t.Fatal("test setup failure: real data has zero rows matching ram=64GB&disk_type=SATA&location=Amsterdam - pick different fixed values")
	}

	got := fetchAllServers(t, ts, fmt.Sprintf("?ram=%s&disk_type=%s&location=%s", ram, diskType, locationSubstr))
	assertSameModels(t, "combined ram+disk_type+location", got, want)
}

// TestRealData_NoFilter_ReturnsFullCatalogCount locks in that meta.total
// with no filters at all equals the exact number of rows the real file
// parsed to - the simplest possible check, and also the one that would
// catch an off-by-one in CSV row handling (e.g. accidentally counting or
// dropping the header) that per-filter checks wouldn't necessarily surface.
func TestRealData_NoFilter_ReturnsFullCatalogCount(t *testing.T) {
	servers := loadRealServers(t)
	ts := newRealDataServer(t, servers)

	page := getServersPage(t, ts, "?limit=1")
	if page.Meta.Total != len(servers) {
		t.Fatalf("expected meta.total=%d (full real catalog), got %d", len(servers), page.Meta.Total)
	}
}
