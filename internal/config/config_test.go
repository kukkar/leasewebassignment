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

// TestLoadYAMLConfig_AdminAndLogging locks in that the auth and logging
// fields actually deserialize - these gate authentication and log
// verbosity respectively, so a silent field-name typo here would be a real
// production issue, not just a cosmetic one.
func TestLoadYAMLConfig_AdminAndLogging(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := `server:
  host: 127.0.0.1
  port: 8080
  timeout: 30
app:
  data_file: data/servers.csv
  upload_dir: data/uploads
  admin_token: super-secret
  logging:
    level: debug
  allowed_ram: ["16GB", "32GB"]
  allowed_disk_types: ["SSD"]
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.App.AdminToken != "super-secret" {
		t.Fatalf("expected admin_token %q, got %q", "super-secret", cfg.App.AdminToken)
	}
	if cfg.App.Logging.Level != "debug" {
		t.Fatalf("expected logging.level %q, got %q", "debug", cfg.App.Logging.Level)
	}
	if len(cfg.App.AllowedRAM) != 2 || cfg.App.AllowedRAM[0] != "16GB" {
		t.Fatalf("expected allowed_ram [16GB 32GB], got %v", cfg.App.AllowedRAM)
	}
	if len(cfg.App.AllowedDiskTypes) != 1 || cfg.App.AllowedDiskTypes[0] != "SSD" {
		t.Fatalf("expected allowed_disk_types [SSD], got %v", cfg.App.AllowedDiskTypes)
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
