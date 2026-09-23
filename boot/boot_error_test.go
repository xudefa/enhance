package boot

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestBootError_Error(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:    ErrCodeStarterStart,
		message: "connection refused",
		phase:   "启动",
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "connection refused") {
		t.Errorf("Expected error to contain message, got: %s", errStr)
	}
}

func TestBootError_Error_WithAnalyzedAndSuggestions(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:        ErrCodeStarterStart,
		message:     "connection refused",
		phase:       "启动",
		analyzed:    "Database port 3306 unreachable",
		suggestions: []string{"Check database status", "Verify port config"},
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "Analysis: Database port 3306 unreachable") {
		t.Error("Expected error to contain analysis section")
	}
	if !strings.Contains(errStr, "Suggestions:") {
		t.Error("Expected error to contain suggestions section")
	}
	if !strings.Contains(errStr, "Check database status") {
		t.Error("Expected error to contain first suggestion")
	}
}

func TestBootError_Error_WithOriginal(t *testing.T) {
	t.Parallel()
	orig := errors.New("underlying cause")
	err := &bootError{
		code:     ErrCodeStarterStart,
		phase:    "启动",
		original: orig,
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "underlying cause") {
		t.Errorf("Expected error to contain cause, got: %s", errStr)
	}
}

func TestBootError_Unwrap(t *testing.T) {
	t.Parallel()
	orig := errors.New("underlying cause")
	err := &bootError{
		code:     ErrCodeStarterStart,
		message:  "starter failed",
		phase:    "启动",
		original: orig,
	}
	unwrapped := err.Unwrap()
	if unwrapped != orig {
		t.Errorf("Expected unwrapped error to be the cause, got: %v", unwrapped)
	}
}

func TestBootError_Unwrap_NilOriginal(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:    ErrCodeStarterStart,
		message: "starter failed",
		phase:   "启动",
	}
	unwrapped := err.Unwrap()
	if unwrapped != nil {
		t.Errorf("Expected unwrapped error to be nil, got: %v", unwrapped)
	}
}

func TestBootError_Cause(t *testing.T) {
	t.Parallel()
	orig := errors.New("underlying cause")
	err := &bootError{
		code:     ErrCodeStarterStart,
		message:  "starter failed",
		phase:    "启动",
		original: orig,
	}
	if err.Cause() != orig {
		t.Errorf("Expected Cause() to return original error, got: %v", err.Cause())
	}
}

func TestBootError_Cause_Nil(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:    ErrCodeStarterStart,
		message: "starter failed",
		phase:   "启动",
	}
	if err.Cause() != nil {
		t.Errorf("Expected Cause() to be nil, got: %v", err.Cause())
	}
}

func TestBootError_Phase(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code:  ErrCodeStarterStart,
		phase: "启动阶段",
	}
	if err.Phase() != "启动阶段" {
		t.Errorf("Expected phase '启动阶段', got '%s'", err.Phase())
	}
}

func TestBootError_Code(t *testing.T) {
	t.Parallel()
	err := &bootError{
		code: ErrCodeStarterStart,
	}
	if err.Code() != ErrCodeStarterStart {
		t.Errorf("Expected code ErrCodeStarterStart, got %v", err.Code())
	}
}

func TestBootError_Message(t *testing.T) {
	t.Parallel()
	err := &bootError{
		message: "starter failed",
	}
	if err.Message() != "starter failed" {
		t.Errorf("Expected message 'starter failed', got %s", err.Message())
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

func TestNewBootErr_WithValidError(t *testing.T) {
	t.Parallel()
	origErr := errors.New("config file missing")
	err := NewBootErr(ErrCodeConfigLoad, "初始化", origErr)
	if err.Code() != ErrCodeConfigLoad {
		t.Errorf("expected code '%s', got '%s'", ErrCodeConfigLoad, err.Code())
	}
	if err.Message() != "config file missing" {
		t.Errorf("expected message 'config file missing', got '%s'", err.Message())
	}
	if err.Cause() != origErr {
		t.Error("expected Cause() to return original error")
	}
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
	bootErr := err.(*bootError)
	if bootErr.Phase() != "启动" {
		t.Errorf("expected phase '启动', got '%s'", bootErr.Phase())
	}
}

func TestBootError_ImplementsErrorInterface(t *testing.T) {
	t.Parallel()
	err := NewBootErr(ErrCodeConfigLoad, "测试", errors.New("test"))
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
			if !strings.Contains(errStr, expectedCode) {
				t.Errorf("expected error to contain '%s'", expectedCode)
			}
			if !strings.Contains(errStr, tc.phase) {
				t.Errorf("expected error to contain phase '%s'", tc.phase)
			}
		})
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
	if !strings.Contains(errStr, "\n\nAnalysis:") {
		t.Error("expected Analysis to be separated by double newline")
	}
	if !strings.Contains(errStr, "\n\nSuggestions:") {
		t.Error("expected Suggestions to be separated by double newline")
	}
	if !strings.Contains(errStr, "\n  - ") {
		t.Error("expected suggestions to be formatted with bullet points")
	}
}
