package service

import (
	"bytes"
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/sahil/leasewebassignment/internal/model"
)

func TestNewLoggingService_NilLoggerReturnsUnwrapped(t *testing.T) {
	repo := &testRepo{}
	svc := NewServerService(repo)
	wrapped := NewLoggingService(svc, nil)
	if wrapped != Service(svc) {
		t.Fatal("expected NewLoggingService(svc, nil) to return svc unwrapped")
	}
}

func TestLoggingService_LogsUploadOutcome(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core).Sugar()

	repo := &testRepo{}
	svc := NewLoggingService(NewServerService(repo), logger)

	csvData := "Model,RAM,HDD,Location,Price\nDell R210,16GB,2x2TBSATA2,AmsterdamAMS-01,49.99\n"
	if err := svc.UploadServerData(context.Background(), "upload.csv", bytes.NewBufferString(csvData)); err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	entries := logs.FilterMessage("UploadServerData").All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 UploadServerData log entry, got %d", len(entries))
	}
	if entries[0].Level != zapcore.InfoLevel {
		t.Fatalf("expected successful upload logged at Info, got %v", entries[0].Level)
	}
}

func TestLoggingService_LogsFailureAtErrorLevel(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core).Sugar()

	svc := NewLoggingService(NewServerService(&testRepo{}), logger)
	if err := svc.LoadServerData(context.Background(), "/nonexistent/path.csv"); err == nil {
		t.Fatal("expected LoadServerData to fail for a nonexistent path")
	}

	entries := logs.FilterMessage("LoadServerData").All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 LoadServerData log entry, got %d", len(entries))
	}
	if entries[0].Level != zapcore.ErrorLevel {
		t.Fatalf("expected failed load logged at Error, got %v", entries[0].Level)
	}
}

// TestLoggingService_GetServersPassesThroughWithoutLogging locks in the
// design decision in logging.go: GetServers is deliberately NOT decorated
// (it's promoted through the embedded Service unchanged) because it's
// always called via HTTP and already fully covered by the access-log
// middleware - logging it here too would just duplicate that.
func TestLoggingService_GetServersPassesThroughWithoutLogging(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core).Sugar()

	want := []model.Server{{Model: "Dell R210"}}
	repo := &testRepo{servers: want}
	svc := NewLoggingService(NewServerService(repo), logger)

	got, err := svc.GetServers(context.Background(), model.ServerFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Model != "Dell R210" {
		t.Fatalf("expected passthrough result %+v, got %+v", want, got)
	}
	if logs.Len() != 0 {
		t.Fatalf("expected GetServers to produce no log entries, got %d: %+v", logs.Len(), logs.All())
	}
}
