package boot

import (
	"errors"
	"fmt"
	"testing"
)

// ==================== BootError 未覆盖路径测试 ====================

func TestBootError_Error_WithAnalyzedAndSuggestions(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:        ErrCodeStarterStart,
		message:     "connection refused",
		phase:       "启动",
		analyzed:    "Database port 3306 is not reachable",
		suggestions: []string{"Check database status", "Verify port configuration"},
	}

	errStr := err.Error()

	if !strContains(errStr, "Analysis: Database port 3306 is not reachable") {
		t.Error("expected error to contain analysis section")
	}
	if !strContains(errStr, "Suggestions:") {
		t.Error("expected error to contain suggestions section")
	}
	if !strContains(errStr, "Check database status") {
		t.Error("expected error to contain first suggestion")
	}
	if !strContains(errStr, "Verify port configuration") {
		t.Error("expected error to contain second suggestion")
	}
}

func TestBootError_Error_WithOnlyOriginal(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:     ErrCodeConfigLoad,
		phase:    "初始化",
		original: errors.New("file not found"),
		message:  "",
	}

	errStr := err.Error()
	if !strContains(errStr, "[BOOT_CONFIG_LOAD]") {
		t.Error("expected error to contain error code")
	}
	if !strContains(errStr, "file not found") {
		t.Error("expected error to contain original error message")
	}
}

func TestBootError_Error_DefaultCase(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:  ErrCodeUnknown,
		phase: "运行",
	}

	errStr := err.Error()
	if !strContains(errStr, "[BOOT_UNKNOWN]") {
		t.Error("expected error to contain error code")
	}
	if !strContains(errStr, "boot failed during 运行") {
		t.Error("expected error to contain phase")
	}
	// 默认情况：没有 message 也没有 original
	if strContains(errStr, "Analysis:") {
		t.Error("expected no analysis when analyzed is empty")
	}
}

func TestBootError_Analyzed(t *testing.T) {
	t.Parallel()
	err := &bootError{
		analyzed: "some analysis",
	}
	if err.Analyzed() != "some analysis" {
		t.Errorf("expected 'some analysis', got '%s'", err.Analyzed())
	}
}

func TestBootError_Suggestions(t *testing.T) {
	t.Parallel()
	suggestions := []string{"fix A", "fix B"}
	err := &bootError{suggestions: suggestions}

	got := err.Suggestions()
	if len(got) != 2 || got[0] != "fix A" || got[1] != "fix B" {
		t.Errorf("expected [fix A, fix B], got %v", got)
	}

	// 验证返回的是副本
	got[0] = "mutated"
	if err.Suggestions()[0] == "mutated" {
		t.Error("Suggestions() should return a copy, not a reference")
	}
}

func TestBootError_Suggestions_Empty(t *testing.T) {
	t.Parallel()
	err := &bootError{}
	if err.Suggestions() != nil {
		t.Error("expected nil for empty suggestions")
	}
}

func TestNewBootErr_NilError(t *testing.T) {
	t.Parallel()
	err := NewBootErr(ErrCodeConfigLoad, "初始化", nil)

	if err.Code() != ErrCodeConfigLoad {
		t.Errorf("expected code '%s', got '%s'", ErrCodeConfigLoad, err.Code())
	}
	if err.Message() != "" {
		t.Errorf("expected empty message for nil error, got '%s'", err.Message())
	}
	if err.Cause() != nil {
		t.Error("expected nil Cause()")
	}
}

func TestNewBootErrf_NilArgs(t *testing.T) {
	t.Parallel()
	err := NewBootErrf(ErrCodeAutoConfig, "初始化", "simple message")
	if err.Message() != "simple message" {
		t.Errorf("expected 'simple message', got '%s'", err.Message())
	}
	if err.Cause() != nil {
		t.Error("expected nil Cause()")
	}
}

// ==================== 补充测试用例 ====================

func TestBootError_Error_WithMessageOnly(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:    ErrCodeStarterStart,
		message: "starter web failed to bind port 8080",
		phase:   "启动",
	}

	errStr := err.Error()
	if !strContains(errStr, "[BOOT_STARTER_START]") {
		t.Error("expected error to contain error code")
	}
	if !strContains(errStr, "starter web failed to bind port 8080") {
		t.Error("expected error to contain message")
	}
	if !strContains(errStr, "启动") {
		t.Error("expected error to contain phase")
	}
}

func TestBootError_Error_WithOriginalAndMessage(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:     ErrCodeConfigLoad,
		message:  "custom message overrides original",
		phase:    "配置加载",
		original: errors.New("original error"),
	}

	errStr := err.Error()
	// 有 message 时优先使用 message
	if !strContains(errStr, "custom message overrides original") {
		t.Error("expected error to contain message, not original")
	}
	if strContains(errStr, "original error") {
		t.Error("expected error NOT to contain original when message is set")
	}
}

