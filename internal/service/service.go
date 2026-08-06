// Package service orchestrates parsing and storing uploaded server catalogs
// and answering filtered catalog queries. It has no logging concerns of its
// own - see logging.go's NewLoggingService for observability, kept as a
// separate decorator so business logic here stays free of that concern.
package service

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sahil/leasewebassignment/internal/model"
	"github.com/sahil/leasewebassignment/internal/store"
)

type Service interface {
	GetServers(ctx context.Context, filter model.ServerFilter) ([]model.Server, error)
	UploadServerData(ctx context.Context, filename string, reader io.Reader) error
	LoadServerData(ctx context.Context, path string) error
}

type ServerService struct {
	repo   store.Repository
	parser Parser
}

// NewServerService constructs a service with the default CSV parser.
func NewServerService(repo store.Repository) *ServerService {
	return NewServerServiceWithParser(repo, NewCSVParser())
}

// NewServerServiceWithParser allows injecting a custom parser strategy.
func NewServerServiceWithParser(repo store.Repository, parser Parser) *ServerService {
	return &ServerService{repo: repo, parser: parser}
}

func (s *ServerService) GetServers(ctx context.Context, filter model.ServerFilter) ([]model.Server, error) {
	return s.repo.ListServers(ctx, filter)
}

func (s *ServerService) UploadServerData(ctx context.Context, filename string, reader io.Reader) error {
	content, err := io.ReadAll(reader)
	if err != nil {
		return &ServiceError{Op: "upload", Err: err}
	}
	servers, err := s.parser.Parse(strings.NewReader(string(content)))
	if err != nil {
		return &ServiceError{Op: "upload", Err: err}
	}
	if _, err := s.repo.SaveUpload(ctx, filename, content); err != nil {
		return &ServiceError{Op: "upload", Err: err}
	}
	if err := s.repo.ReplaceServers(ctx, servers); err != nil {
		return &ServiceError{Op: "upload", Err: err}
	}
	return nil
}

func (s *ServerService) LoadServerData(ctx context.Context, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return &ServiceError{Op: "load", Err: err}
	}
	defer func() { _ = f.Close() }()
	return s.UploadServerData(ctx, filepath.Base(path), f)
}
