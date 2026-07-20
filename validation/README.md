# validation 包 — 数据验证

> **所属层级**: Infrastructure Layer  
> **设计理念**: 声明式验证，灵活扩展  
> **设计灵感**: Spring Validation + Bean Validation

## 概述

`validation` 包提供灵活的数据验证能力，支持 HTTP 请求验证和结构体验证。支持多种验证类型，包括必填、字符串长度、数值范围、邮箱格式、正则表达式、枚举值等。

### 核心功能

| 功能 | 说明 |
|------|------|
| **声明式验证** | 通过配置规则实现验证逻辑与业务代码分离 |
| **多种验证类型** | 支持 required、string、number、email、regex、enum 等 |
| **HTTP 请求验证** | 支持 query、header、body 验证 |
| **快速失败模式** | 支持遇到第一个错误就停止验证 |
| **自定义错误消息** | 支持友好的错误提示 |
| **线程安全** | 验证器创建后只读，支持并发使用 |

---

## 核心接口

### ValidationRule 验证规则

```go
type ValidationRule struct {
    Field     string   // 字段名称
    Type      string   // 验证类型：required, string, number, email, regex, enum, min, max, length
    Value     string   // 验证值（用于 enum, regex 等）
    Min       *float64 // 最小值
    Max       *float64 // 最大值
    MinLength *int     // 最小长度
    MaxLength *int     // 最大长度
    Pattern   string   // 正则表达式
    Message   string   // 自定义错误消息
    In        []string // 枚举值
}
```

### ValidationConfig 验证配置

```go
type ValidationConfig struct {
    Rules    []ValidationRule // 验证规则
    Source   string           // 验证来源：query, header, body
    FailFast bool             // 快速失败（遇到第一个错误就停止）
}
```

### 验证结果

```go
type RuleValidationResult struct {
    Valid  bool                // 是否通过验证
    Errors []RuleValidationError // 错误列表
}

type RuleValidationError struct {
    Field   string // 字段名称
    Message string // 错误消息
    Type    string // 错误类型
}
```

---

## 验证类型

### 1. required — 必填

验证字段不能为空：

```go
rule := validation.ValidationRule{
    Field: "name",
    Type:  "required",
}
```

### 2. string — 字符串验证

验证字符串长度：

```go
minLen := 2
maxLen := 50
rule := validation.ValidationRule{
    Field:     "name",
    Type:      "string",
    MinLength: &minLen,
    MaxLength: &maxLen,
}
```

### 3. number — 数值验证

验证数值范围：

```go
minVal := 1.0
maxVal := 100.0
rule := validation.ValidationRule{
    Field: "age",
    Type:  "number",
    Min:   &minVal,
    Max:   &maxVal,
}
```

### 4. email — 邮箱验证

验证邮箱格式：

```go
rule := validation.ValidationRule{
    Field: "email",
    Type:  "email",
}
```

### 5. regex — 正则表达式验证

使用正则表达式验证格式：

```go
rule := validation.ValidationRule{
    Field:   "phone",
    Type:    "regex",
    Pattern: `^\d{3}-\d{3}-\d{4}$`,
}
```

### 6. enum — 枚举值验证

验证值是否在允许的枚举值中：

```go
rule := validation.ValidationRule{
    Field: "status",
    Type:  "enum",
    In:    []string{"active", "inactive", "pending"},
}
```

### 7. min — 最小值验证

验证数值不小于最小值：

```go
minVal := 0.0
rule := validation.ValidationRule{
    Field: "price",
    Type:  "min",
    Min:   &minVal,
}
```

### 8. max — 最大值验证

验证数值不大于最大值：

```go
maxVal := 1000.0
rule := validation.ValidationRule{
    Field: "quantity",
    Type:  "max",
    Max:   &maxVal,
}
```

### 9. length — 长度验证

验证字符串长度范围：

