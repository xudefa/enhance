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
