// Package devtools 提供开发工具支持，用于 enhance 框架。
package devtools

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ==================== HotReloader 实现 ====================

// hotReloaderImpl HotReloader 接口的默认实现。
type hotReloaderImpl struct {
	mu           sync.RWMutex
	watchDirs    []string
	extensions   map[string]bool
	ignoreDirs   map[string]bool
	pollInterval time.Duration
	callbacks    []ReloadCallback
	fileHashes   map[string]string
	running      bool
	stopChan     chan struct{}
}

// computeFileHash 计算文件 MD5 哈希值。
func computeFileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// shouldWatchFile 检查是否应该监控该文件。
func (r *hotReloaderImpl) shouldWatchFile(path string) bool {
	if len(r.extensions) > 0 {
		ext := filepath.Ext(path)
		if !r.extensions[ext] {
			return false
		}
	}

	for ignoreDir := range r.ignoreDirs {
		if filepath.HasPrefix(path, ignoreDir) {
			return false
		}
	}

	return true
}

// WithWatchDirs 设置监控目录。
func WithWatchDirs(dirs ...string) HotReloaderOption {
	return func(r HotReloader) {
		if impl, ok := r.(*hotReloaderImpl); ok {
			impl.watchDirs = append(impl.watchDirs, dirs...)
		}
	}
}

// WithExtensions 设置监控的文件扩展名。
func WithExtensions(exts ...string) HotReloaderOption {
	return func(r HotReloader) {
		if impl, ok := r.(*hotReloaderImpl); ok {
			impl.extensions = make(map[string]bool)
			for _, ext := range exts {
				impl.extensions[ext] = true
			}
		}
	}
}

// WithInterval 设置轮询间隔。
func WithInterval(interval time.Duration) HotReloaderOption {
	return func(r HotReloader) {
		if impl, ok := r.(*hotReloaderImpl); ok {
			impl.pollInterval = interval
		}
	}
}

// WithIgnoreDirs 设置忽略的目录。
func WithIgnoreDirs(dirs ...string) HotReloaderOption {
	return func(r HotReloader) {
		if impl, ok := r.(*hotReloaderImpl); ok {
			impl.ignoreDirs = make(map[string]bool)
			for _, dir := range dirs {
				impl.ignoreDirs[dir] = true
			}
		}
	}
}

// NewHotReloader 创建热重载管理器。
func NewHotReloader(opts ...HotReloaderOption) HotReloader {
	reloader := &hotReloaderImpl{
		pollInterval: 2 * time.Second,
		extensions:   make(map[string]bool),
		ignoreDirs:   make(map[string]bool),
		fileHashes:   make(map[string]string),
		stopChan:     make(chan struct{}),
	}

	reloader.ignoreDirs[".git"] = true
	reloader.ignoreDirs["node_modules"] = true
	reloader.ignoreDirs["vendor"] = true

	for _, opt := range opts {
		opt(reloader)
	}

	return reloader
}

// OnReload 注册重载回调。
func (r *hotReloaderImpl) OnReload(callback ReloadCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.callbacks = append(r.callbacks, callback)
}

// Start 启动文件监控。
func (r *hotReloaderImpl) Start() error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return fmt.Errorf("hot reloader is already running")
	}
	r.running = true
	r.mu.Unlock()

	if err := r.scanFiles(); err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	go r.poll()

	return nil
}

// Stop 停止文件监控。
func (r *hotReloaderImpl) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return
	}

	r.running = false
	select {
	case <-r.stopChan:
	default:
		close(r.stopChan)
	}
}

// IsRunning 检查是否正在运行。
func (r *hotReloaderImpl) IsRunning() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.running
}

// poll 轮询文件变化。
func (r *hotReloaderImpl) poll() {
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.stopChan:
			return
		case <-ticker.C:
			r.checkForChanges()
		}
	}
}

