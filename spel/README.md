# spel 包 — 表达式语言

> **所属层级**: Core Layer  
> **设计理念**: SpEL 风格表达式，动态求值  
> **设计灵感**: Spring Expression Language (SpEL)

## 概述

`spel` 包提供 SpEL（Spring Expression Language）风格的表达式解析和求值能力，支持属性访问、方法调用和复杂表达式计算。

### 核心功能

| 功能 | 说明 |
|------|------|
| **表达式解析** | 解析 SpEL 风格的表达式 |
| **属性访问** | 支持对象属性访问 |
| **方法调用** | 支持对象方法调用 |
| **上下文求值** | 基于 EvaluationContext 的表达式求值 |
| **拦截器支持** | 支持表达式求值拦截 |

---

## 核心接口

### SpelParser 表达式解析器

```go
type SpelParser struct{}
```

### Expression 表达式接口

```go
type Expression interface {
    GetValue(ctx EvaluationContext) (any, error)
}
```

### EvaluationContext 求值上下文

```go
type EvaluationContext interface {
    GetVariable(name string) (any, bool)
    SetVariable(name string, value any)
}
```

---

## 快速开始

### 基本表达式

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/spel"
)

func main() {
    parser := spel.NewSpelParser()

    // 简单属性表达式
    expr, err := parser.ParseExpression("name")
    if err != nil {
        // 处理错误
    }

    ctx := spel.NewEvaluationContext()
    ctx.SetVariable("name", "Alice")

    value, err := expr.GetValue(ctx)
    fmt.Println(value) // Output: Alice
}
```

---

## API 参考

### 复杂表达式

```go
parser := spel.NewSpelParser()

// 嵌套属性访问
expr, _ := parser.ParseExpression("user.name")

ctx := spel.NewEvaluationContext()
ctx.SetVariable("user", User{Name: "Alice"})

value, _ := expr.GetValue(ctx)
// value == "Alice"
```

### 字面量

```go
parser := spel.NewSpelParser()

// 布尔字面量
expr, _ := parser.ParseExpression("true")
value, _ := expr.GetValue(nil)
// value == true

// 数字字面量
expr, _ := parser.ParseExpression("42")
value, _ := expr.GetValue(nil)
// value == 42
```

### 表达式类型

#### PropertyExpression

简单属性表达式，访问上下文中的变量：

```go
expr := &spel.PropertyExpression{Property: "name"}
```

#### ComplexExpression

复杂表达式，支持更复杂的表达式计算：

```go
expr := &spel.ComplexExpression{Raw: "user.name"}
```

---

## 使用示例

### 条件表达式

```go
parser := spel.NewSpelParser()

// 条件判断
expr, _ := parser.ParseExpression("age > 18")

ctx := spel.NewEvaluationContext()
ctx.SetVariable("age", 20)

result, _ := expr.GetValue(ctx)
// result == true
```

### 方法调用

```go
parser := spel.NewSpelParser()

// 方法调用表达式
expr, _ := parser.ParseExpression("user.GetName()")

ctx := spel.NewEvaluationContext()
ctx.SetVariable("user", &User{Name: "Alice"})

value, _ := expr.GetValue(ctx)
// value == "Alice"
```

### 与 AOP 集成

```go
// 使用表达式解析方法参数
func parseArgs(expr string, args []any) any {
    parser := spel.NewSpelParser()
    parsed, _ := parser.ParseExpression(expr)
    
    ctx := spel.NewEvaluationContext()
    for i, arg := range args {
        ctx.SetVariable(fmt.Sprintf("arg%d", i), arg)
    }
    
    value, _ := parsed.GetValue(ctx)
    return value
}

// 在拦截器中使用
func (i *MyInterceptor) Before(method string, args []any) {
    key := parseArgs("#arg0", args)
    cache.Get(key)
}
```

---

## 最佳实践

### 1. 缓存表达式解析结果

```go
// ✅ 推荐：缓存解析后的表达式
var parsedExpr spel.Expression

func init() {
    parser := spel.NewSpelParser()
    parsedExpr, _ = parser.ParseExpression("user.name")
}

func evaluate(ctx spel.EvaluationContext) any {
    value, _ := parsedExpr.GetValue(ctx)
    return value
}

// ⚠️ 不推荐：每次求值都重新解析
func evaluate(ctx spel.EvaluationContext) any {
    parser := spel.NewSpelParser()
    expr, _ := parser.ParseExpression("user.name")
    return expr.GetValue(ctx)
}
```

### 2. 使用类型安全的求值

```go
// ✅ 推荐：类型断言确保类型安全
func GetUserName(ctx spel.EvaluationContext) string {
    parser := spel.NewSpelParser()
    expr, _ := parser.ParseExpression("user.name")
    
    value, err := expr.GetValue(ctx)
    if err != nil {
        return ""
    }
    
    if name, ok := value.(string); ok {
        return name
    }
    return ""
}

// ⚠️ 不推荐：直接使用 any 类型
func GetUserName(ctx spel.EvaluationContext) any {
    parser := spel.NewSpelParser()
    expr, _ := parser.ParseExpression("user.name")
    value, _ := expr.GetValue(ctx)
    return value
}
```

### 3. 处理求值错误

```go
// ✅ 推荐：检查求值错误
func safeEvaluate(ctx spel.EvaluationContext) (string, error) {
    parser := spel.NewSpelParser()
    expr, err := parser.ParseExpression("user.name")
    if err != nil {
        return "", fmt.Errorf("parse expression failed: %w", err)
    }
    
    value, err := expr.GetValue(ctx)
    if err != nil {
        return "", fmt.Errorf("evaluate expression failed: %w", err)
    }
    
    if name, ok := value.(string); ok {
        return name, nil
    }
    return "", fmt.Errorf("unexpected type: %T", value)
}

// ⚠️ 不推荐：忽略错误
func unsafeEvaluate(ctx spel.EvaluationContext) string {
    parser := spel.NewSpelParser()
    expr, _ := parser.ParseExpression("user.name")
    value, _ := expr.GetValue(ctx)
    return value.(string)
}
```

### 4. 与依赖注入集成

```go
// ✅ 推荐：将 SpelParser 注册为 Bean
container.Register(
    reflect.TypeOf(&spel.SpelParser{}),
    core.Bean(spel.NewSpelParser()),
    core.Singleton(),
)

// 注入使用
type CacheService struct {
    Parser *spel.SpelParser `inject:"spelParser"`
}

func (s *CacheService) GetCacheKey(expr string, args []any) string {
    parsed, _ := s.Parser.ParseExpression(expr)
    ctx := spel.NewEvaluationContext()
    // 设置变量...
    value, _ := parsed.GetValue(ctx)
    return value.(string)
}
```