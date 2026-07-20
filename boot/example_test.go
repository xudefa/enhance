package boot

import (
	"context"
	"fmt"
	"testing"
)

// TestExample_NewApplication_Basic 测试基础应用创建示例
func TestExample_NewApplication_Basic(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithVersion("1.0.0"),
	)
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}

	if app == nil {
		t.Fatal("应用实例不应为 nil")
	}

	if app.config.AppName != "test-app" {
		t.Fatalf("期望应用名为 'test-app'，实际 '%s'", app.config.AppName)
	}
}

// TestExample_NewApplication_WithProfiles 测试带配置文件的应用创建
func TestExample_NewApplication_WithProfiles(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("profile-app"),
		WithProfiles("dev", "local"),
	)
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}

	profiles := app.Environment().GetActiveProfiles()
	if len(profiles) != 2 {
		t.Fatalf("期望 2 个配置文件，实际 %d 个", len(profiles))
	}
}

// TestExample_NewApplication_WithOptions 测试带多种选项的应用创建
func TestExample_NewApplication_WithOptions(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("options-app"),
		WithVersion("2.0.0"),
		WithProfiles("test"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}

	if app.config.AppName != "options-app" {
		t.Fatalf("期望应用名为 'options-app'，实际 '%s'", app.config.AppName)
	}
	if app.config.Version != "2.0.0" {
		t.Fatalf("期望版本为 '2.0.0'，实际 '%s'", app.config.Version)
	}
	if app.config.AutoExecute {
		t.Fatal("AutoExecute 应为 false")
	}
	if app.config.Starters {
		t.Fatal("Starters 应为 false")
	}
}

// TestExample_NewApplication_CustomConfig 测试自定义配置类型
func TestExample_NewApplication_CustomConfig(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("yaml-app"),
		WithConfigType("yaml"),
		WithConfigLocation("/custom/path"),
	)
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}

	if app.config.ConfigType != "yaml" {
		t.Fatalf("期望配置类型为 'yaml'，实际 '%s'", app.config.ConfigType)
	}
	if app.config.ConfigLocation != "/custom/path" {
		t.Fatalf("期望配置路径为 '/custom/path'，实际 '%s'", app.config.ConfigLocation)
	}
}

// TestExample_NewApplication_WithHook 测试带钩子的应用创建
func TestExample_NewApplication_WithHook(t *testing.T) {
	t.Parallel()
	initCalled := false
	startCalled := false
	stopCalled := false

	app, err := NewApplication(
		WithAppName("hook-app"),
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
		t.Fatalf("创建应用失败: %v", err)
	}

	// 启动应用
	if err := app.Start(); err != nil {
		t.Fatalf("启动应用失败: %v", err)
	}

	if !initCalled {
		t.Fatal("OnInit 钩子应该被调用")
	}
	if !startCalled {
		t.Fatal("OnStart 钩子应该被调用")
	}

	// 停止应用
	if err := app.Stop(); err != nil {
		t.Fatalf("停止应用失败: %v", err)
	}

	if !stopCalled {
		t.Fatal("OnStop 钩子应该被调用")
	}
}

// TestExample_NewApplication_MultipleInstances 测试创建多个应用实例
func TestExample_NewApplication_MultipleInstances(t *testing.T) {
	t.Parallel()
	app1, err := NewApplication(WithAppName("app1"))
	if err != nil {
		t.Fatalf("创建 app1 失败: %v", err)
	}

	app2, err := NewApplication(WithAppName("app2"))
	if err != nil {
		t.Fatalf("创建 app2 失败: %v", err)
	}

	// 验证两个应用实例是独立的
	if app1 == app2 {
		t.Fatal("两个应用实例应该是不同的对象")
	}

	if app1.config.AppName == app2.config.AppName {
		t.Fatal("两个应用实例应该有不同的名称")
	}

	if app1.Container() == app2.Container() {
		t.Fatal("两个应用实例应该有独立的容器")
	}
}

// TestExample_NewApplication_AccessContext 测试访问应用上下文
func TestExample_NewApplication_AccessContext(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("context-app"),
		WithVersion("1.0.0"),
	)
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}

	// 访问应用上下文
	ctx := app.Context()
	if ctx == nil {
		t.Fatal("应用上下文不应为 nil")
	}

	// 访问 IoC 容器
	container := app.Container()
	if container == nil {
		t.Fatal("IoC 容器不应为 nil")
	}

	// 访问环境配置
	env := app.Environment()
	if env == nil {
		t.Fatal("环境配置不应为 nil")
	}
}

// TestExample_NewApplication_StartStop 测试完整的启动停止流程
func TestExample_NewApplication_StartStop(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("lifecycle-app"),
		WithVersion("1.0.0"),
		WithoutAutoConfig(), // 禁用自动配置以加快测试
		WithoutStarters(),   // 禁用 starter 以加快测试
	)
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}

	// 启动应用
	if err := app.Start(); err != nil {
		t.Fatalf("启动应用失败: %v", err)
	}

	if !app.IsRunning() {
		t.Fatal("应用启动后应该处于运行状态")
	}

	// 停止应用
	if err := app.Stop(); err != nil {
		t.Fatalf("停止应用失败: %v", err)
	}

	if app.IsRunning() {
		t.Fatal("应用停止后不应处于运行状态")
	}
}

// TestExample_NewApplication_DeprecatedFunction 测试已弃用函数的兼容性
func TestExample_NewApplication_DeprecatedFunction(t *testing.T) {
	t.Parallel()
	// 测试向后兼容的函数
	app, err := NewApplicationWithOptions(
		WithAppName("deprecated-app"),
	)
	if err != nil {
		t.Fatalf("使用已弃用函数创建应用失败: %v", err)
	}

	if app.config.AppName != "deprecated-app" {
		t.Fatalf("期望应用名为 'deprecated-app'，实际 '%s'", app.config.AppName)
	}
}

// TestExample_ApplicationConfig_Access 测试应用配置访问
func TestExample_ApplicationConfig_Access(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("config-access-app"),
		WithVersion("3.0.0"),
		WithProfiles("prod"),
	)
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}

	// 验证配置正确设置
	if app.config.AppName != "config-access-app" {
		t.Errorf("AppName 不匹配")
	}
	if app.config.Version != "3.0.0" {
		t.Errorf("Version 不匹配")
	}
	if len(app.config.Profiles) != 1 || app.config.Profiles[0] != "prod" {
		t.Errorf("Profiles 不匹配")
	}
}

// TestExample_Application_Environment 测试环境配置
func TestExample_Application_Environment(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("env-app"),
		WithProfiles("dev", "test"),
	)
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}

	env := app.Environment()
	if env == nil {
		t.Fatal("环境配置不应为 nil")
	}

	profiles := env.GetActiveProfiles()
	if len(profiles) != 2 {
		t.Fatalf("期望 2 个配置文件，实际 %d 个", len(profiles))
	}

	fmt.Printf("Active profiles: %v\n", profiles)
}
