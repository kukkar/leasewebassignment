package store

import (
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sahil/leasewebassignment/internal/model"
)

func TestListServersFilterFromCSV(t *testing.T) {
	// Try several candidate locations for the sample CSV.
	candidates := []string{
		filepath.Join("data", "servers.csv"),
		filepath.Join("..", "data", "servers.csv"),
		filepath.Join("..", "..", "data", "servers.csv"),
		"/Users/sahil/GolandProjects/leasewebassignment/data/servers.csv",
	}
	var data []byte
	var err error
	var found string
	for _, p := range candidates {
		data, err = os.ReadFile(p)
		if err == nil {
			found = p
			break
		}
	}
	if found == "" {
		t.Skipf("missing sample data file in candidates: %v", candidates)
	}

	r := csv.NewReader(strings.NewReader(string(data)))
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read csv: %v", err)
	}
	if len(rows) < 1 {
		t.Fatal("empty csv")
	}
	// headers may be present; try to detect header row
	start := 0
	headers := rows[0]
	if len(headers) >= 5 && strings.EqualFold(strings.TrimSpace(headers[0]), "Model") {
		start = 1
	}

	parsed := make([]model.Server, 0, len(rows)-start)
	for i, row := range rows[start:] {
		if len(row) < 5 {
			continue
		}
		price, err := model.ParsePrice(row[4])
		if err != nil {
			t.Fatalf("parse price row %d: %v", i+start+1, err)
		}
		parsed = append(parsed, model.Server{
			Model:    strings.TrimSpace(row[0]),
			RAM:      strings.TrimSpace(row[1]),
			HDD:      strings.TrimSpace(row[2]),
			Location: strings.TrimSpace(row[3]),
			Price:    price,
		})
	}

	repo := NewInMemoryRepository(filepath.Join(t.TempDir(), "uploads"))
	if err := repo.ReplaceServers(context.Background(), parsed); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		filter  model.ServerFilter
		matcher func(s model.Server) bool
	}{
		{"filter ram 4GB", model.ServerFilter{RAM: "4GB"}, func(s model.Server) bool { return strings.Contains(strings.ToLower(s.RAM), "4gb") }},
		{"filter location Washington", model.ServerFilter{Location: "Washington"}, func(s model.Server) bool { return strings.Contains(strings.ToLower(s.Location), "washington") }},
		{"filter hdd 4x1TB", model.ServerFilter{HDD: "4x1TB"}, func(s model.Server) bool { return strings.Contains(strings.ToLower(s.HDD), "4x1tb") }},
	}

	for _, tc := range tests {
		// compute expected from parsed data
		want := 0
		for _, s := range parsed {
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
	}
}
