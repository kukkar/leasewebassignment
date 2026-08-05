package service

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/sahil/leasewebassignment/internal/model"
)

type testRepo struct{ servers []model.Server }

func (m *testRepo) ListServers(ctx context.Context, filter model.ServerFilter) ([]model.Server, error) {
	return m.servers, nil
}
func (m *testRepo) ReplaceServers(ctx context.Context, servers []model.Server) error {
	m.servers = servers
	return nil
}
func (m *testRepo) SaveUpload(ctx context.Context, filename string, content []byte) (string, error) {
	return "", nil
}

type testParser struct{}

func (p *testParser) Parse(r io.Reader) ([]model.Server, error) {
	return []model.Server{{Model: "injected", RAM: "8GB", HDD: "1x1TBSATA2", Location: "Y", Price: 2.0}}, nil
}

func TestServerServiceWithParserInjection(t *testing.T) {
	repo := &testRepo{}
	parser := &testParser{}
	svc := NewServerServiceWithParser(repo, parser)

	if err := svc.UploadServerData(context.Background(), "u.csv", bytes.NewBufferString("ignored")); err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if len(repo.servers) != 1 || repo.servers[0].Model != "injected" {
		t.Fatalf("expected injected servers in repo, got %+v", repo.servers)
	}
}
