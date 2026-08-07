// Package config loads and validates the application's server/app settings
// from a YAML or JSON file. Every field here must be read by something -
// an unused config field silently does nothing when set, which is worse
// than not having the field at all, so don't add one without wiring it up.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Host string `json:"host" yaml:"host"`
	Port int    `json:"port" yaml:"port"`
	// Timeout bounds how long a single request is allowed to take, in
	// seconds - applied to both the HTTP server's ReadTimeout and
	// WriteTimeout (see runServer in cmd/main.go). Required and must be
	// positive: a zero value would disable the timeout entirely (net/http
	// treats 0 as "no timeout"), which is never what an unset config value
	// should silently produce.
	Timeout int `json:"timeout" yaml:"timeout"`
}

type AppConfig struct {
	DataFile  string `json:"data_file" yaml:"data_file"`
	UploadDir string `json:"upload_dir" yaml:"upload_dir"`
	// AdminToken gates POST /v1/admin/upload - a plain pre-shared secret
	// compared in constant time (see internal/server/middleware/auth.go),
	// not a JWT. There's no user/claims concept in this service, so a
	// shared secret is the right-sized tool for its one admin action;
	// don't rename this back to anything JWT-flavored unless the auth
	// model actually grows real token verification to match. The
	// ADMIN_TOKEN environment variable always takes priority over this
	// field - see applyEnvOverrides - since a committed config file is the
	// wrong place for a real credential.
	AdminToken       string        `json:"admin_token" yaml:"admin_token"`
	Logging          LoggingConfig `json:"logging" yaml:"logging"`
	AllowedRAM       []string      `json:"allowed_ram" yaml:"allowed_ram"`
	AllowedDiskTypes []string      `json:"allowed_disk_types" yaml:"allowed_disk_types"`
}

// LoggingConfig.Level controls the minimum zap level the server logs at
// (debug/info/warn/error, default info) - see internal/platform/log.NewLogger.
type LoggingConfig struct {
	Level string `json:"level" yaml:"level"`
}

type Config struct {
	Server ServerConfig `json:"server" yaml:"server"`
	App    AppConfig    `json:"app" yaml:"app"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	switch {
	case len(path) >= 5 && path[len(path)-5:] == ".yaml":
		err = yaml.Unmarshal(data, cfg)
	case len(path) >= 5 && path[len(path)-5:] == ".yml":
		err = yaml.Unmarshal(data, cfg)
	default:
		err = json.Unmarshal(data, cfg)
	}
	if err != nil {
		return nil, err
	}
	if err := cfg.applyEnvOverrides(); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// applyEnvOverrides lets deployment-time environment variables take
// priority over the checked-in config file - preference is always env var
// first, falling back to whatever the file says (or its zero value) only
// when the variable isn't set. Two fields use this, for two different
// reasons: ADMIN_TOKEN because a committed file is the wrong place for a
// real credential (config.yaml intentionally ships with no admin_token at
// all), and PORT because most PaaS platforms (Render, Heroku, Railway,
// ...) assign the port at deploy time and route their load balancer to it
// - config.yaml's port is only ever the local-dev default, never the final
// word once deployed.
func (c *Config) applyEnvOverrides() error {
	if token := os.Getenv("ADMIN_TOKEN"); token != "" {
		c.App.AdminToken = token
	}
	if raw := os.Getenv("PORT"); raw != "" {
		port, err := strconv.Atoi(raw)
		if err != nil || port <= 0 {
			return fmt.Errorf("invalid PORT environment variable %q: must be a positive integer", raw)
		}
		c.Server.Port = port
	}
	return nil
}

func (c *Config) Validate() error {
	if c.Server.Host == "" {
		return fmt.Errorf("server.host is required")
	}
	if c.Server.Port == 0 {
		return fmt.Errorf("server.port is required")
	}
	if c.Server.Timeout <= 0 {
		return fmt.Errorf("server.timeout must be a positive number of seconds")
	}
	if c.App.DataFile == "" {
		return fmt.Errorf("app.data_file is required")
	}
	if c.App.UploadDir == "" {
		return fmt.Errorf("app.upload_dir is required")
	}
	return nil
}
