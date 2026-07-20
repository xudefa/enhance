package boot

import (
	"reflect"

	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
	"github.com/xudefa/enhance/lifecycle"
)

// Provide 通过构造函数注册 Bean
//
// 构造函数返回值类型将作为 Bean 的类型。
// 构造函数的参数将自动从容器中注入。
//
// 示例:
//
//	func NewUserService(repo UserRepository) *UserService {
//	    return &UserService{repo: repo}
//	}
//
//	boot.Provide(NewUserService)
func Provide(constructor any) BeanProvider {
	return func(c core.Container) error {
		typ := reflect.TypeOf(constructor)
		if typ.Kind() != reflect.Func {
			return core.ErrInvalidBeanName
		}

		numIn := typ.NumIn()
		def := registry.BeanDef{
			Type: typ.Out(0),
			Factory: func(args ...any) (any, error) {
				fnVal := reflect.ValueOf(constructor)
				callArgs := make([]reflect.Value, numIn)

				for i := 0; i < numIn; i++ {
					paramType := typ.In(i)
					instances, err := c.Get(paramType)
					if err != nil {
						return nil, err
					}
					if len(instances) == 0 {
						return nil, core.ErrBeanNotFound
					}
					callArgs[i] = reflect.ValueOf(instances[0])
				}

				results := fnVal.Call(callArgs)
				return results[0].Interface(), nil
			},
		}

		return c.RegisterBean(def)
	}
}

// Invoke 创建一个安装时立即调用的函数
//
// 用于在模块安装后执行初始化逻辑（如数据库迁移、缓存预热等）。
// 函数的参数会自动从容器中注入。
//
// 示例:
//
//	boot.Invoke(func(db *Database) error {
//	    return db.Migrate()
//	})
func Invoke(fn any) BeanProvider {
	return func(c core.Container) error {
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

		fnVal.Call(args)
		return nil
	}
}

// ProvideBean 注册现有实例
//
// 示例:
//
//	cfg := &Config{Port: 8080}
//	boot.ProvideBean(cfg)
func ProvideBean[T any](bean T, opts ...core.BeanOption) BeanProvider {
	return func(c core.Container) error {
		def := registry.BeanDef{
			Type: reflect.TypeOf((*T)(nil)).Elem(),
			Factory: func(args ...any) (any, error) {
				return bean, nil
			},
		}

		for _, opt := range opts {
			opt(&def)
		}

		return c.RegisterBean(def)
	}
}

// ProvideFactory 注册工厂函数
//
// 示例:
//
//	boot.ProvideFactory(func(c core.Container) (*Database, error) {
//	    return NewDatabase("localhost:5432")
//	})
func ProvideFactory[T any](factory func(core.Container) (T, error), opts ...core.BeanOption) BeanProvider {
	return func(c core.Container) error {
		def := registry.BeanDef{
			Type: reflect.TypeOf((*T)(nil)).Elem(),
			Factory: func(args ...any) (any, error) {
				return factory(c)
			},
		}

		for _, opt := range opts {
			opt(&def)
		}

		return c.RegisterBean(def)
	}
}

// ProvideNamed 注册带名称的 Bean
//
// 示例:
//
//	boot.ProvideNamed("primary", primaryDB)
//	boot.ProvideNamed("readonly", readonlyDB)
func ProvideNamed[T any](name string, bean T, opts ...core.BeanOption) BeanProvider {
	return func(c core.Container) error {
		def := registry.BeanDef{
			Name: name,
			Type: reflect.TypeOf((*T)(nil)).Elem(),
			Factory: func(args ...any) (any, error) {
				return bean, nil
			},
		}

		for _, opt := range opts {
			opt(&def)
		}

		return c.RegisterBean(def)
	}
}

// ProvidePrimary 注册主要 Bean（优先注入）
//
// 示例:
//
//	boot.ProvidePrimary(primaryDB) // 当有多个 Database 时优先注入
func ProvidePrimary[T any](bean T, opts ...core.BeanOption) BeanProvider {
	return func(c core.Container) error {
		def := registry.BeanDef{
			Type: reflect.TypeOf((*T)(nil)).Elem(),
			Factory: func(args ...any) (any, error) {
				return bean, nil
			},
			Primary: true,
		}

		for _, opt := range opts {
			opt(&def)
		}

		return c.RegisterBean(def)
	}
}

