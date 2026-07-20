package boot

import (
	"fmt"
	"testing"

	"github.com/xudefa/enhance/core"
)

// TestBeanNotFoundAnalyzer 测试 Bean 未找到分析器
func TestBeanNotFoundAnalyzer(t *testing.T) {
	t.Parallel()
	analyzer := NewBeanNotFoundAnalyzer()

	// 测试 CanAnalyze
	err1 := core.ErrBeanNotFound
	if !analyzer.CanAnalyze(err1) {
		t.Error("Should analyze ErrBeanNotFound")
	}

	err2 := fmt.Errorf("bean not found: userService")
	if !analyzer.CanAnalyze(err2) {
		t.Error("Should analyze error containing 'bean not found'")
	}

	err3 := fmt.Errorf("some other error")
	if analyzer.CanAnalyze(err3) {
		t.Error("Should not analyze unrelated error")
	}

	// 测试 Analyze
	report := analyzer.Analyze(err1)
	if report == nil || report.Headline == "" || report.Description == "" || report.Action == "" || len(report.PossibleSolutions) == 0 {
		t.Fatalf("Report should be valid, got %+v", report)
	}
}

// TestCircularDependencyAnalyzer 测试循环依赖分析器
func TestCircularDependencyAnalyzer(t *testing.T) {
	t.Parallel()
	analyzer := NewCircularDependencyAnalyzer()

	// 测试 CanAnalyze
	err1 := core.ErrCircularDependency
	if !analyzer.CanAnalyze(err1) {
		t.Error("Should analyze ErrCircularDependency")
	}

	err2 := fmt.Errorf("circular dependency detected between A and B")
	if !analyzer.CanAnalyze(err2) {
		t.Error("Should analyze error containing 'circular dependency'")
	}

	// 测试 Analyze
	report := analyzer.Analyze(err1)
	if report == nil || report.Headline == "" {
		t.Fatalf("Report should be valid, got %+v", report)
	}

	// 验证建议中包含懒加载
	hasLazySuggestion := false
	for _, sol := range report.PossibleSolutions {
		if contains(sol, "lazy") || contains(sol, "懒加载") {
			hasLazySuggestion = true
			break
		}
	}
	if !hasLazySuggestion {
		t.Error("Should suggest lazy injection as a solution")
	}
}

// TestDuplicateBeanAnalyzer 测试重复 Bean 分析器
func TestDuplicateBeanAnalyzer(t *testing.T) {
	t.Parallel()
	analyzer := NewDuplicateBeanAnalyzer()

	// 测试 CanAnalyze
	err1 := core.ErrBeanAlreadyExists
	if !analyzer.CanAnalyze(err1) {
		t.Error("Should analyze ErrBeanAlreadyExists")
	}

	err2 := fmt.Errorf("bean already exists: userService")
	if !analyzer.CanAnalyze(err2) {
		t.Error("Should analyze error containing 'bean already exists'")
	}

	// 测试 Analyze
	report := analyzer.Analyze(err1)
	if report == nil || len(report.PossibleSolutions) == 0 {
		t.Fatalf("Report should be valid, got %+v", report)
	}
}

// TestPortInUseAnalyzer 测试端口占用分析器
func TestPortInUseAnalyzer(t *testing.T) {
	t.Parallel()
	analyzer := NewPortInUseAnalyzer()

	// 测试 CanAnalyze
	err1 := fmt.Errorf("listen tcp :8080: bind: address already in use")
	if !analyzer.CanAnalyze(err1) {
		t.Error("Should analyze 'address already in use' error")
	}

	err2 := fmt.Errorf("port 8080 bind failed")
	if !analyzer.CanAnalyze(err2) {
		t.Error("Should analyze 'port bind' error")
	}

	err3 := fmt.Errorf("some other error")
	if analyzer.CanAnalyze(err3) {
		t.Error("Should not analyze unrelated error")
	}

	// 测试 Analyze
	report := analyzer.Analyze(err1)
	if report == nil || len(report.PossibleSolutions) == 0 {
		t.Fatalf("Report should be valid, got %+v", report)
	}

	// 验证建议中包含 lsof 命令
	hasLsofSuggestion := false
	for _, sol := range report.PossibleSolutions {
		if contains(sol, "lsof") {
			hasLsofSuggestion = true
			break
		}
	}
	if !hasLsofSuggestion {
		t.Error("Should suggest using lsof command")
	}
}

