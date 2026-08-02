package devtools

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestHotReloader_Basic(t *testing.T) {
	t.Parallel()
	// 创建临时目录
	tmpDir := t.TempDir()

	reloader := NewHotReloader(
		WithWatchDirs(tmpDir),
		WithExtensions(".json"),
		WithInterval(100*time.Millisecond),
	)

	err := reloader.Start()
	if err != nil {
		t.Fatalf("Failed to start reloader: %v", err)
	}

	defer reloader.Stop()

	if !reloader.IsRunning() {
		t.Error("expected reloader to be running")
	}
}

func TestHotReloader_DetectFileChange(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "config.json")

	// 创建初始文件
	err := os.WriteFile(testFile, []byte(`{"key": "value1"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var wg sync.WaitGroup
	var events []ReloadEvent
	var mu sync.Mutex

	reloader := NewHotReloader(
		WithWatchDirs(tmpDir),
		WithExtensions(".json"),
		WithInterval(100*time.Millisecond),
	)

	reloader.OnReload(func(event ReloadEvent) {
		mu.Lock()
		events = append(events, event)
		mu.Unlock()
		wg.Done()
	})

	err = reloader.Start()
	if err != nil {
		t.Fatalf("Failed to start reloader: %v", err)
	}
	defer reloader.Stop()

	// 等待初始扫描
	time.Sleep(200 * time.Millisecond)

	// 修改文件
	wg.Add(1)
	err = os.WriteFile(testFile, []byte(`{"key": "value2"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// 等待检测到变化
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 成功检测到变化
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for file change detection")
	}

	mu.Lock()
	if len(events) == 0 {
		t.Fatal("expected at least one reload event")
	}

	event := events[0]
	if event.Type != ReloadTypeModified {
		t.Errorf("expected event type MODIFIED, got %s", event.Type)
	}

	if event.File != testFile {
		t.Errorf("expected file %s, got %s", testFile, event.File)
	}
	mu.Unlock()
}

func TestHotReloader_DetectNewFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	var wg sync.WaitGroup
	var events []ReloadEvent
	var mu sync.Mutex

	reloader := NewHotReloader(
		WithWatchDirs(tmpDir),
		WithExtensions(".json"),
		WithInterval(100*time.Millisecond),
	)

	reloader.OnReload(func(event ReloadEvent) {
		mu.Lock()
		events = append(events, event)
		mu.Unlock()
		wg.Done()
	})

	err := reloader.Start()
	if err != nil {
		t.Fatalf("Failed to start reloader: %v", err)
	}
	defer reloader.Stop()

	// 等待初始扫描
	time.Sleep(200 * time.Millisecond)

	// 创建新文件
	newFile := filepath.Join(tmpDir, "new.json")
	wg.Add(1)
	err = os.WriteFile(newFile, []byte(`{"new": true}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	// 等待检测到变化
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 成功检测到新文件
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for new file detection")
	}

	mu.Lock()
	if len(events) == 0 {
		t.Fatal("expected at least one reload event")
	}

	event := events[0]
	if event.Type != ReloadTypeCreated {
		t.Errorf("expected event type CREATED, got %s", event.Type)
	}
	mu.Unlock()
}

func TestHotReloader_DetectFileDeletion(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "config.json")

	// 创建文件
	err := os.WriteFile(testFile, []byte(`{"key": "value"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var wg sync.WaitGroup
	var events []ReloadEvent
	var mu sync.Mutex

	reloader := NewHotReloader(
		WithWatchDirs(tmpDir),
		WithExtensions(".json"),
		WithInterval(100*time.Millisecond),
	)

	reloader.OnReload(func(event ReloadEvent) {
		mu.Lock()
		events = append(events, event)
		mu.Unlock()
		wg.Done()
	})

	err = reloader.Start()
	if err != nil {
		t.Fatalf("Failed to start reloader: %v", err)
	}
	defer reloader.Stop()

	// 等待初始扫描
	time.Sleep(200 * time.Millisecond)

	// 删除文件
	wg.Add(1)
	err = os.Remove(testFile)
	if err != nil {
		t.Fatalf("Failed to delete test file: %v", err)
	}

	// 等待检测到变化
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 成功检测到删除
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for file deletion detection")
	}

	mu.Lock()
	if len(events) == 0 {
		t.Fatal("expected at least one reload event")
	}

	event := events[0]
	if event.Type != ReloadTypeDeleted {
		t.Errorf("expected event type DELETED, got %s", event.Type)
	}
	mu.Unlock()
}

func TestHotReloader_StopWaitsForCallbacks(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(testFile, []byte(`{"key": "value"}`), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	reloader := NewHotReloader(
		WithWatchDirs(tmpDir),
		WithExtensions(".json"),
		WithInterval(20*time.Millisecond),
	)

	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})

	var startOnce sync.Once
	reloader.OnReload(func(event ReloadEvent) {
		startOnce.Do(func() { close(started) })
		<-release
		close(done)
	})

	if err := reloader.Start(); err != nil {
		t.Fatalf("Failed to start reloader: %v", err)
	}

	// 等待初始扫描
	time.Sleep(100 * time.Millisecond)

	if err := os.WriteFile(testFile, []byte(`{"key": "value2"}`), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for reload callback to start")
	}

	stopDone := make(chan struct{})
	go func() {
		reloader.Stop()
		close(stopDone)
	}()

	// Stop 必须等待正在执行的回调完成，而不是立即返回
	select {
	case <-stopDone:
		t.Fatal("Stop returned while callbacks were still running")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Callback did not complete after release")
	}

	select {
	case <-stopDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not return after callbacks completed")
	}
}

func TestHotReloader_MultipleCallbacks(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "config.json")

	err := os.WriteFile(testFile, []byte(`{"key": "value"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var callback1Count, callback2Count int
	var mu sync.Mutex

	reloader := NewHotReloader(
		WithWatchDirs(tmpDir),
		WithExtensions(".json"),
		WithInterval(100*time.Millisecond),
	)

	reloader.OnReload(func(event ReloadEvent) {
		mu.Lock()
		callback1Count++
		mu.Unlock()
	})

	reloader.OnReload(func(event ReloadEvent) {
		mu.Lock()
		callback2Count++
		mu.Unlock()
	})

	err = reloader.Start()
	if err != nil {
		t.Fatalf("Failed to start reloader: %v", err)
	}
	defer reloader.Stop()

	// 等待初始扫描
	time.Sleep(200 * time.Millisecond)

	// 修改文件
	err = os.WriteFile(testFile, []byte(`{"key": "value2"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// 等待检测
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	if callback1Count == 0 {
		t.Error("expected callback1 to be called")
	}

	if callback2Count == 0 {
		t.Error("expected callback2 to be called")
	}
	mu.Unlock()
}

func TestHotReloader_IgnoreDirs(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	ignoredDir := filepath.Join(tmpDir, ".git")

	err := os.Mkdir(ignoredDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create ignored dir: %v", err)
	}

	ignoredFile := filepath.Join(ignoredDir, "config.json")
	err = os.WriteFile(ignoredFile, []byte(`{"key": "value"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create ignored file: %v", err)
	}

	reloader := NewHotReloader(
		WithWatchDirs(tmpDir),
		WithExtensions(".json"),
		WithIgnoreDirs(".git"),
	)

	err = reloader.Start()
	if err != nil {
		t.Fatalf("Failed to start reloader: %v", err)
	}
	defer reloader.Stop()

	// 等待扫描
	time.Sleep(200 * time.Millisecond)

	files := reloader.GetWatchedFiles()
	for _, file := range files {
		if filepath.Dir(file) == ignoredDir {
			t.Errorf("expected ignored dir files to not be watched, but found %s", file)
		}
	}
}

func TestHotReloader_GetWatchDirs(t *testing.T) {
	t.Parallel()
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	reloader := NewHotReloader(
		WithWatchDirs(dir1, dir2),
	)

	dirs := reloader.GetWatchDirs()
	if len(dirs) != 2 {
		t.Errorf("expected 2 watch dirs, got %d", len(dirs))
	}

	if dirs[0] != dir1 {
		t.Errorf("expected first dir %s, got %s", dir1, dirs[0])
	}

	if dirs[1] != dir2 {
		t.Errorf("expected second dir %s, got %s", dir2, dirs[1])
	}
}

func TestHotReloader_Restart(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	reloader := NewHotReloader(
		WithWatchDirs(tmpDir),
		WithInterval(100*time.Millisecond),
	)

	err := reloader.Start()
	if err != nil {
		t.Fatalf("Failed to start reloader: %v", err)
	}

	if !reloader.IsRunning() {
		t.Error("expected reloader to be running")
	}

	err = reloader.Restart()
	if err != nil {
		t.Fatalf("Failed to restart reloader: %v", err)
	}

	if !reloader.IsRunning() {
		t.Error("expected reloader to be running after restart")
	}

	reloader.Stop()
}

func TestHotReloader_DoubleStart(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	reloader := NewHotReloader(
		WithWatchDirs(tmpDir),
	)

	err := reloader.Start()
	if err != nil {
		t.Fatalf("Failed to start reloader: %v", err)
	}
	defer reloader.Stop()

	err = reloader.Start()
	if err == nil {
		t.Error("expected error when starting already running reloader")
	}
}

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

	// Stop 必须等待正在执行的回调完成，而不是立即返回
	select {
	case <-stopDone:
		t.Fatal("Stop returned while callbacks were still running")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Callback did not complete after release")
	}

	select {
	case <-stopDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not return after callbacks completed")
	}
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
