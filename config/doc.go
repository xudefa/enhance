// Package config 提供配置管理功能，用于 enhance 框架。
//
// 该模块提供配置加载、绑定、验证、热更新等功能。
// 参考 Spring Boot 的配置管理体系设计。
//
// # 架构设计
//
//   - Config: 配置接口，提供统一的配置访问
//   - Binder: 配置绑定器，支持结构体绑定
//   - Validator: 配置验证器接口
//   - WatchManager: 配置热重载管理器
//   - WatchEvent: 配置变更事件
//   - WatchCallback: 配置变更回调函数类型
//   - TypeConverter: 类型转换函数类型
//   - ValidationError: 单个验证错误
//   - ValidationErrors: 验证错误集合
//   - ValidationRule: 验证规则
//
// # 核心功能
//
//   - 配置加载: 支持 JSON、YAML 等多种格式
//   - 配置绑定: 支持结构体自动绑定
//   - 配置验证: 支持配置项验证
//   - 热更新: 支持配置变更监听和自动刷新
//   - 环境变量: 支持环境变量覆盖
//
// # 使用方式
//
// 加载配置：
//
//	cfg := config.NewConfig("application.json")
//
// 绑定到结构体：
//
//	type AppConfig struct {
//	    Port int    `config:"server.port"`
//	    Name string `config:"app.name"`
//	}
//
//	var appCfg AppConfig
//	cfg.Bind(&appCfg)
//
// 监听配置变更：
//
//	cfg.Watch(func(key string, value any) {
//	    fmt.Printf("Config changed: %s = %v\n", key, value)
//	})
//
// # 配置优先级
//
// 配置加载优先级（从高到低）：
//   - 命令行参数
//   - 环境变量
//   - 配置文件
//   - 默认值
package config

import (
	"sync"
	"time"
)

// Config 配置接口。
//
// 提供统一的配置访问和管理功能。
// 支持多种数据类型（字符串、整数、布尔值等），
// 以及配置的加载和持久化。
//
// # 使用示例
//
//	cfg := config.NewConfig()
//	cfg.Set("app.name", "my-app")
//	cfg.Set("server.port", 8080)
//	name := cfg.GetString("app.name")
//	port := cfg.GetInt("server.port")
type Config interface {
	// Get 获取指定键的原始值
	Get(key string) any
	// GetString 获取字符串值，类型不匹配时返回空字符串
	GetString(key string) string
	// GetInt 获取整数值，类型不匹配时返回 0
	GetInt(key string) int
	// GetBool 获取布尔值，类型不匹配时返回 false
	GetBool(key string) bool
	// GetAll 获取所有配置项的副本
	GetAll() map[string]any
	// Set 设置配置值
	Set(key string, value any)
	// Load 从 JSON 文件加载配置
	Load(path string) error
	// Save 保存配置到 JSON 文件
	Save(path string) error
}

// Validator 配置验证器接口。
//
// 用于在配置绑定后验证配置值的合法性。
type Validator interface {
	// Validate 验证配置
	Validate(data map[string]any) error
}

// Loader 配置加载器接口。
//
// 定义配置加载的统一方式,用于支持多种配置源。
type Loader interface {
	// Load 加载配置，返回加载后的配置对象。
	Load(opts ...LoaderOption) (Config, error)
	// Priority 返回加载器优先级，值越大优先级越高。
	Priority() int
	// SupportsWatch 返回是否支持配置热重载。
	SupportsWatch() bool
}

// LoaderOption 配置加载器选项函数类型。
type LoaderOption func(*LoaderModel) error

// LoaderModel 配置加载器模型结构体。
type LoaderModel struct {
	Paths      []string
	FileName   string
	FileType   string
	Env        string
	Prefix     string
	RemoteType string
	Endpoints  []string
	Key        string
}

// ConfigData 配置数据，键值对形式存储从配置中心加载的配置项。
type ConfigData map[string]any

// ConfigCenter 配置中心接口。
//
// 定义远程配置的加载、监听和关闭行为。
type ConfigCenter interface {
	// Load 从配置中心加载配置数据。
	Load() (ConfigData, error)
	// Watch 监听配置变更，变更时调用回调函数。
	Watch(key string, callback func(ConfigData)) error
	Close() error
}

// ConfigCenterConfig 配置中心连接配置。
type ConfigCenterConfig struct {
	Endpoints []string
	Namespace string
	Timeout   time.Duration
	DataID    string
	Group     string
	Prefix    string
}

// ValidationError 单个验证错误。
type ValidationError struct {
	Field   string
	Message string
}

// ValidationErrors 验证错误集合。
type ValidationErrors []ValidationError

// ValidationRule 验证规则。
type ValidationRule struct {
	Field string
	Check func(value any) error
}

// WatchCallback 配置变更回调函数类型。
type WatchCallback func(event WatchEvent)

// WatchEvent 配置变更事件。
type WatchEvent struct {
	Type      string    // modify, delete, create
	Key       string    // 配置键
	Value     any       // 配置值
	Timestamp time.Time // 事件时间
	Source    string    // 事件来源
}

// WatchManager 配置热重载管理器。
//
// 管理配置变更的监听和通知。
type WatchManager struct {
	callbacks map[string]func(WatchEvent)
	mu        sync.RWMutex
	sources   map[string]chan WatchEvent
	closed    bool
}

// memoryConfig 内存配置实现（未导出，实现 Config 接口）。
type memoryConfig struct {
	mu   sync.RWMutex
	data map[string]any
}

// 热重载事件类型常量。
const (
	EventModify = "modify" // 配置修改
	EventDelete = "delete" // 配置删除
	EventCreate = "create" // 配置创建
)
