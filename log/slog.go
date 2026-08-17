package log

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// Option 定义 slog 日志配置选项
type Option func(*SlogLogger)

// WithLevel 设置日志级别
func WithLevel(level Level) Option {
	return func(l *SlogLogger) {
		l.level = level
	}
}

// WithFormat 设置输出格式
func WithFormat(format string) Option {
	return func(l *SlogLogger) {
		l.format = format
	}
}

// WithTimeFormat 设置时间格式（已废弃，slog 使用自己的时间格式）
// Deprecated: slog 使用自己的时间格式，此选项无效
func WithTimeFormat(timeFormat string) Option {
	return func(l *SlogLogger) {
		// 保留以保持 API 兼容性，但不再使用
	}
}

// WithAddSource 设置是否添加源码位置
func WithAddSource(addSource bool) Option {
	return func(l *SlogLogger) {
		l.addSource = addSource
	}
}

// WithOutput 设置输出 writer
func WithOutput(output io.Writer) Option {
	return func(l *SlogLogger) {
		l.output = output
	}
}

// WithOutputPath 设置日志文件输出路径
func WithOutputPath(path string) Option {
	return func(l *SlogLogger) {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return
		}
		// 关闭旧文件句柄（如果存在）
		if l.file != nil {
			_ = l.file.Close()
		}
		l.output = f
		l.file = f
		// 重新创建 handler 和 logger 以使用新的输出
		l.slogLevel = l.toSlogLevel(l.level)
		handlerOptions := &slog.HandlerOptions{
			Level:     l.slogLevel,
			AddSource: l.addSource,
		}
		if l.format == "text" {
			l.logger = slog.New(slog.NewTextHandler(l.output, handlerOptions))
		} else {
			l.logger = slog.New(slog.NewJSONHandler(l.output, handlerOptions))
		}
	}
}

// SlogLogger 是 slog 日志适配器,实现 Logger 接口
type SlogLogger struct {
	logger      *slog.Logger
	level       Level
	slogLevel   slog.Level // 预计算的 slog 级别，避免重复转换
	format      string     // 仅用于初始化时配置，With 方法不复制
	addSource   bool       // 仅用于初始化时配置，With 方法不复制
	output      io.Writer  // 仅用于初始化时配置，With 方法不复制
	file        *os.File   // 仅用于初始化和 Close 方法
	development bool       // 开发模式标志，DPanic 在开发模式下会 panic
}

// WithDevelopment 设置开发模式
func WithDevelopment(development bool) Option {
	return func(l *SlogLogger) {
		l.development = development
	}
}

// NewSlogLogger 创建 slog 日志适配器
func NewSlogLogger(opts ...Option) *SlogLogger {
	l := &SlogLogger{
		level:     InfoLevel,
		format:    "json",
		addSource: false,
		output:    os.Stdout,
	}

	for _, opt := range opts {
		opt(l)
	}

	// 预计算 slog 级别
	l.slogLevel = l.toSlogLevel(l.level)

	var handler slog.Handler
	handlerOptions := &slog.HandlerOptions{
		Level:     l.slogLevel,
		AddSource: l.addSource,
	}

	if l.format == "text" {
		handler = slog.NewTextHandler(l.output, handlerOptions)
	} else {
		handler = slog.NewJSONHandler(l.output, handlerOptions)
	}

	l.logger = slog.New(handler)
	return l
}

// 自定义 slog 级别，用于 panic/fatal（高于 slog.LevelError=8）
const (
	slogLevelDPanic = slog.LevelError + 2
	slogLevelPanic  = slog.LevelError + 4
	slogLevelFatal  = slog.LevelError + 6
)

// toSlogLevel 将日志级别转换为 slog 级别
func (l *SlogLogger) toSlogLevel(level Level) slog.Level {
	switch level {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	case DPanicLevel:
		return slogLevelDPanic
	case PanicLevel:
		return slogLevelPanic
	case FatalLevel:
		return slogLevelFatal
	default:
		return slog.LevelInfo
	}
}

// log 记录日志
func (l *SlogLogger) log(ctx context.Context, level Level, msg string, keys []KeyValue) {
	attrs := make([]any, 0, len(keys)*2)

	for _, kv := range keys {
		attrs = append(attrs, kv.Key, kv.Value)
	}
	l.logger.Log(ctx, l.toSlogLevel(level), msg, attrs...)
}

// logEnabled 快速检查日志级别是否启用
func (l *SlogLogger) logEnabled(level Level) bool {
	return level >= l.level
}

// IsLevelEnabled 检查指定级别是否启用（实现 LoggerLevelChecker 接口）
func (l *SlogLogger) IsLevelEnabled(level Level) bool {
	return l.logEnabled(level)
}

// Debug 记录调试日志
func (l *SlogLogger) Debug(ctx context.Context, msg string, keys ...KeyValue) {
	if l.logEnabled(DebugLevel) {
		l.log(ctx, DebugLevel, msg, keys)
	}
}

// Info 记录信息日志
func (l *SlogLogger) Info(ctx context.Context, msg string, keys ...KeyValue) {
	if l.logEnabled(InfoLevel) {
		l.log(ctx, InfoLevel, msg, keys)
	}
}

// Warn 记录警告日志
func (l *SlogLogger) Warn(ctx context.Context, msg string, keys ...KeyValue) {
	if l.logEnabled(WarnLevel) {
		l.log(ctx, WarnLevel, msg, keys)
	}
}

// Error 记录错误日志
func (l *SlogLogger) Error(ctx context.Context, msg string, keys ...KeyValue) {
	if l.logEnabled(ErrorLevel) {
		l.log(ctx, ErrorLevel, msg, keys)
	}
}

// DPanic 记录致命错误日志
//
// 在开发模式下会 panic，在生产模式下仅记录错误级别日志。
// 这符合 DPanic 的标准语义：用于检测不应发生的编程错误。
func (l *SlogLogger) DPanic(ctx context.Context, msg string, keys ...KeyValue) {
	if l.logEnabled(DPanicLevel) {
		l.log(ctx, DPanicLevel, msg, keys)
	}
	if l.development {
		panic(msg)
	}
}

// Panic 记录日志并 panic
func (l *SlogLogger) Panic(ctx context.Context, msg string, keys ...KeyValue) {
	if l.logEnabled(PanicLevel) {
		l.log(ctx, PanicLevel, msg, keys)
	}
	panic(msg)
}

// Fatal 记录致命级别日志
//
// 注意：与标准 log.Fatal 不同，此方法仅记录日志，不会调用 os.Exit(1)。
// 如需退出程序，调用方需自行处理。
func (l *SlogLogger) Fatal(ctx context.Context, msg string, keys ...KeyValue) {
	if l.logEnabled(FatalLevel) {
		l.log(ctx, FatalLevel, msg, keys)
	}
}

// Sync 同步日志缓冲区
func (l *SlogLogger) Sync() error {
	return nil
}

// Close 关闭日志文件句柄
func (l *SlogLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// With 返回带有额外字段的日志记录器
func (l *SlogLogger) With(ctx context.Context, keys ...KeyValue) Logger {
	attrs := make([]any, 0, len(keys)*2)
	for _, kv := range keys {
		attrs = append(attrs, kv.Key, kv.Value)
	}
	return &SlogLogger{
		logger:      l.logger.With(attrs...),
		level:       l.level,
		slogLevel:   l.slogLevel, // 继承预计算的 slog 级别
		development: l.development,
	}
}

var _ Logger = (*SlogLogger)(nil)
var _ LoggerFatal = (*SlogLogger)(nil)
