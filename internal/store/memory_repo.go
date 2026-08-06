package store

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/sahil/leasewebassignment/internal/model"
)

type memoryRepository struct {
	mu        sync.RWMutex
	servers   []indexedServer
	uploadDir string
}

func newMemoryRepository(uploadDir string) Repository {
	if uploadDir == "" {
		uploadDir = filepath.Join("data", "uploads")
	}
	return &memoryRepository{uploadDir: uploadDir}
}

func (r *memoryRepository) ListServers(ctx context.Context, filter model.ServerFilter) ([]model.Server, error) {
	if r == nil {
		return nil, &StoreError{Op: "list", Err: ErrRepositoryUninitialized}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	pf := prepareFilter(filter)
	result := make([]model.Server, 0, len(r.servers))
	for _, idx := range r.servers {
		if matchesFilter(idx, pf) {
			result = append(result, idx.server)
		}
	}
	return result, nil
}

// ReplaceServers is the only place the catalog changes, so it's the only
// place indexedServer's derived fields get computed - once per server, per
// upload, not once per server per request.
func (r *memoryRepository) ReplaceServers(ctx context.Context, servers []model.Server) error {
	if r == nil {
		return &StoreError{Op: "replace", Err: ErrRepositoryUninitialized}
	}
	indexed := make([]indexedServer, len(servers))
	for i, s := range servers {
		indexed[i] = newIndexedServer(s)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.servers = indexed
	return nil
}

func (r *memoryRepository) SaveUpload(ctx context.Context, filename string, content []byte) (string, error) {
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
	defer func() { _ = tmpFile.Close() }()
	if _, err = tmpFile.Write(content); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", &StoreError{Op: "save_upload", Err: err}
	}
	if err = tmpFile.Sync(); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", &StoreError{Op: "save_upload", Err: err}
	}
	finalPath := filepath.Join(r.uploadDir, finalName)
	if err = os.Rename(tmpFile.Name(), finalPath); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", &StoreError{Op: "save_upload", Err: err}
	}
	return finalPath, nil
}
