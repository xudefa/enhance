package devtools

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ==================== DevModeDetector 实现 ====================

// devModeDetectorImpl DevModeDetector 接口的默认实现。
type devModeDetectorImpl struct {
	envVars []string
}

// NewDevModeDetector 创建开发模式检测器。
func NewDevModeDetector() DevModeDetector {
	return &devModeDetectorImpl{
		envVars: []string{
			"DEV_MODE",
			"DEVELOPMENT",
			"GO_ENV",
		},
	}
}

// IsDevMode 检测是否为开发模式。
func (d *devModeDetectorImpl) IsDevMode() bool {
	for _, envVar := range d.envVars {
		if value := os.Getenv(envVar); value == "true" || value == "development" || value == "dev" {
			return true
		}
	}
	return false
}

// ==================== FileWatcher 实现 ====================

// fileWatcherImpl FileWatcher 接口的默认实现。
type fileWatcherImpl struct {
	mu         sync.RWMutex
	watchDirs  []string
	extensions map[string]bool
	callbacks  []ReloadCallback
	stopChan   chan struct{}
	running    bool
	wg         sync.WaitGroup
	callbackWg sync.WaitGroup
}

// NewFileWatcher 创建文件监控器。
func NewFileWatcher(dirs []string, extensions ...string) FileWatcher {
	watcher := &fileWatcherImpl{
		watchDirs:  dirs,
		extensions: make(map[string]bool),
		stopChan:   make(chan struct{}),
	}

	for _, ext := range extensions {
		watcher.extensions[ext] = true
	}

	return watcher
}

// OnChange 注册文件变化回调。
func (w *fileWatcherImpl) OnChange(callback ReloadCallback) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.callbacks = append(w.callbacks, callback)
}

// Start 启动监控。
func (w *fileWatcherImpl) Start() error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("file watcher is already running")
	}
	w.running = true
	w.stopChan = make(chan struct{})
	stopChan := w.stopChan
	w.mu.Unlock()

	w.wg.Add(1)
	go w.pollFiles(stopChan)

	return nil
}

// Stop 停止监控。
func (w *fileWatcherImpl) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}

	w.running = false
	close(w.stopChan)
	w.mu.Unlock()

	w.wg.Wait()
	w.callbackWg.Wait()
}

// pollFiles 轮询文件。
func (w *fileWatcherImpl) pollFiles(stopChan chan struct{}) {
	defer w.wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fileHashes := make(map[string]string)

	// 获取目录和扩展名的快照
	w.mu.RLock()
	dirs := make([]string, len(w.watchDirs))
	copy(dirs, w.watchDirs)
	extensions := make(map[string]bool)
	for k, v := range w.extensions {
		extensions[k] = v
	}
	w.mu.RUnlock()

	for _, dir := range dirs {
		w.scanDirForWatcher(dir, extensions, fileHashes)
	}

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			currentFiles := make(map[string]string)
			for _, dir := range dirs {
				w.scanDirForWatcher(dir, extensions, currentFiles)
			}

			for file, newHash := range currentFiles {
				if oldHash, exists := fileHashes[file]; !exists || oldHash != newHash {
					eventType := ReloadTypeModified
					if !exists {
						eventType = ReloadTypeCreated
					}

					event := ReloadEvent{
						File:      file,
						Type:      eventType,
						Timestamp: time.Now(),
						OldHash:   oldHash,
						NewHash:   newHash,
					}

					w.mu.RLock()
					cbs := make([]ReloadCallback, len(w.callbacks))
					copy(cbs, w.callbacks)
					w.mu.RUnlock()

					for _, callback := range cbs {
						w.callbackWg.Add(1)
						go func(cb ReloadCallback) {
							defer w.callbackWg.Done()
							defer func() {
								if r := recover(); r != nil {
									fmt.Printf("[devtools] file watcher callback panic: %v\n", r)
								}
							}()
							cb(event)
						}(callback)
					}
				}
			}

			for file, oldHash := range fileHashes {
				if _, exists := currentFiles[file]; !exists {
					event := ReloadEvent{
						File:      file,
						Type:      ReloadTypeDeleted,
						Timestamp: time.Now(),
						OldHash:   oldHash,
					}

					w.mu.RLock()
					cbs := make([]ReloadCallback, len(w.callbacks))
					copy(cbs, w.callbacks)
					w.mu.RUnlock()

					for _, callback := range cbs {
						w.callbackWg.Add(1)
						go func(cb ReloadCallback) {
							defer w.callbackWg.Done()
							defer func() {
								if r := recover(); r != nil {
									fmt.Printf("[devtools] file watcher callback panic: %v\n", r)
								}
							}()
							cb(event)
						}(callback)
					}
				}
			}

			fileHashes = currentFiles
		}
	}
}

// scanDirForWatcher 扫描目录。
func (w *fileWatcherImpl) scanDirForWatcher(dir string, extensions map[string]bool, hashes map[string]string) {
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		if len(extensions) > 0 {
			ext := filepath.Ext(path)
			if !extensions[ext] {
				return nil
			}
		}

		hash, err := computeFileHash(path)
		if err != nil {
			return nil
		}

		hashes[path] = hash
		return nil
	})
}

// ==================== LiveReloadServer 实现 ====================

// LiveReloadServer 实时重载服务器。
type LiveReloadServer struct {
	mu       sync.Mutex
	port     int
	reloader *hotReloaderImpl
	stopChan chan struct{}
	running  bool
}

// NewLiveReloadServer 创建实时重载服务器。
func NewLiveReloadServer(port int, reloader HotReloader) (*LiveReloadServer, error) {
	impl, ok := reloader.(*hotReloaderImpl)
	if !ok {
		return nil, fmt.Errorf("unsupported reloader type: %T", reloader)
	}
	return &LiveReloadServer{
		port:     port,
		reloader: impl,
		stopChan: make(chan struct{}),
	}, nil
}

// Start 启动服务器。
func (s *LiveReloadServer) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("live reload server is already running")
	}
	s.running = true
	s.mu.Unlock()

	s.reloader.OnReload(func(event ReloadEvent) {
		fmt.Printf("[LiveReload] File changed: %s (%s)\n", event.File, event.Type)
	})

	return s.reloader.Start()
}

// Stop 停止服务器。
func (s *LiveReloadServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	s.reloader.Stop()
	close(s.stopChan)
}

// IsRunning 检查是否正在运行。
func (s *LiveReloadServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.running
}