// TestConfigLoadAnalyzer 测试配置加载分析器
func TestConfigLoadAnalyzer(t *testing.T) {
	t.Parallel()
	analyzer := NewConfigLoadAnalyzer()

	// 测试 CanAnalyze
	err1 := ErrPropertyNotFound
	if !analyzer.CanAnalyze(err1) {
		t.Error("Should analyze ErrPropertyNotFound")
	}

	err2 := ErrTypeConversion
	if !analyzer.CanAnalyze(err2) {
		t.Error("Should analyze ErrTypeConversion")
	}

	err3 := fmt.Errorf("failed to load config")
	if !analyzer.CanAnalyze(err3) {
		t.Error("Should analyze error containing 'config'")
	}

	// 测试 Analyze
	report := analyzer.Analyze(err1)
	if report == nil || len(report.PossibleSolutions) == 0 {
		t.Fatalf("Report should be valid, got %+v", report)
	}
}

// TestFailureAnalyzerRegistry 测试分析器注册表
func TestFailureAnalyzerRegistry(t *testing.T) {
	t.Parallel()
	registry := NewFailureAnalyzerRegistry()

	// 注册自定义分析器
	customAnalyzer := &testCustomAnalyzer{}
	registry.Register(customAnalyzer)

	// 测试 CanAnalyze
	customErr := fmt.Errorf("custom error")
	report := registry.Analyze(customErr)
	if report == nil || report.Headline != "Custom Error" {
		t.Fatalf("Expected headline 'Custom Error', got %+v", report)
	}
}

// testCustomAnalyzer 自定义测试分析器
type testCustomAnalyzer struct{}

func (a *testCustomAnalyzer) CanAnalyze(err error) bool {
	return contains(err.Error(), "custom error")
}

func (a *testCustomAnalyzer) Analyze(err error) *FailureReport {
	if !a.CanAnalyze(err) {
		return nil
	}

	return &FailureReport{
		Headline:    "Custom Error",
		Description: "A custom error occurred",
		Action:      "Fix the custom error",
		Cause:       err.Error(),
	}
}

// TestGlobalAnalyzerRegistry 测试全局分析器注册表
func TestGlobalAnalyzerRegistry(t *testing.T) {
	t.Parallel()
	// 全局注册表应该已经注册了内置分析器
	registry := GlobalAnalyzerRegistry()
	if registry == nil {
		t.Fatal("Global registry should not be nil")
	}

	// 测试内置分析器是否工作
	err := core.ErrBeanNotFound
	report := registry.Analyze(err)
	if report == nil {
		t.Error("Global registry should analyze ErrBeanNotFound")
	}
}

// TestReportFailure 测试 ReportFailure 函数
func TestReportFailure(t *testing.T) {
	t.Parallel()
	// 测试可分析的错误
	err1 := core.ErrBeanNotFound
	result1 := ReportFailure(err1)
	if result1 == "" {
		t.Error("ReportFailure should return non-empty string")
	}

	// 测试不可分析的错误
	err2 := fmt.Errorf("unknown error")
	result2 := ReportFailure(err2)
	if result2 == "" {
		t.Error("ReportFailure should return error message even if not analyzable")
	}

	if !contains(result2, "unknown error") {
		t.Error("Should contain original error message")
	}
}

// TestFailureReport_Format 测试失败报告格式化
func TestFailureReport_Format(t *testing.T) {
	t.Parallel()
	report := &FailureReport{
		Headline:    "Test Error",
		Description: "Test description",
		Action:      "Test action",
		Cause:       "Test cause",
		PossibleSolutions: []string{
			"Solution 1",
			"Solution 2",
		},
	}

	formatted := formatFailure(report)

	if !contains(formatted, "Test description") {
		t.Error("Formatted output should contain description")
	}

	if !contains(formatted, "Test action") {
		t.Error("Formatted output should contain action")
	}

	if !contains(formatted, "Solution 1") {
		t.Error("Formatted output should contain solutions")
	}
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
