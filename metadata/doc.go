// Package metadata 提供元数据管理功能，用于 enhance 框架。
//
// 该模块用于存储和管理应用、组件、请求等的元数据信息。
// 提供类型安全的元数据访问接口，支持注解解析和配置元数据生成。
//
// # 架构设计
//
//   - Annotation: 注解结构体，用于标记和描述
//   - PropertyMetadata/GroupMetadata/HintMetadata: 配置元数据结构
//   - ConfigurationMetadata: 完整的配置元数据
//   - MetadataGenerator: 元数据生成器接口
//   - PropertyIndex: 属性索引接口，用于快速查找
//   - TagAnnotationResolver: 基于 struct tag 的注解解析器接口
//
// # 核心功能
//
//   - 注解解析: 支持基于 struct tag 的注解解析
//   - 配置元数据: 自动生成配置元数据（属性、组、提示）
//   - 属性索引: 支持快速查找和按前缀查询
//   - 类型映射: 自动映射 Go 类型到配置类型字符串
//
// # 使用方式
//
// 生成配置元数据：
//
//	type ServerConfig struct {
//	    Port int    `config:"server.port" description:"服务端口"`
//	    Host string `config:"server.host" description:"服务地址"`
//	}
//
//	metadata := metadata.GenerateFromStruct(&ServerConfig{})
//	jsonStr, _ := metadata.ToJSON()
//
// 使用注解解析器：
//
//	resolver := metadata.NewTagAnnotationResolver("metadata")
//	annotations := resolver.ResolveAnnotations(reflect.TypeOf(MyStruct{}))
//
// # 配置提示
//
// 支持为配置属性添加提示值和提供者：
//
//	gen := metadata.NewMetadataGenerator()
//	gen.Register(&config)
//	gen.WithHint("server.port", []metadata.HintValue{
//	    {Value: "8080", Description: "默认端口"},
//	})
package metadata

import (
	"reflect"
)

// Annotation 注解结构体。
type Annotation struct {
	// Name 注解名称。
	Name string
	// Attributes 注解属性。
	Attributes map[string]any
}

// PropertyMetadata 配置属性元数据。
type PropertyMetadata struct {
	// Name 属性名称（如 server.port）。
	Name string `json:"name"`
	// Type 属性类型。
	Type string `json:"type"`
	// Description 属性描述。
	Description string `json:"description,omitempty"`
	// DefaultValue 默认值。
	DefaultValue string `json:"defaultValue,omitempty"`
	// Deprecated 是否已弃用。
	Deprecated bool `json:"deprecated,omitempty"`
	// DeprecationReason 弃用原因。
	DeprecationReason string `json:"deprecationReason,omitempty"`
	// SourceType 来源类型（结构体名称）。
	SourceType string `json:"sourceType,omitempty"`
	// Required 是否必填。
	Required bool `json:"required,omitempty"`
	// Secret 是否是敏感信息（如密码）。
	Secret bool `json:"secret,omitempty"`
}

// GroupMetadata 配置组元数据。
type GroupMetadata struct {
	// Name 组名称（如 server）。
	Name string `json:"name"`
	// Type 组类型（结构体名称）。
	Type string `json:"type"`
	// Description 组描述。
	Description string `json:"description,omitempty"`
	// SourceType 来源类型。
	SourceType string `json:"sourceType,omitempty"`
}

// HintMetadata 配置提示元数据。
type HintMetadata struct {
	// Name 属性名称。
	Name string `json:"name"`
	// Values 可选值。
	Values []HintValue `json:"values,omitempty"`
	// Providers 提供者。
	Providers []HintProvider `json:"providers,omitempty"`
}

// HintValue 提示值。
type HintValue struct {
	// Value 值。
	Value string `json:"value"`
	// Description 描述。
	Description string `json:"description,omitempty"`
}

// HintProvider 提示提供者。
type HintProvider struct {
	// Name 提供者名称。
	Name string `json:"name"`
	// Parameters 参数。
	Parameters map[string]string `json:"parameters,omitempty"`
}

// ConfigurationMetadata 完整的配置元数据。
type ConfigurationMetadata struct {
	// Groups 配置组。
	Groups []GroupMetadata `json:"groups,omitempty"`
	// Properties 配置属性。
	Properties []PropertyMetadata `json:"properties"`
	// Hints 配置提示。
	Hints []HintMetadata `json:"hints,omitempty"`
}

// MetadataGenerator 元数据生成器接口。
//
// 用于从结构体生成配置元数据。
type MetadataGenerator interface {
	// Register 注册配置结构体。
	Register(config any) MetadataGenerator

	// WithHint 添加配置提示。
	WithHint(propertyName string, values []HintValue) MetadataGenerator

	// WithHintProvider 添加提示提供者。
	WithHintProvider(propertyName string, providerName string, params map[string]string) MetadataGenerator

	// Generate 生成配置元数据。
	Generate() *ConfigurationMetadata
}

// PropertyIndex 属性索引接口。
//
// 用于快速查找和按前缀查询属性元数据。
type PropertyIndex interface {
	// Get 获取属性元数据。
	Get(name string) (PropertyMetadata, bool)

	// Has 检查属性是否存在。
	Has(name string) bool

	// GetAll 获取所有属性。
	GetAll() []PropertyMetadata

	// GetByPrefix 按前缀获取属性。
	GetByPrefix(prefix string) []PropertyMetadata
}

// TagAnnotationResolver 基于 struct tag 的注解解析器接口。
//
// 解析格式为 `metadata:"name:attr1=val1,attr2=val2"` 的 tag。
// 支持多种属性类型和自动类型转换。
type TagAnnotationResolver interface {
	// ResolveAnnotations 解析指定类型的注解。
	ResolveAnnotations(typ reflect.Type) []Annotation
}
