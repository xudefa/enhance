# openapi 包 — OpenAPI 文档生成

> **所属层级**: Infrastructure Layer  
> **设计理念**: 自动生成，OpenAPI 3.0 规范  
> **设计灵感**: Spring Boot springdoc-openapi

## 概述

`openapi` 包提供 OpenAPI/Swagger 文档自动生成功能，参考 Spring Boot 的 springdoc-openapi 设计。自动从代码注解和路由信息生成 OpenAPI 3.0 规范的文档。

### 核心功能

| 功能 | 说明 |
|------|------|
| **自动生成** | 从代码注解和路由信息自动生成文档 |
| **OpenAPI 3.0** | 支持 OpenAPI 3.0 规范 |
| **JSON/YAML** | 支持生成 JSON 和 YAML 格式文档 |
| **Swagger UI** | 提供 Swagger UI 端点 |
| **控制器注册** | 支持从控制器自动提取 API 信息 |

---

## 核心接口

### OpenAPIDocument OpenAPI 文档

```go
type OpenAPIDocument struct {
    OpenAPI    string
    Info       InfoObject
    Servers    []ServerObject
    Paths      map[string]PathItem
    Components *ComponentsObject
    Tags       []TagObject
}
```

#### 创建

```go
doc := openapi.NewDocument()
```

#### 设置文档信息

```go
doc.SetInfo("My API", "1.0.0", "A sample API")
doc.SetServer("http://localhost:8080", "Development server")
```

#### 注册控制器

```go
doc.RegisterController(&UserController{})
doc.RegisterController(&OrderController{})
```

#### 生成文档

```go
// 生成 JSON 格式文档
json, err := doc.ToJSON()

// 生成 YAML 格式文档
yaml, err := doc.ToYAML()
```

### InfoObject 文档信息

```go
type InfoObject struct {
    Title          string
    Version        string
    Description    string
    TermsOfService string
    Contact        *ContactObject
    License        *LicenseObject
}
```

### ServerObject 服务器信息

```go
type ServerObject struct {
    URL         string
    Description string
    Variables   map[string]ServerVariable
}
```

### PathItem 路径项

```go
type PathItem struct {
    Summary     string
    Description string
    Get         *OperationObject
    Post        *OperationObject
    Put         *OperationObject
    Delete      *OperationObject
    Patch       *OperationObject
    Parameters  []ParameterObject
    Tags        []string
}
```

### OperationObject 操作对象

```go
type OperationObject struct {
    Summary     string
    Description string
    OperationID string
    Tags        []string
    Parameters  []ParameterObject
    RequestBody *RequestBodyObject
    Responses   map[string]ResponseObject
}
```

---

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/openapi"
)

func main() {
    // 创建文档
    doc := openapi.NewDocument()
    doc.SetInfo("My API", "1.0.0", "A sample API")
    doc.SetServer("http://localhost:8080", "Development server")

    // 注册控制器
    doc.RegisterController(&UserController{})

    // 生成 JSON
    json, err := doc.ToJSON()
    if err != nil {
        panic(err)
    }
    fmt.Println(string(json))
}
```

---

## API 参考

### 提供 Swagger UI

```go
func main() {
    doc := openapi.NewDocument()
    doc.SetInfo("My API", "1.0.0", "A sample API")
    doc.SetServer("http://localhost:8080", "Development server")

    doc.RegisterController(&UserController{})
    doc.RegisterController(&OrderController{})

    // 提供 Swagger UI
    openapi.ServeSwaggerUI(doc, "/swagger", 8080)
}
```

### 添加标签

```go
doc := openapi.NewDocument()

// 添加标签
doc.AddTag("user", "User management APIs")
doc.AddTag("order", "Order management APIs")

// 注册控制器时指定标签
doc.RegisterController(&UserController{})
```

### 添加安全方案

```go
doc := openapi.NewDocument()

// 添加 Bearer Token 认证
doc.AddSecurityScheme("BearerAuth", openapi.SecurityScheme{
    Type:         "http",
    Scheme:       "bearer",
    BearerFormat: "JWT",
    Description:  "JWT Bearer token authentication",
})

