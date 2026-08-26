package zerolog

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/rs/zerolog"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&ZerologAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(ZeroLogEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityInfrastructure)), // 基础设施层，最先执行，提供日志能力
	)
}

// ZerologAutoConfiguration Zerolog 自动配置类。
//
// 当配置文件中启用 Zerolog 时自动生效（zerolog.enabled=true）。
// 负责创建 Zerolog 实例并注册 log.Logger 到 IoC 容器中。
//
// 执行顺序：Order = -3000，确保在其他组件之前执行，提供日志能力。
type ZerologAutoConfiguration struct {
	logger log.Logger
}

// Configure 配置 Zerolog 日志器。
//
// 该方法在自动配置阶段调用，负责：
//  1. 从 Environment 中读取 Zerolog 配置（级别、格式、输出路径等）
//  2. 创建 Zerolog 实例
//  3. 注册 log.Logger 到 IoC 容器
func (c *ZerologAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Zerolog config: %w", err)
	}

	zerologLogger := c.buildZerolog(cfg)

	// 注册 Zerolog 适配器到 IoC 容器（标记为 Primary，确保其他组件优先获取到 zerolog）
	// 使用 Register 泛型函数确保 ConcreteType 是 log.Logger 接口类型，而不是 *ZerologLogger 具体类型
	if err := ctx.Container().RegisterInstance(zerologLogger, reflect.TypeFor[log.Logger]()); err != nil {
		return fmt.Errorf("failed to register Zerolog Logger: %w", err)
	}

	// 注册后使用刚创建的 zerolog 实例打印日志
	zerologLogger.Info(ctx.Context(), "Zerolog logger registered",
		log.KeyValue{Key: LogFieldLevel, Value: cfg.Level},
		log.KeyValue{Key: LogFieldFormat, Value: cfg.Format},
		log.KeyValue{Key: LogFieldOutput, Value: cfg.OutputPath},
	)

	return nil
}

// ZerologConfig Zerolog 日志配置。
type ZerologConfig struct {
	Enabled    bool   `json:"enabled" mapstructure:"enabled"`
	Level      string `json:"level" mapstructure:"level"`
	Format     string `json:"format" mapstructure:"format"`
	TimeFormat string `json:"time-format" mapstructure:"time-format"`
	AddSource  bool   `json:"add-source" mapstructure:"add-source"`
	OutputPath string `json:"output-path" mapstructure:"output-path"`
}

// loadConfig 从 Environment 加载 Zerolog 配置。
func (c *ZerologAutoConfiguration) loadConfig(env *environment.Environment) (*ZerologConfig, error) {
	cfg := &ZerologConfig{
		Level:      DefaultZeroLogLevel,
		Format:     DefaultZeroLogFormat,
		TimeFormat: DefaultZeroLogTimeFormat,
		AddSource:  DefaultZeroLogAddSource,
		OutputPath: DefaultZeroLogOutputPath,
	}

	if err := env.BindPrefix("log.zerolog", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind Zerolog config: %w", err)
	}

	return cfg, nil
}

// buildZerolog 构建 Zerolog 日志器。
func (c *ZerologAutoConfiguration) buildZerolog(cfg *ZerologConfig) log.Logger {
	// 设置输出
	var output io.Writer = os.Stdout
	var file *os.File
	if cfg.OutputPath != "" {
		f, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("[Zerolog] warning: failed to open log file, using stdout: path=%s, error=%v\n", cfg.OutputPath, err)
		} else {
			output = f
			file = f
		}
	}

	// 设置日志级别
	level := zerolog.InfoLevel
	switch cfg.Level {
	case LogLevelDebug:
		level = zerolog.DebugLevel
	case LogLevelInfo:
		level = zerolog.InfoLevel
	case LogLevelWarn, LogLevelWarning:
		level = zerolog.WarnLevel
	case LogLevelError:
		level = zerolog.ErrorLevel
	case LogLevelFatal:
		level = zerolog.FatalLevel
	case LogLevelPanic:
		level = zerolog.PanicLevel
	}

	// 创建 Zerolog 实例
	var logger zerolog.Logger
	if cfg.Format == LogFormatConsole {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: cfg.TimeFormat,
		}).Level(level).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(output).Level(level).With().Timestamp().Logger()
	}

	return &ZerologLogger{
		logger:    logger,
		level:     level,
		format:    cfg.Format,
		addSource: cfg.AddSource,
		output:    output,
		file:      file,
	}
}

// ZerologLogger 是 Zerolog 日志适配器，实现 log.Logger 接口。
type ZerologLogger struct {
	logger    zerolog.Logger
	level     zerolog.Level
	format    string
	addSource bool
	output    io.Writer
	file      *os.File
}

// Debug 记录调试日志。
func (l *ZerologLogger) Debug(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.log(ctx, zerolog.DebugLevel, msg, keys)
}

// Info 记录信息日志。
func (l *ZerologLogger) Info(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.log(ctx, zerolog.InfoLevel, msg, keys)
}

// Warn 记录警告日志。
func (l *ZerologLogger) Warn(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.log(ctx, zerolog.WarnLevel, msg, keys)
}

// Error 记录错误日志。
func (l *ZerologLogger) Error(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.log(ctx, zerolog.ErrorLevel, msg, keys)
}

// DPanic 记录致命错误日志并在开发环境 panic。
func (l *ZerologLogger) DPanic(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.log(ctx, zerolog.FatalLevel, msg, keys)
}

// Panic 记录日志并 panic。
func (l *ZerologLogger) Panic(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.log(ctx, zerolog.PanicLevel, msg, keys)
	panic(msg)
}

// Fatal 记录日志并退出程序。
func (l *ZerologLogger) Fatal(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.log(ctx, zerolog.FatalLevel, msg, keys)
	os.Exit(1)
}

// Sync 同步日志缓冲区。
func (l *ZerologLogger) Sync() error {
	return nil
}

// Close 关闭日志文件。
func (l *ZerologLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// With 返回带有额外字段的日志记录器。
func (l *ZerologLogger) With(ctx context.Context, keys ...log.KeyValue) log.Logger {
	contextLogger := l.logger.With()
	for _, kv := range keys {
		contextLogger = contextLogger.Interface(kv.Key, kv.Value)
	}
	return &ZerologLogger{
		logger:    contextLogger.Logger(),
		level:     l.level,
		format:    l.format,
		addSource: l.addSource,
		output:    l.output,
		file:      l.file,
	}
}

// log 记录日志。
func (l *ZerologLogger) log(ctx context.Context, level zerolog.Level, msg string, keys []log.KeyValue) {
	event := l.logger.WithLevel(level)

	for _, kv := range keys {
		event = event.Interface(kv.Key, kv.Value)
	}

	event.Msg(msg)
}

var _ log.Logger = (*ZerologLogger)(nil)
