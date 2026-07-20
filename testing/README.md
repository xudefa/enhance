# testing 包 — 测试框架

> **所属层级**: Infrastructure Layer  
> **设计理念**: 集成测试支持，Mock 对象  
> **设计灵感**: Spring Test + Mockito

## 概述

`testing` 包提供完整的测试支持，合并了原 `test`、`testcontext` 和 `mock` 包的功能。

### 核心功能

| 功能 | 说明 |
|------|------|
| **TestRunner** | 集成测试运行器，支持自动启动应用上下文和依赖注入 |
| **TestContext** | 测试上下文，集成 IoC 容器和测试配置 |
| **Mock** | Mock 对象支持，用于测试中的依赖隔离 |
| **断言工具** | 丰富的断言函数简化测试验证 |
| **Web 测试客户端** | 简化的 HTTP 请求测试 |
| **泛型辅助** | 类型安全的 Bean 获取和验证 |

---

## 核心接口

### TestRunner 测试运行器

```go
type TestRunner struct {
    // ...
}
```

#### 创建

```go
runner := testing.NewTestRunner(t,
    testing.WithTestAppName("my-test-app"),
    testing.WithProperty("db.url", "localhost:5432"),
    testing.WithMockBean("userService", mockUserService),
)
```

#### 配置选项

| 选项 | 说明 |
|------|------|
| `WithProperty(key, value)` | 设置测试属性 |
| `WithMockBean(name, bean)` | 添加 Mock Bean |
| `WithoutAutoConfig()` | 禁用自动配置 |
| `WithTestAppName(name)` | 设置应用名称 |

#### 运行测试

```go
runner.Run(func(ctx *testing.TestContext) {
    service := ctx.Get("myService").(*MyService)
    result := service.DoSomething()
    testing.AssertEqual(t, "expected", result)
})
```

### TestContext 测试上下文

```go
type TestContext struct {
    // ...
}
```

#### 创建

```go
ctx := testing.NewTestContext(t)
ctx.Register(reflect.TypeOf(&MyService{}), func(c core.Container) (any, error) {
    return &MyService{}, nil
})

service := ctx.Get("myService").(*MyService)
```

#### 便捷函数

| 函数 | 说明 |
|------|------|
| `Test(t, fn)` | 直接运行测试 |
| `TestWithContainer(t, container, fn)` | 使用指定容器运行测试 |
| `SetupTest(t, fn)` | 设置和清理测试 |
| `RunSubtest(t, name, fn)` | 运行子测试 |
| `Parallel(t, tests)` | 并行运行测试 |

```go
// 直接运行测试
testing.Test(t, func(ctx *testing.TestContext) {
    // 测试逻辑
})

// 使用指定容器
testing.TestWithContainer(t, container, func(ctx *testing.TestContext) {
    // 测试逻辑
})

// 设置和清理
ctx := testing.SetupTest(t, func(ctx *testing.TestContext) {
    // 初始化
})

// 子测试
testing.RunSubtest(t, "subtest-name", func(ctx *testing.TestContext) {
    // 子测试逻辑
})

// 并行测试
testing.Parallel(t, map[string]func(ctx *testing.TestContext){
    "test1": func(ctx *testing.TestContext) { /* ... */ },
    "test2": func(ctx *testing.TestContext) { /* ... */ },
})
```

#### 泛型辅助

```go
// 必须获取 Bean
service := testing.MustGet[MyService](ctx, "myService")

// 安全获取 Bean
service, err := testing.GetBean[MyService](ctx, "myService")
```

### Mock 对象

```go
type Mock struct {
    // ...
}
```

#### 创建和使用

```go
m := testing.NewMock()
m.Expect("GetUser", []any{1}, &User{Name: "Alice"}, nil)

result, err := m.Call("GetUser", 1)
testing.AssertNoError(t, err)

// 验证期望
testing.AssertExpectations(t, m)
```

#### 指定调用次数

```go
m.ExpectTimes("Log", []any{"message"}, nil, nil, 3)
```

#### 链式调用

```go
recorder := testing.NewMockRecorder(m)
recorder.Return(&User{}, nil).Times(2)
```

#### 使用 WithMock

```go
testing.WithMock(t, func(ctx *testing.TestContext, mock *testing.MockRecorder) {
    // 测试逻辑
})
```

### 断言工具

| 函数 | 说明 |
|------|------|
| `Assert(t, condition, msg)` | 基础断言 |
| `AssertEqual(t, expected, actual)` | 相等断言 |
| `AssertNoError(t, err)` | 无错误断言 |
| `AssertError(t, err)` | 有错误断言 |
| `AssertNil(t, value)` | 空值断言 |
| `AssertNotNil(t, value)` | 非空值断言 |
| `AssertTrue(t, condition)` | 真值断言 |
| `AssertFalse(t, condition)` | 假值断言 |

```go
testing.Assert(t, condition, "message")
testing.AssertEqual(t, expected, actual)
testing.AssertEqual(t, expected, actual, "custom message")
testing.AssertNoError(t, err)
testing.AssertError(t, err)
testing.AssertNil(t, value)
testing.AssertNotNil(t, value)
testing.AssertTrue(t, condition)
testing.AssertFalse(t, condition)
```

#### 跳过测试

```go
testing.SkipIf(t, runtime.GOOS == "windows", "not supported on Windows")
```

### TestWebClient Web 测试客户端

```go
type TestWebClient struct {
    // ...
}
```

#### 创建和使用

```go
client := testing.NewTestWebClient(t, "http://localhost:8080")

resp := client.Get("/api/users")
resp.AssertStatus(t, 200)
resp.AssertBody(t, `{"status":"ok"}`)
```

---

## 快速开始

### 基本单元测试

