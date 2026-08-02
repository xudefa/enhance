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

	// 调用方的 slice 预留了额外容量，appendContextKeys 可能原地写入
	base := make([]KeyValue, 1, 4)
	base[0] = KeyValue{Key: "a", Value: "1"}

	cl := NewContextLogger(&mockLogger{})
	cl.Info(ctx, "msg", base...)

	// 模拟调用方在日志之后复用底层数组继续追加
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
