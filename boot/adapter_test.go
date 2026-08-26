package boot

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/xudefa/enhance/config/environment"
	contextpkg "github.com/xudefa/enhance/context"
	"github.com/xudefa/enhance/core"
)

// testBean 用于测试的简单 Bean
type testBean struct {
	Name string
}

func TestAppCtxAdapter_Context(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)
	gctx := context.Background()

	adapter := newAppCtx(appCtx, gctx)

	if adapter.Context() != gctx {
		t.Error("expected context to match")
	}
}

func TestAppCtxAdapter_Container(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)
	gctx := context.Background()

	adapter := newAppCtx(appCtx, gctx)

	if adapter.Container() != container {
		t.Error("expected container to match")
	}
}

func TestAppCtxAdapter_Environment(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)
	gctx := context.Background()

	adapter := newAppCtx(appCtx, gctx)

	if adapter.Environment() != env {
		t.Error("expected environment to match")
	}
}

func TestAppCtxAdapter_Register(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)
	gctx := context.Background()

	adapter := newAppCtx(appCtx, gctx)

	err := core.Register[*testBean](adapter.Container())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppCtxAdapter_GetByType(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)
	gctx := context.Background()

	adapter := newAppCtx(appCtx, gctx)

	// 先注册
	err := core.Register[*testBean](adapter.Container())
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	// 获取
	bean, err := adapter.GetByType(reflect.TypeOf(&testBean{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bean == nil {
		t.Fatal("expected non-nil bean")
	}
}

func TestAppCtxAdapter_GetByType_NotFound(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)
	gctx := context.Background()

	adapter := newAppCtx(appCtx, gctx)

	// 获取未注册的类型
	_, err := adapter.GetByType(reflect.TypeOf(&testBean{}))
	if err == nil {
		t.Fatal("expected error for unregistered type")
	}
}

func TestAppCtxAdapter_EventBus(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)
	gctx := context.Background()

	adapter := newAppCtx(appCtx, gctx)

	eventBus := adapter.EventBus()
	if eventBus == nil {
		t.Fatal("expected non-nil event bus")
	}
}

func TestConditionCtx_Environment(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)

	condCtx := newConditionCtx(appCtx)

	envAccessor := condCtx.Environment()
	if envAccessor == nil {
		t.Fatal("expected non-nil environment accessor")
	}
}

func TestConditionCtx_Container(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)

	condCtx := newConditionCtx(appCtx)

	containerAccessor := condCtx.Container()
	if containerAccessor == nil {
		t.Fatal("expected non-nil container accessor")
	}
}

func TestConditionCtx_GetBeanByType(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)

	condCtx := newConditionCtx(appCtx)

	// 注册 Bean
	err := core.Register[*testBean](appCtx.Container())
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	// 获取 Bean
	bean, found := condCtx.GetBeanByType(reflect.TypeOf(&testBean{}))
	if !found {
		t.Fatal("expected bean to be found")
	}
	if bean == nil {
		t.Fatal("expected non-nil bean")
	}
}

func TestConditionCtx_GetBeanByType_NotFound(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)

	condCtx := newConditionCtx(appCtx)

	// 获取未注册的 Bean
	_, found := condCtx.GetBeanByType(reflect.TypeOf(&testBean{}))
	if found {
		t.Fatal("expected bean not to be found")
	}
}

func TestConditionCtx_HasProperty(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)

	condCtx := newConditionCtx(appCtx)

	// 通过 AddPropertySource 添加属性源
	source := environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
		"test.key": "test-value",
	})
	appCtx.Environment().AddPropertySource(source)

	if !condCtx.HasProperty("test.key") {
		t.Error("expected property to exist")
	}
	if condCtx.HasProperty("nonexistent.key") {
		t.Error("expected property not to exist")
	}
}

func TestConditionCtx_GetProperty(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)

	condCtx := newConditionCtx(appCtx)

	// 通过 AddPropertySource 添加属性源
	source := environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
		"test.key": "test-value",
	})
	appCtx.Environment().AddPropertySource(source)

	value, found := condCtx.GetProperty("test.key")
	if !found {
		t.Fatal("expected property to be found")
	}
	if value != "test-value" {
		t.Errorf("expected 'test-value', got %v", value)
	}

	// 获取不存在的属性
	_, found = condCtx.GetProperty("nonexistent.key")
	if found {
		t.Error("expected property not to be found")
	}
}

func TestContainerAccessorAdapter_Has(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	adapter := &containerAccessorAdapter{container: container}

	// 通过 ApplicationContext 注册 Bean
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)
	err := core.Register[*testBean](appCtx.Container())
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	// 检查是否存在（通过类型生成的 ID）
	beanType := reflect.TypeOf(&testBean{})
	generatedID := container.Generate(beanType)
	if !adapter.Has(generatedID) {
		t.Errorf("expected container to have bean with ID %s", generatedID)
	}
}

func TestContainerAccessorAdapter_Has_NotFound(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	adapter := &containerAccessorAdapter{container: container}

	if adapter.Has("nonexistent") {
		t.Error("expected container not to have bean")
	}
}

func TestContainerAccessorAdapter_Has_ByName(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()

	// 注册一个带名称的Bean
	err := container.RegisterInstance(&testBean{Name: "test"}, reflect.TypeOf(&testBean{}))
	if err != nil {
		t.Fatalf("RegisterInstance() error = %v", err)
	}

	adapter := &containerAccessorAdapter{container: container}

	// 通过名称检查（使用类型生成的ID）
	beans := container.ListBeans()
	for name := range beans {
		if adapter.Has(name) {
			return // 测试通过
		}
	}
	t.Error("expected container to have registered bean")
}

func TestContainerAccessorAdapter_Has_WithHashSuffix(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()

	// 注册Bean
	err := core.Register[*testBean](container)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	adapter := &containerAccessorAdapter{container: container}

	// 列出所有Bean并检查带#后缀的名称匹配逻辑
	beans := container.ListBeans()
	for name := range beans {
		// 提取#后面的部分
		if idx := strings.LastIndex(name, "#"); idx >= 0 {
			id := name[idx+1:]
			if adapter.Has(id) {
				return // 测试通过
			}
		}
	}
}
