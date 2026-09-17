package boot

import (
	"errors"
	"testing"
)

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
