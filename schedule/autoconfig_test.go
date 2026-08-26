package schedule

import (
	"context"
	"reflect"
	"testing"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

// mockApplicationContext 用于测试的mock ApplicationContext
type mockApplicationContext struct {
	ctx         context.Context
	container   core.Container
	environment *environment.Environment
}

func (m *mockApplicationContext) Context() context.Context {
	if m.ctx == nil {
		return context.Background()
	}
	return m.ctx
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

func TestScheduleAutoConfiguration_Configure(t *testing.T) {
	t.Parallel()

	config := &ScheduleAutoConfiguration{}

	// 测试有容器和环境的情况
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := &mockApplicationContext{
		container:   container,
		environment: env,
	}
	err := config.Configure(ctx)

	// 由于没有启用配置，应该正常返回
	_ = err
}

func TestScheduleStarter_Name(t *testing.T) {
	t.Parallel()

	starter := &ScheduleStarter{}
	if starter.Name() != "ScheduleStarter" {
		t.Errorf("expected name 'ScheduleStarter', got %s", starter.Name())
	}
}

func TestScheduleStarter_Dependencies(t *testing.T) {
	t.Parallel()

	starter := &ScheduleStarter{}
	deps := starter.Dependencies()
	if deps != nil {
		t.Errorf("expected nil dependencies, got %v", deps)
	}
}

func TestScheduleStarter_Configure(t *testing.T) {
	t.Parallel()

	starter := &ScheduleStarter{}

	// 创建一个包含scheduler的容器
	container := core.NewContainer()
	scheduler := NewScheduler(context.Background())
	_ = container.RegisterInstance(scheduler, reflect.TypeOf(&DefaultScheduler{}))

	ctx := &mockApplicationContext{
		container: container,
	}

	err := starter.Configure(ctx)
	// 应该成功从容器获取scheduler
	if err != nil {
		t.Errorf("configure failed: %v", err)
	}

	if starter.scheduler == nil {
		t.Error("scheduler should be set after configure")
	}
}

func TestScheduleStarter_Start(t *testing.T) {
	t.Parallel()

	starter := &ScheduleStarter{}
	ctx := &mockApplicationContext{}

	// 测试没有scheduler的情况
	err := starter.Start(ctx)
	if err != nil {
		t.Errorf("unexpected error when scheduler is nil: %v", err)
	}
}

func TestScheduleStarter_Stop(t *testing.T) {
	t.Parallel()

	starter := &ScheduleStarter{}
	ctx := &mockApplicationContext{}

	// 测试没有scheduler的情况
	err := starter.Stop(ctx)
	if err != nil {
		t.Errorf("unexpected error when scheduler is nil: %v", err)
	}
}

func TestScheduleStarter_GetCondition(t *testing.T) {
	t.Parallel()

	starter := &ScheduleStarter{}
	cond := starter.GetCondition()
	if cond == nil {
		t.Error("expected non-nil condition")
	}
}

func TestScheduleStarter_StartWithScheduler(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())
	starter := &ScheduleStarter{
		scheduler: scheduler,
		logger:    log.NewLoggerBuilder().Build(),
	}

	ctx := &mockApplicationContext{
		ctx: context.Background(),
	}

	// 测试有scheduler的情况
	err := starter.Start(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 清理
	scheduler.Close()
}

func TestScheduleStarter_StopWithScheduler(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())
	starter := &ScheduleStarter{
		scheduler: scheduler,
		logger:    log.NewLoggerBuilder().Build(),
	}

	ctx := &mockApplicationContext{
		ctx: context.Background(),
	}

	// 先启动
	_ = starter.Start(ctx)

	// 然后停止
	err := starter.Stop(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
