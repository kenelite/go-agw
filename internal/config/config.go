package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	HTTPAddr  string `yaml:"http_addr"`
	AdminAddr string `yaml:"admin_addr"`
}

type UpstreamConfig struct {
	Name    string   `yaml:"name"`
	Targets []string `yaml:"targets"`
	Timeout int      `yaml:"timeout_ms"`
}

type RouteConfig struct {
	Path        string          `yaml:"path"`
	Methods     []string        `yaml:"methods"`
	UpstreamRef string          `yaml:"upstream"`
	RateLimit   RateLimitConfig `yaml:"rate_limit"`
	Plugins     []PluginRef     `yaml:"plugins"`
}

type RateLimitConfig struct {
	RequestsPerSecond int `yaml:"rps"`
	Burst             int `yaml:"burst"`
}

type PluginRef struct {
	Name   string         `yaml:"name"`
	Config map[string]any `yaml:"config"`
}

type ObservabilityConfig struct {
	LogLevel string `yaml:"log_level"`
}

type PluginsConfig struct {
	Available []PluginRef `yaml:"available"`
}

type Config struct {
	Server        ServerConfig        `yaml:"server"`
	Upstreams     []UpstreamConfig    `yaml:"upstreams"`
	Routes        []RouteConfig       `yaml:"routes"`
	Observability ObservabilityConfig `yaml:"observability"`
	Plugins       PluginsConfig       `yaml:"plugins"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	if cfg.Server.HTTPAddr == "" {
		cfg.Server.HTTPAddr = ":8080"
	}
	if cfg.Server.AdminAddr == "" {
		cfg.Server.AdminAddr = ":9000"
	}
	return &cfg, nil
}
