// Package context 提供应用上下文管理，用于 enhance 框架。
package context

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/config/refresh"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
	"github.com/xudefa/enhance/event"
	"github.com/xudefa/enhance/lifecycle"
)

// DefaultApplicationContext 默认应用上下文实现。
//
// 组合了 IoC 容器、环境配置、生命周期管理和事件总线，
// 提供 enhance 框架的核心运行时能力。
type DefaultApplicationContext struct {
	container       core.Container               // IoC 依赖注入容器
	env             *environment.Environment     // 环境配置管理
	lifecycle       *lifecycle.LifecycleManager  // 生命周期管理器
	events          *event.EventBusWithOrdering  // 事件总线（支持优先级和异步）
	asyncPublisher  *event.AsyncPublisher        // 异步事件发布器
	refreshScopeMgr *refresh.RefreshScopeManager // 刷新作用域管理器
}

// asyncEventPublisherAdapter 将 event.AsyncPublisher 适配为 AsyncEventPublisher 接口。
//
// 提供便捷的 PublishAsync 和 PublishAsyncWithCtx 方法。
type asyncEventPublisherAdapter event.AsyncPublisher

func (a *asyncEventPublisherAdapter) PublishAsync(evt event.ApplicationEvent) {
	ctx := context.Background()
	(*event.AsyncPublisher)(a).Publish(ctx, evt)
}

func (a *asyncEventPublisherAdapter) PublishAsyncWithCtx(ctx context.Context, evt event.ApplicationEvent) {
	(*event.AsyncPublisher)(a).Publish(ctx, evt)
}

// NewApplicationContext 创建默认应用上下文实例。
func NewApplicationContext(container core.Container, env *environment.Environment, opts ...refresh.RefreshOption) *DefaultApplicationContext {
	refreshMgr := refresh.NewRefreshScopeManager(container, slog.Default(), opts...)
	events := event.NewEventBusWithOrdering()
	return &DefaultApplicationContext{
		container:       container,
		env:             env,
		lifecycle:       lifecycle.NewLifecycleManager(),
		events:          events,
		asyncPublisher:  event.NewAsyncPublisher(events, event.WithWorkerCount(5)),
		refreshScopeMgr: refreshMgr,
	}
}

func (c *DefaultApplicationContext) Container() core.Container {
	return c.container
}

func (c *DefaultApplicationContext) Environment() *environment.Environment {
	return c.env
}

func (c *DefaultApplicationContext) Lifecycle() *lifecycle.LifecycleManager {
	return c.lifecycle
}

func (c *DefaultApplicationContext) EventBus() EventBusAccess {
	return c.events
}

func (c *DefaultApplicationContext) EventPublisher() EventPublisher {
	return c.events
}

func (c *DefaultApplicationContext) AsyncEventPublisher() AsyncEventPublisher {
	return (*asyncEventPublisherAdapter)(c.asyncPublisher)
}

func (c *DefaultApplicationContext) RefreshScopeManager() *refresh.RefreshScopeManager {
	return c.refreshScopeMgr
}

func (c *DefaultApplicationContext) Register(t reflect.Type, opts ...core.BeanOption) error {
	def := registry.BeanDef{
		Type: t,
	}
	for _, opt := range opts {
		opt(&def)
	}
	return c.container.RegisterBean(def)
}

func (c *DefaultApplicationContext) GetByType(t reflect.Type) (any, error) {
	instances, err := c.container.Get(t)
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return nil, core.ErrBeanNotFound
	}
	return instances[0], nil
}

func (c *DefaultApplicationContext) Invoke(fn any) error {
	rv := reflect.ValueOf(fn)
	if rv.Kind() != reflect.Func {
		return errors.New("Invoke: fn must be a function")
	}

	fnType := rv.Type()
	numIn := fnType.NumIn()
	if numIn == 0 {
		return extractInvokeError(rv.Call(nil), fnType)
	}

	args := make([]reflect.Value, numIn)
	for i := 0; i < numIn; i++ {
		paramType := fnType.In(i)
		instances, err := c.container.Get(paramType)
		if err != nil || len(instances) == 0 {
			return fmt.Errorf("Invoke: cannot resolve parameter %d of type %s", i+1, paramType)
		}
		args[i] = reflect.ValueOf(instances[0])
	}

	return extractInvokeError(rv.Call(args), fnType)
}

// extractInvokeError 从函数调用结果中提取错误，正确处理 typed-nil error。
func extractInvokeError(results []reflect.Value, fnType reflect.Type) error {
	if fnType.NumOut() == 0 {
		return nil
	}
	lastOut := fnType.Out(fnType.NumOut() - 1)
	if !lastOut.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return nil
	}
	errVal := results[len(results)-1]
	if isNilValue(errVal) {
		return nil
	}
	err, ok := errVal.Interface().(error)
	if !ok {
		return nil
	}
	return err
}

// isNilValue 判断 reflect.Value 是否为 nil，支持 interface 包裹的 typed-nil。
func isNilValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return true
		}
		return isNilValue(v.Elem())
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return v.IsNil()
	default:
		return false
	}
}

// Start 启动应用：PhaseInit → PhaseRunning
func (c *DefaultApplicationContext) Start() error {
	c.events.Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})

	if err := c.lifecycle.SetPhase(lifecycle.PhaseRunning); err != nil {
		return err
	}

	c.events.Publish(&event.BaseEvent{EventType: event.EventApplicationReady})
	return nil
}

// Stop 停止应用：PhaseRunning → PhaseStopped
func (c *DefaultApplicationContext) Stop() error {
	if err := c.lifecycle.SetPhase(lifecycle.PhaseStopped); err != nil {
		return err
	}

	c.events.Publish(&event.BaseEvent{EventType: event.EventApplicationStopped})
	return nil
}

// IsRunning 检查应用是否运行中
func (c *DefaultApplicationContext) IsRunning() bool {
	return c.lifecycle.GetPhase() == lifecycle.PhaseRunning
}

func (c *DefaultApplicationContext) HasProperty(key string) bool {
	_, ok := c.env.GetProperty(key)
	return ok
}

func (c *DefaultApplicationContext) GetProperty(key string) (any, bool) {
	return c.env.GetProperty(key)
}
