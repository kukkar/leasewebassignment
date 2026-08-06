package store

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sahil/leasewebassignment/internal/model"
	"github.com/sahil/leasewebassignment/internal/testutil"
)

// TestListServersFilterFromFixture exercises the repository filter logic
// against the shared deterministic fixture (see internal/testutil) rather
// than a data file that can change independently of the test's expectations.
func TestListServersFilterFromFixture(t *testing.T) {
	repo := NewRepository(RepositoryConfig{UploadDir: filepath.Join(t.TempDir(), "uploads")})
	if err := repo.ReplaceServers(context.Background(), testutil.SampleServers); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		filter  model.ServerFilter
		matcher func(s model.Server) bool
	}{
		// RAM is a checkbox filter: matching is by RAM family (case-insensitive,
		// suffix-tolerant), not substring - "4GB" must not also match "64GBDDR3".
		{"filter ram 4GB", model.ServerFilter{RAM: []string{"4GB"}}, func(s model.Server) bool { return model.RAMFamily(s.RAM) == "4GB" }},
		{"filter location Washington", model.ServerFilter{Location: "Washington"}, func(s model.Server) bool {
			return strings.Contains(strings.ToLower(s.Location), "washington")
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			want := 0
			for _, s := range testutil.SampleServers {
				if tc.matcher(s) {
					want++
				}
			}

			got, err := repo.ListServers(context.Background(), tc.filter)
			if err != nil {
				t.Fatalf("%s failed: %v", tc.name, err)
			}
			if len(got) != want {
				t.Fatalf("%s expected %d results, got %d", tc.name, want, len(got))
			}
		})
	}
}

// TestListServersFilterDiskType locks in the disk_type acceptance criterion
// against the assignment's SAS/SATA/SSD dropdown values. Expected counts are
// hardcoded from the fixture's documented disk types (see internal/testutil)
// rather than recomputed with model.ParseHDD, so a regression in ParseHDD's
// normalization can't silently pass by producing a matching wrong answer.
func TestListServersFilterDiskType(t *testing.T) {
	repo := NewRepository(RepositoryConfig{UploadDir: filepath.Join(t.TempDir(), "uploads")})
	if err := repo.ReplaceServers(context.Background(), testutil.SampleServers); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		diskType string
		want     int
	}{
		{"SATA", 7}, // fixture rows labeled SATA2 must normalize to the SATA family
		{"SSD", 2},
		{"SAS", 1},
	}
	for _, tc := range cases {
		t.Run(tc.diskType, func(t *testing.T) {
			got, err := repo.ListServers(context.Background(), model.ServerFilter{DiskType: tc.diskType})
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != tc.want {
				t.Fatalf("disk_type=%s: expected %d results, got %d", tc.diskType, tc.want, len(got))
			}
		})
	}
}
