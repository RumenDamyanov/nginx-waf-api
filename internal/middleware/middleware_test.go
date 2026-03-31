package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RumenDamyanov/nginx-waf-api/internal/config"
)

func testConfig() *config.Config {
	return &config.Config{
		Auth: config.AuthConfig{
			APIKeys: []config.APIKeyConfig{
				{Name: "admin", Key: "admin-key", Permissions: []string{"read", "write"}},
				{Name: "reader", Key: "read-key", Permissions: []string{"read"}},
			},
		},
	}
}

func TestAuthBearer(t *testing.T) {
	cfg := testConfig()
	handler := Auth(cfg, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer admin-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestAuthXAPIKey(t *testing.T) {
	cfg := testConfig()
	handler := Auth(cfg, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "read-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestAuthMissing(t *testing.T) {
	cfg := testConfig()
	handler := Auth(cfg, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuthWriteForbidden(t *testing.T) {
	cfg := testConfig()
	handler := Auth(cfg, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-API-Key", "read-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
