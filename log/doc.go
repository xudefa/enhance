// Package log 提供日志管理功能，用于 enhance 框架。
//
// 该模块提供统一的日志抽象接口，支持多种日志后端集成。
// 包含日志构建器、上下文日志、slog 集成等日志记录支持。
//
// # 架构设计
//
//   - Logger: 日志接口，定义统一的日志操作
//   - LoggerWithLevel: 支持自定义日志级别
//   - LoggerWithName: 支持日志命名
//   - LoggerWithCaller: 支持调用者信息
//   - LoggerWithTimeout: 支持超时日志
//   - Level: 日志级别枚举
//   - KeyValue: 日志键值对
//   - LoggerOption: 日志构建选项
//
// # 核心功能
//
//   - 统一接口: 提供统一的日志抽象接口
//   - 多后端: 支持 zap、zerolog、slog 等多种日志后端
//   - 上下文日志: 支持请求链路追踪和上下文传递
//   - 日志级别: 支持 DEBUG、INFO、WARN、ERROR 等级别
//
// # 使用方式
//
// 使用默认日志器：
//
//	logger := log.Build()
//	logger.Info(context.Background(), "Application started")
//
// 使用上下文日志：
//
//	ctxLogger := log.FromContext(ctx)
//	ctxLogger.Info(context.Background(), "Processing request", log.KeyValue{Key: "request_id", Value: reqID})
//
// # 集成后端
//
// 具体实现位于 starter 子包：
//
//   - starter/zap: Uber Zap 集成
//   - starter/zerolog: Zerolog 集成
//   - log/slog: Go 标准库 slog 集成
package log

import (
	"context"
	"sync/atomic"
	"time"
)

// Level 定义日志级别。
//
// 日志级别从低到高，用于控制日志输出的详细程度。
// 生产环境推荐使用 InfoLevel 或 WarnLevel。
//
// # 级别说明
//
//   - DebugLevel: 调试级别，详细的调试信息，生产环境通常禁用
//   - InfoLevel: 信息级别，一般运行信息，如启动/关闭
//   - WarnLevel: 警告级别，潜在问题，但不影响正常运行
//   - ErrorLevel: 错误级别，错误信息，需要关注但不影响核心功能
//   - DPanicLevel: 致命错误级别，开发环境 panic，生产环境仅记录错误
//   - PanicLevel: panic 级别，记录日志后 panic
//   - FatalLevel: 致命级别，记录日志后退出程序
type Level int8

// 日志级别常量定义。
const (
	DebugLevel  Level = iota // 调试级别：详细的调试信息，生产环境通常禁用
	InfoLevel                // 信息级别：一般运行信息，如启动/关闭
	WarnLevel                // 警告级别：潜在问题，但不影响正常运行
	ErrorLevel               // 错误级别：错误信息，需要关注但不影响核心功能
	DPanicLevel              // 致命错误级别：开发环境 panic，生产环境仅记录错误
	PanicLevel               // panic 级别：记录日志后 panic
	FatalLevel               // 致命级别：记录日志后退出程序
)

// KeyValue 定义日志键值对。
//
// 用于结构化日志记录，支持键值对形式的日志字段。
type KeyValue struct {
	Key   string // 字段名
	Value any    // 字段值
}

// Logger 是日志记录器接口，所有日志库都需实现此接口。
//
// 提供统一的日志记录 API，支持结构化日志和上下文传递。
// 实现可以基于 slog、zap、zerolog 等日志库。
//
// # 使用示例
//
//	logger := log.Build()
//	logger.Info(context.Background(), "Server started",
//	    log.KeyValue{Key: "port", Value: 8080})
//
//	// 带额外字段的日志
//	child := logger.With(context.Background(), log.KeyValue{Key: "module", Value: "auth"})
//	child.Info(context.Background(), "User login")
type Logger interface {
	// Debug 记录调试日志。
	Debug(ctx context.Context, msg string, keys ...KeyValue)
	// Info 记录信息日志。
	Info(ctx context.Context, msg string, keys ...KeyValue)
	// Warn 记录警告日志。
	Warn(ctx context.Context, msg string, keys ...KeyValue)
	// Error 记录错误日志。
	Error(ctx context.Context, msg string, keys ...KeyValue)
	// Sync 同步日志缓冲区，确保日志写入完成。
	Sync() error
	// With 返回带有额外字段的日志记录器。
	With(ctx context.Context, keys ...KeyValue) Logger
}

// LoggerFatal 支持致命日志和 panic。
//
// 提供 Panic、Fatal、DPanic 方法。
type LoggerFatal interface {
	Logger
	// DPanic 记录致命错误日志并在开发环境 panic。
	DPanic(ctx context.Context, msg string, keys ...KeyValue)
	// Panic 记录日志并 panic。
	Panic(ctx context.Context, msg string, keys ...KeyValue)
	// Fatal 记录日志并退出程序。
	Fatal(ctx context.Context, msg string, keys ...KeyValue)
}

// LoggerWithLevel 支持自定义日志级别。
type LoggerWithLevel interface {
	Logger
	Log(ctx context.Context, level Level, msg string, keys ...KeyValue)
}

// LoggerWithName 支持日志命名。
type LoggerWithName interface {
	Logger
	WithName(name string) Logger
}

// LoggerWithCaller 支持调用者信息。
type LoggerWithCaller interface {
	Logger
	WithCaller(skip int) Logger
}

// LoggerWithTimeout 支持超时日志。
type LoggerWithTimeout interface {
	Logger
	WithTimeout(d time.Duration) Logger
}

// Sampler 采样策略接口。
type Sampler interface {
	ShouldSample() bool
}

// RandomSampler 随机采样器。
type RandomSampler struct {
	rate float64
}

// ThresholdSampler 阈值采样器。
type ThresholdSampler struct {
	threshold int64
	counter   atomic.Int64
}

// SampledLogger 带采样的日志器。
type SampledLogger struct {
	logger  Logger
	sampler Sampler
}

// LoggerOption 定义日志记录器配置选项。
type LoggerOption func(*loggerConfig)

// loggerConfig 日志配置结构体（未导出）。
type loggerConfig struct {
	logger Logger
}
