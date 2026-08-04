package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sahil/leasewebassignment/internal/model"
)

func TestListServersFilter(t *testing.T) {
	repo := NewInMemoryRepository(filepath.Join(t.TempDir(), "uploads"))
	repo.servers = []model.Server{
		{Model: "Dell R210", RAM: "16GB", HDD: "2x2TBSATA2", Location: "AmsterdamAMS-01", Price: 49.99},
		{Model: "HP DL180", RAM: "32GB", HDD: "8x2TBSATA2", Location: "AmsterdamAMS-01", Price: 119.00},
	}

	filter := model.ServerFilter{RAM: "16GB"}
	got, err := repo.ListServers(context.Background(), filter)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Model != "Dell R210" {
		t.Fatalf("expected one Dell R210 result, got %+v", got)
	}
}

func TestReplaceServers(t *testing.T) {
	repo := NewInMemoryRepository(filepath.Join(t.TempDir(), "uploads"))
	repo.servers = []model.Server{{Model: "old"}}
	newServers := []model.Server{{Model: "new"}}
	if err := repo.ReplaceServers(context.Background(), newServers); err != nil {
		t.Fatal(err)
	}
	got, _ := repo.ListServers(context.Background(), model.ServerFilter{})
	if len(got) != 1 || got[0].Model != "new" {
		t.Fatalf("replace failed, got %+v", got)
	}
}

func TestSaveUpload(t *testing.T) {
	temp := t.TempDir()
	source := filepath.Join(temp, "upload.csv")
	if err := os.WriteFile(source, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	repo := NewInMemoryRepository(filepath.Join(temp, "uploads"))
	if err := repo.SaveUpload(context.Background(), source); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(temp, "uploads", "upload.csv")); err != nil {
		t.Fatalf("expected saved upload, got %v", err)
	}
}