```go
package main

import (
    "testing"
    "github.com/xudefa/enhance/testing"
)

func TestMyService(t *testing.T) {
    testing.Test(t, func(ctx *testing.TestContext) {
        service := &MyService{}
        result := service.DoSomething()
        testing.AssertEqual(t, "expected", result)
    })
}
```

---

## API 参考

### 完整集成测试

```go
func TestUserService_Integration(t *testing.T) {
    runner := testing.NewTestRunner(t,
        testing.WithTestAppName("user-service-test"),
        testing.WithProperty("db.url", "localhost:5432/test"),
    )

    runner.Run(func(ctx *testing.TestContext) {
        // 获取 Bean
        service := testing.MustGet[UserService](ctx, "userService")

        // 执行测试
        user, err := service.GetUser(1)
        testing.AssertNoError(t, err)
        testing.AssertNotNil(t, user)
        testing.AssertEqual(t, "Alice", user.Name)
    })
}
```

### 单元测试 + Mock

```go
func TestOrderService_WithMock(t *testing.T) {
    // 创建 Mock
    paymentMock := testing.NewMock()
    paymentMock.Expect("ProcessPayment", []any{100.0}, true, nil)

    // 运行测试
    testing.Test(t, func(ctx *testing.TestContext) {
        service := NewOrderService(paymentMock)
        
        err := service.PlaceOrder(100.0)
        testing.AssertNoError(t, err)

        // 验证 Mock 调用
        testing.AssertExpectations(t, paymentMock)
    })
}
```

### 并行测试

```go
func TestParallel(t *testing.T) {
    testing.Parallel(t, map[string]func(ctx *testing.TestContext){
        "test_create": func(ctx *testing.TestContext) {
            // 创建测试
        },
        "test_update": func(ctx *testing.TestContext) {
            // 更新测试
        },
        "test_delete": func(ctx *testing.TestContext) {
            // 删除测试
        },
    })
}
```

---

## 使用示例

### Web API 测试

```go
func TestUserAPI(t *testing.T) {
    // 启动测试服务器
    runner := testing.NewTestRunner(t,
        testing.WithTestAppName("api-test"),
        testing.WithProperty("server.port", "0"), // 随机端口
    )

    runner.Run(func(ctx *testing.TestContext) {
        client := testing.NewTestWebClient(t, ctx.GetServerURL())

        // 测试 GET 请求
        resp := client.Get("/api/users/1")
        resp.AssertStatus(t, 200)
        resp.AssertJSONPath(t, "$.name", "Alice")

        // 测试 POST 请求
        resp = client.Post("/api/users", `{"name": "Bob"}`)
        resp.AssertStatus(t, 201)
    })
}
```

### 子测试

```go
func TestUserService(t *testing.T) {
    testing.RunSubtest(t, "GetUser", func(ctx *testing.TestContext) {
        service := testing.MustGet[UserService](ctx, "userService")
        user, err := service.GetUser(1)
        testing.AssertNoError(t, err)
        testing.AssertNotNil(t, user)
    })

    testing.RunSubtest(t, "CreateUser", func(ctx *testing.TestContext) {
        service := testing.MustGet[UserService](ctx, "userService")
        user, err := service.CreateUser("Bob")
        testing.AssertNoError(t, err)
        testing.AssertEqual(t, "Bob", user.Name)
    })
}
```

---

## 最佳实践

### 1. 使用 TestRunner 管理测试生命周期

```go
// ✅ 推荐：使用 TestRunner 自动管理生命周期
runner := testing.NewTestRunner(t,
    testing.WithTestAppName("my-test"),
)
runner.Run(func(ctx *testing.TestContext) {
    // 测试逻辑
})

// ⚠️ 不推荐：手动管理容器
container := core.NewContainer()
// 手动注册和清理
```

### 2. 使用 Mock 隔离外部依赖

```go
// ✅ 推荐：使用 Mock 隔离外部服务
paymentMock := testing.NewMock()
paymentMock.Expect("ProcessPayment", []any{100.0}, true, nil)

service := NewOrderService(paymentMock)
err := service.PlaceOrder(100.0)
testing.AssertNoError(t, err)
testing.AssertExpectations(t, paymentMock)

// ⚠️ 不推荐：依赖真实外部服务
service := NewOrderService(realPaymentGateway)
```

### 3. 使用泛型获取类型安全的 Bean

```go
// ✅ 推荐：使用泛型获取 Bean
service := testing.MustGet[UserService](ctx, "userService")

// ⚠️ 不推荐：使用类型断言
service := ctx.Get("userService").(*UserService)
```

### 4. 使用并行测试提升性能

```go
// ✅ 推荐：独立测试并行执行
testing.Parallel(t, map[string]func(ctx *testing.TestContext){
    "test_create": func(ctx *testing.TestContext) { /* ... */ },
    "test_update": func(ctx *testing.TestContext) { /* ... */ },
})

// ⚠️ 不推荐：所有测试串行执行
func TestAll(t *testing.T) {
    testCreate(t)
    testUpdate(t)
    testDelete(t)
}
```

### 5. 使用断言工具简化验证

```go
// ✅ 推荐：使用断言工具
testing.AssertEqual(t, "expected", actual)
testing.AssertNoError(t, err)
testing.AssertNotNil(t, user)

// ⚠️ 不推荐：手动编写断言逻辑
if actual != "expected" {
    t.Errorf("expected %s, got %s", "expected", actual)
}
```

### 6. 设计要点

- TestRunner 自动管理应用生命周期
- Mock 对象线程安全，支持并发测试
- 断言函数自动标记为 Helper，显示正确的行号
- 清理函数按注册逆序执行
- 零外部依赖(除框架核心包外)