// Package config loads and validates the application's server/app settings
// from a YAML or JSON file. Every field here must be read by something -
// an unused config field silently does nothing when set, which is worse
// than not having the field at all, so don't add one without wiring it up.
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Host    string `json:"host" yaml:"host"`
	Port    int    `json:"port" yaml:"port"`
	Timeout int    `json:"timeout" yaml:"timeout"`
}

type AppConfig struct {
	DataFile  string `json:"data_file" yaml:"data_file"`
	UploadDir string `json:"upload_dir" yaml:"upload_dir"`
	// AdminToken gates POST /v1/admin/upload - a plain pre-shared secret
	// compared in constant time (see internal/server/middleware/auth.go),
	// not a JWT. There's no user/claims concept in this service, so a
	// shared secret is the right-sized tool for its one admin action;
	// don't rename this back to anything JWT-flavored unless the auth
	// model actually grows real token verification to match.
	AdminToken       string        `json:"admin_token" yaml:"admin_token"`
	Logging          LoggingConfig `json:"logging" yaml:"logging"`
	AllowedRAM       []string      `json:"allowed_ram" yaml:"allowed_ram"`
	AllowedDiskTypes []string      `json:"allowed_disk_types" yaml:"allowed_disk_types"`
}

// LoggingConfig.Level controls the minimum zap level the server logs at
// (debug/info/warn/error, default info) - see internal/log.NewLogger.
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
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Host == "" {
		return fmt.Errorf("server.host is required")
	}
	if c.Server.Port == 0 {
		return fmt.Errorf("server.port is required")
	}
	if c.App.DataFile == "" {
		return fmt.Errorf("app.data_file is required")
	}
	if c.App.UploadDir == "" {
		return fmt.Errorf("app.upload_dir is required")
	}
	return nil
}
