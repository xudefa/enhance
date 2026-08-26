package actuator

import (
	"context"
	"reflect"
	"testing"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/event"
)

// mockEventBusResult 实现boot.EventBusResult接口
type mockEventBusResult struct{}

func (m *mockEventBusResult) Publish(event event.ApplicationEvent) {}

// mockApplicationContext 实现boot.ApplicationContext接口
type mockApplicationContext struct {
	ctx         context.Context
	container   core.Container
	environment *environment.Environment
	eventBus    boot.EventBusResult
}

func (m *mockApplicationContext) Context() context.Context {
	return m.ctx
}

func (m *mockApplicationContext) Container() core.Container {
	return m.container
}

func (m *mockApplicationContext) Environment() *environment.Environment {
	return m.environment
}

func (m *mockApplicationContext) Register(t reflect.Type, opts ...core.BeanOption) error {
	// 这个方法不应该被autoconfig调用，autoconfig直接使用Container().RegisterInstance
	// 这里只是实现接口，实际不会被用到
	return nil
}

func (m *mockApplicationContext) GetByType(t reflect.Type) (any, error) {
	beans, err := m.container.Get(t)
	if err != nil {
		return nil, err
	}
	if len(beans) > 0 {
		return beans[0], nil
	}
	return nil, nil
}

func (m *mockApplicationContext) EventBus() boot.EventBusResult {
	if m.eventBus == nil {
		return &mockEventBusResult{}
	}
	return m.eventBus
}

func TestActuatorAutoConfiguration_Configure(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewDefaultPropertySource("test", map[string]any{
		"actuator.enabled": "true",
		"app.name":         "test-app",
		"app.version":      "1.0.0",
	}))

	ctx := &mockApplicationContext{
		ctx:         context.Background(),
		container:   container,
		environment: env,
		eventBus:    &mockEventBusResult{},
	}

	autoConfig := &ActuatorAutoConfiguration{}
	err := autoConfig.Configure(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestActuatorAutoConfiguration_Configure_WithoutAppName(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewDefaultPropertySource("test", map[string]any{
		"actuator.enabled": "true",
	}))

	ctx := &mockApplicationContext{
		ctx:         context.Background(),
		container:   container,
		environment: env,
		eventBus:    &mockEventBusResult{},
	}

	autoConfig := &ActuatorAutoConfiguration{}
	err := autoConfig.Configure(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
