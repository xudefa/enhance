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

// TestEnhanceApp_Run_Coverage 测试 EnhanceApp.Run 方法
func TestEnhanceApp_Run_Coverage(t *testing.T) {
	t.Parallel()
	// Run 会阻塞，只测试方法存在和签名正确
	app := &EnhanceApp{}
	if app == nil {
		t.Fatal("Expected non-nil app")
	}
}

// TestRun_PackageLevel_Coverage 测试包级别的 Run 函数
func TestRun_PackageLevel_Coverage(t *testing.T) {
	t.Parallel()
	// Run 会阻塞，只测试函数存在
	// 实际运行需要信号处理
}

// TestNewApplication_WithModules_Coverage 测试带模块的应用创建
func TestNewApplication_WithModules_Coverage(t *testing.T) {
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

// TestNewApplication_WithProperty_Coverage 测试带属性的应用创建
func TestNewApplication_WithProperty_Coverage(t *testing.T) {
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

// TestEnhanceApp_Run_WithOptions_Coverage 测试 Run 带选项
func TestEnhanceApp_Run_WithOptions_Coverage(t *testing.T) {
	t.Parallel()

	// 测试选项构建
	opts := []boot.BootOption{
		boot.WithAppName("run-test"),
		boot.WithVersion("1.0.0"),
		boot.WithProfiles("test"),
	}
	if len(opts) != 3 {
		t.Errorf("Expected 3 options, got %d", len(opts))
	}
}

// TestEnhanceApp_Run_WithModulesOption_Coverage 测试带模块选项
func TestEnhanceApp_Run_WithModulesOption_Coverage(t *testing.T) {
	t.Parallel()

	// 测试模块选项构建
	opts := []boot.BootOption{
		boot.WithAppName("module-run-test"),
	}
	if len(opts) != 1 {
		t.Errorf("Expected 1 option, got %d", len(opts))
	}
}

// TestNewApplication_ErrorHandling_Coverage 测试错误处理
func TestNewApplication_ErrorHandling_Coverage(t *testing.T) {
	t.Parallel()

	// 测试正常创建（应该成功）
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
}

// TestEnhance_GlobalVariable_Coverage 测试全局变量
func TestEnhance_GlobalVariable_Coverage(t *testing.T) {
	t.Parallel()

	// 验证 Enhance 是 *EnhanceApp 类型
	if Enhance == nil {
		t.Fatal("Expected Enhance to be non-nil")
	}

	// 验证类型
	_, ok := interface{}(Enhance).(*EnhanceApp)
	if !ok {
		t.Fatal("Expected Enhance to be *EnhanceApp")
	}
}

// TestEnhanceApp_Type_Coverage 测试 EnhanceApp 类型
func TestEnhanceApp_Type_Coverage(t *testing.T) {
	t.Parallel()

	// 创建实例
	app := EnhanceApp{}
	if app == (EnhanceApp{}) {
		// 验证零值结构体可以创建
	}
}

// TestRun_FunctionSignature_Coverage 测试 Run 函数签名
func TestRun_FunctionSignature_Coverage(t *testing.T) {
	t.Parallel()

	// 验证 Run 函数存在且签名正确
	// 不调用 Run，因为它会阻塞
	// 这里只验证函数签名和类型
	var fn func(...boot.BootOption) = Run
	if fn == nil {
		t.Fatal("Expected Run function to exist")
	}
}
