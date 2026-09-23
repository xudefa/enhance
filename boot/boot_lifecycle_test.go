package boot

import (
	"context"
	"errors"
	"syscall"
	"testing"
	"time"

	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/lifecycle"
)

func TestBoot_Stop_DuringInit(t *testing.T) {
	t.Parallel()
	starter := newMockStarter("test")
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	boot.starters = []Starter{starter}
	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if starter.stopped.Load() {
		t.Error("starter.Stop() was called but starter was never started")
	}
	if boot.ctx.Lifecycle().GetPhase() != lifecycle.PhaseStopped {
		t.Errorf("expected PhaseStopped, got %v", boot.ctx.Lifecycle().GetPhase())
	}
}

func TestBoot_Stop_DuringRunning(t *testing.T) {
	t.Parallel()
	starter := newMockStarter("test")
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	boot.starters = []Starter{starter}
	if err := boot.ctx.Lifecycle().SetPhase(lifecycle.PhaseRunning); err != nil {
		t.Fatalf("SetPhase(Running) error = %v", err)
	}
	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if !starter.stopped.Load() {
		t.Error("starter.Stop() should be called when phase was Running")
	}
}

func TestBoot_Stop_DoubleCall(t *testing.T) {
	t.Parallel()
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	if err := boot.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !boot.IsRunning() {
		t.Fatal("expected running after Start()")
	}
	if err := boot.Stop(); err != nil {
		t.Fatalf("first Stop() error = %v", err)
	}
	if boot.IsRunning() {
		t.Fatal("should not be running after Stop()")
	}
	if err := boot.Stop(); err != nil {
		t.Fatalf("second Stop() should return nil, got: %v", err)
	}
}

func TestBoot_Stop_AlreadyStopped(t *testing.T) {
	t.Parallel()
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	if err := boot.ctx.Lifecycle().SetPhase(lifecycle.PhaseRunning); err != nil {
		t.Fatalf("SetPhase(Running) error = %v", err)
	}
	if err := boot.ctx.Lifecycle().SetPhase(lifecycle.PhaseStopped); err != nil {
		t.Fatalf("SetPhase(Stopped) error = %v", err)
	}
	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() during PhaseStopped should return nil, got: %v", err)
	}
}

func TestBoot_Stop_OnlyStopsStartedStarters(t *testing.T) {
	starter := newMockStarter("test")
	orig := globalStarterRegistry.Load()
	testReg := newStarterRegistryImpl()
	globalStarterRegistry.Store(testReg)
	t.Cleanup(func() { globalStarterRegistry.Store(orig) })
	testReg.Register(starter)
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	if err := boot.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !starter.started.Load() {
		t.Fatal("starter should be started after Start()")
	}
	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if !starter.stopped.Load() {
		t.Error("starter.Stop() should be called after proper start")
	}
}

func TestBoot_Stop_WithConditionalStarters(t *testing.T) {
	enabled := newMockStarter("enabled")
	disabled := newMockStarterWithCondition("disabled", condition.OnProperty("never.match"))
	orig := globalStarterRegistry.Load()
	testReg := newStarterRegistryImpl()
	globalStarterRegistry.Store(testReg)
	defer func() { globalStarterRegistry.Store(orig) }()
	testReg.Register(enabled)
	testReg.Register(disabled)
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	if err := boot.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if !enabled.stopped.Load() {
		t.Error("enabled starter should be stopped")
	}
	if disabled.stopped.Load() {
		t.Error("disabled starter should NOT be stopped when condition does not match")
	}
}

type mockStarterWithCondition struct {
	*mockStarter
	cond condition.Condition
}

func newMockStarterWithCondition(name string, cond condition.Condition) *mockStarterWithCondition {
	return &mockStarterWithCondition{
		mockStarter: newMockStarter(name),
		cond:        cond,
	}
}

func (m *mockStarterWithCondition) GetCondition() condition.Condition {
	return m.cond
}

type mockFailingStarter struct {
	*mockStarter
}

func (m *mockFailingStarter) Start(ctx ApplicationContext) error {
	m.started.Store(true)
	return errors.New("boom")
}

func TestBoot_Start_RetryAfterFailure(t *testing.T) {
	s := newMockStarter("failing")
	failing := &mockFailingStarter{mockStarter: s}
	orig := globalStarterRegistry.Load()
	testReg := newStarterRegistryImpl()
	globalStarterRegistry.Store(testReg)
	t.Cleanup(func() { globalStarterRegistry.Store(orig) })
	testReg.Register(failing)
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	if err := boot.Start(); err == nil {
		t.Fatal("first Start() should fail")
	}
	if err := boot.Start(); err == nil {
		t.Fatal("retry Start() should also fail (must not be a no-op after previous failure)")
	}
}

