package mvc

import (
	"testing"

	"github.com/xudefa/enhance/web/core"
)

type testController struct{}

func (c *testController) Routes(router core.Router) {
	// Test implementation - no routes needed for testing
}

func TestRegisterController(t *testing.T) {
	t.Parallel()
	ClearControllers()

	ctrl := &testController{}
	RegisterController(ctrl)

	controllers := GetControllers()
	if len(controllers) != 1 {
		t.Errorf("expected 1 controller, got %d", len(controllers))
	}
}

func TestGetControllers_Empty(t *testing.T) {
	t.Parallel()
	ClearControllers()

	controllers := GetControllers()
	if len(controllers) != 0 {
		t.Errorf("expected 0 controllers, got %d", len(controllers))
	}
}

func TestGetControllers_Multiple(t *testing.T) {
	t.Parallel()
	ClearControllers()

	ctrl1 := &testController{}
	ctrl2 := &testController{}

	RegisterController(ctrl1)
	RegisterController(ctrl2)

	controllers := GetControllers()
	if len(controllers) != 2 {
		t.Errorf("expected 2 controllers, got %d", len(controllers))
	}
}

func TestClearControllers(t *testing.T) {
	t.Parallel()
	RegisterController(&testController{})
	RegisterController(&testController{})

	ClearControllers()

	controllers := GetControllers()
	if len(controllers) != 0 {
		t.Errorf("expected 0 controllers after clear, got %d", len(controllers))
	}
}

func TestGetControllers_ReturnsCopy(t *testing.T) {
	t.Parallel()
	ClearControllers()

	ctrl := &testController{}
	RegisterController(ctrl)

	controllers1 := GetControllers()
	controllers2 := GetControllers()

	// 验证返回的是副本而不是同一个切片
	if &controllers1[0] == &controllers2[0] {
		t.Error("expected GetControllers to return a copy")
	}
}
