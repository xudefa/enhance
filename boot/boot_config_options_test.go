package boot

import (
	"testing"
)

// TestBoot_WithProfiles_Coverage 测试 WithProfiles 选项
func TestBoot_WithProfiles(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithProfiles("custom"),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if len(app.config.Profiles) == 0 {
		t.Fatal("Expected profiles to be set")
	}
}

// TestBoot_WithAutoConfigExclude_Coverage 测试自动配置排除
func TestBoot_WithAutoConfigExclude(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutStarters(),
		WithExclude("TestAutoConfig"),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if len(app.config.ExcludedAutoConfigs) == 0 {
		t.Error("Expected ExcludedAutoConfigs to be set")
	}
}

// TestBoot_CollectAutoConfigReport_Coverage 测试 collectAutoConfigReport 方法
func TestBoot_CollectAutoConfigReport(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// 这个方法在 Start 时会被调用，这里只验证不会 panic
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()
}

// TestBoot_WithExcludeMultiple_Coverage 测试排除多个自动配置
func TestBoot_WithExcludeMultiple(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutStarters(),
		WithExclude("Config1", "Config2", "Config3"),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if len(app.config.ExcludedAutoConfigs) != 3 {
		t.Errorf("Expected 3 excluded configs, got %d", len(app.config.ExcludedAutoConfigs))
	}
}

// TestBoot_CollectAutoConfigReport_WithAutoConfig_Coverage 测试 collectAutoConfigReport 在有自动配置时被调用
func TestBoot_CollectAutoConfigReport_WithAutoConfig(t *testing.T) {
	t.Parallel()

	// 启用自动配置报告
	EnableAutoConfigReport()
	defer DisableAutoConfigReport()

	// 创建一个带自动配置的应用
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// Start 时会调用 collectAutoConfigReport
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 验证应用正常启动
	if !app.IsRunning() {
		t.Error("Expected app to be running")
	}
}

// TestBoot_CollectAutoConfigReport_WithDebug_Coverage 测试 debug 模式下 collectAutoConfigReport 被调用
func TestBoot_CollectAutoConfigReport_WithDebug(t *testing.T) {
	t.Parallel()

	// 确保报告开关关闭，使用 debug 模式
	DisableAutoConfigReport()

	// 创建一个带 debug 模式的应用
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutStarters(),
		WithProperty("enhance.debug", true),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// Start 时会调用 collectAutoConfigReport（因为 debug=true）
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 验证应用正常启动
	if !app.IsRunning() {
		t.Error("Expected app to be running")
	}
}

// TestBoot_WithProfilesMultiple_Coverage 测试多个 Profile
func TestBoot_WithProfilesMultiple(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithProfiles("dev", "local", "test"),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if len(app.config.Profiles) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(app.config.Profiles))
	}
}