func TestBoot_Start_WithLifecyclePhase(t *testing.T) {
	t.Parallel()
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	if err := boot.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	phase := boot.ctx.Lifecycle().GetPhase()
	if phase != lifecycle.PhaseRunning {
		t.Errorf("expected PhaseRunning, got %v", phase)
	}
	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	phase = boot.ctx.Lifecycle().GetPhase()
	if phase != lifecycle.PhaseStopped {
		t.Errorf("expected PhaseStopped, got %v", phase)
	}
}

func TestBoot_Start_WithStarter_boot_Register(t *testing.T) {
	starter := newMockStarter("test-starter")
	orig := globalStarterRegistry.Load()
	testReg := newStarterRegistryImpl()
	globalStarterRegistry.Store(testReg)
	defer func() { globalStarterRegistry.Store(orig) }()
	testReg.Register(starter)
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	if err := boot.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer boot.Stop()
	if !starter.configured.Load() {
		t.Error("starter.Configure() should be called")
	}
	if !starter.started.Load() {
		t.Error("starter.Start() should be called")
	}
}

func TestBoot_Start_WithGlobalStarter(t *testing.T) {
	starter := newMockStarter("global-starter")
	orig := globalStarterRegistry.Load()
	testReg := newStarterRegistryImpl()
	globalStarterRegistry.Store(testReg)
	defer func() { globalStarterRegistry.Store(orig) }()
	testReg.Register(starter)
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	if err := boot.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer boot.Stop()
	if !starter.configured.Load() {
		t.Error("global starter.Configure() should be called")
	}
	if !starter.started.Load() {
		t.Error("global starter.Start() should be called")
	}
}

type TestAutoConfiguration struct {
	onConfigure func()
}

func (a *TestAutoConfiguration) Configure(ctx ApplicationContext) error {
	if a.onConfigure != nil {
		a.onConfigure()
	}
	return nil
}

func (a *TestAutoConfiguration) Dependencies() []string {
	return nil
}

func (a *TestAutoConfiguration) GetCondition() condition.Condition {
	return condition.OnPropertyOrDefault("test.enabled", "true", "true")
}

func TestBoot_Start_WithHooks(t *testing.T) {
	t.Parallel()
	initCalled := false
	startCalled := false
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithHook(lifecycle.NewHookFunc(
			func(ctx context.Context) error { initCalled = true; return nil },
			func(ctx context.Context) error { startCalled = true; return nil },
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

func TestBoot_Start_Stop(t *testing.T) {
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
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !app.started.Load() {
		t.Error("Expected started to be true after Start")
	}
	if err := app.Start(); err != nil {
		t.Errorf("Second Start should return nil, got %v", err)
	}
	if err := app.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if !app.stopped.Load() {
		t.Error("Expected stopped to be true after Stop")
	}
	if err := app.Stop(); err != nil {
		t.Errorf("Second Stop should return nil, got %v", err)
	}
}

func TestBoot_Start_WithAutoConfig(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutStarters())
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
}

func TestBoot_StartFailed(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithHook(lifecycle.NewHookFunc(
			func(ctx context.Context) error { return errors.New("intentional failure") },
			nil, nil,
		)),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err == nil {
		t.Error("Expected Start to fail")
	}
	if app.started.Load() {
		t.Error("Expected started to be false after failed Start")
	}
}

func TestBoot_IsRunning(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters())
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if app.IsRunning() {
		t.Error("Expected IsRunning to be false before Start")
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !app.IsRunning() {
		t.Error("Expected IsRunning to be true after Start")
	}
	defer app.Stop()
}

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

func TestBoot_StartMultipleTimes(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters())
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("First Start failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Second Start failed: %v", err)
	}
	if err := app.Stop(); err != nil {
		t.Fatalf("First Stop failed: %v", err)
	}
	if err := app.Stop(); err != nil {
		t.Fatalf("Second Stop failed: %v", err)
	}
}

func TestBoot_WaitForSignal(t *testing.T) {
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters())
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()
	app.WaitForSignal()
	if app.IsRunning() {
		t.Error("Expected app to be stopped after WaitForSignal")
	}
}

func TestBoot_WaitForSignal_StopError(t *testing.T) {
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters())
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()
	app.WaitForSignal()
}
