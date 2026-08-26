package log

import (
	"context"
	"testing"
)

func TestWithTraceID_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		traceID string
	}{
		{"basic", "abc-123"},
		{"empty", ""},
		{"long", "a-very-long-trace-id-value-for-testing"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := WithTraceID(context.Background(), tt.traceID)
			got := GetTraceID(ctx)
			if got != tt.traceID {
				t.Errorf("GetTraceID() = %q, want %q", got, tt.traceID)
			}
		})
	}
}

func TestGetTraceID_NoValue(t *testing.T) {
	t.Parallel()

	got := GetTraceID(context.Background())
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestGetTraceID_WrongType(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), TraceContextKey, 12345)
	got := GetTraceID(ctx)
	if got != "" {
		t.Errorf("expected empty string for wrong type, got %q", got)
	}
}

func TestContextLogger_WithoutTraceID(t *testing.T) {
	t.Parallel()

	base := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(base)

	ctx := context.Background()
	logger.Info(ctx, "no trace id")
}

func TestContextLogger_WithFields(t *testing.T) {
	t.Parallel()

	base := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(base)

	ctx := context.Background()
	child := logger.With(ctx, KeyValue{Key: "k", Value: "v"})
	if child == nil {
		t.Fatal("expected child logger")
	}

	child.Info(ctx, "child log")
}

func TestDynamicLevelLogger_SetGetLevel(t *testing.T) {
	t.Parallel()

	base := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(base, WarnLevel)

	if logger.GetLevel() != WarnLevel {
		t.Errorf("expected WarnLevel, got %v", logger.GetLevel())
	}

	logger.SetLevel(DebugLevel)
	if logger.GetLevel() != DebugLevel {
		t.Errorf("expected DebugLevel after SetLevel, got %v", logger.GetLevel())
	}
}

func TestDynamicLevelLogger_SharedLevel(t *testing.T) {
	t.Parallel()

	base := NewSlogLogger(WithLevel(DebugLevel))
	parent := NewDynamicLevelLogger(base, InfoLevel)
	child := parent.With(context.Background(), KeyValue{Key: "k", Value: "v"})

	childDynamic, ok := child.(*DynamicLevelLogger)
	if !ok {
		t.Fatal("expected DynamicLevelLogger child")
	}

	parent.SetLevel(ErrorLevel)
	if childDynamic.GetLevel() != ErrorLevel {
		t.Errorf("expected child to share parent level, got %v", childDynamic.GetLevel())
	}
}

func TestAppendContextKeys_EmptyTraceID(t *testing.T) {
	t.Parallel()

	keys := []KeyValue{{Key: "a", Value: "b"}}
	result := appendContextKeys(context.Background(), keys)

	if len(result) != 1 {
		t.Errorf("expected 1 key, got %d", len(result))
	}
}

func TestAppendContextKeys_WithTraceID(t *testing.T) {
	t.Parallel()

	ctx := WithTraceID(context.Background(), "trace-1")
	keys := []KeyValue{{Key: "a", Value: "b"}}
	result := appendContextKeys(ctx, keys)

	if len(result) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result))
	}
	if result[1].Key != "trace_id" {
		t.Errorf("expected trace_id key, got %q", result[1].Key)
	}
}

func TestAppendContextKeys_NoMutation(t *testing.T) {
	t.Parallel()

	ctx := WithTraceID(context.Background(), "trace-1")
	keys := []KeyValue{{Key: "a", Value: "b"}}
	_ = appendContextKeys(ctx, keys)

	if len(keys) != 1 {
		t.Errorf("original keys should not be mutated, got %d", len(keys))
	}
}
