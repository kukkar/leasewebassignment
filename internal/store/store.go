package store

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/sahil/leasewebassignment/internal/model"
)

type Repository interface {
	ListServers(ctx context.Context, filter model.ServerFilter) ([]model.Server, error)
	ReplaceServers(ctx context.Context, servers []model.Server) error
	SaveUpload(ctx context.Context, sourcePath string) error
}

type InMemoryRepository struct {
	mu        sync.RWMutex
	servers   []model.Server
	uploadDir string
}

func NewInMemoryRepository(uploadDir string) *InMemoryRepository {
	if uploadDir == "" {
		uploadDir = filepath.Join("data", "uploads")
	}
	return &InMemoryRepository{uploadDir: uploadDir}
}

func (r *InMemoryRepository) ListServers(ctx context.Context, filter model.ServerFilter) ([]model.Server, error) {
	if r == nil {
		return nil, &StoreError{Op: "list", Err: ErrRepositoryUninitialized}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]model.Server, 0, len(r.servers))
	for _, s := range r.servers {
		if filter.Model != "" && filter.Model != s.Model {
			continue
		}
		if filter.RAM != "" && filter.RAM != s.RAM {
			continue
		}
		if filter.HDD != "" && filter.HDD != s.HDD {
			continue
		}
		if filter.Location != "" && filter.Location != s.Location {
			continue
		}
		if filter.PriceMin != nil && s.Price < *filter.PriceMin {
			continue
		}
		if filter.PriceMax != nil && s.Price > *filter.PriceMax {
			continue
		}
		result = append(result, s)
	}
	return result, nil
}

func (r *InMemoryRepository) ReplaceServers(ctx context.Context, servers []model.Server) error {
	if r == nil {
		return &StoreError{Op: "replace", Err: ErrRepositoryUninitialized}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.servers = append([]model.Server(nil), servers...)
	return nil
}

func (r *InMemoryRepository) SaveUpload(ctx context.Context, sourcePath string) error {
	if r == nil {
		return &StoreError{Op: "save_upload", Err: ErrRepositoryUninitialized}
	}
	if sourcePath == "" {
		return &StoreError{Op: "save_upload", Err: ErrSourcePathRequired}
	}
	if err := os.MkdirAll(r.uploadDir, 0o755); err != nil {
		return &StoreError{Op: "save_upload", Err: err}
	}
	dest := filepath.Join(r.uploadDir, filepath.Base(sourcePath))
	input, err := os.ReadFile(sourcePath)
	if err != nil {
		return &StoreError{Op: "save_upload", Err: err}
	}
	if err := os.WriteFile(dest, input, 0o644); err != nil {
		return &StoreError{Op: "save_upload", Err: err}
	}
	return nil
}
