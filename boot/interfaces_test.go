package boot

import (
	"errors"
	"testing"
)

// ==================== StarterRegistry 接口测试 ====================

func TestStarterRegistryInterface_Register(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()
	s := newMockStarter("test-starter")

	registry.Register(s)

	all := registry.GetAll()
	if len(all) != 1 {
		t.Fatalf("expected 1 starter, got %d", len(all))
	}
	if all[0].Name() != "test-starter" {
		t.Fatalf("expected name 'test-starter', got '%s'", all[0].Name())
	}
}

func TestStarterRegistryInterface_RegisterMultiple(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()
	registry.Register(newMockStarter("s1"))
	registry.Register(newMockStarter("s2"))
	registry.Register(newMockStarter("s3"))

	if len(registry.GetAll()) != 3 {
		t.Fatalf("expected 3 starters, got %d", len(registry.GetAll()))
	}
}

func TestStarterRegistryInterface_Get(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()
	registry.Register(newMockStarter("alpha"))
	registry.Register(newMockStarter("beta"))

	alpha := registry.Get("alpha")
	if alpha == nil {
		t.Fatal("expected to find starter 'alpha'")
	}
	if alpha.Name() != "alpha" {
		t.Fatalf("expected name 'alpha', got '%s'", alpha.Name())
	}

	beta := registry.Get("beta")
	if beta == nil {
		t.Fatal("expected to find starter 'beta'")
	}

	missing := registry.Get("nonexistent")
	if missing != nil {
		t.Fatal("expected nil for nonexistent starter")
	}
}

func TestStarterRegistryInterface_GetOrdered(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()
	registry.Register(newMockStarter("web", "db"))
	registry.Register(newMockStarter("db"))
	registry.Register(newMockStarter("cache", "db"))

	ordered := registry.GetOrdered()
	if len(ordered) != 3 {
		t.Fatalf("expected 3 ordered starters, got %d", len(ordered))
	}

	// db should come before web and cache
	dbIndex, webIndex, cacheIndex := -1, -1, -1
	for i, s := range ordered {
		switch s.Name() {
		case "db":
			dbIndex = i
		case "web":
			webIndex = i
		case "cache":
			cacheIndex = i
		}
	}
	if dbIndex >= webIndex {
		t.Error("db should come before web")
	}
	if dbIndex >= cacheIndex {
		t.Error("db should come before cache")
	}
}

func TestStarterRegistryInterface_GetOrdered_CircularDependency(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()
	registry.Register(newMockStarter("a", "b"))
	registry.Register(newMockStarter("b", "a"))

	ordered := registry.GetOrdered()
	// Should fall back to original order when circular dependency detected
	if len(ordered) != 2 {
		t.Fatalf("expected 2 starters, got %d", len(ordered))
	}
}

func TestStarterRegistryInterface_Empty(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()

	if len(registry.GetAll()) != 0 {
		t.Error("expected empty GetAll")
	}
	if len(registry.GetOrdered()) != 0 {
		t.Error("expected empty GetOrdered")
	}
	if registry.Get("anything") != nil {
		t.Error("expected nil from Get on empty registry")
	}
}

func TestStarterRegistryInterface_ReturnsCopy(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()
	registry.Register(newMockStarter("s1"))

	all1 := registry.GetAll()
	registry.Register(newMockStarter("s2"))
	all2 := registry.GetAll()

	if len(all1) != 1 {
		t.Fatalf("expected first GetAll to return 1, got %d", len(all1))
	}
	if len(all2) != 2 {
		t.Fatalf("expected second GetAll to return 2, got %d", len(all2))
	}
}

// ==================== BootError 接口测试 ====================

func TestBootErrorInterface_Code(t *testing.T) {
	t.Parallel()
	err := NewBootErr(ErrCodeConfigLoad, "初始化", errors.New("file not found"))
	if err.Code() != ErrCodeConfigLoad {
		t.Fatalf("expected code '%s', got '%s'", ErrCodeConfigLoad, err.Code())
	}
}

