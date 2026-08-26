package boot

import (
	"fmt"
	"reflect"

	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/lifecycle"
)

// NewModule 创建模块（便捷构造函数）
//
// 支持两种调用方式：
//  1. NewModule() - 返回构建器，支持链式调用
//  2. NewModule(name, Bean(...), Starter(...)) - 直接创建模块
//
// 示例（链式调用）:
//
//	var DBModule = boot.NewModule().
//	    Name("database").
//	    Bean(NewDatabase).
//	    Starter(&MigrationStarter{})
//
// 示例（直接创建）:
//
//	var DBModule = boot.NewModule("database",
//	    boot.Bean(boot.Provide(NewDatabase)),
//	    boot.Starter(&MigrationStarter{}),
//	)
func NewModule(args ...any) *ModuleBuilder {
	b := &ModuleBuilder{}
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			b.name = v
		case BeanProvider:
			b.beans = append(b.beans, v)
		case Starter:
			b.starters = append(b.starters, v)
		case condition.Condition:
			b.conditions = append(b.conditions, v)
		}
	}
	return b
}

// ModuleBuilder 模块构建器，支持链式调用
type ModuleBuilder struct {
	name       string
	beans      []BeanProvider
	invokes    []func(core.Container) error
	hooks      []lifecycle.Hook
	starters   []Starter
	conditions []condition.Condition
}

// Name 设置模块名称
func (b *ModuleBuilder) Name(name string) *ModuleBuilder {
	b.name = name
	return b
}

// Bean 添加一个 Bean 提供者
func (b *ModuleBuilder) Bean(provider BeanProvider) *ModuleBuilder {
	b.beans = append(b.beans, provider)
	return b
}

// Invoke 添加一个安装时立即调用的函数
func (b *ModuleBuilder) Invoke(fn any) *ModuleBuilder {
	b.invokes = append(b.invokes, func(c core.Container) error {
		if fn == nil {
			return fmt.Errorf("invoke function must not be nil")
		}
		fnVal := reflect.ValueOf(fn)
		fnType := fnVal.Type()

		if fnType.Kind() != reflect.Func {
			return core.ErrInvalidBeanName
		}

		numIn := fnType.NumIn()
		args := make([]reflect.Value, numIn)

		for i := 0; i < numIn; i++ {
			paramType := fnType.In(i)
			instances, err := c.Get(paramType)
			if err != nil {
				return err
			}
			if len(instances) == 0 {
				return core.ErrBeanNotFound
			}
			args[i] = reflect.ValueOf(instances[0])
		}

		results := fnVal.Call(args)
		if fnType.NumOut() >= 1 {
			if !isNilReflectValue(results[len(results)-1]) {
				if err, ok := results[len(results)-1].Interface().(error); ok {
					return err
				}
			}
		}
		return nil
	})
	return b
}

// Hook 添加一个生命周期钩子
func (b *ModuleBuilder) Hook(hook lifecycle.Hook) *ModuleBuilder {
	b.hooks = append(b.hooks, hook)
	return b
}

// Hooks 添加多个生命周期钩子
func (b *ModuleBuilder) Hooks(hooks ...lifecycle.Hook) *ModuleBuilder {
	b.hooks = append(b.hooks, hooks...)
	return b
}

// Starter 添加一个 Starter
func (b *ModuleBuilder) Starter(s Starter) *ModuleBuilder {
	b.starters = append(b.starters, s)
	return b
}

// Condition 添加一个条件
func (b *ModuleBuilder) Condition(c condition.Condition) *ModuleBuilder {
	b.conditions = append(b.conditions, c)
	return b
}

// Conditions 添加多个条件
func (b *ModuleBuilder) Conditions(conds ...condition.Condition) *ModuleBuilder {
	b.conditions = append(b.conditions, conds...)
	return b
}

// Build 构建为 Module
func (b *ModuleBuilder) Build() Module {
	return Module{
		moduleName: b.name,
		beans:      b.beans,
		invokes:    b.invokes,
		hooks:      b.hooks,
		starters:   b.starters,
		conditions: b.conditions,
	}
}

// Module 将构建器转换为 Module（Build 的别名）
func (b *ModuleBuilder) Module() Module {
	return b.Build()
}

// Install 直接安装模块（便捷方法，无需先 Build）
func (b *ModuleBuilder) Install(c core.Container) error {
	return b.Build().Install(c)
}
