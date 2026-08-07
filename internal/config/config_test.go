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

// TestLoadEnvOverride_AdminToken locks in that ADMIN_TOKEN always wins over
// app.admin_token in the file - the whole point is that a real credential
// should come from the environment, never a committed config file, so the
// override direction (env beats file, not the reverse) matters as much as
// the override existing at all.
func TestLoadEnvOverride_AdminToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := `server:
  host: 127.0.0.1
  port: 8080
  timeout: 30
app:
  data_file: data/servers.csv
  upload_dir: data/uploads
  admin_token: from-file
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADMIN_TOKEN", "from-env")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.App.AdminToken != "from-env" {
		t.Fatalf("expected ADMIN_TOKEN to override the file, got %q", cfg.App.AdminToken)
	}
}

// TestLoadEnvOverride_Port mirrors the admin-token override for the other
// field a PaaS deployment (Render, Heroku, ...) needs to control at deploy
// time rather than through a committed file.
func TestLoadEnvOverride_Port(t *testing.T) {
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
	t.Setenv("PORT", "9090")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != 9090 {
		t.Fatalf("expected PORT to override the file, got %d", cfg.Server.Port)
	}
}

// TestLoadEnvOverride_InvalidPort_Rejected locks in that a malformed PORT
// fails Load loudly instead of silently falling back to the file's port -
// a deploy platform that injects a PORT the app then ignores would bind the
// wrong port and fail its health check anyway, just far less legibly.
func TestLoadEnvOverride_InvalidPort_Rejected(t *testing.T) {
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
	t.Setenv("PORT", "not-a-number")

	if _, err := Load(path); err == nil {
		t.Fatal("expected load error for invalid PORT environment variable")
	}
}

// TestLoadZeroTimeout_Rejected locks in that server.timeout is required and
// must be positive - net/http treats a zero ReadTimeout/WriteTimeout as "no
// timeout at all", so a missing/zero value must fail loudly at load time
// rather than silently produce an unbounded-request-duration server.
func TestLoadZeroTimeout_Rejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{
        "server": {"host": "127.0.0.1", "port": 8080},
        "app": {"data_file": "data/servers.csv", "upload_dir": "data/uploads"}
    }`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected load error for missing/zero server.timeout")
	}
}