func TestBootErrorInterface_Message(t *testing.T) {
	t.Parallel()
	err := NewBootErr(ErrCodeAutoConfig, "初始化", errors.New("config failed"))
	if err.Message() != "config failed" {
		t.Fatalf("expected message 'config failed', got '%s'", err.Message())
	}
}

func TestBootErrorInterface_Cause(t *testing.T) {
	t.Parallel()
	original := errors.New("root cause")
	err := NewBootErr(ErrCodeStarterStart, "启动", original)
	if !errors.Is(err, original) {
		t.Fatal("expected errors.Is to find original error")
	}
	if err.Cause() != original {
		t.Fatal("expected Cause() to return original error")
	}
}

func TestBootErrorInterface_Error(t *testing.T) {
	t.Parallel()
	err := NewBootErr(ErrCodeConfigLoad, "初始化", errors.New("file missing"))
	errStr := err.Error()
	if errStr == "" {
		t.Fatal("expected non-empty error string")
	}
	if !strContains(errStr, ErrCodeConfigLoad) {
		t.Fatalf("expected error to contain code '%s'", ErrCodeConfigLoad)
	}
	if !strContains(errStr, "file missing") {
		t.Fatal("expected error to contain original message")
	}
}

func TestBootErrorInterface_Unwrap(t *testing.T) {
	t.Parallel()
	original := errors.New("underlying error")
	err := NewBootErr(ErrCodeAutoConfig, "初始化", original)
	if !errors.Is(err, original) {
		t.Fatal("expected errors.Is to find underlying error via Unwrap")
	}
}

func TestBootErrorInterface_Phase(t *testing.T) {
	t.Parallel()
	err := NewBootErr(ErrCodeStarterConfig, "配置阶段", errors.New("fail"))
	bootErr := err.(*bootError)
	if bootErr.Phase() != "配置阶段" {
		t.Fatalf("expected phase '配置阶段', got '%s'", bootErr.Phase())
	}
}

func TestBootErrorInterface_ErrorsAs(t *testing.T) {
	t.Parallel()
	err := NewBootErr(ErrCodeAutoConfig, "初始化", errors.New("config failed"))

	var bootErr BootError
	if !errors.As(err, &bootErr) {
		t.Fatal("expected errors.As to extract BootError")
	}
	if bootErr.Code() != ErrCodeAutoConfig {
		t.Fatalf("expected code '%s', got '%s'", ErrCodeAutoConfig, bootErr.Code())
	}
}

func TestBootErrorInterface_ErrorCodes(t *testing.T) {
	t.Parallel()
	codes := []string{
		ErrCodeConfigLoad,
		ErrCodeConfigCenter,
		ErrCodeAutoConfig,
		ErrCodeModuleInstall,
		ErrCodeStarterConfig,
		ErrCodeStarterStart,
		ErrCodeLifecycle,
		ErrCodeUnknown,
	}
	for _, code := range codes {
		if code == "" {
			t.Fatal("error code should not be empty")
		}
	}
}

func TestNewBootErrf(t *testing.T) {
	t.Parallel()
	err := NewBootErrf(ErrCodeAutoConfig, "初始化", "config %s failed", "database")
	if err.Code() != ErrCodeAutoConfig {
		t.Fatalf("expected code '%s', got '%s'", ErrCodeAutoConfig, err.Code())
	}
	if err.Message() != "config database failed" {
		t.Fatalf("expected message 'config database failed', got '%s'", err.Message())
	}
}

func TestNewBootErrf_WithCause(t *testing.T) {
	t.Parallel()
	original := errors.New("connection refused")
	err := NewBootErrf(ErrCodeStarterStart, "启动", "starter %s failed: %v", "web", original)

	if err.Code() != ErrCodeStarterStart {
		t.Fatalf("expected code '%s', got '%s'", ErrCodeStarterStart, err.Code())
	}
	if err.Cause() != original {
		t.Fatal("expected Cause() to return original error")
	}
	if !errors.Is(err, original) {
		t.Fatal("expected errors.Is to find original error via optional error arg")
	}
}

