package store

import (
    "testing"
)

func TestNewRepositoryFactoryKinds(t *testing.T) {
    cfg := RepositoryConfig{UploadDir: "data/uploads", Kind: "memory"}
    r := NewRepository(cfg)
    if _, ok := r.(*InMemoryRepository); !ok {
        t.Fatalf("expected InMemoryRepository for kind=memory, got %T", r)
    }

    cfg.Kind = "file"
    r2 := NewRepository(cfg)
    if _, ok := r2.(*InMemoryRepository); !ok {
        t.Fatalf("expected InMemoryRepository (placeholder) for kind=file, got %T", r2)
    }
}
