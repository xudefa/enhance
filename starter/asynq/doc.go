// Package asynq 提供 Asynq 异步任务队列自动配置。
//
// Asynq 是基于 Redis 的异步任务队列库，支持任务重试和定时任务。
// 本模块提供自动配置支持，内置任务调度器。
//
// 功能特性：
//   - 自动配置 Asynq 客户端和调度器
//   - 支持任务重试
//   - 支持定时任务
//   - 支持任务优先级
//
// 配置示例：
//
//	{
//	  "asynq": {
//	    "enabled": true,
//	    "host": "localhost",
//	    "port": 6379,
//	    "enable-scheduler": false,
//	    "concurrency": 10,
//	    "retry-limit": 25,
//	    "timeout": "30s"
//	  }
//	}
//
// 使用示例：
//
//	client := core.MustGetBean[*asynq.Client](app.Container())
//	task := asynq.NewTask("email:send", []byte(`{"to":"user@example.com"}`))
//	client.Enqueue(task)
package asynq

import (
	"errors"
	"time"
)

// Doc 包文档说明。
const Doc = "Asynq 异步任务队列自动配置 - 当 asynq.enabled=true 时自动生效"

// AsynqConfig Asynq 异步任务队列配置。
// 包含 Asynq 客户端和调度器的所有可配置参数。
type AsynqConfig struct {
	Enabled         bool          `json:"enabled" mapstructure:"enabled"`                   // 是否启用 Asynq
	Host            string        `json:"host" mapstructure:"host"`                         // Redis 主机地址
	Port            int           `json:"port" mapstructure:"port"`                         // Redis 端口
	Password        string        `json:"password" mapstructure:"password"`                 // Redis 密码
	DB              int           `json:"db" mapstructure:"db"`                             // Redis 数据库
	PoolSize        int           `json:"pool-size" mapstructure:"pool-size"`               // 连接池大小
	EnableScheduler bool          `json:"enable-scheduler" mapstructure:"enable-scheduler"` // 是否启用调度器
	Concurrency     int           `json:"concurrency" mapstructure:"concurrency"`           // Worker 并发数
	RetryLimit      int           `json:"retry-limit" mapstructure:"retry-limit"`           // 任务最大重试次数
	Timeout         time.Duration `json:"timeout" mapstructure:"timeout"`                   // 任务超时时间
}

const (
	// ConfigPrefix Asynq 配置前缀。
	ConfigPrefix = "asynq" // 配置前缀
	// AsynqEnabled 启用 Asynq 的条件配置键。
	AsynqEnabled = "asynq.enabled" // 启用条件配置键
	// DefaultAsynqHost 默认 Redis 主机。
	DefaultAsynqHost = "localhost" // 默认 Redis 主机
	// DefaultAsynqPort 默认 Redis 端口。
	DefaultAsynqPort = 6379 // 默认 Redis 端口
	// DefaultAsynqDB 默认 Redis 数据库编号。
	DefaultAsynqDB = 0 // 默认 Redis 数据库
	// DefaultAsynqPoolSize 默认连接池大小。
	DefaultAsynqPoolSize = 10 // 默认连接池大小
	// DefaultEnableScheduler 默认不启用调度器。
	DefaultEnableScheduler = false // 默认不启用调度器
	// DefaultConcurrency 默认 Worker 并发数。
	DefaultConcurrency = 10 // 默认 Worker 并发数
	// DefaultRetryLimit 默认任务最大重试次数。
	DefaultRetryLimit = 25 // 默认任务最大重试次数
	// DefaultTimeout 默认任务超时（0 表示不超时）。
	DefaultTimeout = 0 // 默认任务超时（0 表示不超时）
	// ConditionTrue 条件真值。
	ConditionTrue = "true" // 条件真值
	// StarterName Asynq 启动器名称。
	StarterName = "AsynqStarter" // 启动器名称
)

// ErrClientNotConfigured 客户端未配置错误。
// 在 Configure 未成功执行时调用 Enqueue 系列方法会返回此错误。
var (
	ErrClientNotConfigured = errors.New("asynq: client not configured")
)

var (
	// ErrInvalidHost 主机无效错误。
	ErrInvalidHost = errors.New("asynq: invalid host, must not be empty")
	// ErrInvalidPort 端口无效错误。
	ErrInvalidPort = errors.New("asynq: invalid port, must be in [1, 65535]")
	// ErrInvalidDB 数据库编号无效错误。
	ErrInvalidDB = errors.New("asynq: invalid db, must be >= 0")
	// ErrInvalidPoolSize 连接池大小无效错误。
	ErrInvalidPoolSize = errors.New("asynq: invalid pool-size, must be >= 0")
	// ErrInvalidConcurrency 并发数无效错误。
	ErrInvalidConcurrency = errors.New("asynq: invalid concurrency, must be >= 0")
	// ErrInvalidRetryLimit 重试次数无效错误。
	ErrInvalidRetryLimit = errors.New("asynq: invalid retry-limit, must be >= 0")
	// ErrInvalidTimeout 超时无效错误。
	ErrInvalidTimeout = errors.New("asynq: invalid timeout, must be >= 0")
)
