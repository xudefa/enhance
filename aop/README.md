# aop 包 — 面向切面编程

> **所属层级**: Core Layer  
> **设计理念**: 横切关注点分离，动态代理  
> **设计灵感**: Spring AOP

## 概述

`aop` 包提供一个轻量级的面向切面编程（AOP）框架，允许开发者通过切面（Aspect）将横切关注点（如日志、事务、权限检查）与业务逻辑分离。

### 核心概念

| 概念 | 说明 |
|------|------|
| **Advice（通知）** | 在特定连接点执行的增强逻辑 |
| **PointCut（切点）** | 定义哪些方法需要被拦截的匹配规则 |
| **Advisor（顾问）** | 切点 + 通知的组合单元 |
| **AspectMeta（切面元数据）** | 切面的完整描述，包含切点、通知和执行顺序 |
| **Weaver（织入器）** | 将切面织入目标对象，生成代理对象 |
| **JoinPoint（连接点）** | 程序执行的某个位置（如方法调用） |

### 通知类型

| 类型 | 常量 | 执行时机 | 用途 |
|------|------|----------|------|
| 前置通知 | `AdviceBefore` | 目标方法执行之前 | 日志、权限检查 |
| 后置通知 | `AdviceAfter` | 目标方法执行之后（无论是否异常） | 资源清理 |
| 返回通知 | `AdviceAfterReturning` | 目标方法正常返回之后 | 结果处理 |
| 异常通知 | `AdviceAfterThrowing` | 目标方法抛出异常之后 | 异常处理 |
| 环绕通知 | `AdviceAround` | 包裹整个目标方法，可控制方法执行 | 事务、性能监控 |

---

## 核心接口

### Advice 接口

```go
type Advice interface {
    Type() AdviceType                              // 返回通知类型
    Apply(jp JoinPoint, proceed ProceedFunc) any   // 应用通知逻辑
}
```

### PointCut 接口

```go
type PointCut interface {
    MatchClass(c reflect.Type) bool     // 匹配类
    MatchMethod(m reflect.Method) bool  // 匹配方法
}
```

### Advisor 接口

```go
type Advisor interface {
    GetPointCut() PointCut
    GetAdvice() Advice
    Order() int
}
```

### AspectMeta 结构体

```go
type AspectMeta struct {
    Instance any        // 切面实例
    PointCut PointCut  // 切点
    Advice   Advice    // 通知
    Order    int       // 执行顺序，值越小优先级越高
}
```

### JoinPoint 接口

```go
type JoinPoint interface {
    Method() any                // 获取被拦截的方法
    Args() []any                // 获取方法调用参数
    Signature() MethodSignature // 获取方法签名
    This() any                  // 获取代理对象
    Target() any                // 获取目标对象
}

type MethodSignature interface {
    Name() string               // 方法名
    DeclaringType() reflect.Type // 声明方法的类型
}
```

### Weaver 接口

```go
type Weaver interface {
    Weave(target any) any                        // 织入目标，返回代理
    AddAspects(aspects ...*AspectMeta)           // 添加切面
}
```

---

## 快速开始

### 创建通知

```go
// 前置通知
beforeAdvice := aop.Before(func(jp aop.JoinPoint) {
    fmt.Println("方法执行前:", jp.Signature().Name())
})

// 后置通知
afterAdvice := aop.After(func(jp aop.JoinPoint) {
    fmt.Println("方法执行后:", jp.Signature().Name())
})

// 返回通知
returningAdvice := aop.AfterReturning(func(jp aop.JoinPoint, result any) {
    fmt.Println("方法返回:", result)
})

// 异常通知
throwingAdvice := aop.AfterThrowing(func(jp aop.JoinPoint, err error) {
    fmt.Println("方法异常:", err)
})

// 环绕通知
aroundAdvice := aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
    fmt.Println("方法执行前")
    result := proceed()
    fmt.Println("方法执行后")
    return result
})
```

### 创建切点