// checkForChanges 检查文件变化。
func (r *hotReloaderImpl) checkForChanges() {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentFiles := make(map[string]string)

	for _, dir := range r.watchDirs {
		r.scanDir(dir, currentFiles)
	}

	for file, oldHash := range r.fileHashes {
		if _, exists := currentFiles[file]; !exists {
			event := ReloadEvent{
				File:      file,
				Type:      ReloadTypeDeleted,
				Timestamp: time.Now(),
				OldHash:   oldHash,
			}
			r.triggerCallbacks(event)
		}
	}

	for file, newHash := range currentFiles {
		oldHash, exists := r.fileHashes[file]
		if !exists {
			event := ReloadEvent{
				File:      file,
				Type:      ReloadTypeCreated,
				Timestamp: time.Now(),
				NewHash:   newHash,
			}
			r.triggerCallbacks(event)
		} else if oldHash != newHash {
			event := ReloadEvent{
				File:      file,
				Type:      ReloadTypeModified,
				Timestamp: time.Now(),
				OldHash:   oldHash,
				NewHash:   newHash,
			}
			r.triggerCallbacks(event)
		}
	}

	r.fileHashes = currentFiles
}

// scanFiles 扫描所有文件。
func (r *hotReloaderImpl) scanFiles() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.fileHashes = make(map[string]string)

	for _, dir := range r.watchDirs {
		r.scanDir(dir, r.fileHashes)
	}

	return nil
}

// scanDir 扫描目录。
func (r *hotReloaderImpl) scanDir(dir string, hashes map[string]string) {
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			baseName := filepath.Base(path)
			if r.ignoreDirs[baseName] {
				return filepath.SkipDir
			}
			return nil
		}

		if len(r.extensions) > 0 {
			ext := filepath.Ext(path)
			if !r.extensions[ext] {
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

// triggerCallbacks 触发回调。
func (r *hotReloaderImpl) triggerCallbacks(event ReloadEvent) {
	for _, callback := range r.callbacks {
		go callback(event)
	}
}

// GetWatchedFiles 获取所有被监控的文件。
func (r *hotReloaderImpl) GetWatchedFiles() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	files := make([]string, 0, len(r.fileHashes))
	for file := range r.fileHashes {
		files = append(files, file)
	}
	return files
}

// GetWatchDirs 获取监控目录。
func (r *hotReloaderImpl) GetWatchDirs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dirs := make([]string, len(r.watchDirs))
	copy(dirs, r.watchDirs)
	return dirs
}

// Restart 重启热重载。
func (r *hotReloaderImpl) Restart() error {
	r.Stop()
	return r.Start()
}

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
	mu         sync.Mutex
	watchDirs  []string
	extensions map[string]bool
	callbacks  []ReloadCallback
	stopChan   chan struct{}
	running    bool
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
	w.mu.Unlock()

	go w.pollFiles()

	return nil
}

// Stop 停止监控。
func (w *fileWatcherImpl) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	w.running = false
	close(w.stopChan)
}

// pollFiles 轮询文件。
func (w *fileWatcherImpl) pollFiles() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fileHashes := make(map[string]string)

	for _, dir := range w.watchDirs {
		w.scanDirForWatcher(dir, fileHashes)
	}

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			currentFiles := make(map[string]string)
			for _, dir := range w.watchDirs {
				w.scanDirForWatcher(dir, currentFiles)
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

					w.mu.Lock()
					for _, callback := range w.callbacks {
						go callback(event)
					}
					w.mu.Unlock()
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

					w.mu.Lock()
					for _, callback := range w.callbacks {
						go callback(event)
					}
					w.mu.Unlock()
				}
			}

			fileHashes = currentFiles
		}
	}
}

// scanDirForWatcher 扫描目录。
func (w *fileWatcherImpl) scanDirForWatcher(dir string, hashes map[string]string) {
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if len(w.extensions) > 0 {
			ext := filepath.Ext(path)
			if !w.extensions[ext] {
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
func NewLiveReloadServer(port int, reloader HotReloader) *LiveReloadServer {
	impl, _ := reloader.(*hotReloaderImpl)
	return &LiveReloadServer{
		port:     port,
		reloader: impl,
		stopChan: make(chan struct{}),
	}
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
