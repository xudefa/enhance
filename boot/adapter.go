package boot

import (
	"reflect"

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
	ctx *contextpkg.DefaultApplicationContext
}

func newAppCtx(ctx *contextpkg.DefaultApplicationContext) *appCtxAdapter {
	return &appCtxAdapter{ctx: ctx}
}

func (a *appCtxAdapter) Container() core.Container {
	return a.ctx.Container()
}

func (a *appCtxAdapter) Environment() *environment.Environment {
	return a.ctx.Environment()
}

func (a *appCtxAdapter) Register(t reflect.Type, opts ...core.BeanOption) error {
	return a.ctx.Register(t, opts...)
}

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

func (c *conditionCtx) Environment() condition.EnvironmentAccessor {
	return c.ctx.Environment()
}

// containerAccessorAdapter 适配 core.Container 到 condition.ContainerAccessor
type containerAccessorAdapter struct {
	container core.Container
}

func (a *containerAccessorAdapter) Has(id string) bool {
	// 由于 core.Container.Has 需要类型参数，这里使用反射遍历来检查
	allBeans := a.container.GetAll()
	for _, bean := range allBeans {
		beanType := reflect.TypeOf(bean)
		beanID := a.container.Generate(beanType)
		if beanID == id {
			return true
		}
	}
	return false
}

func (c *conditionCtx) Container() condition.ContainerAccessor {
	return &containerAccessorAdapter{container: c.ctx.Container()}
}

func (c *conditionCtx) GetBeanByType(t reflect.Type) (any, bool) {
	instances, err := c.ctx.Container().Get(t)
	if err != nil || len(instances) == 0 {
		return nil, false
	}
	return instances[0], true
}

func (c *conditionCtx) HasProperty(key string) bool {
	return c.ctx.HasProperty(key)
}

func (c *conditionCtx) GetProperty(key string) (any, bool) {
	return c.ctx.GetProperty(key)
}

var _ condition.ConditionContext = (*conditionCtx)(nil)
