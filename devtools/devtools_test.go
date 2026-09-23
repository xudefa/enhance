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

	// 等待检测到变化并校验事件
	testHotReloaderDetectEvent(hotReloaderDetectEventArgs{
		t:          t,
		wg:         &wg,
		events:     &events,
		mu:         &mu,
		wantType:   ReloadTypeModified,
		wantFile:   testFile,
		timeoutMsg: "Timeout waiting for file change detection",
	})
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

	// 等待检测到变化并校验事件
	testHotReloaderDetectEvent(hotReloaderDetectEventArgs{
		t:          t,
		wg:         &wg,
		events:     &events,
		mu:         &mu,
		wantType:   ReloadTypeCreated,
		timeoutMsg: "Timeout waiting for new file detection",
	})
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

	// 等待检测到变化并校验事件
	testHotReloaderDetectEvent(hotReloaderDetectEventArgs{
		t:          t,
		wg:         &wg,
		events:     &events,
		mu:         &mu,
		wantType:   ReloadTypeDeleted,
		timeoutMsg: "Timeout waiting for file deletion detection",
	})
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

	started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})

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

	testStopWaitsForCallbacks(t, done, stopDone, release)
}

func TestHotReloader_MultipleCallbacks(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(testFile, []byte(`{"key": "value"}`), 0644); err != nil {
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

	if err := reloader.Start(); err != nil {
		t.Fatalf("Failed to start reloader: %v", err)
	}
	defer reloader.Stop()

	// 等待初始扫描
	time.Sleep(200 * time.Millisecond)

	// 修改文件
	if err := os.WriteFile(testFile, []byte(`{"key": "value2"}`), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// 等待检测
	time.Sleep(300 * time.Millisecond)

	testHotReloaderExpectCallbacks(t, &mu, &callback1Count, &callback2Count)
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
