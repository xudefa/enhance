package metrics

import (
	"context"
	"reflect"
	"testing"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
)

type mockApplicationContext struct {
	ctx         context.Context
	container   core.Container
	environment *environment.Environment
}

func (m *mockApplicationContext) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}

func (m *mockApplicationContext) Container() core.Container {
	return m.container
}

func (m *mockApplicationContext) Environment() *environment.Environment {
	return m.environment
}

func (m *mockApplicationContext) Register(t reflect.Type, opts ...core.BeanOption) error {
	return nil
}

func (m *mockApplicationContext) GetByType(t reflect.Type) (any, error) {
	if m.container != nil {
		instances, err := m.container.Get(t)
		if err != nil {
			return nil, err
		}
		if len(instances) > 0 {
			return instances[0], nil
		}
	}
	return nil, nil
}

func (m *mockApplicationContext) EventBus() boot.EventBusResult {
	return nil
}

func TestMetricsAutoConfiguration_Configure(t *testing.T) {
	t.Parallel()

	config := &MetricsAutoConfiguration{}
	container := core.NewContainer()
	env := environment.NewEnvironment()

	ctx := &mockApplicationContext{
		container:   container,
		environment: env,
	}

	err := config.Configure(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
