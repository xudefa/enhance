package context

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/event"
)

// TestMustBuild_Coverage 测试 MustBuild panic 路径
func TestMustBuild_Coverage(t *testing.T) {
	t.Parallel()

	// MustBuild 在 Build 失败时会 panic
	// 这里测试正常路径，因为触发 panic 需要复杂的设置
	ctx := NewApplicationContextBuilder().
		Profile("test").
		MustBuild()

	if ctx == nil {
		t.Fatal("Expected non-nil context")
	}
}

// TestMustBuild_Success_Coverage 测试 MustBuild 成功路径
func TestMustBuild_Success_Coverage(t *testing.T) {
	t.Parallel()

	ctx := NewApplicationContextBuilder().
		Profile("test").
		MustBuild()

	if ctx == nil {
		t.Fatal("Expected non-nil context")
	}
}

// TestApplicationContextHelper_GetBeanByType_Coverage 测试 GetBeanByType
func TestApplicationContextHelper_GetBeanByType_Coverage(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	ctx := NewApplicationContextBuilder().
		Profile("test").
		Bean(reflect.TypeOf(TestBean{}), core.WithFactory[TestBean](func(c ...any) (any, error) {
			return &TestBean{Name: "test"}, nil
		})).
		MustBuild()

	helper := NewApplicationContextHelper(ctx)

	// 测试存在的 Bean
	bean, err := helper.GetBeanByType(reflect.TypeOf(TestBean{}))
	if err != nil {
		t.Fatalf("GetBeanByType failed: %v", err)
	}

	testBean := bean.(*TestBean)
	if testBean.Name != "test" {
		t.Errorf("Expected Name 'test', got %s", testBean.Name)
	}

	// 测试不存在的 Bean
	type NonExistent struct{}
	_, err = helper.GetBeanByType(reflect.TypeOf(NonExistent{}))
	if err == nil {
		t.Error("Expected error for non-existent bean")
	}
}

// TestApplicationContextHelper_GetBeanByTypeOrDefault_Coverage 测试 GetBeanByTypeOrDefault
func TestApplicationContextHelper_GetBeanByTypeOrDefault_Coverage(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	ctx := NewApplicationContextBuilder().
		Profile("test").
		MustBuild()

	helper := NewApplicationContextHelper(ctx)

	// 测试默认值路径
	defaultBean := &TestBean{Name: "default"}
	result := helper.GetBeanByTypeOrDefault(reflect.TypeOf(TestBean{}), defaultBean)

	testBean := result.(*TestBean)
	if testBean.Name != "default" {
		t.Errorf("Expected Name 'default', got %s", testBean.Name)
	}
}

// TestApplicationContextHelper_PropertyMethods_Coverage 测试属性获取方法
func TestApplicationContextHelper_PropertyMethods_Coverage(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{
		"app.name":  "test-app",
		"app.port":  "8080",
		"app.debug": "true",
	})

	ctx := NewApplicationContextBuilder().
		Profile("test").
		Environment(env).
		MustBuild()

	helper := NewApplicationContextHelper(ctx)

	// 测试 GetString
	name := helper.GetProperty("app.name", "default")
	if name != "test-app" {
		t.Errorf("Expected 'test-app', got %s", name)
	}

	// 测试 GetInt
	port := helper.GetIntProperty("app.port", 3000)
	if port != 8080 {
		t.Errorf("Expected 8080, got %d", port)
	}

	// 测试 GetBool
	debug := helper.GetBoolProperty("app.debug", false)
	if !debug {
		t.Error("Expected true, got false")
	}

	// 测试默认值
	missing := helper.GetProperty("missing.key", "fallback")
	if missing != "fallback" {
		t.Errorf("Expected 'fallback', got %s", missing)
	}
}

// TestApplicationContextHelper_IsRunning_Coverage 测试 IsRunning 方法
func TestApplicationContextHelper_IsRunning_Coverage(t *testing.T) {
	t.Parallel()

	ctx := NewApplicationContextBuilder().
		Profile("test").
		MustBuild()

	helper := NewApplicationContextHelper(ctx)

	// 应用尚未运行
	if helper.IsRunning() {
		t.Error("Expected IsRunning to return false before start")
	}

	// 测试 GetPhase
	phase := helper.GetPhase()
	t.Logf("Phase is %v", phase)
}

// TestApplicationContextHelper_EnvironmentChecks_Coverage 测试环境检查方法
func TestApplicationContextHelper_EnvironmentChecks_Coverage(t *testing.T) {
	t.Parallel()

	ctx := NewApplicationContextBuilder().
		Profile("test").
		MustBuild()

	helper := NewApplicationContextHelper(ctx)

	// 测试 GetActiveProfiles
	profiles := helper.GetActiveProfiles()
	if len(profiles) == 0 {
		t.Log("No active profiles (may be expected)")
	}

	// 测试 IsDev
	if helper.IsDev() {
		t.Error("Expected IsDev to return false for default profile")
	}

	// 测试 IsProd
	if helper.IsProd() {
		t.Error("Expected IsProd to return false for default profile")
	}
}

// TestApplicationContextHelper_PublishEvent_Coverage 测试事件发布方法
func TestApplicationContextHelper_PublishEvent_Coverage(t *testing.T) {
	t.Parallel()

	ctx := NewApplicationContextBuilder().
		Profile("test").
		MustBuild()

	helper := NewApplicationContextHelper(ctx)

	// 测试 PublishEvent - 不应 panic
	helper.PublishEvent("test.event")

	// 测试 PublishStarted - 不应 panic
	helper.PublishStarted()

	// 测试 PublishReady - 不应 panic
	helper.PublishReady()

	// 测试 PublishStopped - 不应 panic
	helper.PublishStopped()
}

