package boot

import (
	"context"
	"reflect"
	"strings"

	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	contextpkg "github.com/xudefa/enhance/context"
	"github.com/xudefa/enhance/core"
)

// appCtxAdapter 适配 DefaultApplicationContext 到 boot.ApplicationContext
//
// DefaultApplicationContext.EventBus() 返回 *event.EventBus，
// 而 boot.ApplicationContext.EventBus() 要求返回 interface{ Publish(...) }，
// 在 Go 中这被视为不同签名，需要显式适配。
type appCtxAdapter struct {
	ctx  *contextpkg.DefaultApplicationContext
	gctx context.Context
}

func newAppCtx(ctx *contextpkg.DefaultApplicationContext, gctx context.Context) *appCtxAdapter {
	return &appCtxAdapter{ctx: ctx, gctx: gctx}
}

// Context 返回应用上下文关联的 Go context.Context。
func (a *appCtxAdapter) Context() context.Context {
	return a.gctx
}

// Container 返回 IoC 容器实例。
func (a *appCtxAdapter) Container() core.Container {
	return a.ctx.Container()
}

// Environment 返回环境配置实例。
func (a *appCtxAdapter) Environment() *environment.Environment {
	return a.ctx.Environment()
}

// Register 向 IoC 容器注册指定类型的 Bean。
func (a *appCtxAdapter) Register(t reflect.Type, opts ...core.BeanOption) error {
	return a.ctx.Register(t, opts...)
}

// GetByType 根据类型从 IoC 容器获取 Bean 实例。
func (a *appCtxAdapter) GetByType(t reflect.Type) (any, error) {
	instances, err := a.ctx.Container().Get(t)
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return nil, core.ErrBeanNotFound
	}
	return instances[0], nil
}

// EventBus 返回事件总线实例。
func (a *appCtxAdapter) EventBus() EventBusResult {
	return a.ctx.EventBus()
}

// conditionCtx 适配 DefaultApplicationContext 到 condition.ConditionContext
//
// DefaultApplicationContext 的方法签名（如 Environment() *environment.Environment）
// 与 condition.ConditionContext 要求的签名不完全一致，因此需要此适配器桥接。
type conditionCtx struct {
	ctx *contextpkg.DefaultApplicationContext
}

func newConditionCtx(ctx *contextpkg.DefaultApplicationContext) *conditionCtx {
	return &conditionCtx{ctx: ctx}
}

// Environment 返回环境访问器，用于条件判断。
func (c *conditionCtx) Environment() condition.EnvironmentAccessor {
	return c.ctx.Environment()
}

// containerAccessorAdapter 适配 core.Container 到 condition.ContainerAccessor
type containerAccessorAdapter struct {
	container core.Container
}

// Has 检查容器中是否存在指定 ID 的 Bean。
func (a *containerAccessorAdapter) Has(id string) bool {
	// 优先通过 Bean 注册表按名称匹配（含自定义名称）
	if beans := a.container.ListBeans(); len(beans) > 0 {
		if _, ok := beans[id]; ok {
			return true
		}
		for name := range beans {
			if idx := strings.LastIndex(name, "#"); idx >= 0 && name[idx+1:] == id {
				return true
			}
		}
	}

	// 兼容旧行为：通过类型生成 ID 匹配
	for _, bean := range a.container.GetAll() {
		beanType := reflect.TypeOf(bean)
		if beanType == nil {
			continue
		}
		if a.container.Generate(beanType) == id {
			return true
		}
	}
	return false
}

// Container 返回容器访问器，用于条件判断。
func (c *conditionCtx) Container() condition.ContainerAccessor {
	return &containerAccessorAdapter{container: c.ctx.Container()}
}

// GetBeanByType 根据类型从容器中获取 Bean 实例。
func (c *conditionCtx) GetBeanByType(t reflect.Type) (any, bool) {
	instances, err := c.ctx.Container().Get(t)
	if err != nil || len(instances) == 0 {
		return nil, false
	}
	return instances[0], true
}

// HasProperty 检查环境中是否存在指定配置属性。
func (c *conditionCtx) HasProperty(key string) bool {
	return c.ctx.HasProperty(key)
}

// GetProperty 从环境中获取指定配置属性的值。
func (c *conditionCtx) GetProperty(key string) (any, bool) {
	return c.ctx.GetProperty(key)
}

var _ condition.ConditionContext = (*conditionCtx)(nil)
