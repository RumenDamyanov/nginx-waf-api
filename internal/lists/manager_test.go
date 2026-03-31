package lists

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) string {
	dir := t.TempDir()
	content := "# test list\n192.168.1.1\n10.0.0.0/8\n2001:db8::1\n"
	if err := os.WriteFile(filepath.Join(dir, "test.txt"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestList(t *testing.T) {
	dir := setupTestDir(t)
	mgr := NewManager(dir)
	lists, err := mgr.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lists) != 1 {
		t.Fatalf("expected 1, got %d", len(lists))
	}
	if lists[0].Entries != 3 {
		t.Errorf("entries = %d, want 3", lists[0].Entries)
	}
}

func TestGet(t *testing.T) {
	dir := setupTestDir(t)
	mgr := NewManager(dir)
	detail, err := mgr.Get("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(detail.IPs) != 3 {
		t.Fatalf("ips = %d, want 3", len(detail.IPs))
	}
}

func TestGetInvalidName(t *testing.T) {
	mgr := NewManager(t.TempDir())
	_, err := mgr.Get("../etc/passwd")
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestAddEntry(t *testing.T) {
	dir := setupTestDir(t)
	mgr := NewManager(dir)
	if err := mgr.AddEntry("test", "172.16.0.0/12"); err != nil {
		t.Fatalf("add: %v", err)
	}
	detail, _ := mgr.Get("test")
	if len(detail.IPs) != 4 {
		t.Errorf("expected 4, got %d", len(detail.IPs))
	}
}

func TestAddDuplicate(t *testing.T) {
	dir := setupTestDir(t)
	mgr := NewManager(dir)
	err := mgr.AddEntry("test", "192.168.1.1")
	if err == nil {
		t.Fatal("expected error for duplicate")
	}
}

func TestAddInvalidIP(t *testing.T) {
	dir := setupTestDir(t)
	mgr := NewManager(dir)
	err := mgr.AddEntry("test", "not-an-ip")
	if err == nil {
		t.Fatal("expected error for invalid IP")
	}
}

func TestRemoveEntry(t *testing.T) {
	dir := setupTestDir(t)
	mgr := NewManager(dir)
	if err := mgr.RemoveEntry("test", "192.168.1.1"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	detail, _ := mgr.Get("test")
	if len(detail.IPs) != 2 {
		t.Errorf("expected 2, got %d", len(detail.IPs))
	}
}

func TestRemoveNotFound(t *testing.T) {
	dir := setupTestDir(t)
	mgr := NewManager(dir)
	err := mgr.RemoveEntry("test", "1.1.1.1")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestAddCreatesNewList(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)
	if err := mgr.AddEntry("newlist", "10.0.0.1"); err != nil {
		t.Fatalf("add to new list: %v", err)
	}
	detail, err := mgr.Get("newlist")
	if err != nil {
		t.Fatalf("get new list: %v", err)
	}
	if len(detail.IPs) != 1 {
		t.Errorf("expected 1, got %d", len(detail.IPs))
	}
}
