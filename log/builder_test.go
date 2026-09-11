package log

import (
	"context"
	"strings"
	"testing"
)

func TestLoggerBuilder_Build(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().
		Level(DebugLevel).
		Format("json").
		AddSource(false).
		Build()

	if logger == nil {
		t.Fatal("expected logger to be created")
	}

	ctx := context.Background()
	// 应该能正常记录日志
	logger.Debug(ctx, "test debug message")
	logger.Info(ctx, "test info message")
}

func TestLoggerBuilder_WithName(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().
		Name("my-service").
		Build()

	if logger == nil {
		t.Fatal("expected logger to be created")
	}

	ctx := context.Background()
	logger.Info(ctx, "test message with name")
}

func TestLoggerBuilder_WithSampler(t *testing.T) {
	t.Parallel()
	// 10% 采样率
	sampler := NewRandomSampler(0.1)
	logger := NewLoggerBuilder().
		Level(DebugLevel).
		Sampler(sampler).
		Build()

	if logger == nil {
		t.Fatal("expected logger to be created")
	}

	ctx := context.Background()
	// 多次调用，部分应该被采样
	for range 100 {
		logger.Debug(ctx, "sampled debug message")
	}
}

func TestLoggerBuilder_OutputPath(t *testing.T) {
	t.Parallel()
	// 测试无效路径（应该回退到 stdout）
	logger := NewLoggerBuilder().
		Level(InfoLevel).
		OutputPath("/nonexistent/path/test.log").
		Build()

	if logger == nil {
		t.Fatal("expected logger to be created even with invalid path")
	}
}

func TestLoggerBuilder_DefaultValues(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().Build()

	if logger == nil {
		t.Fatal("expected logger with default values")
	}

	ctx := context.Background()
	logger.Info(ctx, "default config log")
}

func TestLoggerBuilder_ComplexConfig(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().
		Name("complex-service").
		Level(WarnLevel).
		Format("text").
		AddSource(false).
		Sampler(NewThresholdSampler(10)).
		Build()

	if logger == nil {
		t.Fatal("expected logger with complex config")
	}

	ctx := WithTraceID(context.Background(), "complex-trace")
	logger.Warn(ctx, "complex config log")
}

func TestLoggerBuilder_StringFormat(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().
		Format("text").
		Build()

	if logger == nil {
		t.Fatal("expected text format logger")
	}

	ctx := context.Background()
	logger.Info(ctx, "text format log")
}

func TestLoggerBuilder_JSONFormat(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().
		Format("json").
		Build()

	if logger == nil {
		t.Fatal("expected json format logger")
	}

	ctx := context.Background()
	logger.Info(ctx, "json format log")
}

func TestLoggerBuilder_WithOutputPath(t *testing.T) {
	t.Parallel()
	// 测试临时文件路径
	logger := NewLoggerBuilder().
		OutputPath("/tmp/test-log-builder.log").
		Build()

	if logger == nil {
		t.Fatal("expected logger with output path")
	}

	ctx := context.Background()
	logger.Info(ctx, "output path log")

	// 清理
	if syncer, ok := logger.(LoggerWithSync); ok {
		_ = syncer.Sync()
	}
}

func TestLoggerBuilder_LevelString(t *testing.T) {
	t.Parallel()
	// 测试级别字符串
	levels := []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel, PanicLevel, FatalLevel}
	expected := []string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"}

	for i, level := range levels {
		if level.String() != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], level.String())
		}
	}
}

func TestToLevel_Conversion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", DebugLevel},
		{"info", InfoLevel},
		{"warn", WarnLevel},
		{"warning", WarnLevel},
		{"error", ErrorLevel},
		{"dpanic", DPanicLevel},
		{"panic", PanicLevel},
		{"fatal", FatalLevel},
		{"unknown", InfoLevel}, // 默认
	}

	for _, tt := range tests {
		if ToLevel(tt.input) != tt.expected {
			t.Errorf("ToLevel(%s) = %v, expected %v", tt.input, ToLevel(tt.input), tt.expected)
		}
	}
}

func TestLoggerBuilder_NilLogger(t *testing.T) {
	t.Parallel()
	// 测试 Build 函数在没有指定 logger 时的默认行为
	logger := Build()

	if logger == nil {
		t.Fatal("expected default logger")
	}

	ctx := context.Background()
	logger.Info(ctx, "default logger")
}

func TestLoggerBuilder_WithCustomLogger(t *testing.T) {
	t.Parallel()
	customLogger := NewSlogLogger(WithLevel(DebugLevel), WithFormat("text"))
	logger := Build(WithLogger(customLogger))

	if logger == nil {
		t.Fatal("expected custom logger")
	}

	ctx := context.Background()
	logger.Info(ctx, "custom logger")
}

