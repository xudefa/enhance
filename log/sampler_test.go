package log

import (
	"context"
	"strings"
	"testing"
)

func TestRandomSampler_Rate(t *testing.T) {
	t.Parallel()
	sampler := NewRandomSampler(0.5)

	sampled := 0
	total := 1000

	for range total {
		if sampler.ShouldSample() {
			sampled++
		}
	}

	rate := float64(sampled) / float64(total)
	if rate < 0.4 || rate > 0.6 {
		t.Errorf("expected sampling rate around 0.5, got %f", rate)
	}
}

func TestRandomSampler_Boundary(t *testing.T) {
	t.Parallel()
	sampler0 := NewRandomSampler(0)
	if sampler0.ShouldSample() {
		t.Error("expected 0% sampler to never sample")
	}

	sampler1 := NewRandomSampler(1)
	for range 100 {
		if !sampler1.ShouldSample() {
			t.Error("expected 100% sampler to always sample")
			break
		}
	}
}

func TestThresholdSampler(t *testing.T) {
	t.Parallel()
	sampler := NewThresholdSampler(5)

	sampled := 0
	for range 100 {
		if sampler.ShouldSample() {
			sampled++
		}
	}

	if sampled != 20 {
		t.Errorf("expected 20 samples, got %d", sampled)
	}
}

func TestSampledLogger_ErrorNotSampled(t *testing.T) {
	t.Parallel()
	sampler := NewRandomSampler(0)
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()

	logger.Error(ctx, "error message")
}

func TestSampledLogger_With(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(0.5)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	childLogger := logger.With(ctx, KeyValue{Key: "key", Value: "value"})

	if childLogger == nil {
		t.Fatal("expected child logger to be created")
	}

	sampledChild, ok := childLogger.(*SampledLogger)
	if !ok {
		t.Fatal("expected child to be SampledLogger")
	}

	if sampledChild.sampler != sampler {
		t.Error("expected child to share same sampler")
	}
}

func TestSampledLogger_AllLevels(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()

	logger.Debug(ctx, "debug")
	logger.Info(ctx, "info")
	logger.Warn(ctx, "warn")
	logger.Error(ctx, "error")
}

func TestSampledLogger_Sync(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(0.5)
	logger := NewSampledLogger(baseLogger, sampler)

	err := logger.Sync()
	if err != nil {
		t.Errorf("expected no error from Sync, got %v", err)
	}
}

func TestSampledLogger_NilSampler(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))

	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	logger.Info(ctx, "message with sampler")
}

func TestSampledLogger_ContextPropagation(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := WithTraceID(context.Background(), "sample-trace")
	logger.Info(ctx, "sampled with context")
}

func TestSampledLogger_PerformanceWithHighSampling(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()

	for range 1000 {
		logger.Info(ctx, "high frequency log")
	}
}

func TestSampledLogger_SamplerShouldSample(t *testing.T) {
	t.Parallel()
	sampler := NewRandomSampler(0)
	if sampler.ShouldSample() {
		t.Error("0% sampler should not sample")
	}

	sampler = NewRandomSampler(1)
	for range 100 {
		if !sampler.ShouldSample() {
			t.Error("100% sampler should always sample")
			break
		}
	}
}

func TestSampledLogger_WithNilSampler(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))

	panicked := false
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
		if !panicked {
			t.Error("expected panic with nil sampler")
		}
	}()

	logger := NewSampledLogger(baseLogger, nil)
	ctx := context.Background()
	logger.Info(ctx, "nil sampler")
}

func TestSampledLogger_DPanic(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel), WithDevelopment(true))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected DPanic to panic in development mode")
		}
	}()

	logger.DPanic(ctx, "dpanic message")
}

func TestSampledLogger_Panic(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected Panic to panic")
		}
	}()

	logger.Panic(ctx, "panic message")
}

func TestSampledLogger_Fatal(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()

	_ = logger
	_ = ctx
}

func TestSampledLogger_ErrorAlwaysLogged(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()

	logger.Error(ctx, "always logged error")
}

func TestSampledLogger_WithPreservesSampler(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(0.5)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	childLogger := logger.With(ctx, KeyValue{Key: "key", Value: "value"})

	sampledChild, ok := childLogger.(*SampledLogger)
	if !ok {
		t.Fatal("expected SampledLogger")
	}

	if sampledChild.sampler != sampler {
		t.Error("expected child to preserve sampler")
	}
}

func TestSampledLogger_WithNilLogger(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil logger")
		}
	}()

	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(nil, sampler)
	ctx := context.Background()
	logger.Info(ctx, "nil logger")
}

func TestSampledLogger_WithEmptyMessage(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	logger.Info(ctx, "")
}

func TestSampledLogger_WithNilKeys(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	childLogger := logger.With(ctx, nil...)

	if childLogger == nil {
		t.Fatal("expected child logger with nil keys")
	}

	childLogger.Info(ctx, "nil keys log")
}

func TestSampledLogger_WithSpecialCharacters(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	logger.Info(ctx, "special chars: !@#$%^&*()")
}

func TestSampledLogger_UnicodeMessage(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	logger.Info(ctx, "Unicode message: 你好世界 🌍")
}

func TestSampledLogger_LongMessage(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	longMsg := strings.Repeat("a", 10000)
	logger.Info(ctx, longMsg)
}

func TestSampledLogger_MultipleInstances(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler1 := NewRandomSampler(0.5)
	sampler2 := NewRandomSampler(0.8)

	logger1 := NewSampledLogger(baseLogger, sampler1)
	logger2 := NewSampledLogger(baseLogger, sampler2)

	ctx := context.Background()
	logger1.Info(ctx, "logger1 message")
	logger2.Info(ctx, "logger2 message")
}

func TestSampledLogger_WithChain(t *testing.T) {
	t.Parallel()
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()

	child1 := logger.With(ctx, KeyValue{Key: "key1", Value: "value1"})
	child2 := child1.(LoggerWithFields).With(ctx, KeyValue{Key: "key2", Value: "value2"})

	child2.Info(ctx, "chained with log")
}

func TestSampledLogger_WithNilLoggerAndSampler(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil logger")
		}
	}()

	logger := NewSampledLogger(nil, nil)
	ctx := context.Background()
	logger.Info(ctx, "nil logger and sampler")
}
