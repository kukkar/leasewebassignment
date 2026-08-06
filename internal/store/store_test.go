package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sahil/leasewebassignment/internal/model"
)

func TestListServersFilter(t *testing.T) {
	repo := NewRepository(RepositoryConfig{UploadDir: filepath.Join(t.TempDir(), "uploads")})
	seed := []model.Server{
		{Model: "Dell R210", RAM: "16GB", HDD: "2x2TBSATA2", Location: "AmsterdamAMS-01", Price: 49.99},
		{Model: "HP DL180", RAM: "32GB", HDD: "8x2TBSATA2", Location: "AmsterdamAMS-01", Price: 119.00},
	}
	if err := repo.ReplaceServers(context.Background(), seed); err != nil {
		t.Fatal(err)
	}

	filter := model.ServerFilter{RAM: []string{"16GB"}}
	got, err := repo.ListServers(context.Background(), filter)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Model != "Dell R210" {
		t.Fatalf("expected one Dell R210 result, got %+v", got)
	}
}

func TestReplaceServers(t *testing.T) {
	repo := NewRepository(RepositoryConfig{UploadDir: filepath.Join(t.TempDir(), "uploads")})
	if err := repo.ReplaceServers(context.Background(), []model.Server{{Model: "old"}}); err != nil {
		t.Fatal(err)
	}
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
	content := []byte("data")
	repo := NewRepository(RepositoryConfig{UploadDir: filepath.Join(temp, "uploads")})
	path, err := repo.SaveUpload(context.Background(), "upload.csv", content)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := filepath.Base(path), "upload.csv"; got != want {
		t.Fatalf("expected saved filename %q, got %q", want, got)
	}
	if got, err := os.ReadFile(path); err != nil {
		t.Fatal(err)
	} else if string(got) != string(content) {
		t.Fatalf("expected content %q, got %q", content, got)
	}
}
