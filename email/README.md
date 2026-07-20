# email 包 — 邮件发送

> **所属层级**: Infrastructure Layer  
> **设计理念**: 零外部依赖，原生 SMTP  
> **设计灵感**: Spring Mail + JavaMail

## 概述

`email` 包提供基于 Go 原生 `net/smtp` 的邮件发送功能，默认使用 163 邮箱服务器。

### 核心功能

| 功能 | 说明 |
|------|------|
| **原生 SMTP** | 使用 Go 标准库，无外部依赖 |
| **默认配置** | 开箱即用的 163 邮箱配置 |
| **灵活选项** | 支持自定义 SMTP 服务器、端口、认证 |
| **环境变量** | 密码从环境变量读取，安全配置 |
| **多格式支持** | 支持纯文本和 HTML 邮件 |

---

## 核心接口

### Message 邮件结构

```go
type Message struct {
    From        string            // 发件人（为空时使用默认发件人）
    To          []string          // 收件人列表
    Subject     string            // 邮件主题
    Body        string            // 纯文本内容
    HTML        string            // HTML 内容（优先于 Body）
    Attachments []Attachment      // 附件列表
    Headers     map[string]string // 自定义邮件头
}
```

### Attachment 附件结构

```go
type Attachment struct {
    Filename    string // 文件名
    ContentType string // MIME 类型
    Data        []byte // 文件数据
}
```

### Sender 邮件发送器

```go
type Sender struct {
    // ...
}
```

#### 创建

```go
// 使用默认配置
sender := email.NewSender()

// 使用自定义配置
sender := email.NewSender(
    email.WithHost("smtp.gmail.com"),
    email.WithPort(587),
    email.WithAuth("your@gmail.com", "your-password"),
)
```

#### 发送邮件

```go
err := sender.Send(ctx, &email.Message{
    To:      []string{"recipient@example.com"},
    Subject: "测试邮件",
    Body:    "这是一封测试邮件",
})
```

---

## 默认配置

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| SMTP Host | `smtp.163.com` | 163 邮箱 SMTP 服务器 |
| SMTP Port | `25` | SMTP 端口 |
| From | `xudefa_163mail@163.com` | 默认发件人 |
| Password | 环境变量 `ENHANCE_EMAIL_PASSWORD` | 密码从环境变量读取 |

---

## 配置选项

### SenderOption 配置选项

| 选项 | 说明 | 示例 |
|------|------|------|
| `WithHost(host)` | 设置 SMTP 服务器 | `WithHost("smtp.gmail.com")` |
| `WithPort(port)` | 设置 SMTP 端口 | `WithPort(587)` |
| `WithAuth(user, pass)` | 设置认证信息 | `WithAuth("user@example.com", "password")` |
| `WithPasswordFromEnv()` | 从环境变量读取密码 | `WithPasswordFromEnv()` |

---

## 快速开始

### 1. 设置环境变量

```bash
export ENHANCE_EMAIL_PASSWORD="your-163-password"
```

### 2. 发送邮件

```go
package main

import (
    "context"
    "github.com/xudefa/enhance/email"
)

func main() {
    // 使用默认配置创建发送器
    sender := email.NewSender()

    // 发送邮件
    err := sender.Send(context.Background(), &email.Message{
        To:      []string{"recipient@example.com"},
        Subject: "测试邮件",
        Body:    "这是一封测试邮件",
    })
    if err != nil {
        panic(err)
    }
}
```

---

## API 参考

### 使用自定义 SMTP 服务器

```go
sender := email.NewSender(
    email.WithHost("smtp.gmail.com"),
    email.WithPort(587),
    email.WithAuth("your@gmail.com", "your-password"),
)
```

### 发送 HTML 邮件

```go
err := sender.Send(ctx, &email.Message{
    To:      []string{"user@example.com"},
    Subject: "欢迎注册",
    HTML:    "<h1>欢迎!</h1><p>感谢您的注册。</p>",
})
```

### 发送带附件的邮件

```go
err := sender.Send(ctx, &email.Message{
    To:      []string{"user@example.com"},
    Subject: "带附件的邮件",
    Body:    "请查看附件",
    Attachments: []email.Attachment{
        {
            Filename:    "report.pdf",
            ContentType: "application/pdf",
            Data:        pdfData,
        },
    },
})
```

### 发送带自定义邮件头的邮件

