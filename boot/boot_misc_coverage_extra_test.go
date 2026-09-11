package boot

import (
	"context"
	"testing"
)

// TestBoot_WithVersion_Coverage 测试 WithVersion 选项
func TestBoot_WithVersion(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithVersion("2.0.0"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if app.config.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got %s", app.config.Version)
	}
}

// TestBoot_Context_Coverage 测试 Context 方法
func TestBoot_Context_Coverage(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	ctx := app.Context()
	if ctx == nil {
		t.Fatal("Expected non-nil context")
	}

	if app.Container() == nil {
		t.Fatal("Expected non-nil container")
	}

	if app.Environment() == nil {
		t.Fatal("Expected non-nil environment")
	}
}

// TestBoot_WithHookFunc_Coverage 测试 WithHookFunc 选项
func TestBoot_WithHookFunc(t *testing.T) {
	t.Parallel()

	initCalled := false
	startCalled := false
	stopCalled := false

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithHookFunc(
			func(ctx context.Context) error {
				initCalled = true
				return nil
			},
			func(ctx context.Context) error {
				startCalled = true
				return nil
			},
			func(ctx context.Context) error {
				stopCalled = true
				return nil
			},
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !initCalled {
		t.Error("Expected OnInit to be called")
	}
	if !startCalled {
		t.Error("Expected OnStart to be called")
	}

	if err := app.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if !stopCalled {
		t.Error("Expected OnStop to be called")
	}
}

// TestBoot_EventBus_Coverage 测试 EventBus 方法
func TestBoot_EventBus(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 验证 EventBus 返回非 nil
	ctx := app.Context()
	eventBus := ctx.EventBus()
	if eventBus == nil {
		t.Error("Expected non-nil EventBus")
	}
}
