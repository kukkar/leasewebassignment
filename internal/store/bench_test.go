package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/sahil/leasewebassignment/internal/model"
)

// generateServers builds a synthetic catalog for benchmarking - large enough
// (50k rows, ~100x the real dataset) that any accidental reintroduction of
// per-request parsing in matchesFilter would show up clearly in ns/op.
func generateServers(n int) []model.Server {
	ramValues := []string{"16GBDDR3", "32GBDDR3", "64GBDDR4", "128GBDDR4"}
	hddValues := []string{"2x2TBSATA2", "8x300GBSAS", "4x480GBSSD", "24x1TBSATA2"}
	locations := []string{"AmsterdamAMS-01", "Washington D.C.WDC-01", "DallasDAL-10", "SingaporeSIN-11"}

	servers := make([]model.Server, n)
	for i := 0; i < n; i++ {
		servers[i] = model.Server{
			Model:    fmt.Sprintf("Server-%d", i),
			RAM:      ramValues[i%len(ramValues)],
			HDD:      hddValues[i%len(hddValues)],
			Location: locations[i%len(locations)],
			Price:    float64(i%1000) + 0.99,
		}
	}
	return servers
}

// BenchmarkListServers measures the read path, which should stay cheap
// regardless of catalog size since it only compares precomputed fields
// (see indexedServer in store.go) rather than re-parsing HDD/RAM strings
// on every call - this benchmark guards against that regressing.
func BenchmarkListServers(b *testing.B) {
	repo := NewRepository(RepositoryConfig{UploadDir: b.TempDir()})
	if err := repo.ReplaceServers(context.Background(), generateServers(50_000)); err != nil {
		b.Fatal(err)
	}
	filter := model.ServerFilter{RAM: []string{"16GB", "32GB"}, DiskType: "SATA"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := repo.ListServers(context.Background(), filter); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReplaceServers(b *testing.B) {
	repo := NewRepository(RepositoryConfig{UploadDir: b.TempDir()})
	servers := generateServers(50_000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := repo.ReplaceServers(context.Background(), servers); err != nil {
			b.Fatal(err)
		}
	}
}
