package boot

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
)

func TestPluginManager_Register(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	plugin := &mockPlugin{name: "test", version: "1.0.0"}

	err := pm.Register(plugin)
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	// 重复注册应该失败
	err = pm.Register(plugin)
	if err == nil {
		t.Error("Register() duplicate should return error")
	}

	// 获取插件
	got, ok := pm.Get("test")
	if !ok {
		t.Error("Get() should return true")
	}
	if got.Name() != "test" {
		t.Errorf("Get().Name() = %q, want %q", got.Name(), "test")
	}
}

func TestPluginManager_List(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	pm.Register(&mockPlugin{name: "a", version: "1.0.0"})
	pm.Register(&mockPlugin{name: "b", version: "2.0.0"})

	plugins := pm.List()
	if len(plugins) != 2 {
		t.Errorf("List() = %d, want 2", len(plugins))
	}
}

func TestPluginManager_Get_NotFound(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	_, ok := pm.Get("nonexistent")
	if ok {
		t.Error("Get() should return false for nonexistent plugin")
	}
}

func TestPluginManager_Register_Nil(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Register() nil should panic")
		}
	}()
	_ = pm.Register(nil)
}

func TestPluginManager_InitAll_Error(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	plugin := &mockPlugin{name: "fail", initErr: context.Canceled}
	pm.Register(plugin)
	ctx := &mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	}
	pm.SetContext(ctx)
	err := pm.InitAll()
	if err == nil {
		t.Error("InitAll() should return error")
	}
}

func TestPluginManager_StartAll_Error(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	plugin := &mockPlugin{name: "fail", startErr: context.Canceled}
	pm.Register(plugin)
	ctx := &mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	}
	pm.SetContext(ctx)
	pm.InitAll()
	err := pm.StartAll()
	if err == nil {
		t.Error("StartAll() should return error")
	}
}

func TestPluginManager_StopAll_Error(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	plugin := &mockPlugin{name: "fail", stopErr: context.Canceled}
	pm.Register(plugin)
	ctx := &mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	}
	pm.SetContext(ctx)
	pm.InitAll()
	pm.StartAll()
	err := pm.StopAll()
	if err == nil {
		t.Error("StopAll() should return error")
	}
}
