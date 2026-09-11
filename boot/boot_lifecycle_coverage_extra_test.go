package boot

import (
	"context"
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/xudefa/enhance/lifecycle"
)

// TestBoot_Start_Stop_Coverage 测试 Boot 的 Start 和 Stop 完整流程
func TestBoot_Start_Stop_Coverage(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithVersion("1.0.0"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// 测试 Start
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// 验证启动状态
	if !app.started.Load() {
		t.Error("Expected started to be true after Start")
	}

	// 测试重复 Start（应返回 nil）
	if err := app.Start(); err != nil {
		t.Errorf("Second Start should return nil, got %v", err)
	}

	// 测试 Stop
	if err := app.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// 验证停止状态
	if !app.stopped.Load() {
		t.Error("Expected stopped to be true after Stop")
	}

	// 测试重复 Stop（应返回 nil）
	if err := app.Stop(); err != nil {
		t.Errorf("Second Stop should return nil, got %v", err)
	}
}

// TestBoot_Start_WithAutoConfig_Coverage 测试带自动配置的 Start
func TestBoot_Start_WithAutoConfig_Coverage(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()
}

// TestBoot_Start_WithHooks_Coverage 测试带钩子的 Start
func TestBoot_Start_WithHooks_Coverage(t *testing.T) {
	t.Parallel()

	initCalled := false
	startCalled := false

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithHook(lifecycle.NewHookFunc(
			func(ctx context.Context) error {
				initCalled = true
				return nil
			},
			func(ctx context.Context) error {
				startCalled = true
				return nil
			},
			nil,
		)),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	if !initCalled {
		t.Error("Expected OnInit hook to be called")
	}
	if !startCalled {
		t.Error("Expected OnStart hook to be called")
	}
}

// TestBoot_StartFailed_Coverage 测试启动失败时的状态重置
func TestBoot_StartFailed(t *testing.T) {
	t.Parallel()

	// 创建一个会失败的钩子
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithHook(lifecycle.NewHookFunc(
			func(ctx context.Context) error {
				return fmt.Errorf("intentional failure")
			},
			nil,
			nil,
		)),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// 启动应失败
	if err := app.Start(); err == nil {
		t.Error("Expected Start to fail")
	}

	// 验证状态已重置
	if app.started.Load() {
		t.Error("Expected started to be false after failed Start")
	}
}

// TestBoot_IsRunning_Coverage 测试 IsRunning 方法
func TestBoot_IsRunning(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// 启动前不应运行
	if app.IsRunning() {
		t.Error("Expected IsRunning to be false before Start")
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// 启动后应运行
	if !app.IsRunning() {
		t.Error("Expected IsRunning to be true after Start")
	}

	defer app.Stop()
}

// TestBoot_WithoutStartupReport_Coverage 测试 WithoutStartupReport 选项
func TestBoot_WithoutStartupReport(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithoutStartupReport(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()
}

// TestBoot_StartMultipleTimes_Coverage 测试多次启动和停止
func TestBoot_StartMultipleTimes(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// 第一次启动
	if err := app.Start(); err != nil {
		t.Fatalf("First Start failed: %v", err)
	}

	// 第二次启动（应该成功，幂等）
	if err := app.Start(); err != nil {
		t.Fatalf("Second Start failed: %v", err)
	}

	// 第一次停止
	if err := app.Stop(); err != nil {
		t.Fatalf("First Stop failed: %v", err)
	}

	// 第二次停止（应该成功，幂等）
	if err := app.Stop(); err != nil {
		t.Fatalf("Second Stop failed: %v", err)
	}
}

// TestBoot_WaitForSignal_Coverage 测试 WaitForSignal 方法
func TestBoot_WaitForSignal(t *testing.T) {
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

	// 在 goroutine 中调用 WaitForSignal
	go func() {
		// 等待应用启动
		time.Sleep(100 * time.Millisecond)
		// 发送 SIGTERM 信号
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	// 调用 WaitForSignal（会阻塞直到收到信号）
	app.WaitForSignal()

	// 验证应用已停止
	if app.IsRunning() {
		t.Error("Expected app to be stopped after WaitForSignal")
	}
}

// TestBoot_WaitForSignal_StopError_Coverage 测试 WaitForSignal 中 Stop 返回错误的情况
func TestBoot_WaitForSignal_StopError(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// 不启动应用，直接调用 WaitForSignal
	// 这样 Stop 会返回错误（因为应用未启动）
	go func() {
		time.Sleep(100 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	// 调用 WaitForSignal
	app.WaitForSignal()
}
