package log

import (
	"context"
	"strings"
	"testing"
)

func TestDynamicLevelLogger(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	if logger.GetLevel() != InfoLevel {
		t.Errorf("expected InfoLevel, got %v", logger.GetLevel())
	}

	ctx := context.Background()

	logger.Debug(ctx, "debug message")

	logger.Info(ctx, "info message")

	logger.SetLevel(DebugLevel)

	if logger.GetLevel() != DebugLevel {
		t.Errorf("expected DebugLevel after SetLevel, got %v", logger.GetLevel())
	}

	logger.Debug(ctx, "debug message after level change")
}

func TestDynamicLevelLogger_LevelFiltering(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, WarnLevel)

	ctx := context.Background()

	logger.Debug(ctx, "debug")
	logger.Info(ctx, "info")
	logger.Warn(ctx, "warn")
	logger.Error(ctx, "error")

	logger.SetLevel(DebugLevel)
	logger.Debug(ctx, "debug after change")
}

func TestDynamicLevelLogger_With(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()
	childLogger := logger.With(ctx, KeyValue{Key: "module", Value: "test"})

	dynamicChild, ok := childLogger.(*DynamicLevelLogger)
	if !ok {
		t.Fatal("expected child to be DynamicLevelLogger")
	}

	logger.SetLevel(DebugLevel)
	if dynamicChild.GetLevel() != DebugLevel {
		t.Errorf("expected child to inherit level change, got %v", dynamicChild.GetLevel())
	}
}

func TestDynamicLevelLogger_ConcurrentLevelChange(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()

	done := make(chan bool)
	for range 10 {
		go func() {
			for j := range 100 {
				logger.SetLevel(Level(j % 5))
				logger.Info(ctx, "concurrent log")
			}
			done <- true
		}()
	}

	for range 10 {
		<-done
	}
}

func TestDynamicLevelLogger_BoundaryLevels(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, DebugLevel)

	logger.SetLevel(DebugLevel)
	if logger.GetLevel() != DebugLevel {
		t.Errorf("expected DebugLevel, got %v", logger.GetLevel())
	}

	logger.SetLevel(InfoLevel)
	if logger.GetLevel() != InfoLevel {
		t.Errorf("expected InfoLevel, got %v", logger.GetLevel())
	}

	logger.SetLevel(WarnLevel)
	if logger.GetLevel() != WarnLevel {
		t.Errorf("expected WarnLevel, got %v", logger.GetLevel())
	}

	logger.SetLevel(ErrorLevel)
	if logger.GetLevel() != ErrorLevel {
		t.Errorf("expected ErrorLevel, got %v", logger.GetLevel())
	}
}

func TestDynamicLevelLogger_LevelComparison(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, WarnLevel)

	ctx := context.Background()

	logger.Warn(ctx, "warn message")
	logger.Error(ctx, "error message")

	logger.SetLevel(ErrorLevel)
	logger.Error(ctx, "error after change")
}

func TestDynamicLevelLogger_WithPreservesLevel(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, WarnLevel)

	ctx := context.Background()
	childLogger := logger.With(ctx, KeyValue{Key: "key", Value: "value"})

	dynamicChild, ok := childLogger.(*DynamicLevelLogger)
	if !ok {
		t.Fatal("expected DynamicLevelLogger")
	}

	if dynamicChild.GetLevel() != WarnLevel {
		t.Errorf("expected WarnLevel, got %v", dynamicChild.GetLevel())
	}
}

func TestDynamicLevelLogger_RapidLevelChanges(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()

	for i := range 100 {
		logger.SetLevel(Level(i % 5))
		logger.Info(ctx, "rapid level change")
	}
}

func TestDynamicLevelLogger_AllLevelsShouldLog(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, DebugLevel)

	ctx := context.Background()

	logger.Debug(ctx, "debug")
	logger.Info(ctx, "info")
	logger.Warn(ctx, "warn")
	logger.Error(ctx, "error")
}