```go
// 匹配所有方法
aop.MatchAll()

// 按方法名精确匹配
aop.MatchByName("DoSomething")

// 按方法名前缀匹配
aop.MatchByNamePrefix("Get")

// 按正则表达式匹配
aop.MatchByRegex("(?i)^do.*")

// 自定义匹配器组合
aop.MatchClassMethod(
    func(t reflect.Type) bool {
        return t.Name() == "UserService"
    },
    func(m reflect.Method) bool {
        return m.Name == "DoSomething"
    },
)
```

### 创建 Advisor

```go
advisor := aop.NewAdvisor(
    aop.MatchByName("DoSomething"),
    aop.Before(func(jp aop.JoinPoint) {
        fmt.Println("执行前:", jp.Signature().Name())
    }),
    1, // 可选顺序参数，值越小优先级越高
)
```

### 织入目标对象

```go
// 创建织入器
weaver := aop.NewWeaver()

// 添加切面
weaver.AddAspects(
    &aop.AspectMeta{
        PointCut: aop.MatchByName("DoSomething"),
        Advice:   aop.Before(func(jp aop.JoinPoint) { fmt.Println("before") }),
        Order:    1,
    },
)

// 织入目标对象，返回代理
target := &UserService{}
proxy := weaver.Weave(target)

// 使用代理对象，通知会自动执行
proxy.(*UserService).DoSomething()
// 输出: before
// 输出: DoSomething did it
```

---

## API 参考

### 通知创建函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `Before(fn)` | 创建前置通知 | `aop.Before(func(jp JoinPoint) { ... })` |
| `After(fn)` | 创建后置通知 | `aop.After(func(jp JoinPoint) { ... })` |
| `AfterReturning(fn)` | 创建返回通知 | `aop.AfterReturning(func(jp JoinPoint, result any) { ... })` |
| `AfterThrowing(fn)` | 创建异常通知 | `aop.AfterThrowing(func(jp JoinPoint, err error) { ... })` |
| `Around(fn)` | 创建环绕通知 | `aop.Around(func(jp JoinPoint, proceed ProceedFunc) any { ... })` |

### 切点匹配器函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `MatchAll()` | 匹配所有类和方法 | `aop.MatchAll()` |
| `MatchByName(name)` | 按方法名精确匹配 | `aop.MatchByName("GetUser")` |
| `MatchByNamePrefix(prefix)` | 按方法名前缀匹配 | `aop.MatchByNamePrefix("Get")` |
| `MatchByRegex(pattern)` | 按正则表达式匹配方法名 | `aop.MatchByRegex("^Get.*$")` |
| `MatchClass(matcher)` | 自定义类匹配器 | `aop.MatchClass(func(t reflect.Type) bool { ... })` |
| `MatchMethod(matcher)` | 自定义方法匹配器 | `aop.MatchMethod(func(m reflect.Method) bool { ... })` |
| `MatchClassMethod(class, method)` | 同时匹配类和方法 | `aop.MatchClassMethod(classMatcher, methodMatcher)` |
| `MatchByAnnotation(annotationType)` | 按注解类型匹配 | `aop.MatchByAnnotation(reflect.TypeOf((*Transactional)(nil)).Elem())` |
| `MatchInterface(iface)` | 匹配实现指定接口的类 | `aop.MatchInterface((*ServiceInterface)(nil))` |

### 排序函数

```go
// 按 Order 值升序排列切面列表
aop.SortAspectsByOrder(aspects)
```

### InterfaceProxyWrapper

接口代理包装器，通过反射转发接口方法调用，支持 AOP 切面织入：

```go
type InterfaceProxyWrapper struct {
    target      any
    advisors    []*AspectMeta
    iface       reflect.Type
    methodCache map[string]reflect.Method
    cacheMu     sync.RWMutex
    executor    ChainExecutor
}
```

主要方法：

```go
// 调用指定方法（不带上下文）
func (w *InterfaceProxyWrapper) Invoke(methodName string, args ...any) (any, error)

// 带上下文的方法调用
func (w *InterfaceProxyWrapper) InvokeContext(ctx context.Context, methodName string, args ...any) (any, error)

// 返回原始目标对象
func (w *InterfaceProxyWrapper) Unwrap() any
```

---

## 使用示例

### 日志切面

