package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	content := `
server:
  listen: ":9090"
auth:
  api_keys:
    - name: test
      key: secret123
      permissions: [read, write]
nginx:
  lists_dir: /tmp/lists
  reload_debounce: 3
logging:
  level: debug
`
	if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Listen != ":9090" {
		t.Errorf("listen = %q", cfg.Server.Listen)
	}
	if len(cfg.Auth.APIKeys) != 1 {
		t.Fatalf("api_keys count = %d", len(cfg.Auth.APIKeys))
	}
	if cfg.Auth.APIKeys[0].Key != "secret123" {
		t.Errorf("api key = %q", cfg.Auth.APIKeys[0].Key)
	}
}

func TestHasPermission(t *testing.T) {
	cfg := &Config{
		Auth: AuthConfig{
			APIKeys: []APIKeyConfig{
				{Name: "admin", Key: "key1", Permissions: []string{"read", "write"}},
				{Name: "reader", Key: "key2", Permissions: []string{"read"}},
			},
		},
	}
	if !cfg.HasPermission("key1", "write") {
		t.Error("admin should have write")
	}
	if cfg.HasPermission("key2", "write") {
		t.Error("reader should not have write")
	}
	if cfg.HasPermission("badkey", "read") {
		t.Error("unknown key should not have any permission")
	}
}

func TestValidateBadPermission(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	content := `
auth:
  api_keys:
    - name: bad
      key: k
      permissions: [admin]
nginx:
  lists_dir: /tmp
`
	if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(cfgFile)
	if err == nil {
		t.Fatal("expected error for invalid permission")
	}
}

func TestEnvExpansion(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	t.Setenv("TEST_API_KEY", "expanded-key")
	content := `
auth:
  api_keys:
    - name: env
      key: ${TEST_API_KEY}
      permissions: [read]
nginx:
  lists_dir: /tmp/lists
`
	if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Auth.APIKeys[0].Key != "expanded-key" {
		t.Errorf("env expansion: key = %q", cfg.Auth.APIKeys[0].Key)
	}
}
