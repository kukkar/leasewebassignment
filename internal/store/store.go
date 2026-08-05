package store

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sahil/leasewebassignment/internal/model"
)

type Repository interface {
	ListServers(ctx context.Context, filter model.ServerFilter) ([]model.Server, error)
	ReplaceServers(ctx context.Context, servers []model.Server) error
	SaveUpload(ctx context.Context, filename string, content []byte) (string, error)
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
		if filter.Model != "" && !stringMatches(s.Model, filter.Model) {
			continue
		}
		if filter.RAM != "" && !stringMatches(s.RAM, filter.RAM) {
			continue
		}
		if filter.HDD != "" && !stringMatches(s.HDD, filter.HDD) {
			continue
		}
		// parse HDD to support storage range and disk-type filtering
		totalGB := 0
		diskType := ""
		if s.HDD != "" {
			if tb, dt, err := model.ParseHDD(s.HDD); err == nil {
				totalGB = tb
				diskType = dt
			}
		}
		if filter.Location != "" && !stringMatches(s.Location, filter.Location) {
			continue
		}
		if filter.PriceMin != nil && s.Price < *filter.PriceMin {
			continue
		}
		if filter.PriceMax != nil && s.Price > *filter.PriceMax {
			continue
		}
		if filter.DiskType != "" && !strings.EqualFold(filter.DiskType, diskType) {
			continue
		}
		if filter.StorageMin != nil && totalGB < *filter.StorageMin {
			continue
		}
		if filter.StorageMax != nil && totalGB > *filter.StorageMax {
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

func stringMatches(value, filter string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	filter = strings.ToLower(strings.TrimSpace(filter))
	return strings.Contains(value, filter)
}

func (r *InMemoryRepository) SaveUpload(ctx context.Context, filename string, content []byte) (string, error) {
	if r == nil {
		return "", &StoreError{Op: "save_upload", Err: ErrRepositoryUninitialized}
	}
	if filename == "" {
		return "", &StoreError{Op: "save_upload", Err: ErrSourcePathRequired}
	}
	if err := os.MkdirAll(r.uploadDir, 0o755); err != nil {
		return "", &StoreError{Op: "save_upload", Err: err}
	}
	finalName := filepath.Base(filename)
	tmpFile, err := os.CreateTemp(r.uploadDir, finalName+"-*.tmp")
	if err != nil {
		return "", &StoreError{Op: "save_upload", Err: err}
	}
	defer tmpFile.Close()
	if _, err = tmpFile.Write(content); err != nil {
		os.Remove(tmpFile.Name())
		return "", &StoreError{Op: "save_upload", Err: err}
	}
	if err = tmpFile.Sync(); err != nil {
		os.Remove(tmpFile.Name())
		return "", &StoreError{Op: "save_upload", Err: err}
	}
	finalPath := filepath.Join(r.uploadDir, finalName)
	if err = os.Rename(tmpFile.Name(), finalPath); err != nil {
		os.Remove(tmpFile.Name())
		return "", &StoreError{Op: "save_upload", Err: err}
	}
	return finalPath, nil
}
