package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	if r == nil {
		t.Fatal("expected non-nil Registry")
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register("UserService", "proxy/user_service.go")

	path, ok := r.Get("UserService")
	if !ok {
		t.Fatal("expected to find UserService")
	}
	if path != "proxy/user_service.go" {
		t.Errorf("expected 'proxy/user_service.go', got %s", path)
	}
}

func TestRegistry_GetNonExistent(t *testing.T) {
	t.Parallel()

	r := NewRegistry()

	_, ok := r.Get("NonExistent")
	if ok {
		t.Error("expected false for non-existent key")
	}
}

func TestRegistry_List_Extended(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register("UserService", "proxy/user_service.go")
	r.Register("OrderService", "proxy/order_service.go")

	list := r.List()
	if len(list) != 2 {
		t.Errorf("expected 2 entries, got %d", len(list))
	}
	if list["UserService"] != "proxy/user_service.go" {
		t.Errorf("expected UserService path, got %s", list["UserService"])
	}
	if list["OrderService"] != "proxy/order_service.go" {
		t.Errorf("expected OrderService path, got %s", list["OrderService"])
	}
}

func TestRegistry_SaveAndLoad_Extended(t *testing.T) {
	t.Parallel()

	r1 := NewRegistry()
	r1.Register("UserService", "proxy/user_service.go")
	r1.Register("OrderService", "proxy/order_service.go")

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "registry.json")

	err := r1.Save(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("expected registry file to exist")
	}

	// Load into new registry
	r2 := NewRegistry()
	err = r2.Load(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path, ok := r2.Get("UserService")
	if !ok {
		t.Fatal("expected to find UserService after load")
	}
	if path != "proxy/user_service.go" {
		t.Errorf("expected 'proxy/user_service.go', got %s", path)
	}
}

func TestRegistry_LoadNonExistentFile(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	err := r.Load("/non/existent/file.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestRegistry_LoadInvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid.json")
	err := os.WriteFile(filePath, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	r := NewRegistry()
	err = r.Load(filePath)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestRegistry_Clear_Extended(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register("UserService", "proxy/user_service.go")
	r.Register("OrderService", "proxy/order_service.go")

	r.Clear()

	list := r.List()
	if len(list) != 0 {
		t.Errorf("expected 0 entries after clear, got %d", len(list))
	}
}

func TestRegistry_Overwrite(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register("UserService", "proxy/user_service_v1.go")
	r.Register("UserService", "proxy/user_service_v2.go")

	path, ok := r.Get("UserService")
	if !ok {
		t.Fatal("expected to find UserService")
	}
	if path != "proxy/user_service_v2.go" {
		t.Errorf("expected 'proxy/user_service_v2.go', got %s", path)
	}
}

func TestRegistry_LoadOverwritesExisting(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register("OldService", "proxy/old_service.go")

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "registry.json")

	r2 := NewRegistry()
	r2.Register("NewService", "proxy/new_service.go")
	r2.Save(filePath)

	err := r.Load(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Old entry should be removed
	_, ok := r.Get("OldService")
	if ok {
		t.Error("expected OldService to be removed after load")
	}

	// New entry should be present
	_, ok = r.Get("NewService")
	if !ok {
		t.Error("expected NewService to be present after load")
	}
}
