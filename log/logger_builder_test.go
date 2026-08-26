package log

import (
	"context"
	"testing"
)

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
	s := NewRandomSampler(0.5)
	b.Sampler(s)
	if b.sampler != s {
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
