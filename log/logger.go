// Package log 提供日志管理功能。
package log

// String 返回日志级别的字符串表示。
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case DPanicLevel:
		return "dpanic"
	case PanicLevel:
		return "panic"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// WithLogger 设置日志记录器。
func WithLogger(logger Logger) LoggerOption {
	return func(c *loggerConfig) {
		c.logger = logger
	}
}

// Build 使用选项构建日志记录器。
// 未指定 WithLogger 时默认创建基于 slog 的日志器。
func Build(opts ...LoggerOption) Logger {
	cfg := &loggerConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.logger == nil {
		cfg.logger = NewSlogLogger(WithFormat("text"), WithLevel(InfoLevel))
	}
	return cfg.logger
}

// ToLevel 将字符串转换为日志级别。
func ToLevel(level string) Level {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "dpanic":
		return DPanicLevel
	case "panic":
		return PanicLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}