```go
// 创建日志切面
logAspect := &aop.AspectMeta{
    PointCut: aop.MatchByNamePrefix("Get"),
    Advice: aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
        start := time.Now()
        fmt.Printf("调用方法: %s\n", jp.Signature().Name())
        result := proceed()
        fmt.Printf("方法执行耗时: %v\n", time.Since(start))
        return result
    }),
    Order: 1,
}

// 织入目标
weaver := aop.NewWeaver()
weaver.AddAspects(logAspect)
proxy := weaver.Weave(&UserService{})
```

### 事务切面

```go
// 创建事务切面
txAspect := &aop.AspectMeta{
    PointCut: aop.MatchByAnnotation(reflect.TypeOf((*Transactional)(nil)).Elem()),
    Advice: aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
        tx := db.Begin()
        defer func() {
            if r := recover(); r != nil {
                tx.Rollback()
                panic(r)
            }
        }()
        result := proceed()
        tx.Commit()
        return result
    }),
    Order: 1,
}
```

### 权限检查切面

```go
// 创建权限检查切面
authAspect := &aop.AspectMeta{
    PointCut: aop.MatchByNamePrefix("Admin"),
    Advice: aop.Before(func(jp aop.JoinPoint) {
        // 检查用户是否有管理员权限
        if !hasAdminPermission() {
            panic("access denied")
        }
    }),
    Order: 1,
}
```

### 性能监控切面

```go
// 创建性能监控切面
perfAspect := &aop.AspectMeta{
    PointCut: aop.MatchAll(),
    Advice: aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
        start := time.Now()
        result := proceed()
        duration := time.Since(start)
        
        // 记录性能指标
        metrics.RecordMethodDuration(jp.Signature().Name(), duration)
        
        return result
    }),
    Order: 1,
}
```

---

## 最佳实践

### 1. 合理使用通知类型

- **Before**：用于方法执行前的准备工作（参数验证、权限检查）
- **After**：用于资源清理（无论是否异常都会执行）
- **AfterReturning**：用于处理正常返回值（结果转换、缓存）
- **AfterThrowing**：用于异常处理（日志记录、告警）
- **Around**：用于完全控制方法执行（事务、性能监控）

### 2. 注意切面执行顺序

Order 值越小，优先级越高。合理设置 Order 确保切面按预期顺序执行：

```go
// 权限检查应该最先执行
authAspect := &aop.AspectMeta{Order: 1, ...}

// 日志记录其次
logAspect := &aop.AspectMeta{Order: 2, ...}

// 性能监控最后执行
perfAspect := &aop.AspectMeta{Order: 3, ...}
```

### 3. 避免在通知中抛出异常

在 Before/After/AfterReturning 通知中抛出异常会中断后续通知和目标方法的执行。建议在通知中捕获并处理异常：

```go
aop.Before(func(jp aop.JoinPoint) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("通知执行异常: %v", r)
        }
    }()
    // 通知逻辑
})
```

### 4. 使用 Around 通知时务必调用 proceed

环绕通知必须调用 `proceed()` 继续执行目标方法或下一个通知，否则目标方法不会被执行：

```go
aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
    // 前置逻辑
    result := proceed() // ⚠️ 必须调用
    // 后置逻辑
    return result
})
```

### 5. 切点匹配规则建议

- 使用 `MatchByNamePrefix` 匹配一组相关方法
- 使用 `MatchByRegex` 进行复杂模式匹配
- 使用 `MatchClassMethod` 精确控制匹配的类和方法
- 避免使用 `MatchAll()` 除非确实需要拦截所有方法

## ChainExecutor（链执行器）

`ChainExecutor` 负责执行 AOP 通知链，按照正确的顺序执行各种类型的通知：

```go
type ChainExecutor struct{}

// Execute 执行通知链
func (e *ChainExecutor) Execute(inv Invocation, advisors []*AspectMeta, targetFunc func(...any) any) any
```

### 通知执行顺序

```
Before 通知（按 Order 升序）
    ↓
Around 通知链（最外层到最内层）
    ↓
目标方法执行
    ↓
AfterReturning 通知（正常返回）或 AfterThrowing 通知（异常返回）
    ↓
After 通知（无论是否异常都执行）
```