func TestBootError_Error_WithAnalysisOnly(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:     ErrCodeAutoConfig,
		phase:    "自动配置",
		analyzed: "Circular dependency detected between A and B",
	}

	errStr := err.Error()
	if !strContains(errStr, "Analysis: Circular dependency detected between A and B") {
		t.Error("expected error to contain analysis")
	}
	if strContains(errStr, "Suggestions:") {
		t.Error("expected error NOT to contain suggestions when empty")
	}
}

func TestBootError_Error_Complete(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:        ErrCodeModuleInstall,
		message:     "module database failed",
		phase:       "模块安装",
		original:    errors.New("connection timeout"),
		analyzed:    "Database service is not running",
		suggestions: []string{"Start database service", "Check connection string", "Verify network connectivity"},
	}

	errStr := err.Error()

	// 验证错误码
	if !strContains(errStr, "[BOOT_MODULE_INSTALL]") {
		t.Error("expected error to contain error code")
	}
	// 验证消息
	if !strContains(errStr, "module database failed") {
		t.Error("expected error to contain message")
	}
	// 验证阶段
	if !strContains(errStr, "模块安装") {
		t.Error("expected error to contain phase")
	}
	// 验证分析
	if !strContains(errStr, "Analysis: Database service is not running") {
		t.Error("expected error to contain analysis")
	}
	// 验证建议
	if !strContains(errStr, "Suggestions:") {
		t.Error("expected error to contain suggestions header")
	}
	if !strContains(errStr, "Start database service") {
		t.Error("expected error to contain first suggestion")
	}
	if !strContains(errStr, "Check connection string") {
		t.Error("expected error to contain second suggestion")
	}
	if !strContains(errStr, "Verify network connectivity") {
		t.Error("expected error to contain third suggestion")
	}
}

func TestBootError_Unwrap(t *testing.T) {
	t.Parallel()
	originalErr := errors.New("underlying cause")
	err := &bootError{
		code:     ErrCodeStarterConfig,
		phase:    "配置",
		original: originalErr,
	}

	if err.Unwrap() != originalErr {
		t.Error("expected Unwrap() to return original error")
	}

	// 验证 errors.Is 可用
	if !errors.Is(err, originalErr) {
		t.Error("expected errors.Is to find original error")
	}
}

func TestBootError_Unwrap_Nil(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:  ErrCodeUnknown,
		phase: "未知",
	}

	if err.Unwrap() != nil {
		t.Error("expected Unwrap() to return nil when no original error")
	}
}

func TestBootError_Phase(t *testing.T) {
	t.Parallel()
	err := &bootError{
		phase: "启动阶段",
	}

	if err.Phase() != "启动阶段" {
		t.Errorf("expected phase '启动阶段', got '%s'", err.Phase())
	}
}

func TestBootError_Phase_Empty(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code: ErrCodeConfigLoad,
	}

	if err.Phase() != "" {
		t.Errorf("expected empty phase, got '%s'", err.Phase())
	}
}

func TestBootError_Code(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code: ErrCodeLifecycle,
	}

	if err.Code() != ErrCodeLifecycle {
		t.Errorf("expected code '%s', got '%s'", ErrCodeLifecycle, err.Code())
	}
}

func TestBootError_Message(t *testing.T) {
	t.Parallel()
	err := &bootError{
		message: "detailed error message",
	}

	if err.Message() != "detailed error message" {
		t.Errorf("expected message 'detailed error message', got '%s'", err.Message())
	}
}

func TestBootError_Message_Empty(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code: ErrCodeConfigCenter,
	}

	if err.Message() != "" {
		t.Errorf("expected empty message, got '%s'", err.Message())
	}
}

func TestNewBootErr_WithValidError(t *testing.T) {
	t.Parallel()
	originalErr := errors.New("config file missing")
	err := NewBootErr(ErrCodeConfigLoad, "初始化", originalErr)

	if err.Code() != ErrCodeConfigLoad {
		t.Errorf("expected code '%s', got '%s'", ErrCodeConfigLoad, err.Code())
	}
	if err.Message() != "config file missing" {
		t.Errorf("expected message 'config file missing', got '%s'", err.Message())
	}
	if err.Cause() != originalErr {
		t.Error("expected Cause() to return original error")
	}
	// Phase 是内部方法，需要通过具体类型访问
	bootErr := err.(*bootError)
	if bootErr.Phase() != "初始化" {
		t.Errorf("expected phase '初始化', got '%s'", bootErr.Phase())
	}
}

func TestNewBootErrf_WithFormatting(t *testing.T) {
	t.Parallel()
	err := NewBootErrf(ErrCodeStarterStart, "启动", "starter %s failed on port %d", "web", 8080)

	if err.Code() != ErrCodeStarterStart {
		t.Errorf("expected code '%s', got '%s'", ErrCodeStarterStart, err.Code())
	}
	if err.Message() != "starter web failed on port 8080" {
		t.Errorf("expected formatted message, got '%s'", err.Message())
	}
	// Phase 是内部方法
	bootErr := err.(*bootError)
	if bootErr.Phase() != "启动" {
		t.Errorf("expected phase '启动', got '%s'", bootErr.Phase())
	}
}

