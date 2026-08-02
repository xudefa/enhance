// Package devtools 提供开发工具支持，用于 enhance 框架。
package devtools

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
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
	wg           sync.WaitGroup
	callbackWg   sync.WaitGroup
}

// computeFileHash 计算文件 MD5 哈希值。
func computeFileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
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
func NewHotReloader(opts ...HotReloaderOption) HotReloaderInfo {
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
	r.stopChan = make(chan struct{})
	stopChan := r.stopChan
	r.mu.Unlock()

	if err := r.scanFiles(); err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	r.wg.Add(1)
	go r.poll(stopChan)

	return nil
}

// Stop 停止文件监控。
func (r *hotReloaderImpl) Stop() {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return
	}

	r.running = false
	select {
	case <-r.stopChan:
	default:
		close(r.stopChan)
	}
	r.mu.Unlock()

	r.wg.Wait()
	r.callbackWg.Wait()
}

// IsRunning 检查是否正在运行。
func (r *hotReloaderImpl) IsRunning() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.running
}

// poll 轮询文件变化。
func (r *hotReloaderImpl) poll(stopChan chan struct{}) {
	defer r.wg.Done()
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			r.checkForChanges()
		}
	}
}

// checkForChanges 检查文件变化。
func (r *hotReloaderImpl) checkForChanges() {
	// 在锁外执行文件系统扫描，避免锁内I/O
	currentFiles := make(map[string]string)
	r.mu.RLock()
	dirs := make([]string, len(r.watchDirs))
	copy(dirs, r.watchDirs)
	r.mu.RUnlock()

	for _, dir := range dirs {
		r.scanDir(dir, currentFiles)
	}

	// 在锁内收集事件并更新状态
	r.mu.Lock()

	var events []ReloadEvent

	for file, oldHash := range r.fileHashes {
		if _, exists := currentFiles[file]; !exists {
			events = append(events, ReloadEvent{
				File:      file,
				Type:      ReloadTypeDeleted,
				Timestamp: time.Now(),
				OldHash:   oldHash,
			})
		}
	}

	for file, newHash := range currentFiles {
		oldHash, exists := r.fileHashes[file]
		if !exists {
			events = append(events, ReloadEvent{
				File:      file,
				Type:      ReloadTypeCreated,
				Timestamp: time.Now(),
				NewHash:   newHash,
			})
		} else if oldHash != newHash {
			events = append(events, ReloadEvent{
				File:      file,
				Type:      ReloadTypeModified,
				Timestamp: time.Now(),
				OldHash:   oldHash,
				NewHash:   newHash,
			})
		}
	}

	r.fileHashes = currentFiles
	r.mu.Unlock()

	// 在锁外触发回调，避免锁内执行用户代码
	for i := range events {
		r.triggerCallbacks(events[i])
	}
}

// scanFiles 扫描所有文件。
func (r *hotReloaderImpl) scanFiles() error {
	r.mu.RLock()
	dirs := make([]string, len(r.watchDirs))
	copy(dirs, r.watchDirs)
	r.mu.RUnlock()

	newHashes := make(map[string]string)
	for _, dir := range dirs {
		r.scanDir(dir, newHashes)
	}

	r.mu.Lock()
	r.fileHashes = newHashes
	r.mu.Unlock()

	return nil
}

// scanDir 扫描目录。
func (r *hotReloaderImpl) scanDir(dir string, hashes map[string]string) {
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
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
	r.mu.RLock()
	callbacks := make([]ReloadCallback, len(r.callbacks))
	copy(callbacks, r.callbacks)
	r.mu.RUnlock()

	for _, callback := range callbacks {
		r.callbackWg.Add(1)
		go func(cb ReloadCallback) {
			defer r.callbackWg.Done()
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("[devtools] hot reload callback panic: %v\n", r)
				}
			}()
			cb(event)
		}(callback)
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