### 设计优化

- **指针版本**：统一使用 `[]*AspectMeta` 指针版本，避免值类型到指针的转换开销
- **错误查找**：AfterThrowing 通知从后往前检查多返回值中的 error，符合 Go 错误返回惯例

---

## ProxyFactory（代理工厂）

`ProxyFactory` 负责创建 AOP 代理对象，支持缓存和通知链：

```go
// 创建代理工厂
factory := aop.NewProxyFactory(target)

// 设置切面
factory.SetAspects(aspects)

// 获取代理对象
proxy := factory.GetProxy()
```

### 缓存机制

`ProxyFactory` 会缓存方法匹配的切面列表，避免每次调用都重新匹配和排序。使用 `sync.RWMutex` 保证并发安全。

---

## AopRegistry（AOP 注册表）

`AopRegistry` 是一个全局 AOP 配置中心，用于在 IoC 容器中集成 AOP：

```go
registry := aop.NewAopRegistry()

// 注册切面
registry.RegisterAspect(aspectMeta)

// 注册织入器
registry.RegisterWeaver("userService", weaver)

// 按需织入
proxy := registry.WeaveIfNeeded("userService", &UserService{})

// 为类型匹配切面
matched := registry.MatchAspectsForType(reflect.TypeOf(&UserService{}))
```

---

## 通知执行顺序

当目标方法被调用时，通知按以下顺序执行：

```
Before 通知（按 Order 升序）
    ↓
Around 通知链（最外层到最内层）
    ↓
目标方法本身
    ↓
After / AfterReturning / AfterThrowing 通知（按 Order 升序）
```

多个切面之间按 `Order` 值升序排列。Order 值相同的切面按照添加顺序执行。

---

## 完整示例

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/aop"
)

// LogAspect 日志切面
type LogAspect struct{}

func (l *LogAspect) BeforeLog(jp aop.JoinPoint) {
    fmt.Printf("[日志] 方法 %s 被调用，参数: %v\n",
        jp.Signature().Name(), jp.Args())
}

func (l *LogAspect) AfterLog(jp aop.JoinPoint) {
    fmt.Printf("[日志] 方法 %s 执行完毕\n",
        jp.Signature().Name())
}

// UserService 业务服务
type UserService struct {
    Name string
}

func (u *UserService) CreateUser(name string) string {
    fmt.Println("创建用户:", name)
    return "user_" + name
}

func (u *UserService) DeleteUser(id string) {
    fmt.Println("删除用户:", id)
}

func main() {
    aspect := &LogAspect{}

    // 创建切点 — 匹配所有以 User 结尾的方法
    pointCut := aop.MatchByRegex(".*User$")

    // 创建前置通知
    beforeAdvice := aop.Before(aspect.BeforeLog)

    // 创建后置通知
    afterAdvice := aop.After(aspect.AfterLog)

    // 创建切面元数据
    beforeAspect := &aop.AspectMeta{
        PointCut: pointCut,
        Advice:   beforeAdvice,
        Order:    1,
    }
    afterAspect := &aop.AspectMeta{
        PointCut: pointCut,
        Advice:   afterAdvice,
        Order:    2,
    }

    // 创建织入器
    weaver := aop.NewWeaver()
    weaver.AddAspects(beforeAspect, afterAspect)

    // 织入目标
    proxy := weaver.Weave(&UserService{Name: "test"})

    // 使用代理
    userSvc := proxy.(*UserService)
    result := userSvc.CreateUser("Alice") // 触发 before → 原方法 → after
    fmt.Println("结果:", result)

    userSvc.DeleteUser("123") // 同样触发 before → 原方法 → after
}
```

预计输出：

```
[日志] 方法 CreateUser 被调用，参数: [Alice]
创建用户: Alice
[日志] 方法 CreateUser 执行完毕
结果: user_Alice
[日志] 方法 DeleteUser 被调用，参数: [123]
删除用户: 123
[日志] 方法 DeleteUser 执行完毕
```

---

## 与 IoC 容器集成

`AopRegistry` 可以轻松与 `core.Container` 集成，实现 Bean 自动代理：

```go
registry := aop.NewAopRegistry()

