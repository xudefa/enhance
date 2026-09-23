package devtools

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestDevModeDetector(t *testing.T) {
	t.Parallel()
	detector := NewDevModeDetector()

	// 默认不是开发模式
	if detector.IsDevMode() {
		t.Error("expected not to be in dev mode by default")
	}

	// 设置环境变量
	_ = os.Setenv("DEV_MODE", "true")
	defer func() { _ = os.Unsetenv("DEV_MODE") }()

	if !detector.IsDevMode() {
		t.Error("expected to be in dev mode when DEV_MODE=true")
	}
}

func TestDevModeDetector_GoEnv(t *testing.T) {
	t.Parallel()
	detector := NewDevModeDetector()

	_ = os.Setenv("GO_ENV", "development")
	defer func() { _ = os.Unsetenv("GO_ENV") }()

	if !detector.IsDevMode() {
		t.Error("expected to be in dev mode when GO_ENV=development")
	}
}

func TestFileWatcher_Basic(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	watcher := NewFileWatcher([]string{tmpDir}, ".json")

	err := watcher.Start()
	if err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	defer watcher.Stop()
}

func TestFileWatcher_StopWaitsForCallbacks(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "config.json")

	watcher := NewFileWatcher([]string{tmpDir}, ".json")

	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})

	var startOnce sync.Once
	watcher.OnChange(func(event ReloadEvent) {
		startOnce.Do(func() { close(started) })
		<-release
		close(done)
	})

	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// 等待初始扫描完成
	time.Sleep(1200 * time.Millisecond)

	if err := os.WriteFile(testFile, []byte(`{"key": "value"}`), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for watcher callback to start")
	}

	stopDone := make(chan struct{})
	go func() {
		watcher.Stop()
		close(stopDone)
	}()

	testStopWaitsForCallbacks(t, done, stopDone, release)
}

func TestCalculateFileHash(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content")
	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash1, err := computeFileHash(testFile)
	if err != nil {
		t.Fatalf("Failed to calculate file hash: %v", err)
	}

	if hash1 == "" {
		t.Error("expected non-empty hash")
	}

	// 相同内容应该产生相同哈希
	hash2, err := computeFileHash(testFile)
	if err != nil {
		t.Fatalf("Failed to calculate file hash: %v", err)
	}

	if hash1 != hash2 {
		t.Error("expected same hash for same content")
	}

	// 修改内容应该产生不同哈希
	err = os.WriteFile(testFile, []byte("different content"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	hash3, err := computeFileHash(testFile)
	if err != nil {
		t.Fatalf("Failed to calculate file hash: %v", err)
	}

	if hash1 == hash3 {
		t.Error("expected different hash for different content")
	}
}

func TestLiveReloadServer_Basic(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	reloader := NewHotReloader(
		WithWatchDirs(tmpDir),
		WithInterval(100*time.Millisecond),
	)

	server, err := NewLiveReloadServer(35729, reloader)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	err = server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	defer server.Stop()

	if !server.IsRunning() {
		t.Error("expected server to be running")
	}
}

func TestLiveReloadServer_DoubleStart(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	reloader := NewHotReloader(
		WithWatchDirs(tmpDir),
	)

	server, err := NewLiveReloadServer(35729, reloader)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	err = server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	err = server.Start()
	if err == nil {
		t.Error("expected error when starting already running server")
	}
}
