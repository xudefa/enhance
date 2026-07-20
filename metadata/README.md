# metadata 包 — 元数据编程

> **所属层级**: Core Layer  
> **设计理念**: 反射驱动，注解解析  
> **设计灵感**: Spring MetadataReader

## 概述

`metadata` 包提供元数据编程和注解处理能力，参考 Spring `MetadataReader` 设计，支持通过 struct tag 或自定义解析器读取类型的元数据信息。

### 核心功能

| 功能 | 说明 |
|------|------|
| **元数据读取** | 通过反射读取类型的元数据信息 |
| **注解解析** | 支持 struct tag 注解解析 |
| **缓存机制** | 避免重复解析同一类型 |
| **可扩展** | 支持自定义注解解析器 |

---

## 核心接口

### MetadataReader 元数据读取器

```go
type MetadataReader interface {
    // GetAnnotations 获取目标的所有注解
    GetAnnotations(target any) []Annotation
    
    // GetAnnotation 获取指定名称的注解
    GetAnnotation(target any, name string) Annotation
    
    // HasAnnotation 检查是否存在指定注解
    HasAnnotation(target any, name string) bool
}
```

### AnnotationResolver 注解解析器

```go
type AnnotationResolver interface {
    ResolveAnnotations(t reflect.Type) []Annotation
}
```

### Annotation 注解结构体

```go
type Annotation struct {
    Name       string
    Attributes map[string]any
}
```

---

## 内置实现

### ReflectMetadataReader 基于反射的元数据读取器

基于反射的元数据读取器，带缓存机制：

- 避免重复解析同一类型
- 支持自定义 AnnotationResolver
- 线程安全的缓存访问

### TagAnnotationResolver 基于 struct tag 的注解解析器

基于 struct tag 的注解解析器：

- 解析格式为 `metadata:"name:attr1=val1,attr2=val2"` 的 tag
- 支持多种属性类型
- 自动类型转换

---

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/metadata"
)

type User struct {
    Name string `metadata:"field:name=fullName,required=true"`
    Age  int    `metadata:"field:name=age,type=number"`
}

func main() {
    reader := metadata.NewReflectMetadataReader(nil)
    anns := reader.GetAnnotations(User{})

    for _, ann := range anns {
        fmt.Printf("Annotation: %s\n", ann.Name)
        for key, val := range ann.Attributes {
            fmt.Printf("  %s: %v\n", key, val)
        }
    }
}
```

---

## API 参考

### 获取指定注解

```go
ann := reader.GetAnnotation(User{}, "field")
if ann.Name != "" {
    name, _ := ann.GetStringAttribute("name")
    required, _ := ann.GetAttribute("required")
    fmt.Printf("Field name: %s, Required: %v\n", name, required)
}
```

### 检查注解存在性

```go
if reader.HasAnnotation(User{}, "field") {
    // User 结构体有 field 注解
}
```

### 自定义注解解析器

```go
type CustomResolver struct{}

func (r *CustomResolver) ResolveAnnotations(t reflect.Type) []metadata.Annotation {
    // 自定义解析逻辑
    return []metadata.Annotation{
        {
            Name: "custom",
            Attributes: map[string]any{
                "key": "value",
            },
        },
    }
}

reader := metadata.NewReflectMetadataReader(&CustomResolver{})
```

---

## 使用示例

### Struct Tag 格式

```go
`metadata:"name:attr1=val1,attr2=val2"`
```

- **name**：注解名称
- **attr1=val1**：属性键值对
- 多个属性用逗号分隔

#### 示例

```go
type User struct {
    Name string `metadata:"field:name=fullName,required=true,minLength=1"`
    Age  int    `metadata:"field:name=age,type=number,min=0,max=150"`
}
```

### 与验证框架集成

```go
type Validator struct {
    reader metadata.MetadataReader
}

func (v *Validator) Validate(obj any) error {
    anns := v.reader.GetAnnotations(obj)
    
    for _, ann := range anns {
        if ann.Name == "field" {
            required, _ := ann.GetAttribute("required")
            if required == true {
                // 检查字段是否为空
                fieldValue := reflect.ValueOf(obj).FieldByName(ann.Attributes["name"].(string))
                if fieldValue.IsZero() {
                    return fmt.Errorf("field %s is required", ann.Attributes["name"])
                }
            }
        }
    }
    
    return nil
}
```

### 与依赖注入集成

```go
type BeanFactory struct {
    reader metadata.MetadataReader
}

func (f *BeanFactory) CreateBean(t reflect.Type) any {
    anns := f.reader.GetAnnotations(t)
    
    for _, ann := range anns {
        if ann.Name == "bean" {
            name, _ := ann.GetStringAttribute("name")
            scope, _ := ann.GetStringAttribute("scope")
            
            // 根据注解创建 Bean
            return f.createBeanInstance(t, name, scope)
        }
    }
    
    return nil
}
```

---

## 最佳实践

### 1. 使用缓存提升性能

```go
// ✅ 推荐：使用内置缓存机制
reader := metadata.NewReflectMetadataReader(nil)
// 第一次解析会缓存，后续直接使用缓存
anns := reader.GetAnnotations(User{})

// ⚠️ 不推荐：每次都创建新的读取器
func getAnnotations() []metadata.Annotation {
    reader := metadata.NewReflectMetadataReader(nil)
    return reader.GetAnnotations(User{})
}
```

### 2. 自定义注解解析器

```go
// ✅ 推荐：根据需求自定义解析器
type DatabaseResolver struct{}

func (r *DatabaseResolver) ResolveAnnotations(t reflect.Type) []metadata.Annotation {
    var anns []metadata.Annotation
    
    // 解析数据库相关注解
    if t.Kind() == reflect.Struct {
        if tableTag, ok := t.FieldByName("ID").Tag.Lookup("db"); ok {
            anns = append(anns, metadata.Annotation{
                Name: "table",
                Attributes: map[string]any{
                    "name": tableTag,
                },
            })
        }
    }
    
    return anns
}

reader := metadata.NewReflectMetadataReader(&DatabaseResolver{})

// ⚠️ 不推荐：使用默认解析器处理所有场景
reader := metadata.NewReflectMetadataReader(nil)
```

### 3. 类型安全的属性访问

```go
// ✅ 推荐：使用类型安全的访问方法
ann := reader.GetAnnotation(User{}, "field")
name, ok := ann.GetStringAttribute("name")
if !ok {
    // 处理类型错误
}

required, ok := ann.GetBoolAttribute("required")
if !ok {
    // 处理类型错误
}

// ⚠️ 不推荐：直接访问 Attributes map
name := ann.Attributes["name"].(string) // 可能 panic
```

### 4. 与依赖注入集成

```go
// ✅ 推荐：将 MetadataReader 注册为 Bean
container.Register(
    reflect.TypeOf(&metadata.MetadataReader{}),
    core.Bean(createMetadataReader()),
    core.Singleton(),
)

// 注入使用
type BeanFactory struct {
    Reader metadata.MetadataReader `inject:"metadataReader"`
}

func (f *BeanFactory) CreateBean(t reflect.Type) any {
    anns := f.Reader.GetAnnotations(t)
    // 根据注解创建 Bean
}
```

### 5. 设计原则

- **参考 Spring MetadataReader**：借鉴 Spring 的元数据编程设计理念
- **缓存优化**：避免重复解析，提高性能
- **类型安全**：利用 Go 的反射机制实现类型安全的元数据读取
- **可扩展**：支持自定义注解解析器
- **零外部依赖**：核心框架仅使用 Go 标准库