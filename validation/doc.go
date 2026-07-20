// Package validation 提供参数校验功能，用于 enhance 框架。
//
// 该模块提供字段级校验、跨字段校验、校验规则注册等功能，支持 HTTP 中间件集成。
// 参考 Jakarta Bean Validation (JSR 380) 的设计理念。
//
// # 架构设计
//
//   - Validator: 校验器接口，定义校验操作
//   - ValidationRule: 校验规则接口，定义单个校验逻辑
//   - ValidationContext: 校验上下文，包含被校验对象和错误信息
//   - ValidationBuilder: 校验构建器，支持链式配置
//   - ValidationMiddleware: HTTP 校验中间件
//   - CustomValidator: 自定义验证器接口
//   - ValidatorRegistry: 验证器注册表，支持并发安全地注册和获取自定义验证器
//   - MiddlewareValidator: 中间件验证器接口
//
// # 核心功能
//
//   - 字段级校验: 支持 @Required, @Min, @Max, @Email 等常用校验
//   - 跨字段校验: 支持比较两个字段的值（如密码确认）
//   - 规则注册: 支持自定义校验规则
//   - 错误收集: 收集所有校验错误并返回
//   - HTTP 集成: 提供 HTTP 中间件自动校验请求参数
//
// # 使用方式
//
// 定义校验规则：
//
//	type User struct {
//	    Name  string `validate:"required,min=2,max=50"`
//	    Email string `validate:"required,email"`
//	    Age   int    `validate:"required,min=18,max=100"`
//	}
//
// 校验对象：
//
//	validator := validation.NewValidator()
//	errs := validator.Validate(user)
//	if len(errs) > 0 {
//	    // 处理校验错误
//	}
//
// 使用校验构建器：
//
//	builder := validation.NewBuilder()
//	builder.Rule("name").Required().Min(2).Max(50)
//	builder.Rule("email").Required().Email()
//	validator := builder.Build()
//
// # 内置校验规则
//
//   - Required: 必填
//   - NotBlank: 非空字符串
//   - Min/Max: 最小/最大值
//   - Size: 字符串长度或集合大小
//   - Email: 邮箱格式
//   - URL: URL 格式
//   - Regex: 正则表达式匹配
//   - In: 值在指定列表中
//   - NotIn: 值不在指定列表中
package validation

import (
	"net/http"
	"reflect"
	"sync"
)

// Validator 验证器接口。
//
// 定义了验证操作的标准接口。
type Validator interface {
	// Validate 验证对象。
	Validate(obj any) error
}

// ValidationError 验证错误结构。
//
// 包含字段名称、错误消息和实际值。
type ValidationError struct {
	Field   string `json:"field"`           // 字段名称
	Message string `json:"message"`         // 错误消息
	Value   any    `json:"value,omitempty"` // 实际值
}

// ValidationErrors 验证错误集合。
//
// 实现了错误接口，包含多个验证错误。
type ValidationErrors []ValidationError

// CustomValidator 自定义验证器接口。
type CustomValidator interface {
	// Validate 验证字段值。
	Validate(field reflect.Value, param string) (bool, string)
}

// ValidatorRegistry 验证器注册表。
//
// 支持并发安全地注册和获取自定义验证器。
// 使用 sync.Map 优化读多写少场景的并发性能。
type ValidatorRegistry struct {
	validators     sync.Map // map[string]CustomValidator
	funcValidators sync.Map // map[string]func(reflect.Value, string) (bool, string)
}

// RuleBuilder 验证规则构建器。
//
// 支持链式配置校验规则。
type RuleBuilder struct {
	rules    []string
	messages map[string]string
}

// MiddlewareValidator 中间件验证器接口。
type MiddlewareValidator interface {
	// ValidateRequest 验证请求对象。
	ValidateRequest(c any, obj any) error

	// HandleValidationError 处理验证错误。
	HandleValidationError(c any, err error)
}

// MiddlewareConfig 中间件配置。
type MiddlewareConfig struct {
	Validator    Validator
	Groups       []string
	ErrorHandler func(c any, err error)
	SkipPaths    []string
}

// TagValidator 基于标签的验证器。
//
// 支持多种验证规则，通过结构体标签定义校验规则。
type TagValidator struct {
	registry *ValidatorRegistry
}

// ErrorResponse 错误响应结构。
//
// 用于 HTTP 中间件中的错误响应格式。
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ResponseWriterInterface 响应写入器接口。
//
// 抽象层，适配不同 HTTP 框架的响应写入。
type ResponseWriterInterface interface {
	// SetStatusCode 设置 HTTP 状态码。
	SetStatusCode(code int)
	// SetHeader 设置响应头。
	SetHeader(key, value string)
	// Write 写入响应体。
	Write(data []byte) error
}

// Binder 参数绑定接口。
//
// 定义了将 HTTP 请求参数绑定到结构体的标准方法。
type Binder interface {
	// Bind 将请求参数绑定到目标对象。
	Bind(req *http.Request, obj any) error
}

// 以下类型定义在其他文件中，此处仅作文档说明：
// - ValidationMiddleware: middleware.go（包含完整实现）
// - ValidationGroup: groups.go（包含完整实现）
// - CrossFieldValidator: crossfield.go（包含完整实现）
// - BindingValidator: binding.go（包含完整实现）