// ConditionalModule 创建条件化模块
//
// 仅当所有条件匹配时，模块才会生效。
//
// 示例:
//
//	var RedisModule = boot.ConditionalModule(
//	    []condition.Condition{
//	        condition.OnProperty("cache.type", "redis"),
//	    },
//	    boot.Module{
//	        Beans: []boot.BeanProvider{
//	            boot.Provide(NewRedisClient),
//	        },
//	    },
//	)
func ConditionalModule(conds []condition.Condition, mod Module) Module {
	mod.conditions = append(mod.conditions, conds...)
	return mod
}

// NamedModule 创建带名称的模块
//
// 示例:
//
//	var DB = boot.NamedModule("database", boot.Module{
//	    Beans: []boot.BeanProvider{
//	        boot.Provide(NewDatabase),
//	    },
//	})
func NamedModule(name string, mod Module) Module {
	mod.moduleName = name
	return mod
}

// MergeModules 合并多个模块为一个
//
// 示例:
//
//	var AppModule = boot.MergeModules(DatabaseModule, WebModule, CacheModule)
func MergeModules(modules ...Module) Module {
	result := Module{}
	for _, mod := range modules {
		result.beans = append(result.beans, mod.beans...)
		result.starters = append(result.starters, mod.starters...)
		result.conditions = append(result.conditions, mod.conditions...)
		if mod.moduleName != "" {
			if result.moduleName == "" {
				result.moduleName = mod.moduleName
				continue
			}
			result.moduleName += "+" + mod.moduleName
		}
	}
	return result
}

// ModuleName 返回模块名称
func (m Module) ModuleName() string {
	return m.moduleName
}

// Install 将模块的 Bean 注册到容器
func (m Module) Install(c core.Container) error {
	for _, provider := range m.beans {
		if err := provider(c); err != nil {
			return err
		}
	}
	// 执行 invokes
	for _, fn := range m.invokes {
		if err := fn(c); err != nil {
			return err
		}
	}
	return nil
}

// ModuleConditions 返回模块生效的条件
func (m Module) ModuleConditions() []condition.Condition {
	return m.conditions
}

// ModuleHooks 返回模块包含的生命周期钩子
func (m Module) ModuleHooks() []lifecycle.Hook {
	return m.hooks
}

// ModuleStarters 返回模块包含的 Starter
func (m Module) ModuleStarters() []Starter {
	return m.starters
}

// WithModule 添加单个模块
func WithModule(mod Module) BootOption {
	return func(cfg *BootConfig) {
		cfg.Modules = append(cfg.Modules, mod)
	}
}

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

		fnVal.Call(args)
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

// ApplicationOption 应用级选项函数
//
// 用于 NewApplicationWithOptions 中，在应用创建后执行自定义逻辑。
type ApplicationOption func(*Boot) error

// WithModulesOption 通过 ApplicationOption 方式添加模块
//
// 支持 Module 和 *ModuleBuilder 混合使用。
//
// 示例:
//
//	app, err := boot.NewApplicationWithOptions(
//	    boot.WithAppName("my-app"),
//	    boot.WithModulesOption(DatabaseModule, WebModule),
//	)
func WithModulesOption(modules ...any) ApplicationOption {
	return func(b *Boot) error {
		for _, mod := range modules {
			switch m := mod.(type) {
			case Module:
				b.config.Modules = append(b.config.Modules, m)
			case *ModuleBuilder:
				b.config.Modules = append(b.config.Modules, m.Build())
			}
		}
		return nil
	}
}

// WithModules 通过 BootOption 方式添加模块
//
// 支持 Module 和 *ModuleBuilder 混合使用。
//
// 示例:
//
//	app := boot.New(
//	    boot.WithAppName("my-app"),
//	    boot.WithModules(DatabaseModule, WebModule),
//	)
func WithModules(modules ...any) BootOption {
	return func(cfg *BootConfig) {
		for _, mod := range modules {
			switch m := mod.(type) {
			case Module:
				cfg.Modules = append(cfg.Modules, m)
			case *ModuleBuilder:
				cfg.Modules = append(cfg.Modules, m.Build())
			}
		}
	}
}
