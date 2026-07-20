# i18n 包 — 国际化

> **所属层级**: Infrastructure Layer  
> **设计理念**: 多区域支持，回退机制  
> **设计灵感**: Spring MessageSource

## 概述

`i18n` 包提供国际化（i18n）消息源支持，参考 Spring `MessageSource` 设计，支持多区域消息查找、回退机制和消息格式化。

### 核心功能

| 功能 | 说明 |
|------|------|
| **多区域消息查找** | 支持按语言和区域获取消息 |
| **回退机制** | 精确匹配 → 语言匹配 → 父消息源 → 返回代码 |
| **消息格式化** | 支持参数化的消息模板 |
| **多种实现** | ResourceBundle、Static、Delegating 三种消息源实现 |
| **线程安全** | 所有实现都支持并发访问 |

---

## 核心接口

### MessageSource 消息源接口

```go
type MessageSource interface {
    // GetMessage 使用默认区域获取消息，args 用于格式化消息模板
    GetMessage(code string, args ...any) string
    
    // GetMessageWithLocale 使用指定区域获取消息
    GetMessageWithLocale(code string, locale Locale, args ...any) string
}
```

### Locale 语言和区域设置

```go
type Locale struct {
    Language string  // 语言代码（如 "en", "zh"）
    Country  string  // 国家/地区代码（如 "US", "CN"）
    Variant  string  // 变体标识（可选）
}
```

---

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/i18n"
)

func main() {
    source := i18n.NewResourceBundleMessageSource()

    // 注册中文资源
    source.AddResourceBundle(i18n.Locale{Language: "zh"}, map[string]string{
        "greeting": "你好, %s!",
        "farewell": "再见, %s!",
    })

    // 注册英文资源
    source.AddResourceBundle(i18n.Locale{Language: "en"}, map[string]string{
        "greeting": "Hello, %s!",
        "farewell": "Goodbye, %s!",
    })

    // 使用默认区域获取消息
    msg := source.GetMessage("greeting", "World")
    fmt.Println(msg) // Output: Hello, World!
}
```

---

## API 参考

### ResourceBundleMessageSource

基于资源包的消息源，支持按区域注册消息映射表：

```go
source := i18n.NewResourceBundleMessageSource()

// 注册中文资源
source.AddResourceBundle(i18n.Locale{Language: "zh"}, map[string]string{
    "greeting": "你好, %s!",
    "farewell": "再见, %s!",
})

// 注册英文资源
source.AddResourceBundle(i18n.Locale{Language: "en"}, map[string]string{
    "greeting": "Hello, %s!",
    "farewell": "Goodbye, %s!",
})
```

### StaticMessageSource

静态消息源，适用于测试场景或简单应用：

```go
source := i18n.NewStaticMessageSource()
source.AddMessage("greeting", "Hello, %s!")
source.AddMessage("farewell", "Goodbye, %s!")
```

### DelegatingMessageSource

委托消息源，支持父子层级结构：

```go
parent := i18n.NewStaticMessageSource()
parent.AddMessage("common.message", "Common message")

child := i18n.NewDelegatingMessageSource(parent)
child.AddMessage("child.message", "Child message")

// 先在子消息源中查找，未找到时回退到父消息源
msg := child.GetMessage("common.message")
// Output: Common message
```

### 指定区域获取消息

```go
msg := source.GetMessageWithLocale("greeting", i18n.Locale{Language: "zh"}, "世界")
// Output: 你好, 世界!
```

### 使用 Locale 对象

```go
locale := i18n.Locale{
    Language: "zh",
    Country:  "CN",
}

msg := source.GetMessageWithLocale("greeting", locale, "世界")
```

---

## 使用示例

### 多区域消息

```go
source := i18n.NewResourceBundleMessageSource()

// 注册简体中文
source.AddResourceBundle(i18n.Locale{Language: "zh", Country: "CN"}, map[string]string{
    "welcome": "欢迎使用, %s!",
    "error.not_found": "资源未找到",
})

// 注册繁体中文
source.AddResourceBundle(i18n.Locale{Language: "zh", Country: "TW"}, map[string]string{
    "welcome": "歡迎使用, %s!",
    "error.not_found": "資源未找到",
})

// 注册英文
source.AddResourceBundle(i18n.Locale{Language: "en"}, map[string]string{
    "welcome": "Welcome, %s!",
    "error.not_found": "Resource not found",
})

