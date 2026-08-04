package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJSONConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{
        "server": {"host": "127.0.0.1", "port": 8080, "timeout": 30},
        "app": {"data_file": "data/servers.csv", "upload_dir": "data/uploads"}
    }`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Fatalf("expected host 127.0.0.1, got %q", cfg.Server.Host)
	}
}

func TestLoadYAMLConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := `server:
  host: 127.0.0.1
  port: 8080
  timeout: 30
app:
  data_file: data/servers.csv
  upload_dir: data/uploads
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != 8080 {
		t.Fatalf("expected port 8080, got %d", cfg.Server.Port)
	}
}

func TestLoadMissingRequiredValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "badconfig.json")
	content := `{"server": {"host": "127.0.0.1"}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected load error")
	}
}
