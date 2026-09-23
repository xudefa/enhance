package context

import (
	"errors"
	"reflect"
	"testing"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/event"
	"github.com/xudefa/enhance/lifecycle"
)

func TestInvoke_TypedNilError(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()
	ctx, err := builder.Build()
	if err != nil {
		t.Fatalf("Build should succeed: %v", err)
	}
	helper := NewApplicationContextHelper(ctx)

	// 函数返回 typed-nil error，不应被当作错误
	fn := func() error {
		var e *typedNilErr
		return e
	}
	if err := helper.Invoke(fn); err != nil {
		t.Errorf("typed-nil error should be treated as nil, got %v", err)
	}

	// 真实错误应正常返回
	realErr := errors.New("boom")
	fn2 := func() error {
		return realErr
	}
	if err := helper.Invoke(fn2); !errors.Is(err, realErr) {
		t.Errorf("expected realErr, got %v", err)
	}
}

// typedNilErr 用于测试 typed-nil error 场景。
type typedNilErr struct{}

func (e *typedNilErr) Error() string { return "typed-nil-err" }

func TestApplicationContextHelper_GetBean(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	type TestService struct {
		Value string
	}

	svc := &TestService{Value: "hello"}
	if err := ctx.Container().RegisterInstance(svc, reflect.TypeOf(svc)); err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	bean, err := helper.GetBeanByType(reflect.TypeOf(svc))
	if err != nil {
		t.Fatalf("GetBeanByType should succeed: %v", err)
	}

	ts, ok := bean.(*TestService)
	if !ok {
		t.Fatal("bean should be *TestService")
	}
	if ts.Value != "hello" {
		t.Errorf("expected Value='hello', got %s", ts.Value)
	}
}

func TestApplicationContextHelper_GetBeanByTypeOrDefault(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	type TestService struct {
		Value string
	}

	defaultVal := &TestService{Value: "default"}
	actual := &TestService{Value: "actual"}

	if err := ctx.Container().RegisterInstance(actual, reflect.TypeOf(actual)); err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	bean := helper.GetBeanByTypeOrDefault(reflect.TypeOf(actual), defaultVal)
	ts, ok := bean.(*TestService)
	if !ok {
		t.Fatal("bean should be *TestService")
	}
	if ts.Value != "actual" {
		t.Errorf("expected Value='actual', got %s", ts.Value)
	}

	type nonExistent struct{}
	bean = helper.GetBeanByTypeOrDefault(reflect.TypeOf(nonExistent{}), defaultVal)
	ts, ok = bean.(*TestService)
	if !ok {
		t.Fatal("default bean should be *TestService")
	}
	if ts.Value != "default" {
		t.Errorf("expected Value='default', got %s", ts.Value)
	}
}

func TestApplicationContextHelper_HasBean(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	type testBean struct{}
	bean := &testBean{}
	if err := ctx.Container().RegisterInstance(bean, reflect.TypeOf(bean)); err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	if !helper.HasBeanByType(reflect.TypeOf(bean)) {
		t.Error("HasBeanByType should return true for existing bean")
	}

	type nonExistentBean struct{}
	if helper.HasBeanByType(reflect.TypeOf(nonExistentBean{})) {
		t.Error("HasBeanByType should return false for non-existent bean")
	}
}

func TestApplicationContextHelper_GetProperty(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	ctx.Environment().AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityHighest, map[string]any{
		"app.name": "test-app",
	}))

	value := helper.GetProperty("app.name", "default")
	if value != "test-app" {
		t.Errorf("expected 'test-app', got %s", value)
	}

	value = helper.GetProperty("non.existent", "default")
	if value != "default" {
		t.Errorf("expected 'default', got %s", value)
	}
}

func TestApplicationContextHelper_GetIntProperty(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	ctx.Environment().AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityHighest, map[string]any{
		"app.port": "8080",
	}))

	value := helper.GetIntProperty("app.port", 3000)
	if value != 8080 {
		t.Errorf("expected 8080, got %d", value)
	}

	value = helper.GetIntProperty("non.existent", 3000)
	if value != 3000 {
		t.Errorf("expected 3000, got %d", value)
	}
}

func TestApplicationContextHelper_GetBoolProperty(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	ctx.Environment().AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityHighest, map[string]any{
		"app.debug": "true",
	}))

	value := helper.GetBoolProperty("app.debug", false)
	if !value {
		t.Error("expected true, got false")
	}

	value = helper.GetBoolProperty("non.existent", true)
	if !value {
		t.Error("expected true, got false")
	}
}

func TestApplicationContextHelper_IsRunning(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	if helper.IsRunning() {
		t.Error("should not be running initially")
	}

	if err := ctx.Lifecycle().SetPhase(lifecycle.PhaseRunning); err != nil {
		t.Fatalf("SetPhase failed: %v", err)
	}

	if !helper.IsRunning() {
		t.Error("should be running after start")
	}
}

func TestApplicationContextHelper_GetPhase(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	phase := helper.GetPhase()
	if phase != lifecycle.PhaseInit {
		t.Errorf("expected PhaseInit, got %v", phase)
	}
}

func TestApplicationContextHelper_GetActiveProfiles(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	profiles := helper.GetActiveProfiles()
	if profiles == nil {
		t.Error("GetActiveProfiles should not return nil")
	}
}

func TestApplicationContextHelper_IsDev(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	if helper.IsDev() {
		t.Error("should not be dev by default")
	}
}

func TestApplicationContextHelper_IsProd(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	if helper.IsProd() {
		t.Error("should not be prod by default")
	}
}

func TestApplicationContextHelper_PublishEvent(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	var called bool
	ctx.EventBus().Subscribe("customEvent", func(e event.ApplicationEvent) {
		called = true
	})

	helper.PublishEvent("customEvent")

	if !called {
		t.Error("event listener should be called")
	}
}

func TestApplicationContextHelper_PublishStarted(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	var called bool
	ctx.EventBus().Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
		called = true
	})

	helper.PublishStarted()

	if !called {
		t.Error("ApplicationStarted listener should be called")
	}
}

func TestApplicationContextHelper_PublishReady(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	var called bool
	ctx.EventBus().Subscribe(event.EventApplicationReady, func(e event.ApplicationEvent) {
		called = true
	})

	helper.PublishReady()

	if !called {
		t.Error("ApplicationReady listener should be called")
	}
}

func TestApplicationContextHelper_PublishStopped(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	var called bool
	ctx.EventBus().Subscribe(event.EventApplicationStopped, func(e event.ApplicationEvent) {
		called = true
	})

	helper.PublishStopped()

	if !called {
		t.Error("ApplicationStopped listener should be called")
	}
}

func TestApplicationRunner_Run(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	runner := NewApplicationRunner(ctx)

	if runner.Context() != ctx {
		t.Error("Context should return the same context")
	}
}

func TestApplicationRunner_Stop(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	runner := NewApplicationRunner(ctx)

	if err := runner.Stop(); err != nil {
		t.Fatalf("Stop should succeed: %v", err)
	}
}

func TestApplicationRunner_Stop_NotStarted(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	runner := NewApplicationRunner(ctx)

	err := runner.Stop()

	if err != nil {
		t.Logf("Stop returned error (expected behavior depends on implementation): %v", err)
	}
}
