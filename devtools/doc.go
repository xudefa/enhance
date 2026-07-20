// Package devtools 提供开发工具支持，用于 enhance 框架。
//
// 该模块包含开发环境下的调试工具、性能分析、热重载等开发辅助功能。
// 仅在开发模式下启用，生产环境自动禁用。
//
// # 架构设计
//
//   - HotReloader: 热重载管理器接口，监控文件变更并触发重载
//   - FileWatcher: 文件监控器接口，监控文件变化
//   - DevModeDetector: 开发模式检测器接口
//   - ReloadEvent: 重载事件结构体，记录文件变更信息
//   - ReloadType: 重载类型枚举
//   - ReloadCallback: 重载回调函数类型
//   - HotReloaderOption: 热重载配置选项函数
//
// # 核心功能
//
//   - 文件监控: 监控指定目录的文件变更
//   - 热重载: 文件变更时触发回调函数
//   - 灵活配置: 支持自定义监控目录、扩展名、轮询间隔等
//
// # 使用方式
//
// 创建热重载器：
//
//	reloader := devtools.NewHotReloader(
//	    devtools.WithWatchDirs("./internal", "./pkg"),
//	    devtools.WithExtensions(".go"),
//	    devtools.WithInterval(time.Second),
//	)
//
// 启动监控：
//
//	reloader.Start(func(event devtools.ReloadEvent) {
//	    fmt.Printf("文件变更: %s (%s)\n", event.File, event.Type)
//	})
//
// # 环境变量
//
// 开发工具仅在开发模式下启用，通过环境变量控制：
//
//	export ENHANCE_DEV_MODE=true
package devtools

import (
	"time"
)

// ReloadType 重载类型。
type ReloadType string

// 重载类型常量。
const (
	// ReloadTypeCreated 文件创建。
	ReloadTypeCreated ReloadType = "CREATED"
	// ReloadTypeModified 文件修改。
	ReloadTypeModified ReloadType = "MODIFIED"
	// ReloadTypeDeleted 文件删除。
	ReloadTypeDeleted ReloadType = "DELETED"
)

// ReloadEvent 重载事件。
type ReloadEvent struct {
	// File 触发重载的文件。
	File string
	// Type 事件类型。
	Type ReloadType
	// Timestamp 事件时间。
	Timestamp time.Time
	// OldHash 旧文件哈希。
	OldHash string
	// NewHash 新文件哈希。
	NewHash string
}

// ReloadCallback 重载回调函数类型。
type ReloadCallback func(event ReloadEvent)

// HotReloaderOption 热重载器配置选项函数。
type HotReloaderOption func(HotReloader)

// HotReloader 热重载管理器接口。
//
// 通过轮询方式监控指定目录的文件变更，
// 当检测到文件创建、修改或删除时触发回调函数。
type HotReloader interface {
	// OnReload 注册重载回调。
	OnReload(callback ReloadCallback)

	// Start 启动文件监控。
	Start() error

	// Stop 停止文件监控。
	Stop()

	// IsRunning 检查是否正在运行。
	IsRunning() bool

	// GetWatchedFiles 获取所有被监控的文件。
	GetWatchedFiles() []string

	// GetWatchDirs 获取监控目录。
	GetWatchDirs() []string

	// Restart 重启热重载。
	Restart() error
}

// FileWatcher 文件监控器接口。
//
// 监控指定目录的文件变化，支持自定义扩展名过滤。
type FileWatcher interface {
	// OnChange 注册文件变化回调。
	OnChange(callback ReloadCallback)

	// Start 启动监控。
	Start() error

	// Stop 停止监控。
	Stop()
}

// DevModeDetector 开发模式检测器接口。
//
// 检测当前是否处于开发模式。
type DevModeDetector interface {
	// IsDevMode 检测是否为开发模式。
	IsDevMode() bool
}