```go
minLen := 6
maxLen := 20
rule := validation.ValidationRule{
    Field:     "password",
    Type:      "length",
    MinLength: &minLen,
    MaxLength: &maxLen,
}
```

---

## 快速开始

### 基本验证

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/validation"
)

func main() {
    rules := []validation.ValidationRule{
        {Field: "name", Type: "required"},
        {Field: "email", Type: "email"},
        {Field: "age", Type: "number", Min: floatPtr(18), Max: floatPtr(100)},
    }

    body := []byte(`{"name":"test","email":"test@example.com","age":25}`)
    result := validation.ValidateJSONBody(body, rules)
    
    if !result.Valid {
        for _, e := range result.Errors {
            fmt.Printf("Field: %s, Error: %s\n", e.Field, e.Message)
        }
    }
}
```

---

## API 参考

### RequestValidator 请求验证器

创建可复用的请求验证器：

```go
config := validation.ValidationConfig{
    Source: "query",
    Rules: []validation.ValidationRule{
        {Field: "page", Type: "required"},
        {Field: "size", Type: "number", Min: floatPtr(1), Max: floatPtr(100)},
    },
    FailFast: false,
}

validator, err := validation.NewRequestValidator(config)
if err != nil {
    // 处理配置错误（如无效的正则表达式）
}

// 验证请求
result := validator.Validate(req)
if !result.Valid {
    for _, e := range result.Errors {
        fmt.Printf("Field: %s, Error: %s\n", e.Field, e.Message)
    }
}
```

### ValidateQuery 快速验证查询参数

```go
rules := []validation.ValidationRule{
    {Field: "page", Type: "required"},
    {Field: "size", Type: "number", Min: floatPtr(1), Max: floatPtr(100)},
}

result := validation.ValidateQuery(req, rules)
if !result.Valid {
    // 处理验证错误
}
```

### ValidateHeaders 快速验证请求头

```go
rules := []validation.ValidationRule{
    {Field: "X-Api-Key", Type: "required"},
    {Field: "X-Request-Id", Type: "required"},
}

result := validation.ValidateHeaders(req, rules)
if !result.Valid {
    // 处理验证错误
}
```

### ValidateJSONBody 验证 JSON Body

```go
rules := []validation.ValidationRule{
    {Field: "name", Type: "required"},
    {Field: "email", Type: "email"},
    {Field: "age", Type: "number", Min: floatPtr(18), Max: floatPtr(100)},
}

