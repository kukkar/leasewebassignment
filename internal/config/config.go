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
	DataFile      string        `json:"data_file" yaml:"data_file"`
	UploadDir     string        `json:"upload_dir" yaml:"upload_dir"`
	JWTSigningKey string        `json:"jwt_signing_key" yaml:"jwt_signing_key"`
	Observability OTelConfig    `json:"observability" yaml:"observability"`
	Logging       LoggingConfig `json:"logging" yaml:"logging"`
}

type OTelConfig struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	Service  string `json:"service" yaml:"service"`
}

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
