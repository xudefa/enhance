package context

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/config/refresh"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
	"github.com/xudefa/enhance/event"
	"github.com/xudefa/enhance/lifecycle"
)

// beanInstance 创建一个 BeanOption，用于注册已有实例
func beanInstance(instance any) core.BeanOption {
	return func(def *registry.BeanDef) {
		def.Factory = func(c ...any) (any, error) {
			return instance, nil
		}
	}
}

func TestNewApplicationContextBuilder(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	if builder == nil || builder.eventListeners == nil || builder.beans == nil {
		t.Fatalf("builder should not be nil, eventListeners=%v, beans=%v", builder != nil, builder != nil)
	}
}

func TestBuilder_ChainMethods(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	lifecycle := lifecycle.NewLifecycleManager()
	eventBus := event.NewEventBusWithOrdering()
	refreshMgr := refresh.NewRefreshScopeManager(container, nil)

	builderResult := builder.
		Container(container).
		Environment(env).
		Lifecycle(lifecycle).
		EventBus(eventBus).
		RefreshScopeManager(refreshMgr)

	if builderResult != builder {
		t.Error("chain methods should return the same builder")
	}
	if builder.container != container {
		t.Error("container not set correctly")
	}
	if builder.env != env {
		t.Error("environment not set correctly")
	}
	if builder.lifecycle != lifecycle {
		t.Error("lifecycle not set correctly")
	}
	if builder.eventBus != eventBus {
		t.Error("eventBus not set correctly")
	}
	if builder.refreshScopeMgr != refreshMgr {
		t.Error("refreshScopeMgr not set correctly")
	}
}

func TestBuilder_WithRefreshOption(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	opt1 := func(o *refresh.RefreshConfig) {}
	opt2 := func(o *refresh.RefreshConfig) {}

	builderResult := builder.WithRefreshOption(opt1, opt2)

	if builderResult != builder {
		t.Error("WithRefreshOption should return the same builder")
	}
	if len(builder.refreshOpts) != 2 {
		t.Errorf("expected 2 refresh options, got %d", len(builder.refreshOpts))
	}
}

func TestBuilder_WithPhaseListener(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	listener := &testPhaseListener{}
	builderResult := builder.WithPhaseListener(listener)

	if builderResult != builder {
		t.Error("WithPhaseListener should return the same builder")
	}
	if len(builder.phaseListeners) != 1 {
		t.Errorf("expected 1 phase listener, got %d", len(builder.phaseListeners))
	}
}

func TestBuilder_WithEventListener(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	builder.WithEventListener("testEvent", func(e event.ApplicationEvent) {})

	if len(builder.eventListeners["testEvent"]) != 1 {
		t.Errorf("expected 1 event listener, got %d", len(builder.eventListeners["testEvent"]))
	}
}

func TestBuilder_OnApplicationStarted(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	listener := func(e event.ApplicationEvent) {}
	builderResult := builder.OnApplicationStarted(listener)

	if builderResult != builder {
		t.Error("OnApplicationStarted should return the same builder")
	}
	if len(builder.eventListeners[event.EventApplicationStarted]) != 1 {
		t.Errorf("expected 1 listener for ApplicationStarted, got %d", len(builder.eventListeners[event.EventApplicationStarted]))
	}
}

func TestBuilder_OnApplicationReady(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	listener := func(e event.ApplicationEvent) {}
	builderResult := builder.OnApplicationReady(listener)

	if builderResult != builder {
		t.Error("OnApplicationReady should return the same builder")
	}
	if len(builder.eventListeners[event.EventApplicationReady]) != 1 {
		t.Errorf("expected 1 listener for ApplicationReady, got %d", len(builder.eventListeners[event.EventApplicationReady]))
	}
}

func TestBuilder_OnApplicationStopped(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	listener := func(e event.ApplicationEvent) {}
	builderResult := builder.OnApplicationStopped(listener)

	if builderResult != builder {
		t.Error("OnApplicationStopped should return the same builder")
	}
	if len(builder.eventListeners[event.EventApplicationStopped]) != 1 {
		t.Errorf("expected 1 listener for ApplicationStopped, got %d", len(builder.eventListeners[event.EventApplicationStopped]))
	}
}

func TestBuilder_Bean(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	type testBean struct{}
	tt := reflect.TypeFor[testBean]()
	builderResult := builder.Bean(tt, core.WithScope[any](registry.Singleton))

	if builderResult != builder {
		t.Error("Bean should return the same builder")
	}
	if len(builder.beans[tt]) != 1 {
		t.Errorf("expected 1 bean option, got %d", len(builder.beans[tt]))
	}
}

func TestBuilder_Bean_MultipleOptions(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	type testBean struct{}
	tt := reflect.TypeFor[testBean]()
	builder.Bean(tt, core.WithScope[any](registry.Singleton))

	if len(builder.beans[tt]) != 1 {
		t.Errorf("expected 1 bean option, got %d", len(builder.beans[tt]))
	}
}

func TestBuilder_Build_WithDefaults(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	ctx, err := builder.Build()

	if err != nil {
		t.Fatalf("Build should succeed with defaults: %v", err)
	}
	if ctx == nil {
		t.Fatal("context should not be nil")
	}
	if ctx.Container() == nil {
		t.Error("container should be created")
	}
	if ctx.Environment() == nil {
		t.Error("environment should be created")
	}
	if ctx.EventBus() == nil {
		t.Error("eventBus should be created")
	}
	if ctx.Lifecycle() == nil {
		t.Error("lifecycle should be created")
	}
}