func TestNewBootErrf_WithoutCause(t *testing.T) {
	t.Parallel()
	err := NewBootErrf(ErrCodeAutoConfig, "初始化", "config %s failed", "database")
	if err.Cause() != nil {
		t.Fatal("expected Cause() to be nil when no error arg provided")
	}
}

// ==================== BootConfig 重构测试 ====================

func TestBootConfig_NewFields(t *testing.T) {
	t.Parallel()
	cfg := &BootConfig{
		AppName:     "test-app",
		Version:     "1.0.0",
		Profiles:    []string{"dev"},
		ConfigPaths: []string{"/config/app.json"},
		Port:        8080,
		Debug:       true,
	}

	if cfg.AppName != "test-app" {
		t.Fatalf("expected AppName 'test-app', got '%s'", cfg.AppName)
	}
	if cfg.Version != "1.0.0" {
		t.Fatalf("expected Version '1.0.0', got '%s'", cfg.Version)
	}
	if len(cfg.Profiles) != 1 || cfg.Profiles[0] != "dev" {
		t.Fatalf("expected Profiles ['dev'], got %v", cfg.Profiles)
	}
	if len(cfg.ConfigPaths) != 1 || cfg.ConfigPaths[0] != "/config/app.json" {
		t.Fatalf("expected ConfigPaths ['/config/app.json'], got %v", cfg.ConfigPaths)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected Port 8080, got %d", cfg.Port)
	}
	if !cfg.Debug {
		t.Fatal("expected Debug true")
	}
}

func TestBootConfig_BackwardCompatible(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()

	// 新字段有合理默认值
	if cfg.Port != 0 {
		t.Logf("Port default: %d", cfg.Port)
	}
	if cfg.Debug {
		t.Fatal("Debug should default to false")
	}

	// 旧字段仍然有效
	if cfg.AppName != "enhance-app" {
		t.Fatalf("expected AppName 'enhance-app', got '%s'", cfg.AppName)
	}
	if cfg.ConfigType != "json" {
		t.Fatalf("expected ConfigType 'json', got '%s'", cfg.ConfigType)
	}
	if !cfg.AutoExecute {
		t.Fatal("expected AutoExecute true")
	}
}

func TestBootConfig_WithOptionPattern(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()

	opt := WithAppName("custom-app")
	opt(cfg)

	if cfg.AppName != "custom-app" {
		t.Fatalf("expected AppName 'custom-app', got '%s'", cfg.AppName)
	}
}

// ==================== Application 接口兼容性测试 ====================

func TestApplicationInterface_StartStop(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("app-test"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	// Boot 实现了 Application 接口的核心方法
	if err := app.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := app.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestApplicationInterface_Config(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("config-test"),
		WithVersion("2.0.0"),
		WithProfiles("prod"),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	cfg := app.Config()
	if cfg == nil {
		t.Fatal("Config() should not return nil")
	}
	if cfg.AppName != "config-test" {
		t.Fatalf("expected AppName 'config-test', got '%s'", cfg.AppName)
	}
	if cfg.Version != "2.0.0" {
		t.Fatalf("expected Version '2.0.0', got '%s'", cfg.Version)
	}
}

func TestApplicationInterface_Context(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("ctx-test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	ctx := app.Context()
	if ctx == nil {
		t.Fatal("Context() should not return nil")
	}
	if ctx.Container() == nil {
		t.Fatal("Context().Container() should not return nil")
	}
	if ctx.Environment() == nil {
		t.Fatal("Context().Environment() should not return nil")
	}
}

// ==================== NewStarterRegistry 返回接口类型测试 ====================

func TestNewStarterRegistry_ReturnsInterface(t *testing.T) {
	t.Parallel()
	var registry StarterRegistry = NewStarterRegistry()
	registry.Register(newMockStarter("test"))

	if len(registry.GetAll()) != 1 {
		t.Fatal("expected 1 starter")
	}
}

func TestGlobalStarterRegistry_ReturnsInterface(t *testing.T) {
	t.Parallel()
	var registry StarterRegistry = GlobalStarterRegistry()
	if registry == nil {
		t.Fatal("GlobalStarterRegistry() should not return nil")
	}
}

// ==================== 辅助函数 ====================

func strContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
