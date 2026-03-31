package handler

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/RumenDamyanov/nginx-waf-api/internal/lists"
	"github.com/RumenDamyanov/nginx-waf-api/internal/reload"
)

func setup(t *testing.T) (*http.ServeMux, string) {
	dir := t.TempDir()
	content := "# test\n192.168.1.1\n10.0.0.0/8\n"
	if err := os.WriteFile(filepath.Join(dir, "test.txt"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mgr := lists.NewManager(dir)
	reloader := reload.New("echo reload", 0, logger)
	h := New(mgr, reloader, logger)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux, dir
}

func TestListAll(t *testing.T) {
	mux, _ := setup(t)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/api/v1/lists", nil))
	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}

func TestGetList(t *testing.T) {
	mux, _ := setup(t)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/api/v1/lists/test", nil))
	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}

func TestGetListNotFound(t *testing.T) {
	mux, _ := setup(t)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/api/v1/lists/nonexistent", nil))
	if rr.Code != 404 {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}

func TestAddEntry(t *testing.T) {
	mux, _ := setup(t)
	body, _ := json.Marshal(map[string]string{"ip": "172.16.0.0/12"})
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("POST", "/api/v1/lists/test/entries", bytes.NewReader(body)))
	if rr.Code != 201 {
		t.Fatalf("status = %d, want 201", rr.Code)
	}
}

func TestAddDuplicate(t *testing.T) {
	mux, _ := setup(t)
	body, _ := json.Marshal(map[string]string{"ip": "192.168.1.1"})
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("POST", "/api/v1/lists/test/entries", bytes.NewReader(body)))
	if rr.Code != 409 {
		t.Fatalf("status = %d, want 409", rr.Code)
	}
}

func TestRemoveEntry(t *testing.T) {
	mux, _ := setup(t)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("DELETE", "/api/v1/lists/test/entries/192.168.1.1", nil))
	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}

func TestHealth(t *testing.T) {
	mux, _ := setup(t)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))
	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}
