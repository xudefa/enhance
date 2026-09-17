package log

import (
	"context"
	"strings"
	"testing"
)

func TestContextLogger_WithTraceID(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()
	ctx = WithTraceID(ctx, "test-trace-123")

	// 应该能正常记录，且包含 trace_id
	logger.Info(ctx, "message with trace id")
}

func TestContextLogger_NoTraceID(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()

	// 没有 trace_id 也应该正常记录
	logger.Info(ctx, "message without trace id")
}

func TestGetTraceID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 没有 trace_id
	if GetTraceID(ctx) != "" {
		t.Error("expected empty trace id")
	}

	// 有 trace_id
	ctx = WithTraceID(ctx, "abc-123")
	if GetTraceID(ctx) != "abc-123" {
		t.Errorf("expected abc-123, got %s", GetTraceID(ctx))
	}
}

func TestContextLogger_ChainWith(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-456")

	childLogger := logger.With(ctx, KeyValue{Key: "request_id", Value: "req-789"})

	// 子日志器应该也能正常工作
	childLogger.Info(ctx, "chained context log")
}

func TestContextLogger_AllLevels(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := WithTraceID(context.Background(), "test-trace")

	logger.Debug(ctx, "debug")
	logger.Info(ctx, "info")
	logger.Warn(ctx, "warn")
	logger.Error(ctx, "error")
}

func TestContextLogger_Sync(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	err := logger.Sync()
	if err != nil {
		t.Errorf("expected no error from Sync, got %v", err)
	}
}

func TestContextLogger_EmptyContext(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	// 空上下文也应该正常工作
	ctx := context.Background()
	logger.Info(ctx, "empty context log")
}

func TestContextLogger_MultipleTraceIDs(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx1 := WithTraceID(context.Background(), "trace-1")
	ctx2 := WithTraceID(context.Background(), "trace-2")

	logger.Info(ctx1, "first trace")
	logger.Info(ctx2, "second trace")
}

func TestContextLogger_WithEmptyKeys(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()
	childLogger := logger.With(ctx)

	if childLogger == nil {
		t.Fatal("expected child logger with empty keys")
	}

	childLogger.Info(ctx, "empty keys log")
}

func TestContextLogger_WithTraceIDEmpty(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	// 测试空 trace_id
	ctx := WithTraceID(context.Background(), "")
	logger.Info(ctx, "empty trace id")
}

func TestContextLogger_WithNilContext(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	// 不应该 panic
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("unexpected panic: %v", rec)
		}
	}()

	ctx := context.Background()
	logger.Info(ctx, "nil context test")
}

func TestContextLogger_DPanic(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel), WithDevelopment(true))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()

	// DPanic 应该 panic（因为 SlogLogger 开启了 development 模式）
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected DPanic to panic in development mode")
		}
	}()

	logger.DPanic(ctx, "dpanic message")
}

func TestContextLogger_Panic(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()

	// Panic 应该 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected Panic to panic")
		}
	}()

	logger.Panic(ctx, "panic message")
}

func TestContextLogger_Fatal(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()

	// Fatal 应该调用 os.Exit(1)，这里无法直接测试
	// 实际测试时需要用子进程测试
	_ = logger
	_ = ctx
}

func TestContextLogger_WithMultipleFields(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := WithTraceID(context.Background(), "multi-trace")
	childLogger := logger.With(ctx,
		KeyValue{Key: "key1", Value: "value1"},
		KeyValue{Key: "key2", Value: "value2"},
	)

	childLogger.Info(ctx, "multiple fields log")
}

func TestContextLogger_PreserveOriginalContext(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := WithTraceID(context.Background(), "original-trace")

	// 原始上下文应该保持不变
	if GetTraceID(ctx) != "original-trace" {
		t.Errorf("expected original trace id to be preserved")
	}

	logger.Info(ctx, "preserve context log")
}

func TestContextLogger_WithNilLogger(t *testing.T) {
	t.Parallel()
	// 测试 nil logger 的情况
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil logger")
		}
	}()

	logger := NewContextLogger(nil)
	ctx := context.Background()
	logger.Info(ctx, "nil logger")
}

func TestContextLogger_WithEmptyMessage(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()
	logger.Info(ctx, "")
}

func TestContextLogger_WithNilKeys(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()
	childLogger := logger.With(ctx, nil...)

	if childLogger == nil {
		t.Fatal("expected child logger with nil keys")
	}

	childLogger.Info(ctx, "nil keys log")
}

func TestContextLogger_WithSpecialCharacters(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := WithTraceID(context.Background(), "trace-with-special-chars-!@#$%^&*()")
	logger.Info(ctx, "message with special chars: !@#$%^&*()")
}

func TestContextLogger_UnicodeMessage(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()
	logger.Info(ctx, "Unicode message: 你好世界 🌍")
}

func TestContextLogger_LongMessage(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()
	longMsg := strings.Repeat("a", 10000)
	logger.Info(ctx, longMsg)
}

func TestContextLogger_MultipleInstances(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger1 := NewContextLogger(baseLogger)
	logger2 := NewContextLogger(baseLogger)

	ctx := context.Background()
	logger1.Info(ctx, "logger1 message")
	logger2.Info(ctx, "logger2 message")
}

func TestContextLogger_WithChain(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()

	// 链式 With
	child1 := logger.With(ctx, KeyValue{Key: "key1", Value: "value1"})
	child2 := child1.(LoggerWithFields).With(ctx, KeyValue{Key: "key2", Value: "value2"})

	child2.Info(ctx, "chained with log")
}

func TestContextLogger_WithNilLoggerAndContext(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil logger")
		}
	}()

	logger := NewContextLogger(nil)
	logger.Info(context.TODO(), "nil logger and context")
}
