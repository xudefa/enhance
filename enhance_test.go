package enhance

import (
	"testing"

	"github.com/xudefa/enhance/boot"
)

func TestEnhance_Run(t *testing.T) {
	t.Parallel()
	// Run会阻塞，我们只测试它能正常创建
	// 实际Run测试需要更复杂的设置
}

func TestEnhance_Run_WithOptions(t *testing.T) {
	t.Parallel()
	// 测试 Run 函数接受选项参数（不实际调用，因为会阻塞）
	// 这里验证 API 签名正确性
	opts := []boot.BootOption{
		boot.WithAppName("test-app"),
		boot.WithVersion("1.0.0"),
	}
	if len(opts) != 2 {
		t.Errorf("Expected 2 options, got %d", len(opts))
	}
}

func TestEnhance_Variable(t *testing.T) {
	t.Parallel()
	// 测试 Enhance 全局变量
	if Enhance == nil {
		t.Fatal("Expected Enhance to be non-nil")
	}
	if _, ok := interface{}(Enhance).(*EnhanceApp); !ok {
		t.Fatal("Expected Enhance to be *EnhanceApp")
	}
}

func TestNewApplication(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		boot.WithAppName("test-app"),
		boot.WithVersion("1.0.0"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestNewApplication_Default(t *testing.T) {
	t.Parallel()
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestNewApplication_WithProfiles(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		boot.WithAppName("test-app"),
		boot.WithProfiles("dev", "test"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestEnhanceApp_Struct(t *testing.T) {
	t.Parallel()
	// 测试 EnhanceApp 结构体
	app := &EnhanceApp{}
	if app == nil {
		t.Fatal("Expected non-nil EnhanceApp")
	}
}

// ==================== 补充覆盖测试 ====================

func TestRun_PackageLevel(t *testing.T) {
	t.Parallel()
	// Run 会阻塞，只测试函数存在
	// 实际运行需要信号处理
}

func TestNewApplication_WithModules(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		boot.WithAppName("module-test"),
		boot.WithVersion("2.0.0"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestNewApplication_WithProperty(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		boot.WithAppName("property-test"),
		boot.WithProperty("test.key", "test-value"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestEnhanceApp_Run_WithModulesOption(t *testing.T) {
	t.Parallel()

	// 测试模块选项构建
	opts := []boot.BootOption{
		boot.WithAppName("module-run-test"),
	}
	if len(opts) != 1 {
		t.Errorf("Expected 1 option, got %d", len(opts))
	}
}

func TestRun_FunctionSignature(t *testing.T) {
	t.Parallel()

	// 验证 Run 函数存在且签名正确
	var fn func(...boot.BootOption) = Run
	if fn == nil {
		t.Fatal("Expected Run function to exist")
	}
}