// 注册日志切面
registry.RegisterAspect(&aop.AspectMeta{
    PointCut: aop.MatchByRegex(".*Service\\..*"),
    Advice:   aop.Before(func(jp aop.JoinPoint) { fmt.Println("before") }),
    Order:    0,
})

// 注册 Bean 时织入
container.Register(
    reflect.TypeOf(&UserService{}),
    core.Factory(func(c core.Container) (any, error) {
        svc := &UserService{}
        // 通过 AopRegistry 按需织入
        return registry.WeaveIfNeeded("userService", svc), nil
    }, reflect.TypeOf(&UserService{})),
)
```

---

## 通知链执行机制

当多个 Around 通知作用于同一方法时，它们会组成一个嵌套的调用链。每个 Around 通知可以通过 `proceed` 控制是否以及如何调用下一个通知：

```go
// 最外层 Around（Order=1）
chain1 := aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
    fmt.Println("chain1 before")
    result := proceed()           // 调用 chain2
    fmt.Println("chain1 after")
    return result
})

// 内层 Around（Order=2）
chain2 := aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
    fmt.Println("chain2 before")
    result := proceed()           // 调用目标方法
    fmt.Println("chain2 after")
    return result
})
```

执行顺序：

```
chain1 before → chain2 before → 目标方法 → chain2 after → chain1 after
```

---

## 使用场景

### 场景 1：日志记录

**描述**：记录方法调用日志，包括方法名、参数、返回值等。

```go
factory := aop.NewProxyFactory(service)
factory.SetAspects([]*aop.AspectMeta{
	{
		PointCut: aop.MatchAll(),
		Advice: aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
			start := time.Now()
			result := proceed()
			duration := time.Since(start)
			log.Printf("Method %s called, args: %v, result: %v, duration: %v",
				jp.Signature().Name(), jp.Args(), result, duration)
			return result
		}),
		Order: 1,
	},
})
```

**最佳实践**：
- 使用环绕通知记录完整调用信息
- 记录方法执行时间用于性能分析
- 避免记录敏感信息

### 场景 2：事务管理

**描述**：声明式事务管理，自动提交或回滚事务。

```go
factory := aop.NewProxyFactory(service)
factory.SetAspects([]*aop.AspectMeta{
	{
		PointCut: aop.MatchByNamePrefix("Update"),
		Advice: aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
			tx := db.Begin()
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
					panic(r)
				}
			}()

			result := proceed()

			if err := getError(result); err != nil {
				tx.Rollback()
				return result
			}

			tx.Commit()
			return result
		}),
		Order: 1,
	},
})
```

**最佳实践**：
- 使用环绕通知控制事务边界
- 确保异常时回滚事务
- 避免在事务中执行耗时操作

### 场景 3：安全控制

**描述**：检查用户权限，拦截未授权访问。

```go
factory := aop.NewProxyFactory(service)
factory.SetAspects([]*aop.AspectMeta{
	{
		PointCut: aop.MatchByNamePrefix("Delete"),
		Advice: aop.Before(func(jp aop.JoinPoint) {
			if !hasPermission(jp.Args()[0].(string), "delete") {
				panic("permission denied")
			}
		}),
		Order: 1,
	},
})
```

**最佳实践**：
- 使用前置通知进行权限检查
- 权限检查逻辑应高效
- 记录权限拒绝日志

### 场景 4：缓存管理

**描述**：缓存方法结果，减少重复计算。

```go
cache := make(map[string]any)

factory := aop.NewProxyFactory(service)
factory.SetAspects([]*aop.AspectMeta{
	{
		PointCut: aop.MatchByNamePrefix("Get"),
		Advice: aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
			key := fmt.Sprintf("%s:%v", jp.Signature().Name(), jp.Args())
			if cached, ok := cache[key]; ok {
				return cached
			}

			result := proceed()
			cache[key] = result
			return result
		}),
		Order: 1,
	},
})
```

**最佳实践**：
- 使用环绕通知实现缓存逻辑
- 缓存键应包含方法名和参数
- 考虑缓存失效策略