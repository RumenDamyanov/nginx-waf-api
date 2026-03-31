package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Auth    AuthConfig    `yaml:"auth"`
	Nginx   NginxConfig   `yaml:"nginx"`
	Logging LogConfig     `yaml:"logging"`
	Metrics MetricsConfig `yaml:"metrics"`
}

// ServerConfig holds the HTTP server settings.
type ServerConfig struct {
	Listen string    `yaml:"listen"`
	TLS    TLSConfig `yaml:"tls"`
}

// TLSConfig holds optional TLS settings.
type TLSConfig struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	APIKeys []APIKeyConfig `yaml:"api_keys"`
}

// APIKeyConfig defines a single API key.
type APIKeyConfig struct {
	Name        string   `yaml:"name"`
	Key         string   `yaml:"key"`
	Permissions []string `yaml:"permissions"`
}

// NginxConfig holds nginx integration settings.
type NginxConfig struct {
	ListsDir       string        `yaml:"lists_dir"`
	ReloadCommand  string        `yaml:"reload_command"`
	ReloadDebounce time.Duration `yaml:"reload_debounce"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// MetricsConfig holds metrics endpoint settings.
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`
}

// UnmarshalYAML handles duration parsing for NginxConfig.
func (n *NginxConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type raw struct {
		ListsDir       string `yaml:"lists_dir"`
		ReloadCommand  string `yaml:"reload_command"`
		ReloadDebounce int    `yaml:"reload_debounce"`
	}
	var r raw
	if err := unmarshal(&r); err != nil {
		return err
	}
	n.ListsDir = r.ListsDir
	n.ReloadCommand = r.ReloadCommand
	n.ReloadDebounce = time.Duration(r.ReloadDebounce) * time.Second
	return nil
}

// Load reads and parses the configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	expanded := os.ExpandEnv(string(data))

	cfg := &Config{
		Server: ServerConfig{
			Listen: ":8080",
		},
		Nginx: NginxConfig{
			ListsDir:       "/etc/nginx/waf-lists",
			ReloadDebounce: 5 * time.Second,
		},
		Logging: LogConfig{
			Level:  "info",
			Format: "text",
		},
		Metrics: MetricsConfig{
			Listen: ":9102",
		},
	}

	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if c.Server.Listen == "" {
		return fmt.Errorf("server.listen is required")
	}
	if c.Nginx.ListsDir == "" {
		return fmt.Errorf("nginx.lists_dir is required")
	}
	for i, k := range c.Auth.APIKeys {
		if k.Key == "" {
			return fmt.Errorf("auth.api_keys[%d]: key is required", i)
		}
		for _, p := range k.Permissions {
			if p != "read" && p != "write" {
				return fmt.Errorf("auth.api_keys[%d]: invalid permission %q (use read or write)", i, p)
			}
		}
	}
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[strings.ToLower(c.Logging.Level)] {
		return fmt.Errorf("logging.level must be one of: debug, info, warn, error")
	}
	return nil
}

// HasPermission checks if any configured API key matches and has the required permission.
func (c *Config) HasPermission(key, permission string) bool {
	for _, k := range c.Auth.APIKeys {
		if k.Key == key {
			for _, p := range k.Permissions {
				if p == permission {
					return true
				}
			}
			return false
		}
	}
	return false
}
