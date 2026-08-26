package boot

import (
	"errors"
	"testing"

	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/lifecycle"
)

func TestBoot_Stop_DuringInit(t *testing.T) {
	t.Parallel()

	s := newMockStarter("test")
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	// 手动设置 starter 绕过 Start() 完整流程
	boot.starters = []Starter{s}

	// 在 PhaseInit 时调用 Stop()
	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	// Stop() 不应调用 starter.Stop()，因为 starter 尚未启动
	if s.stopped.Load() {
		t.Error("starter.Stop() was called but starter was never started")
	}
	if boot.ctx.Lifecycle().GetPhase() != lifecycle.PhaseStopped {
		t.Errorf("expected PhaseStopped, got %v", boot.ctx.Lifecycle().GetPhase())
	}
}

func TestBoot_Stop_DuringRunning(t *testing.T) {
	t.Parallel()

	s := newMockStarter("test")
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	boot.starters = []Starter{s}

	// 模拟已运行状态
	if err := boot.ctx.Lifecycle().SetPhase(lifecycle.PhaseRunning); err != nil {
		t.Fatalf("SetPhase(Running) error = %v", err)
	}

	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	// PhaseRunning 时 starter 已启动，应调用 Stop
	if !s.stopped.Load() {
		t.Error("starter.Stop() should be called when phase was Running")
	}
}

func TestBoot_Stop_DoubleCall(t *testing.T) {
	t.Parallel()
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	// 正常启动
	if err := boot.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !boot.IsRunning() {
		t.Fatal("expected running after Start()")
	}

	// 第一次 Stop
	if err := boot.Stop(); err != nil {
		t.Fatalf("first Stop() error = %v", err)
	}
	if boot.IsRunning() {
		t.Fatal("should not be running after Stop()")
	}

	// 第二次 Stop — 应安全返回 nil
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

	// 模拟已停止状态
	if err := boot.ctx.Lifecycle().SetPhase(lifecycle.PhaseRunning); err != nil {
		t.Fatalf("SetPhase(Running) error = %v", err)
	}
	if err := boot.ctx.Lifecycle().SetPhase(lifecycle.PhaseStopped); err != nil {
		t.Fatalf("SetPhase(Stopped) error = %v", err)
	}

	// 在 PhaseStopped 调用 Stop() — 应直接返回 nil
	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() during PhaseStopped should return nil, got: %v", err)
	}
}

func TestBoot_Stop_OnlyStopsStartedStarters(t *testing.T) {
	s := newMockStarter("test")
	// 通过全局注册表注册，Start() 会从全局注册表加载
	orig := globalStarterRegistry.Load()
	testReg := newStarterRegistryImpl()
	globalStarterRegistry.Store(testReg)
	t.Cleanup(func() { globalStarterRegistry.Store(orig) })
	testReg.Register(s)

	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	// 正常启动 — starter 会经历 Configure → Start
	if err := boot.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !s.started.Load() {
		t.Fatal("starter should be started after Start()")
	}

	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if !s.stopped.Load() {
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

// mockStarterWithCondition 支持条件的 mock starter
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

// mockFailingStarter 的 Start 总是失败，用于验证失败后重试
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

	// 第一次启动失败
	if err := boot.Start(); err == nil {
		t.Fatal("first Start() should fail")
	}

	// 重试不应是 no-op：若 started 未重置，第二次 Start() 会直接返回 nil
	if err := boot.Start(); err == nil {
		t.Fatal("retry Start() should also fail (must not be a no-op after previous failure)")
	}
}

func TestBoot_BindConfig(t *testing.T) {
	t.Parallel()

	type TestConfig struct {
		Host string `config:"server.host"`
	}

	boot, err := NewApplication(
		WithAppName("test-app"),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	// 添加属性源
	boot.ctx.Environment().AddPropertySource(environment.NewDefaultPropertySource("test", map[string]any{
		"server.host": "localhost",
	}))

	cfg, err := BindConfig[TestConfig](boot)
	if err != nil {
		t.Fatalf("BindConfig() error = %v", err)
	}
	if cfg.Host != "localhost" {
		t.Errorf("expected Host 'localhost', got '%s'", cfg.Host)
	}
}

func TestBoot_BindConfig_WithPrefix(t *testing.T) {
	t.Parallel()

	type ServerConfig struct {
		Port int `config:"port"`
	}

	boot, err := NewApplication(
		WithAppName("test-app"),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	// 设置带前缀的配置
	boot.ctx.Environment().AddPropertySource(environment.NewDefaultPropertySource("test", map[string]any{
		"server.port": "8080",
	}))

	cfg, err := BindConfig[ServerConfig](boot, WithConfigPrefix("server"))
	if err != nil {
		t.Fatalf("BindConfig() error = %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected Port 8080, got %d", cfg.Port)
	}
}

func TestWithConfigPrefix(t *testing.T) {
	t.Parallel()

	opt := WithConfigPrefix("test")
	if opt == nil {
		t.Error("expected non-nil option")
	}
}

func TestBoot_NewApplicationFromRunOptions(t *testing.T) {
	t.Parallel()

	app, err := NewApplicationFromRunOptions(WithAppName("test-app"))
	if err != nil {
		t.Fatalf("NewApplicationFromRunOptions() error = %v", err)
	}
	if app == nil {
		t.Error("expected non-nil application")
	}
}
