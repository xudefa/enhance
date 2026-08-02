// Package zap 提供 Zap 高性能日志自动配置。
package zap

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&ZapAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(ZapEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityInfrastructure)),
	)
}

// ZapAutoConfiguration Zap 日志自动配置类。
type ZapAutoConfiguration struct{}

// Configure 配置 Zap 日志器。
func (c *ZapAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Zap config: %w", err)
	}

	logger := c.buildZap(cfg)

	if err := ctx.Container().RegisterInstance(logger, reflect.TypeFor[log.Logger]()); err != nil {
		return fmt.Errorf("failed to register Zap Logger: %w", err)
	}

	return nil
}

// ZapConfig Zap 日志配置。
type ZapConfig struct {
	Enabled    bool   `json:"enabled" mapstructure:"enabled"`
	Level      string `json:"level" mapstructure:"level"`
	Format     string `json:"format" mapstructure:"format"`
	OutputPath string `json:"output-path" mapstructure:"output-path"`
}

// ZapLogger 是 Zap 日志适配器，实现 log.Logger 接口。
type ZapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// Debug 记录调试日志。
func (l *ZapLogger) Debug(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.logger.Debug(msg, l.toFields(keys)...)
}

// Info 记录信息日志。
func (l *ZapLogger) Info(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.logger.Info(msg, l.toFields(keys)...)
}

// Warn 记录警告日志。
func (l *ZapLogger) Warn(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.logger.Warn(msg, l.toFields(keys)...)
}

// Error 记录错误日志。
func (l *ZapLogger) Error(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.logger.Error(msg, l.toFields(keys)...)
}

// DPanic 记录致命错误日志并在开发环境 panic。
func (l *ZapLogger) DPanic(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.logger.DPanic(msg, l.toFields(keys)...)
}

// Panic 记录日志并 panic。
func (l *ZapLogger) Panic(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.logger.Panic(msg, l.toFields(keys)...)
}

// Fatal 记录日志并退出程序。
func (l *ZapLogger) Fatal(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.logger.Fatal(msg, l.toFields(keys)...)
}

// Sync 同步日志缓冲区。
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

// With 返回带有额外字段的日志记录器。
func (l *ZapLogger) With(ctx context.Context, keys ...log.KeyValue) log.Logger {
	return &ZapLogger{
		logger: l.logger.With(l.toFields(keys)...),
		sugar:  l.sugar.With(l.toSugarArgs(keys)...),
	}
}

func (l *ZapLogger) toFields(keys []log.KeyValue) []zap.Field {
	fields := make([]zap.Field, len(keys))
	for i, kv := range keys {
		fields[i] = zap.Any(kv.Key, kv.Value)
	}
	return fields
}

func (l *ZapLogger) toSugarArgs(keys []log.KeyValue) []interface{} {
	args := make([]interface{}, 0, len(keys)*2)
	for _, kv := range keys {
		args = append(args, kv.Key, kv.Value)
	}
	return args
}

// 配置常量。
const (
	ZapEnabled           = "zap.enabled"
	DefaultZapLevel      = "info"
	DefaultZapFormat     = "json"
	DefaultZapOutputPath = "stdout"
	ConditionTrue        = "true"
)

// loadConfig 从 Environment 加载 Zap 配置。
func (c *ZapAutoConfiguration) loadConfig(env *environment.Environment) (*ZapConfig, error) {
	cfg := &ZapConfig{
		Level:      DefaultZapLevel,
		Format:     DefaultZapFormat,
		OutputPath: DefaultZapOutputPath,
	}

	if err := env.BindPrefix("log.zap", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind Zap config: %w", err)
	}

	return cfg, nil
}

// buildZap 构建 Zap 日志器。
func (c *ZapAutoConfiguration) buildZap(cfg *ZapConfig) log.Logger {
	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn", "warning":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	}

	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	if cfg.Format == "console" {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	var writer zapcore.WriteSyncer
	if cfg.OutputPath == "stdout" {
		writer = zapcore.AddSync(os.Stdout)
	} else {
		writer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.OutputPath,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		})
	}

	core := zapcore.NewCore(encoder, writer, level)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &ZapLogger{
		logger: logger,
		sugar:  logger.Sugar(),
	}
}