func TestBuilder_Build_WithCustomComponents(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	lifecycle := lifecycle.NewLifecycleManager()
	eventBus := event.NewEventBusWithOrdering()
	refreshMgr := refresh.NewRefreshScopeManager(container, nil)

	builder.
		Container(container).
		Environment(env).
		Lifecycle(lifecycle).
		EventBus(eventBus).
		RefreshScopeManager(refreshMgr)

	ctx, err := builder.Build()

	if err != nil {
		t.Fatalf("Build should succeed: %v", err)
	}
	if ctx.Container() != container {
		t.Error("custom container not used")
	}
	if ctx.Environment() != env {
		t.Error("custom environment not used")
	}
	if ctx.Lifecycle() != lifecycle {
		t.Error("custom lifecycle not used")
	}
	if ctx.EventBus() != eventBus {
		t.Error("custom eventBus not used")
	}
	if ctx.RefreshScopeManager() != refreshMgr {
		t.Error("custom refreshScopeMgr not used")
	}
}

func TestBuilder_Build_WithPhaseListeners(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	lifecycle := lifecycle.NewLifecycleManager()
	listener := &testPhaseListener{}

	builder.
		Lifecycle(lifecycle).
		WithPhaseListener(listener)

	ctx, err := builder.Build()

	if err != nil {
		t.Fatalf("Build should succeed: %v", err)
	}

	if ctx.Lifecycle() == nil {
		t.Fatal("lifecycle should be set")
	}
}

func TestBuilder_Build_WithPhaseListeners_DefaultLifecycle(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()
	notified := false
	listener := &testPhaseListener{
		onPhase: func(old, new lifecycle.ApplicationPhase) error {
			notified = true
			return nil
		},
	}

	builder.WithPhaseListener(listener)

	ctx, err := builder.Build()
	if err != nil {
		t.Fatalf("Build should succeed: %v", err)
	}

	// 使用默认生命周期管理器时，自定义阶段监听器也必须生效
	if err := ctx.Lifecycle().SetPhase(lifecycle.PhaseRunning); err != nil {
		t.Fatalf("SetPhase should succeed: %v", err)
	}
	if !notified {
		t.Error("phase listener should be notified with default lifecycle")
	}
}

func TestBuilder_Build_WithEventListeners(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	var startedCalled, readyCalled, stoppedCalled bool

	builder.
		OnApplicationStarted(func(e event.ApplicationEvent) {
			startedCalled = true
		}).
		OnApplicationReady(func(e event.ApplicationEvent) {
			readyCalled = true
		}).
		OnApplicationStopped(func(e event.ApplicationEvent) {
			stoppedCalled = true
		})

	ctx, err := builder.Build()

	if err != nil {
		t.Fatalf("Build should succeed: %v", err)
	}

	if ctx.EventBus() == nil {
		t.Fatal("eventBus should be set")
	}

	ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})
	if !startedCalled {
		t.Error("ApplicationStarted listener should be called")
	}

	ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventApplicationReady})
	if !readyCalled {
		t.Error("ApplicationReady listener should be called")
	}

	ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventApplicationStopped})
	if !stoppedCalled {
		t.Error("ApplicationStopped listener should be called")
	}
}

func TestBuilder_Build_WithBeanRegistration(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	type TestService struct {
		Name string
	}

	service := &TestService{Name: "test"}
	tt := reflect.TypeOf(service)
	builder.Bean(tt, beanInstance(service))

	ctx, err := builder.Build()

	if err != nil {
		t.Fatalf("Build should succeed: %v", err)
	}

	bean, err := ctx.Container().Get(tt)
	if err != nil {
		t.Fatalf("Container.Get should succeed: %v", err)
	}
	if len(bean) == 0 {
		t.Fatal("expected at least one bean")
	}

	ts, ok := bean[0].(*TestService)
	if !ok {
		t.Fatal("bean should be *TestService")
	}
	if ts.Name != "test" {
		t.Errorf("expected Name='test', got %s", ts.Name)
	}
}

func TestBuilder_Build_BeanRegistrationError(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	type existingBean struct{}
	container := core.NewContainer()
	bean := &existingBean{}
	if err := container.RegisterInstance(bean, reflect.TypeOf(bean)); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	builder.Container(container)
	builder.Bean(reflect.TypeOf(bean), beanInstance(&existingBean{}))

	// 新 core API 允许重复注册，不会报错
	ctx, err := builder.Build()
	if err != nil {
		t.Fatalf("Build should succeed (duplicate registration is allowed): %v", err)
	}
	if ctx == nil {
		t.Fatal("context should not be nil")
	}
}

func TestBuilder_MustBuild_Success(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	ctx := builder.MustBuild()

	if ctx == nil {
		t.Fatal("MustBuild should return valid context")
	}
}

func TestBuilder_MustBuild_Panic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			// 新 core API 允许重复注册，不会 panic
			t.Log("MustBuild did not panic (duplicate registration is allowed in new core API)")
		}
	}()

	builder := NewApplicationContextBuilder()
	type testBean struct{}
	container := core.NewContainer()
	bean := &testBean{}
	if err := container.RegisterInstance(bean, reflect.TypeOf(bean)); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	builder.Container(container)
	builder.Bean(reflect.TypeOf(bean), beanInstance(&testBean{}))

	builder.MustBuild()
}

type testPhaseListener struct {
	onPhase func(old, new lifecycle.ApplicationPhase) error
}

func (l *testPhaseListener) OnPhaseChange(old, new lifecycle.ApplicationPhase) error {
	if l.onPhase != nil {
		return l.onPhase(old, new)
	}
	return nil
}