body := []byte(`{"name":"test","email":"test@example.com","age":25}`)
result := validation.ValidateJSONBody(body, rules)
if !result.Valid {
    // 处理验证错误
}
```

---

## 使用示例

### HTTP 处理器中的验证

```go
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    // 验证请求头
    headerRules := []validation.ValidationRule{
        {Field: "Content-Type", Type: "required"},
        {Field: "X-Request-Id", Type: "required"},
    }
    
    headerResult := validation.ValidateHeaders(r, headerRules)
    if !headerResult.Valid {
        http.Error(w, fmt.Sprintf("Header validation failed: %v", headerResult.Errors), http.StatusBadRequest)
        return
    }
    
    // 读取并验证 body
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read request body", http.StatusBadRequest)
        return
    }
    
    bodyRules := []validation.ValidationRule{
        {Field: "name", Type: "required", Message: "Name is required"},
        {Field: "email", Type: "email", Message: "Valid email is required"},
        {Field: "age", Type: "number", Min: floatPtr(18), Max: floatPtr(100)},
    }
    
    bodyResult := validation.ValidateJSONBody(body, bodyRules)
    if !bodyResult.Valid {
        http.Error(w, fmt.Sprintf("Body validation failed: %v", bodyResult.Errors), http.StatusBadRequest)
        return
    }
    
    // 处理创建用户逻辑
}
```

### 中间件集成

```go
func ValidationMiddleware(rules []validation.ValidationRule) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            result := validation.ValidateQuery(r, rules)
            if !result.Valid {
                http.Error(w, fmt.Sprintf("Validation failed: %v", result.Errors), http.StatusBadRequest)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// 使用中间件
rules := []validation.ValidationRule{
    {Field: "page", Type: "required"},
    {Field: "size", Type: "number", Min: floatPtr(1), Max: floatPtr(100)},
}

handler := ValidationMiddleware(rules)(myHandler)
```

### 自定义错误消息

```go
rule := validation.ValidationRule{
    Field:   "email",
    Type:    "email",
    Message: "请输入有效的邮箱地址",
}
```

### 快速失败模式

```go
config := validation.ValidationConfig{
    Source: "query",
    Rules: []validation.ValidationRule{
        {Field: "name", Type: "required"},
        {Field: "email", Type: "email"},
        {Field: "age", Type: "number"},
    },
    FailFast: true, // 遇到第一个错误就停止验证
}
```

---

## 最佳实践

### 1. 使用声明式验证规则

```go
// ✅ 推荐：声明式规则配置
rules := []validation.ValidationRule{
    {Field: "name", Type: "required"},
    {Field: "email", Type: "email"},
    {Field: "age", Type: "number", Min: floatPtr(18)},
}

// ⚠️ 不推荐：硬编码验证逻辑
if name == "" {
    return errors.New("name is required")
}
if !isValidEmail(email) {
    return errors.New("invalid email")
}
```

### 2. 使用自定义错误消息

```go
// ✅ 推荐：友好的错误消息
rules := []validation.ValidationRule{
    {Field: "email", Type: "email", Message: "请输入有效的邮箱地址"},
    {Field: "password", Type: "length", MinLength: intPtr(8), Message: "密码长度至少8位"},
}

// ⚠️ 不推荐：默认错误消息
rules := []validation.ValidationRule{
    {Field: "email", Type: "email"},
    {Field: "password", Type: "length", MinLength: intPtr(8)},
}
```

### 3. 复用验证器提升性能

```go
// ✅ 推荐：创建可复用的验证器
var userValidator *validation.RequestValidator

func init() {
    config := validation.ValidationConfig{
        Source: "body",
        Rules: []validation.ValidationRule{
            {Field: "name", Type: "required"},
            {Field: "email", Type: "email"},
        },
    }
    userValidator, _ = validation.NewRequestValidator(config)
}

func handler(w http.ResponseWriter, r *http.Request) {
    result := userValidator.Validate(r)
    // ...
}

// ⚠️ 不推荐：每次请求都创建新验证器
func handler(w http.ResponseWriter, r *http.Request) {
    validator, _ := validation.NewRequestValidator(config)
    result := validator.Validate(r)
}
```

### 4. 根据场景选择验证模式

```go
// ✅ 推荐：快速失败适用于严格验证
config := validation.ValidationConfig{
    FailFast: true, // 第一个错误就停止
}

// ✅ 推荐：收集所有错误适用于表单验证
config := validation.ValidationConfig{
    FailFast: false, // 收集所有错误
}
```

### 5. 与 Web 框架集成

```go
// ✅ 推荐：使用中间件统一验证
router.Use(ValidationMiddleware([]validation.ValidationRule{
    {Field: "X-Api-Key", Type: "required"},
}))

// ⚠️ 不推荐：每个处理器重复验证逻辑
func handler1(w http.ResponseWriter, r *http.Request) {
    // 重复验证代码
}

func handler2(w http.ResponseWriter, r *http.Request) {
    // 重复验证代码
}
```

### 6. 注意事项

- **空值处理**: 除 `required` 类型外，其他验证类型在值为空时会跳过验证
- **正则表达式**: 建议在创建验证器时预编译正则表达式，提高性能
- **JSON Body**: `ValidateJSONBody` 会自动处理 JSON 中的数字类型（float64）
- **线程安全**: `RequestValidator` 创建后是只读的，可以安全地在多个 goroutine 中使用
- **自定义消息**: 使用 `Message` 字段可以提供更友好的错误提示