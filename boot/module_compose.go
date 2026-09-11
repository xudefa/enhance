package boot

import (
	"context"

	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/lifecycle"
)

// ModuleDef 创建一个带名称的模块定义（DSL 风格）。
//
// 这是模块化组合 API 的入口点，提供链式调用风格。
//
// 示例:
//
//	var DatabaseModule = boot.ModuleDef("database").
//	    DependsOn("config").
//	    Beans(
//	        boot.Provide(NewDatabase),
//	        boot.Provide(NewRepository),
//	    ).
//	    Starters(
//	        &MigrationStarter{},
//	    ).
//	    Hooks(
//	        lifecycle.OnInit(initDB),
//	    ).
//	    Build()
func ModuleDef(name string) *ModuleComposer {
	return &ModuleComposer{
		module: Module{
			moduleName: name,
		},
	}
}

// ModuleComposer 模块组合器，提供 DSL 风格的模块定义。
//
// 通过链式调用构建模块，使模块定义更加清晰和易读。
type ModuleComposer struct {
	module    Module
	dependsOn []string
}

// DependsOn 声明模块依赖的其他模块名称。
//
// 注意：当前版本仅用于文档和日志，不强制执行依赖顺序。
// 未来版本将实现真正的依赖解析和排序。
func (mc *ModuleComposer) DependsOn(modules ...string) *ModuleComposer {
	mc.dependsOn = append(mc.dependsOn, modules...)
	return mc
}

// Beans 添加 Bean 提供者列表。
func (mc *ModuleComposer) Beans(providers ...BeanProvider) *ModuleComposer {
	mc.module.beans = append(mc.module.beans, providers...)
	return mc
}

// Starters 添加 Starter 列表。
func (mc *ModuleComposer) Starters(starters ...Starter) *ModuleComposer {
	mc.module.starters = append(mc.module.starters, starters...)
	return mc
}

// Hooks 添加生命周期钩子列表。
func (mc *ModuleComposer) Hooks(hooks ...lifecycle.Hook) *ModuleComposer {
	mc.module.hooks = append(mc.module.hooks, hooks...)
	return mc
}

// Invoke 添加安装时立即调用的函数。
func (mc *ModuleComposer) Invoke(fn func(core.Container) error) *ModuleComposer {
	mc.module.invokes = append(mc.module.invokes, fn)
	return mc
}

// Conditions 添加模块生效的条件。
func (mc *ModuleComposer) Conditions(conds ...condition.Condition) *ModuleComposer {
	mc.module.conditions = append(mc.module.conditions, conds...)
	return mc
}

// OnProperty 添加基于配置属性的条件。
//
// 当指定属性存在且不为空时，模块才会生效。
func (mc *ModuleComposer) OnProperty(key string) *ModuleComposer {
	mc.module.conditions = append(mc.module.conditions, condition.OnProperty(key))
	return mc
}

// OnPropertyEquals 添加基于配置属性等于指定值的条件。
//
// 当指定属性等于期望值时，模块才会生效。
func (mc *ModuleComposer) OnPropertyEquals(key, value string) *ModuleComposer {
	mc.module.conditions = append(mc.module.conditions, condition.OnProperty(key, value))
	return mc
}

// OnPropertyMissing 添加基于配置属性不存在的条件。
//
// 当指定属性不存在时，模块才会生效。
func (mc *ModuleComposer) OnPropertyMissing(key string) *ModuleComposer {
	mc.module.conditions = append(mc.module.conditions, condition.OnMissingProperty(key))
	return mc
}

// Build 构建并返回模块。
func (mc *ModuleComposer) Build() Module {
	return mc.module
}

// ModuleGroup 模块分组，用于组织和管理多个相关模块。
//
// 示例:
//
//	var BackendModules = boot.ModuleGroup("backend").
//	    Include(DatabaseModule, CacheModule, WebModule).
//	    Build()
func ModuleGroup(name string) *ModuleGroupComposer {
	return &ModuleGroupComposer{
		name:    name,
		modules: make([]Module, 0),
	}
}

// ModuleGroupComposer 模块分组组合器。
type ModuleGroupComposer struct {
	name    string
	modules []Module
}

// Include 添加模块到分组中。
func (mgc *ModuleGroupComposer) Include(modules ...Module) *ModuleGroupComposer {
	mgc.modules = append(mgc.modules, modules...)
	return mgc
}

// Build 构建并返回模块列表。
func (mgc *ModuleGroupComposer) Build() []Module {
	return mgc.modules
}

// ModuleList 便捷函数，直接返回模块列表。
//
// 示例:
//
//	app := boot.New(
//	    boot.WithAppName("my-app"),
//	    boot.WithModulesList(boot.ModuleList(DatabaseModule, WebModule)),
//	)
func ModuleList(modules ...Module) []Module {
	return modules
}

// EmptyModule 创建一个空模块，用于占位或测试。
func EmptyModule() Module {
	return Module{}
}

// ModuleFromFunc 从函数创建模块。
//
// 函数在安装时调用，参数自动从容器注入。
//
// 示例:
//
//	var DBModule = boot.ModuleFromFunc("database", func(c core.Container) error {
//	    db, err := NewDatabase("localhost:5432")
//	    if err != nil {
//	        return err
//	    }
//	    return c.RegisterBean(db)
//	})
func ModuleFromFunc(name string, fn func(core.Container) error, opts ...ModuleOption) Module {
	mod := Module{
		moduleName: name,
		invokes:    []func(core.Container) error{fn},
	}
	for _, opt := range opts {
		opt(&mod)
	}
	return mod
}

// ModuleOption 模块选项函数。
type ModuleOption func(*Module)

// WithModuleConditions 添加模块条件。
func WithModuleConditions(conds ...condition.Condition) ModuleOption {
	return func(m *Module) {
		m.conditions = append(m.conditions, conds...)
	}
}

// WithModuleHooks 添加模块钩子。
func WithModuleHooks(hooks ...lifecycle.Hook) ModuleOption {
	return func(m *Module) {
		m.hooks = append(m.hooks, hooks...)
	}
}

// WithModuleStarters 添加模块 Starter。
func WithModuleStarters(starters ...Starter) ModuleOption {
	return func(m *Module) {
		m.starters = append(m.starters, starters...)
	}
}

// WithModuleBeans 添加 Bean 提供者。
func WithModuleBeans(providers ...BeanProvider) ModuleOption {
	return func(m *Module) {
		m.beans = append(m.beans, providers...)
	}
}

// LifecycleHook 便捷函数，创建生命周期钩子。
//
// 示例:
//
//	boot.LifecycleHook(
//	    func(ctx context.Context) error { /* init */ return nil },
//	    func(ctx context.Context) error { /* start */ return nil },
//	    func(ctx context.Context) error { /* stop */ return nil },
//	)
func LifecycleHook(onInit, onStart, onStop func(context.Context) error) lifecycle.Hook {
	return lifecycle.NewHookFunc(onInit, onStart, onStop)
}

// OnInitHook 便捷函数，创建仅 OnInit 的钩子。
func OnInitHook(fn func(context.Context) error) lifecycle.Hook {
	return lifecycle.NewHookFunc(fn, nil, nil)
}

// OnStartHook 便捷函数，创建仅 OnStart 的钩子。
func OnStartHook(fn func(context.Context) error) lifecycle.Hook {
	return lifecycle.NewHookFunc(nil, fn, nil)
}

// OnStopHook 便捷函数，创建仅 OnStop 的钩子。
func OnStopHook(fn func(context.Context) error) lifecycle.Hook {
	return lifecycle.NewHookFunc(nil, nil, fn)
}