func TestNewBootErrf_WithComplexFormatting(t *testing.T) {
	t.Parallel()
	type MyConfig struct {
		Name string
	}
	cfg := &MyConfig{Name: "DatabaseConfig"}
	causeErr := errors.New("timeout")
	err := NewBootErrf(ErrCodeAutoConfig, "自动配置", "配置 %T 失败: %v", cfg, causeErr)

	expectedMsg := "配置 *boot.MyConfig 失败: timeout"
	if err.Message() != expectedMsg {
		t.Errorf("expected '%s', got '%s'", expectedMsg, err.Message())
	}
}

func TestBootError_Error_AllErrorCodes(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		code  string
		phase string
	}{
		{ErrCodeConfigLoad, "配置加载"},
		{ErrCodeConfigCenter, "配置中心"},
		{ErrCodeAutoConfig, "自动配置"},
		{ErrCodeModuleInstall, "模块安装"},
		{ErrCodeStarterConfig, "Starter配置"},
		{ErrCodeStarterStart, "Starter启动"},
		{ErrCodeLifecycle, "生命周期"},
		{ErrCodeUnknown, "未知阶段"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.code, func(t *testing.T) {
			t.Parallel()
			err := &bootError{
				code:  tc.code,
				phase: tc.phase,
			}

			errStr := err.Error()
			expectedCode := fmt.Sprintf("[%s]", tc.code)
			if !strContains(errStr, expectedCode) {
				t.Errorf("expected error to contain '%s'", expectedCode)
			}
			if !strContains(errStr, tc.phase) {
				t.Errorf("expected error to contain phase '%s'", tc.phase)
			}
		})
	}
}

func TestBootError_Suggestions_SingleItem(t *testing.T) {
	t.Parallel()
	err := &bootError{
		suggestions: []string{"Single suggestion"},
	}

	got := err.Suggestions()
	if len(got) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(got))
	}
	if got[0] != "Single suggestion" {
		t.Errorf("expected 'Single suggestion', got '%s'", got[0])
	}
}

func TestBootError_Suggestions_MutationSafety(t *testing.T) {
	t.Parallel()
	original := []string{"fix A", "fix B", "fix C"}
	err := &bootError{suggestions: original}

	// 第一次获取
	got1 := err.Suggestions()
	// 修改返回值
	got1[0] = "mutated"
	got1[1] = "also mutated"

	// 第二次获取应该不受影响
	got2 := err.Suggestions()
	if got2[0] == "mutated" {
		t.Error("Suggestions() should return independent copies")
	}
	if got2[1] == "also mutated" {
		t.Error("Suggestions() should not be affected by mutations")
	}

	// 验证原始数据也未受影响
	if original[0] == "mutated" {
		t.Error("Original suggestions should not be mutated")
	}
}

func TestBootError_Error_NoAnalysisWhenEmpty(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:    ErrCodeStarterStart,
		message: "some error",
		phase:   "启动",
	}

	errStr := err.Error()
	if strContains(errStr, "Analysis:") {
		t.Error("expected no Analysis section when analyzed is empty")
	}
	if strContains(errStr, "Suggestions:") {
		t.Error("expected no Suggestions section when suggestions is empty")
	}
}

func TestBootError_ImplementsErrorInterface(t *testing.T) {
	t.Parallel()
	err := NewBootErr(ErrCodeConfigLoad, "测试", errors.New("test"))

	// 验证可以作为 error 接口使用
	var e error = err
	if e.Error() == "" {
		t.Error("expected non-empty Error() implementation")
	}
}

func TestBootError_AssetsAsBootError(t *testing.T) {
	t.Parallel()
	err := NewBootErr(ErrCodeStarterStart, "启动", errors.New("start failed"))

	var bootErr BootError
	if !errors.As(err, &bootErr) {
		t.Fatal("expected errors.As to extract BootError")
	}

	if bootErr.Code() != ErrCodeStarterStart {
		t.Errorf("expected code '%s', got '%s'", ErrCodeStarterStart, bootErr.Code())
	}
	if bootErr.Message() != "start failed" {
		t.Errorf("expected message 'start failed', got '%s'", bootErr.Message())
	}
}

func TestBootError_ErrorOutputFormat(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:        ErrCodeConfigLoad,
		message:     "config.json not found",
		phase:       "初始化",
		analyzed:    "Configuration file is missing or inaccessible",
		suggestions: []string{"Create config.json", "Check file permissions"},
	}

	errStr := err.Error()

	// 验证格式包含必要的分隔符
	if !strContains(errStr, "\n\nAnalysis:") {
		t.Error("expected Analysis to be separated by double newline")
	}
	if !strContains(errStr, "\n\nSuggestions:") {
		t.Error("expected Suggestions to be separated by double newline")
	}
	if !strContains(errStr, "\n  - ") {
		t.Error("expected suggestions to be formatted with bullet points")
	}
}
