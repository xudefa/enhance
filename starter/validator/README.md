# Validator Starter

Validator 数据验证自动配置模块，提供灵活的数据验证支持。

## 功能特性

- ✅ 自动配置验证器
- ✅ 支持结构体标签验证
- ✅ 自定义验证器支持
- ✅ 内置常用验证规则
- ✅ 支持嵌套验证

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/validator"
)
```

### 2. 配置文件

在 `application.json` 中添加 Validator 配置：

```json
{
  "validator": {
    "enabled": true,
    "enable-custom-validators": true
  }
}
```

### 3. 使用示例

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/go-playground/validator/v10"
)

type User struct {
    Name  string `validate:"required,min=3,max=50"`
    Email string `validate:"required,email"`
    Age   int    `validate:"required,min=1,max=130"`
    Phone string `validate:"phone"`
}

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("validator-demo"),
    )
    defer app.Stop()
    
    // 获取验证器实例
    v := core.MustGetBean[*validator.Validate](app.Container())
    
    user := User{
        Name:  "John",
        Email: "john@example.com",
        Age:   30,
        Phone: "13800138000",
    }
    
    // 验证数据结构
    if err := v.Struct(user); err != nil {
        fmt.Println("验证失败:", err)
    }
    
    // 验证单个字段
    if err := v.Var("test@email.com", "email"); err != nil {
        fmt.Println("邮箱格式错误:", err)
    }
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `validator.enabled` | bool | false | 是否启用 Validator |
| `validator.enable-custom-validators` | bool | true | 是否启用自定义验证器 |

## 内置验证规则

### 常用规则

| 规则 | 说明 | 示例 |
|------|------|------|
| `required` | 必填 | `validate:"required"` |
| `min` | 最小值 | `validate:"min=1"` |
| `max` | 最大值 | `validate:"max=100"` |
| `len` | 固定长度 | `validate:"len=11"` |
| `email` | 邮箱格式 | `validate:"email"` |
| `url` | URL 格式 | `validate:"url"` |
| `numeric` | 数字 | `validate:"numeric"` |
| `alpha` | 字母 | `validate:"alpha"` |
| `alphanum` | 字母数字 | `validate:"alphanum"` |

### 自定义验证器

模块内置以下自定义验证器：

| 验证器 | 说明 | 示例 |
|--------|------|------|
| `phone` | 手机号 | `validate:"phone"` |
| `idcard` | 身份证号 | `validate:"idcard"` |

## 高级用法

### 自定义错误消息

```go
type User struct {
    Name  string `validate:"required,min=3" label:"用户名"`
    Email string `validate:"required,email" label:"邮箱"`
}

// 自定义错误处理
func formatValidationError(err error) string {
    if errs, ok := err.(validator.ValidationErrors); ok {
        var messages []string
        for _, e := range errs {
            messages = append(messages, fmt.Sprintf(
                "%s 验证失败: %s",
                e.Field(),
                e.Tag(),
            ))
        }
        return strings.Join(messages, ", ")
    }
    return err.Error()
}
```

### 嵌套验证

```go
type Address struct {
    City    string `validate:"required"`
    Street  string `validate:"required"`
}

type User struct {
    Name    string  `validate:"required"`
    Address Address `validate:"required,dive"`
}
```

## 启动顺序

- **优先级**: `OrderPriorityInfrastructure` (-4000)
- **触发条件**: `validator.enabled=true`

## 依赖

- `github.com/go-playground/validator/v10`