// 设置为全局安全要求
doc.SetSecurityRequirement("BearerAuth", []string{})
```

---

## 使用示例

### 场景 1: API 文档自动生成

从代码自动生成 API 文档，保持文档与代码同步：

```go
func main() {
    doc := openapi.NewDocument()
    doc.SetInfo("User Service API", "2.0.0", "User management service")
    
    // 添加服务器
    doc.SetServer("https://api.example.com", "Production")
    doc.SetServer("https://staging-api.example.com", "Staging")
    
    // 注册所有控制器
    doc.RegisterController(&UserController{})
    doc.RegisterController(&RoleController{})
    doc.RegisterController(&PermissionController{})
    
    // 保存文档
    json, _ := doc.ToJSON()
    os.WriteFile("openapi.json", json, 0644)
}
```

**最佳实践**:
- 在构建流程中自动生成文档
- 文档版本与 API 版本保持一致
- 提供详细的描述信息

### 场景 2: 开发环境 Swagger UI

开发环境提供交互式 API 文档，方便调试和测试：

```go
func main() {
    doc := buildAPIDocument()
    
    // 仅开发环境启用
    if isDevMode() {
        openapi.ServeSwaggerUI(doc, "/swagger", 8080)
    }
    
    startServer()
}
```

**最佳实践**:
- 仅开发环境启用 Swagger UI
- 生产环境禁用交互式文档
- 使用独立的文档端口

### 场景 3: API 网关集成

生成 OpenAPI 文档供 API 网关使用：

```go
func generateGatewayConfig() {
    doc := openapi.NewDocument()
    doc.SetInfo("Gateway API", "1.0.0", "API Gateway")
    
    // 注册所有后端服务
    doc.RegisterController(&UserService{})
    doc.RegisterController(&OrderService{})
    doc.RegisterController(&PaymentService{})
    
    // 生成 YAML 供网关使用
    yaml, _ := doc.ToYAML()
    os.WriteFile("gateway-config.yaml", yaml, 0644)
}
```

**最佳实践**:
- 使用 YAML 格式便于阅读
- 包含所有后端服务 API
- 定期更新网关配置

---

## 最佳实践

### 1. 文档版本管理

```go
// ✅ 推荐：文档版本与 API 版本一致
doc.SetInfo("User Service API", "v2.0.0", "User management service")

// ⚠️ 不推荐：版本不一致
doc.SetInfo("User Service API", "1.0.0", "User management service")
```

### 2. 环境隔离

```go
// ✅ 推荐：根据环境配置服务器
if isProduction() {
    doc.SetServer("https://api.example.com", "Production")
} else if isStaging() {
    doc.SetServer("https://staging-api.example.com", "Staging")
} else {
    doc.SetServer("http://localhost:8080", "Development")
}

// ⚠️ 不推荐：硬编码服务器地址
doc.SetServer("http://localhost:8080", "Development server")
```

### 3. 安全方案配置

```go
// ✅ 推荐：配置安全方案
doc.AddSecurityScheme("BearerAuth", openapi.SecurityScheme{
    Type:         "http",
    Scheme:       "bearer",
    BearerFormat: "JWT",
    Description:  "JWT Bearer token authentication",
})
doc.SetSecurityRequirement("BearerAuth", []string{})

// ⚠️ 不推荐：不配置安全方案
// 文档中没有认证信息
```

### 4. 标签组织

```go
// ✅ 推荐：使用标签组织 API
doc.AddTag("user", "User management APIs")
doc.AddTag("order", "Order management APIs")
doc.AddTag("auth", "Authentication APIs")

// ⚠️ 不推荐：不使用标签
// 所有 API 混在一起，难以查找
```

### 5. 与依赖注入集成

```go
// ✅ 推荐：将 OpenAPI 文档注册为 Bean
container.Register(
    reflect.TypeOf(&openapi.OpenAPIDocument{}),
    core.Bean(createOpenAPIDocument()),
    core.Singleton(),
)

// 注入使用
type APIService struct {
    Document *openapi.OpenAPIDocument `inject:"openapiDocument"`
}

func (s *APIService) Start() {
    openapi.ServeSwaggerUI(s.Document, "/swagger", 8080)
}
```

### 6. 设计要点

- 支持 OpenAPI 3.0 规范
- 从控制器自动提取路由信息
- 支持 JSON 和 YAML 两种格式
- 提供 Swagger UI 集成
- 零外部依赖，仅使用 Go 标准库