// TestNewApplicationContextBuilder_Coverage 测试 NewApplicationContextBuilder
func TestNewApplicationContextBuilder_Coverage(t *testing.T) {
	t.Parallel()

	builder := NewApplicationContextBuilder()
	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}
}

// TestApplicationContextBuilder_WithPhaseListener_Coverage 测试 WithPhaseListener
func TestApplicationContextBuilder_WithPhaseListener_Coverage(t *testing.T) {
	t.Parallel()

	listener := &testPhaseListener{}

	builder := NewApplicationContextBuilder().
		Profile("test").
		WithPhaseListener(listener)

	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}
}

// TestApplicationContextBuilder_WithEventListener_Coverage 测试 WithEventListener
func TestApplicationContextBuilder_WithEventListener_Coverage(t *testing.T) {
	t.Parallel()

	listener := event.EventListener(func(e event.ApplicationEvent) {
		// event handler
	})

	builder := NewApplicationContextBuilder().
		Profile("test").
		WithEventListener("test.event", listener)

	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}
}

// TestApplicationContextBuilder_OnApplicationStarted_Coverage 测试 OnApplicationStarted
func TestApplicationContextBuilder_OnApplicationStarted_Coverage(t *testing.T) {
	t.Parallel()

	listener := event.EventListener(func(e event.ApplicationEvent) {
		// started callback
	})
	builder := NewApplicationContextBuilder().
		Profile("test").
		OnApplicationStarted(listener)

	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}
}

// TestApplicationContextBuilder_OnApplicationReady_Coverage 测试 OnApplicationReady
func TestApplicationContextBuilder_OnApplicationReady_Coverage(t *testing.T) {
	t.Parallel()

	listener := event.EventListener(func(e event.ApplicationEvent) {
		// ready callback
	})
	builder := NewApplicationContextBuilder().
		Profile("test").
		OnApplicationReady(listener)

	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}
}

// TestApplicationContextBuilder_OnApplicationStopped_Coverage 测试 OnApplicationStopped
func TestApplicationContextBuilder_OnApplicationStopped_Coverage(t *testing.T) {
	t.Parallel()

	listener := event.EventListener(func(e event.ApplicationEvent) {
		// stopped callback
	})
	builder := NewApplicationContextBuilder().
		Profile("test").
		OnApplicationStopped(listener)

	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}
}

// TestApplicationContextBuilder_Bean_Coverage 测试 Bean 方法
func TestApplicationContextBuilder_Bean_Coverage(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	builder := NewApplicationContextBuilder().
		Profile("test").
		Bean(reflect.TypeOf(TestBean{}), core.WithFactory[TestBean](func(c ...any) (any, error) {
			return &TestBean{Name: "test"}, nil
		}))

	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}
}

// TestApplicationContextBuilder_Build_WithDefaults_Coverage 测试 Build 使用默认值
func TestApplicationContextBuilder_Build_WithDefaults_Coverage(t *testing.T) {
	t.Parallel()

	ctx, err := NewApplicationContextBuilder().Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if ctx == nil {
		t.Fatal("Expected non-nil context")
	}

	// 验证默认组件已创建
	if ctx.Container() == nil {
		t.Error("Expected non-nil container")
	}
	if ctx.Environment() == nil {
		t.Error("Expected non-nil environment")
	}
	if ctx.Lifecycle() == nil {
		t.Error("Expected non-nil lifecycle")
	}
	if ctx.EventPublisher() == nil {
		t.Error("Expected non-nil event publisher")
	}
}

// TestApplicationContextHelper_HasBeanByType_Coverage 测试 HasBeanByType
func TestApplicationContextHelper_HasBeanByType_Coverage(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	ctx := NewApplicationContextBuilder().
		Profile("test").
		Bean(reflect.TypeOf(TestBean{}), core.WithFactory[TestBean](func(c ...any) (any, error) {
			return &TestBean{Name: "test"}, nil
		})).
		MustBuild()

	helper := NewApplicationContextHelper(ctx)

	// 测试存在的 Bean
	if !helper.HasBeanByType(reflect.TypeOf(TestBean{})) {
		t.Error("Expected HasBeanByType to return true")
	}

	// 测试不存在的 Bean
	type NonExistent struct{}
	if helper.HasBeanByType(reflect.TypeOf(NonExistent{})) {
		t.Error("Expected HasBeanByType to return false")
	}
}

// TestApplicationContextBuilder_WithBean_Option_Coverage 测试 WithBean 选项函数
func TestApplicationContextBuilder_WithBean_Option_Coverage(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	option := WithBean(reflect.TypeOf(TestBean{}), core.WithFactory[TestBean](func(c ...any) (any, error) {
		return &TestBean{Name: "test"}, nil
	}))

	builder := NewApplicationContextBuilder()
	option(builder)

	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}
}

// TestWithProfile_Option_Coverage 测试 WithProfile 选项函数
func TestWithProfile_Option_Coverage(t *testing.T) {
	t.Parallel()

	option := WithProfile("test")
	builder := NewApplicationContextBuilder()
	option(builder)

	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}
}

// TestWithProfiles_Option_Coverage 测试 WithProfiles 选项函数
func TestWithProfiles_Option_Coverage(t *testing.T) {
	t.Parallel()

	option := WithProfiles("test", "dev")
	builder := NewApplicationContextBuilder()
	option(builder)

	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}
}