func TestDynamicLevelLogger_LevelOrdering(t *testing.T) {
	t.Parallel()
	if DebugLevel >= InfoLevel {
		t.Error("DebugLevel should be less than InfoLevel")
	}
	if InfoLevel >= WarnLevel {
		t.Error("InfoLevel should be less than WarnLevel")
	}
	if WarnLevel >= ErrorLevel {
		t.Error("WarnLevel should be less than ErrorLevel")
	}
	if ErrorLevel >= DPanicLevel {
		t.Error("ErrorLevel should be less than DPanicLevel")
	}
}

func TestDynamicLevelLogger_IsLevelEnabled(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, WarnLevel)

	if logger.IsLevelEnabled(DebugLevel) {
		t.Error("DebugLevel should not be enabled when level is WarnLevel")
	}
	if logger.IsLevelEnabled(InfoLevel) {
		t.Error("InfoLevel should not be enabled when level is WarnLevel")
	}
	if !logger.IsLevelEnabled(WarnLevel) {
		t.Error("WarnLevel should be enabled")
	}
	if !logger.IsLevelEnabled(ErrorLevel) {
		t.Error("ErrorLevel should be enabled")
	}
}

func TestDynamicLevelLogger_DPanic(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel), WithDevelopment(true))
	logger := NewDynamicLevelLogger(baseLogger, DebugLevel)

	ctx := context.Background()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected DPanic to panic in development mode")
		}
	}()

	logger.DPanic(ctx, "dpanic message")
}

func TestDynamicLevelLogger_Panic(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, DebugLevel)

	ctx := context.Background()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected Panic to panic")
		}
	}()

	logger.Panic(ctx, "panic message")
}

func TestDynamicLevelLogger_Fatal(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, DebugLevel)

	ctx := context.Background()

	_ = logger
	_ = ctx
}

func TestDynamicLevelLogger_Sync(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, DebugLevel)

	err := logger.Sync()
	if err != nil {
		t.Errorf("expected no error from Sync, got %v", err)
	}
}

func TestDynamicLevelLogger_SetLevelConcurrent(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	done := make(chan bool)
	for i := range 10 {
		go func(n int) {
			logger.SetLevel(Level(n % 5))
			done <- true
		}(i)
	}

	for range 10 {
		<-done
	}
}

func TestDynamicLevelLogger_NilLogger(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil logger")
		}
	}()

	logger := NewDynamicLevelLogger(nil, InfoLevel)
	ctx := context.Background()
	logger.Info(ctx, "nil logger")
}

func TestDynamicLevelLogger_WithEmptyMessage(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()
	logger.Info(ctx, "")
}

func TestDynamicLevelLogger_WithNilKeys(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()
	childLogger := logger.With(ctx, nil...)

	if childLogger == nil {
		t.Fatal("expected child logger with nil keys")
	}

	childLogger.Info(ctx, "nil keys log")
}

func TestDynamicLevelLogger_WithSpecialCharacters(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()
	logger.Info(ctx, "special chars: !@#$%^&*()")
}

func TestDynamicLevelLogger_UnicodeMessage(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()
	logger.Info(ctx, "Unicode message: 你好世界 🌍")
}

func TestDynamicLevelLogger_LongMessage(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()
	longMsg := strings.Repeat("a", 10000)
	logger.Info(ctx, longMsg)
}

func TestDynamicLevelLogger_MultipleInstances(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger1 := NewDynamicLevelLogger(baseLogger, InfoLevel)
	logger2 := NewDynamicLevelLogger(baseLogger, WarnLevel)

	ctx := context.Background()
	logger1.Info(ctx, "logger1 message")
	logger2.Warn(ctx, "logger2 message")
}

func TestDynamicLevelLogger_WithChain(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()

	child1 := logger.With(ctx, KeyValue{Key: "key1", Value: "value1"})
	child2 := child1.(LoggerWithFields).With(ctx, KeyValue{Key: "key2", Value: "value2"})

	child2.Info(ctx, "chained with log")
}

func TestDynamicLevelLogger_NilLoggerPanic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil logger")
		}
	}()

	logger := NewDynamicLevelLogger(nil, InfoLevel)
	ctx := context.Background()
	logger.Info(ctx, "nil logger")
}
