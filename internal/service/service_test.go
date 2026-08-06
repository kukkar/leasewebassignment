package service

import (
	"bytes"
	"context"
	"testing"

	"github.com/sahil/leasewebassignment/internal/model"
)

type mockRepo struct {
	servers []model.Server
}

func (m *mockRepo) ListServers(ctx context.Context, filter model.ServerFilter) ([]model.Server, error) {
	return m.servers, nil
}

func (m *mockRepo) ReplaceServers(ctx context.Context, servers []model.Server) error {
	m.servers = servers
	return nil
}

func (m *mockRepo) SaveUpload(ctx context.Context, filename string, content []byte) (string, error) {
	return "", nil
}

func TestUploadServerData(t *testing.T) {
	csvData := "Model,RAM,HDD,Location,Price\nDell R210,16GB,2x2TBSATA2,AmsterdamAMS-01,€49.99\n"
	repo := &mockRepo{}
	service := NewServerService(repo)

	if err := service.UploadServerData(context.Background(), "upload.csv", bytes.NewBufferString(csvData)); err != nil {
		t.Fatal(err)
	}
	if len(repo.servers) != 1 || repo.servers[0].Model != "Dell R210" {
		t.Fatalf("expected uploaded server in repo, got %+v", repo.servers)
	}
}

func TestUploadServerDataWithBareQuote(t *testing.T) {
	csvData := "Model,RAM,HDD,Location,Price\nDell R210,16GB,2x2TBSATA2,AmsterdamAMS-01\"test,€49.99\n"
	repo := &mockRepo{}
	service := NewServerService(repo)

	if err := service.UploadServerData(context.Background(), "upload.csv", bytes.NewBufferString(csvData)); err != nil {
		t.Fatalf("expected upload to succeed with bare quote, got %v", err)
	}
	if len(repo.servers) != 1 || repo.servers[0].Model != "Dell R210" {
		t.Fatalf("expected uploaded server in repo, got %+v", repo.servers)
	}
}