```go
err := sender.Send(ctx, &email.Message{
    To:      []string{"user@example.com"},
    Subject: "通知",
    Body:    "您有新的消息",
    Headers: map[string]string{
        "X-Priority": "1",
        "X-Custom":   "value",
    },
})
```

### 多个收件人

```go
err := sender.Send(ctx, &email.Message{
    To:      []string{"user1@example.com", "user2@example.com"},
    Subject: "群发通知",
    Body:    "大家好！",
})
```

---

## 使用示例

### 发送欢迎邮件

```go
func SendWelcomeEmail(userEmail string, userName string) error {
    sender := email.NewSender()
    
    message := &email.Message{
        To:      []string{userEmail},
        Subject: "欢迎使用我们的服务",
        HTML: fmt.Sprintf(`
            <h1>欢迎, %s!</h1>
            <p>感谢您的注册。</p>
            <p>如果您有任何问题，请随时联系我们。</p>
        `, userName),
    }
    
    return sender.Send(context.Background(), message)
}
```

### 发送密码重置邮件

```go
func SendPasswordResetEmail(userEmail string, resetToken string) error {
    sender := email.NewSender()
    
    resetLink := fmt.Sprintf("https://example.com/reset?token=%s", resetToken)
    
    message := &email.Message{
        To:      []string{userEmail},
        Subject: "密码重置请求",
        HTML: fmt.Sprintf(`
            <h1>密码重置</h1>
            <p>请点击以下链接重置您的密码：</p>
            <a href="%s">重置密码</a>
            <p>此链接将在 24 小时后过期。</p>
        `, resetLink),
    }
    
    return sender.Send(context.Background(), message)
}
```

### 与依赖注入集成

```go
// 注册 Email Sender
container.Register(
    reflect.TypeOf(&email.Sender{}),
    core.Bean(email.NewSender(
        email.WithHost("smtp.gmail.com"),
        email.WithPort(587),
        email.WithAuth("your@gmail.com", os.Getenv("EMAIL_PASSWORD")),
    )),
    core.Singleton(),
)

// 注入使用
type UserService struct {
    EmailSender *email.Sender `inject:"sender"`
}

func (s *UserService) CreateUser(name, emailAddr string) error {
    // 创建用户...
    
    // 发送欢迎邮件
    return s.EmailSender.Send(context.Background(), &email.Message{
        To:      []string{emailAddr},
        Subject: "欢迎",
        Body:    fmt.Sprintf("欢迎, %s!", name),
    })
}
```

---

## 最佳实践

### 1. 使用环境变量存储密码

```go
// ✅ 推荐：从环境变量读取密码
sender := email.NewSender(
    email.WithAuth("user@gmail.com", os.Getenv("EMAIL_PASSWORD")),
)

// ⚠️ 不推荐：硬编码密码
sender := email.NewSender(
    email.WithAuth("user@gmail.com", "my-secret-password"),
)
```

### 2. 使用 HTML 邮件提升用户体验

```go
// ✅ 推荐：使用 HTML 格式
message := &email.Message{
    HTML: "<h1>欢迎!</h1><p>感谢您的注册。</p>",
}

// ⚠️ 不推荐：仅使用纯文本
message := &email.Message{
    Body: "欢迎! 感谢您的注册。",
}
```

### 3. 设置合理的超时和重试

```go
// ✅ 推荐：使用 context 控制超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := sender.Send(ctx, message)

// ⚠️ 不推荐：不设置超时
err := sender.Send(context.Background(), message)
```

### 4. 错误处理和日志记录

```go
// ✅ 推荐：记录错误日志
err := sender.Send(ctx, message)
if err != nil {
    log.Error(ctx, "Failed to send email",
        log.KeyValue{Key: "to", Value: message.To},
        log.KeyValue{Key: "error", Value: err.Error()},
    )
    return err
}

// ⚠️ 不推荐：忽略错误
sender.Send(ctx, message)
```

### 5. 异步发送邮件

```go
// ✅ 推荐：使用异步发送
go func() {
    err := sender.Send(context.Background(), message)
    if err != nil {
        log.Error(context.Background(), "Failed to send email", log.KeyValue{Key: "error", Value: err.Error()})
    }
}()

// ⚠️ 不推荐：同步发送阻塞请求
err := sender.Send(ctx, message)
if err != nil {
    return err
}
```