func TestLoggerBuilder_SamplerNil(t *testing.T) {
	t.Parallel()
	// 测试 sampler 为 nil 的情况
	logger := NewLoggerBuilder().
		Level(InfoLevel).
		Build()

	if logger == nil {
		t.Fatal("expected logger without sampler")
	}

	ctx := context.Background()
	logger.Info(ctx, "no sampler")
}

func TestLoggerBuilder_ChainMultipleOptions(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().
		Name("chain-test").
		Level(DebugLevel).
		Format("json").
		AddSource(false).
		Sampler(NewRandomSampler(0.5)).
		Build()

	if logger == nil {
		t.Fatal("expected chained options logger")
	}

	ctx := WithTraceID(context.Background(), "chain-trace")
	logger.Info(ctx, "chained options log")
}

func TestLoggerBuilder_DefaultFormat(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().Build()

	// 默认应该是 json 格式
	if logger == nil {
		t.Fatal("expected default format logger")
	}

	ctx := context.Background()
	logger.Info(ctx, "default format log")
}

func TestLoggerBuilder_OutputPathEmpty(t *testing.T) {
	t.Parallel()
	// 测试空输出路径
	logger := NewLoggerBuilder().
		OutputPath("").
		Build()

	if logger == nil {
		t.Fatal("expected logger with empty output path")
	}

	ctx := context.Background()
	logger.Info(ctx, "empty output path log")
}

func TestLoggerBuilder_ComplexConfigWithAllOptions(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().
		Name("full-config").
		Level(DebugLevel).
		Format("json").
		AddSource(false).
		OutputPath("/tmp/test-full-config.log").
		Sampler(NewRandomSampler(0.8)).
		Build()

	if logger == nil {
		t.Fatal("expected full config logger")
	}

	ctx := WithTraceID(context.Background(), "full-trace")
	logger.Info(ctx, "full config log")

	if syncer, ok := logger.(LoggerWithSync); ok {
		_ = syncer.Sync()
	}
}

func TestLoggerBuilder_InvalidFormat(t *testing.T) {
	t.Parallel()
	// 测试无效格式（应该回退到默认格式）
	logger := NewLoggerBuilder().
		Format("invalid").
		Build()

	if logger == nil {
		t.Fatal("expected logger with invalid format")
	}

	ctx := context.Background()
	logger.Info(ctx, "invalid format log")
}

func TestLoggerBuilder_LevelDebug(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().
		Level(DebugLevel).
		Build()

	ctx := context.Background()
	logger.Debug(ctx, "debug level log")
}

func TestLoggerBuilder_LevelInfo(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().
		Level(InfoLevel).
		Build()

	ctx := context.Background()
	logger.Info(ctx, "info level log")
}

func TestLoggerBuilder_LevelWarn(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().
		Level(WarnLevel).
		Build()

	ctx := context.Background()
	logger.Warn(ctx, "warn level log")
}

func TestLoggerBuilder_LevelError(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().
		Level(ErrorLevel).
		Build()

	ctx := context.Background()
	logger.Error(ctx, "error level log")
}

func TestLoggerBuilder_UnicodeMessage(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().
		Level(InfoLevel).
		Build()

	ctx := context.Background()
	logger.Info(ctx, "Unicode message: 你好世界 🌍")
}

func TestLoggerBuilder_LongMessage(t *testing.T) {
	t.Parallel()
	logger := NewLoggerBuilder().
		Level(InfoLevel).
		Build()

	ctx := context.Background()
	longMsg := strings.Repeat("a", 10000)
	logger.Info(ctx, longMsg)
}

func TestLoggerBuilder_MultipleBuilds(t *testing.T) {
	t.Parallel()
	logger1 := NewLoggerBuilder().Name("logger1").Build()
	logger2 := NewLoggerBuilder().Name("logger2").Build()

	if logger1 == logger2 {
		t.Error("expected different logger instances")
	}

	ctx := context.Background()
	logger1.Info(ctx, "logger1 message")
	logger2.Info(ctx, "logger2 message")
}

func TestLoggerBuilder_BuildIdempotent(t *testing.T) {
	t.Parallel()
	builder := NewLoggerBuilder().Name("idempotent")

	logger1 := builder.Build()
	logger2 := builder.Build()

	if logger1 == logger2 {
		t.Error("expected different logger instances from same builder")
	}

	ctx := context.Background()
	logger1.Info(ctx, "logger1 message")
	logger2.Info(ctx, "logger2 message")
}

func TestLoggerBuilder_WithNilOptions(t *testing.T) {
	t.Parallel()
	// 测试 nil options
	logger := NewLoggerBuilder().Build()

	if logger == nil {
		t.Fatal("expected logger with nil options")
	}

	ctx := context.Background()
	logger.Info(ctx, "nil options log")
}