// 根据用户区域获取消息
func getUserMessage(userLocale string, code string, args ...any) string {
    locale := parseLocale(userLocale)
    return source.GetMessageWithLocale(code, locale, args...)
}
```

### Web 请求国际化

```go
func i18nMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 从请求头获取语言
        lang := r.Header.Get("Accept-Language")
        locale := parseAcceptLanguage(lang)
        
        // 将 locale 存入 context
        ctx := context.WithValue(r.Context(), "locale", locale)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func parseAcceptLanguage(lang string) i18n.Locale {
    // 解析 Accept-Language 头
    // 例如: "zh-CN,zh;q=0.9,en;q=0.8"
    parts := strings.Split(lang, ",")
    if len(parts) > 0 {
        langParts := strings.Split(parts[0], "-")
        if len(langParts) == 2 {
            return i18n.Locale{
                Language: langParts[0],
                Country:  langParts[1],
            }
        }
        return i18n.Locale{Language: langParts[0]}
    }
    return i18n.Locale{Language: "en"}
}
```

---

## 回退机制

消息查找按以下顺序回退：

1. **精确匹配区域**：如 "zh_CN"
2. **语言匹配**：如 "zh"
3. **回退到父消息源**：如果设置了 fallback
4. **返回消息代码本身**：作为最后的回退

```go
// 示例：用户请求 zh_CN，但未注册该区域
source.AddResourceBundle(i18n.Locale{Language: "zh"}, map[string]string{
    "greeting": "你好!",
})

// 查找顺序：
// 1. zh_CN (未找到)
// 2. zh (找到: "你好!")
msg := source.GetMessageWithLocale("greeting", i18n.Locale{Language: "zh", Country: "CN"})
```

---

## 最佳实践

### 1. 使用 ResourceBundle 管理多语言资源

```go
// ✅ 推荐：按区域组织资源
source := i18n.NewResourceBundleMessageSource()
source.AddResourceBundle(i18n.Locale{Language: "zh"}, zhMessages)
source.AddResourceBundle(i18n.Locale{Language: "en"}, enMessages)

// ⚠️ 不推荐：硬编码消息
if lang == "zh" {
    msg = "你好"
} else {
    msg = "Hello"
}
```

### 2. 使用消息代码而非硬编码文本

```go
// ✅ 推荐：使用消息代码
msg := source.GetMessage("user.welcome", username)

// ⚠️ 不推荐：硬编码文本
msg := fmt.Sprintf("Welcome, %s!", username)
```

### 3. 设置合理的回退机制

```go
// ✅ 推荐：设置父消息源作为默认回退
parent := i18n.NewStaticMessageSource()
parent.AddMessage("error.default", "An error occurred")

child := i18n.NewDelegatingMessageSource(parent)
child.AddResourceBundle(i18n.Locale{Language: "zh"}, zhMessages)

// ⚠️ 不推荐：不设置回退，可能返回消息代码
source := i18n.NewResourceBundleMessageSource()
```

### 4. 与依赖注入集成

```go
// ✅ 推荐：将 MessageSource 注册为 Bean
container.Register(
    reflect.TypeOf(&i18n.ResourceBundleMessageSource{}),
    core.Bean(createMessageSource()),
    core.Singleton(),
)

// 注入使用
type UserService struct {
    I18n i18n.MessageSource `inject:"messageSource"`
}

func (s *UserService) GreetUser(locale i18n.Locale, name string) string {
    return s.I18n.GetMessageWithLocale("user.greeting", locale, name)
}
```

### 5. 从配置文件加载资源

```go
// ✅ 推荐：从配置文件加载资源
func loadResources(source *i18n.ResourceBundleMessageSource) {
    // 从 YAML/JSON 文件加载
    zhMessages := loadYAML("locales/zh.yaml")
    enMessages := loadYAML("locales/en.yaml")
    
    source.AddResourceBundle(i18n.Locale{Language: "zh"}, zhMessages)
    source.AddResourceBundle(i18n.Locale{Language: "en"}, enMessages)
}

// ⚠️ 不推荐：硬编码所有资源
source.AddResourceBundle(i18n.Locale{Language: "zh"}, map[string]string{
    "key1": "value1",
    "key2": "value2",
    // ... 大量硬编码
})
```