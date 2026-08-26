package devtools

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewDevModeDetector_Helper(t *testing.T) {
	t.Parallel()
	d := NewDevModeDetector()
	if d == nil {
		t.Fatal("NewDevModeDetector returned nil")
	}
}

func TestDevModeDetector_IsDevMode_Helper(t *testing.T) {
	t.Parallel()
	d := NewDevModeDetector()

	// Ensure env vars are cleared
	for _, env := range []string{"DEV_MODE", "DEVELOPMENT", "GO_ENV"} {
		os.Unsetenv(env)
	}
	if d.IsDevMode() {
		t.Error("should not be in dev mode without env vars")
	}
}

func TestDevModeDetector_IsDevMode_TrueHelper(t *testing.T) {
	t.Parallel()
	d := NewDevModeDetector()

	tests := []struct {
		name  string
		value string
	}{
		{"DEV_MODE true", "true"},
		{"DEV_MODE development", "development"},
		{"DEV_MODE dev", "dev"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			os.Setenv("DEV_MODE", tt.value)
			defer os.Unsetenv("DEV_MODE")

			if !d.IsDevMode() {
				t.Errorf("should be in dev mode with %s=%s", "DEV_MODE", tt.value)
			}
		})
	}
}

func TestNewFileWatcher_Helper(t *testing.T) {
	t.Parallel()
	w := NewFileWatcher([]string{"/tmp"}, ".go", ".json")
	if w == nil {
		t.Fatal("NewFileWatcher returned nil")
	}
}

func TestFileWatcher_StartStop_Helper(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	w := NewFileWatcher([]string{tmpDir}, ".go")

	if err := w.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	w.Stop()
}

func TestFileWatcher_DoubleStart_Helper(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	w := NewFileWatcher([]string{tmpDir})

	if err := w.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer w.Stop()

	err := w.Start()
	if err == nil {
		t.Error("second Start should return error")
	}
}

func TestFileWatcher_StopWithoutStart_Helper(t *testing.T) {
	t.Parallel()
	w := NewFileWatcher([]string{"/tmp"})
	w.Stop()
}

func TestFileWatcher_OnChange_Helper(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatal(err)
	}

	w := NewFileWatcher([]string{tmpDir}, ".txt")

	var mu sync.Mutex
	var events []ReloadEvent

	w.OnChange(func(event ReloadEvent) {
		mu.Lock()
		events = append(events, event)
		mu.Unlock()
	})

	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	// Modify file
	time.Sleep(1500 * time.Millisecond)
	if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	mu.Lock()
	count := len(events)
	mu.Unlock()

	if count == 0 {
		t.Error("expected at least one change event")
	}
}

func TestLiveReloadServer_BasicHelper(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	reloader := NewHotReloader(
		WithWatchDirs(tmpDir),
		WithExtensions(".go"),
		WithInterval(100*time.Millisecond),
	)

	server, err := NewLiveReloadServer(0, reloader)
	if err != nil {
		t.Fatalf("NewLiveReloadServer error: %v", err)
	}

	if server.IsRunning() {
		t.Error("should not be running initially")
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	if !server.IsRunning() {
		t.Error("should be running after Start")
	}

	server.Stop()

	if server.IsRunning() {
		t.Error("should not be running after Stop")
	}
}

func TestLiveReloadServer_DoubleStartHelper(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	reloader := NewHotReloader(
		WithWatchDirs(tmpDir),
		WithExtensions(".go"),
		WithInterval(100*time.Millisecond),
	)

	server, _ := NewLiveReloadServer(0, reloader)

	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	defer server.Stop()

	err := server.Start()
	if err == nil {
		t.Error("double Start should return error")
	}
}

func TestLiveReloadServer_StopWithoutStart_Helper(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	reloader := NewHotReloader(
		WithWatchDirs(tmpDir),
		WithExtensions(".go"),
		WithInterval(100*time.Millisecond),
	)

	server, _ := NewLiveReloadServer(0, reloader)
	server.Stop()
}

func TestLiveReloadServer_UnsupportedReloader_Helper(t *testing.T) {
	t.Parallel()
	_, err := NewLiveReloadServer(0, &unsupportedReloaderHelper{})
	if err == nil {
		t.Error("expected error for unsupported reloader type")
	}
}

type unsupportedReloaderHelper struct{}

func (u *unsupportedReloaderHelper) OnReload(_ ReloadCallback) {}
func (u *unsupportedReloaderHelper) Start() error              { return nil }
func (u *unsupportedReloaderHelper) Stop()                     {}
