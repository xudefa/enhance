package log

import (
	"context"
	"testing"
)

func TestLevelString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "debug"},
		{InfoLevel, "info"},
		{WarnLevel, "warn"},
		{ErrorLevel, "error"},
		{DPanicLevel, "dpanic"},
		{PanicLevel, "panic"},
		{FatalLevel, "fatal"},
		{Level(100), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

type mockLogger struct {
	lastMsg  string
	lastKeys []KeyValue
}

func (m *mockLogger) Debug(ctx context.Context, msg string, keys ...KeyValue) {
	m.lastMsg = msg
	m.lastKeys = keys
}

func (m *mockLogger) Info(ctx context.Context, msg string, keys ...KeyValue) {
	m.lastMsg = msg
	m.lastKeys = keys
}

func (m *mockLogger) Warn(ctx context.Context, msg string, keys ...KeyValue) {
	m.lastMsg = msg
	m.lastKeys = keys
}

func (m *mockLogger) Error(ctx context.Context, msg string, keys ...KeyValue) {
	m.lastMsg = msg
	m.lastKeys = keys
}

func (m *mockLogger) DPanic(ctx context.Context, msg string, keys ...KeyValue) {
	m.lastMsg = msg
	m.lastKeys = keys
}

func (m *mockLogger) Panic(ctx context.Context, msg string, keys ...KeyValue) {
	m.lastMsg = msg
	m.lastKeys = keys
}

func (m *mockLogger) Fatal(ctx context.Context, msg string, keys ...KeyValue) {
	m.lastMsg = msg
	m.lastKeys = keys
}

func (m *mockLogger) Sync() error {
	return nil
}

func (m *mockLogger) With(ctx context.Context, keys ...KeyValue) Logger {
	m.lastKeys = keys
	return m
}

func TestAppendContextKeys_DoesNotAliasCallerSlice(t *testing.T) {
	t.Parallel()

	ctx := WithTraceID(context.Background(), "trace-123")

	base := make([]KeyValue, 1, 4)
	base[0] = KeyValue{Key: "a", Value: "1"}

	cl := NewContextLogger(&mockLogger{})
	cl.Info(ctx, "msg", base...)

	_ = append(base, KeyValue{Key: "b", Value: "2"})

	mock := cl.logger.(*mockLogger)
	if len(mock.lastKeys) < 2 {
		t.Fatalf("expected trace_id key to be appended, got %d keys", len(mock.lastKeys))
	}
	if mock.lastKeys[1].Key != "trace_id" || mock.lastKeys[1].Value != "trace-123" {
		t.Errorf("trace_id key corrupted by caller's slice reuse: got %+v", mock.lastKeys[1])
	}
}

func TestLoggerOption(t *testing.T) {
	t.Parallel()
	mock := &mockLogger{}
	opt := WithLogger(mock)
	cfg := &loggerConfig{}
	opt(cfg)

	if cfg.logger != mock {
		t.Error("WithLogger option did not set logger correctly")
	}
}

func TestBuildOptions(t *testing.T) {
	t.Parallel()
	mock := &mockLogger{}
	logger := Build(WithLogger(mock))

	if logger != mock {
		t.Error("Build() did not apply options correctly")
	}
}

func TestNewLoggerBuilder_Defaults(t *testing.T) {
	t.Parallel()

	b := NewLoggerBuilder()
	if b == nil {
		t.Fatal("expected non-nil builder")
	}

	if b.level != InfoLevel {
		t.Errorf("expected default InfoLevel, got %v", b.level)
	}
	if b.format != "json" {
		t.Errorf("expected default format 'json', got %q", b.format)
	}
	if b.addSource != false {
		t.Error("expected default addSource false")
	}
	if b.outputPath != "" {
		t.Errorf("expected default empty outputPath, got %q", b.outputPath)
	}
	if b.sampler != nil {
		t.Error("expected default nil sampler")
	}
	if b.name != "" {
		t.Errorf("expected default empty name, got %q", b.name)
	}
}

func TestLoggerBuilder_Level(t *testing.T) {
	t.Parallel()

	b := NewLoggerBuilder()
	b.Level(WarnLevel)
	if b.level != WarnLevel {
		t.Errorf("expected WarnLevel, got %v", b.level)
	}
}

func TestLoggerBuilder_Format(t *testing.T) {
	t.Parallel()

	b := NewLoggerBuilder()
	b.Format("text")
	if b.format != "text" {
		t.Errorf("expected 'text', got %q", b.format)
	}
}

func TestLoggerBuilder_AddSource(t *testing.T) {
	t.Parallel()

	b := NewLoggerBuilder()
	b.AddSource(true)
	if !b.addSource {
		t.Error("expected addSource true")
	}
}

func TestLoggerBuilder_SetOutputPath(t *testing.T) {
	t.Parallel()

	b := NewLoggerBuilder()
	b.OutputPath("/tmp/test.log")
	if b.outputPath != "/tmp/test.log" {
		t.Errorf("expected '/tmp/test.log', got %q", b.outputPath)
	}
}

func TestLoggerBuilder_Sampler(t *testing.T) {
	t.Parallel()

	b := NewLoggerBuilder()
	sampler := NewRandomSampler(0.5)
	b.Sampler(sampler)
	if b.sampler != sampler {
		t.Error("expected sampler to be set")
	}
}

func TestLoggerBuilder_Name(t *testing.T) {
	t.Parallel()

	b := NewLoggerBuilder()
	b.Name("my-logger")
	if b.name != "my-logger" {
		t.Errorf("expected 'my-logger', got %q", b.name)
	}
}

func TestLoggerBuilder_Build_WithAllOptions(t *testing.T) {
	t.Parallel()

	logger := NewLoggerBuilder().
		Level(DebugLevel).
		Format("text").
		AddSource(true).
		OutputPath("/tmp/test-builder.log").
		Sampler(NewRandomSampler(1.0)).
		Name("test-logger").
		Build()

	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	ctx := context.Background()
	logger.Debug(ctx, "test debug message")
	logger.Info(ctx, "test info message")
}

func TestLoggerBuilder_Build_MultipleBuilds(t *testing.T) {
	t.Parallel()

	b := NewLoggerBuilder().Name("shared-builder")
	l1 := b.Build()
	l2 := b.Build()

	if l1 == l2 {
		t.Error("expected different logger instances from same builder")
	}
}

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
	sampler := NewRandomSampler(0.1)
	logger := NewLoggerBuilder().
		Level(DebugLevel).
		Sampler(sampler).
		Build()

	if logger == nil {
		t.Fatal("expected logger to be created with sampler")
	}

	ctx := context.Background()
	logger.Info(ctx, "test message with sampler")
}
