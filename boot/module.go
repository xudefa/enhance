package boot

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
	"github.com/xudefa/enhance/lifecycle"
)

// ==================== Module 结构体 ====================

// Module 可组合的配置单元。
//
// Module 是 Go 风格的显式组合方式，替代全局 init() 注册。
// 每个 Module 可以独立测试、独立复用。
type Module struct {
	moduleName string                       // 模块名称，用于日志和调试
	beans      []BeanProvider               // 要注册的 Bean 提供者列表
	invokes    []func(core.Container) error // 安装时立即调用的函数列表
	hooks      []lifecycle.Hook             // 生命周期钩子列表
	starters   []Starter                    // 要启动的 Starter 列表
	conditions []condition.Condition        // 模块生效的条件
}

// Provide 注册 Bean（泛型工厂方式，Go 惯用）。
//
// 构造函数接收容器并返回实例，依赖显式注入。
//
// 示例:
//
//	boot.Provide(func(c core.Container) (*UserService, error) {
//	    repo, err := core.GetBean[*UserRepository](c)
//	    if err != nil {
//	        return nil, err
//	    }
//	    return NewUserService(repo), nil
//	})
func Provide[T any](factory func(core.Container) (T, error), opts ...core.BeanOption) BeanProvider {
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

// ProvideReflect 通过反射构造函数注册 Bean（低层 API）。
//
// 构造函数参数自动从容器注入。新代码推荐使用泛型 Provide。
//
// Deprecated: 使用 Provide[T] 替代。
func ProvideReflect(constructor any) BeanProvider {
	return func(c core.Container) error {
		if constructor == nil {
			return fmt.Errorf("constructor must not be nil")
		}
		typ := reflect.TypeOf(constructor)
		if typ.Kind() != reflect.Func {
			return core.ErrInvalidBeanName
		}
		if typ.NumOut() < 1 {
			return fmt.Errorf("constructor must return at least one value")
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
				if typ.NumOut() >= 2 {
					if !isNilReflectValue(results[1]) {
						if err, ok := results[1].Interface().(error); ok {
							return nil, err
						}
					}
				}
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
			var sb strings.Builder
			sb.WriteString(result.moduleName)
			sb.WriteString("+")
			sb.WriteString(mod.moduleName)
			result.moduleName = sb.String()
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

// ApplicationOption 已合并为 BootOption 别名，保留用于向后兼容。
//
// Deprecated: 请直接使用 BootOption（如 WithModules）。
type ApplicationOption = BootOption

// WithModulesOption 通过 BootOption 方式添加模块。
//
// 支持 Module 和 *ModuleBuilder 混合使用。
//
// Deprecated: 使用 WithModules 替代。
func WithModulesOption(modules ...any) BootOption {
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

// isNilReflectValue 判断反射值是否为 nil，兼容接口包裹 typed-nil（如 (*MyErr)(nil)）的情况。
func isNilReflectValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	case reflect.Interface:
		if v.IsNil() {
			return true
		}
		inner := v.Elem()
		if !inner.IsValid() {
			return true
		}
		switch inner.Kind() {
		case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice, reflect.Interface:
			return inner.IsNil()
		}
		return false
	}
	return false
}
