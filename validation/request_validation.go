package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// ValidationRule 验证规则。
type ValidationRule struct {
	// Field 字段名称
	Field string `json:"field"`

	// Type 验证类型：required, string, number, email, regex, enum, min, max, length
	Type string `json:"type"`

	// Value 验证值（用于 enum, regex 等）
	Value string `json:"value,omitempty"`

	// Min 最小值
	Min *float64 `json:"min,omitempty"`

	// Max 最大值
	Max *float64 `json:"max,omitempty"`

	// MinLength 最小长度
	MinLength *int `json:"min_length,omitempty"`

	// MaxLength 最大长度
	MaxLength *int `json:"max_length,omitempty"`

	// Pattern 正则表达式
	Pattern string `json:"pattern,omitempty"`

	// Message 自定义错误消息
	Message string `json:"message,omitempty"`

	// In 枚举值
	In []string `json:"in,omitempty"`
}

// ValidationConfig 验证配置。
type ValidationConfig struct {
	// Rules 验证规则
	Rules []ValidationRule `json:"rules"`

	// Source 验证来源：query, header, body
	Source string `json:"source"`

	// FailFast 快速失败（遇到第一个错误就停止）
	FailFast bool `json:"fail_fast"`
}

// RuleValidationError 验证错误。
type RuleValidationError struct {
	// Field 字段名称
	Field string `json:"field"`

	// Message 错误消息
	Message string `json:"message"`

	// Type 错误类型
	Type string `json:"type"`
}

// RuleValidationResult 验证结果。
type RuleValidationResult struct {
	// Valid 是否通过验证
	Valid bool `json:"valid"`

	// Errors 错误列表
	Errors []RuleValidationError `json:"errors,omitempty"`
}

// RequestValidator HTTP 请求验证器。
type RequestValidator struct {
	config ValidationConfig
	regexs map[string]*regexp.Regexp
}

// NewRequestValidator 创建请求验证器
func NewRequestValidator(config ValidationConfig) (*RequestValidator, error) {
	validator := &RequestValidator{
		config: config,
		regexs: make(map[string]*regexp.Regexp),
	}

	// 预编译正则表达式
	for i, rule := range config.Rules {
		if rule.Pattern != "" {
			regex, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regular expression for field %s: %w", rule.Field, err)
			}
			validator.regexs[rule.Field] = regex
			config.Rules[i].Pattern = "" // 清空，使用预编译的
		}
	}

	return validator, nil
}

// GetConfig 获取验证配置。
func (v *RequestValidator) GetConfig() ValidationConfig {
	return v.config
}

// Validate 验证请求。
func (v *RequestValidator) Validate(req *http.Request) *RuleValidationResult {
	validationResult := &RuleValidationResult{Valid: true}

	// body 验证需要读取并解析请求体
	var bodyFields map[string]any
	if v.config.Source == "body" {
		parsed, err := decodeRequestBody(req)
		if err != nil {
			return invalidJSONResult(err)
		}
		bodyFields = parsed
	}

	for _, rule := range v.config.Rules {
		value := v.extractFieldValue(req, rule.Field, bodyFields)

		// 执行验证
		if err := v.validateRule(rule, value); err != nil {
			validationResult.Valid = false
			validationResult.Errors = append(validationResult.Errors, RuleValidationError{
				Field:   rule.Field,
				Message: err.Error(),
				Type:    rule.Type,
			})

			if v.config.FailFast {
				return validationResult
			}
		}
	}

	return validationResult
}

// decodeRequestBody 解析 JSON 请求体。
func decodeRequestBody(req *http.Request) (map[string]any, error) {
	if req.Body == nil {
		return nil, nil
	}
	var bodyFields map[string]any
	if err := json.NewDecoder(req.Body).Decode(&bodyFields); err != nil {
		return nil, err
	}
	return bodyFields, nil
}

// invalidJSONResult 构造 JSON 解析失败结果。
func invalidJSONResult(err error) *RuleValidationResult {
	return &RuleValidationResult{
		Valid: false,
		Errors: []RuleValidationError{{
			Field:   "body",
			Message: "Invalid JSON body: " + err.Error(),
			Type:    "json",
		}},
	}
}

// extractFieldValue 根据配置来源提取字段值。
func (v *RequestValidator) extractFieldValue(req *http.Request, field string, bodyFields map[string]any) string {
	switch v.config.Source {
	case "query":
		return req.URL.Query().Get(field)
	case "header":
		return req.Header.Get(field)
	case "body":
		if bodyFields != nil {
			if fieldValue, exists := bodyFields[field]; exists {
				return formatValueToString(fieldValue)
			}
		}
	}
	return ""
}

// formatValueToString 将任意值格式化为字符串
func formatValueToString(v any) string {
	switch typed := v.(type) {
	case string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// validateRule 验证单个规则
func (v *RequestValidator) validateRule(rule ValidationRule, value string) error {
	switch rule.Type {
	case "required":
		return validateRequiredValue(rule, value)
	case "string":
		return validateStringRule(rule, value)
	case "number":
		return validateNumberRule(rule, value)
	case "email":
		return validateEmailRule(rule, value)
	case "regex":
		return v.validateRegexRule(rule, value)
	case "enum":
		return validateEnumRule(rule, value)
	case "min":
		return validateMinRule(rule, value)
	case "max":
		return validateMaxRule(rule, value)
	case "length":
		return validateLengthRule(rule, value)
	default:
		return nil
	}
}

// validateRequiredValue 验证必填字段。
func validateRequiredValue(rule ValidationRule, value string) error {
	if value == "" {
		return fmt.Errorf("%s", rule.MessageOrDefault("%s is required", rule.Field))
	}
	return nil
}

// validateStringRule 验证字符串长度限制。
func validateStringRule(rule ValidationRule, value string) error {
	if value == "" {
		return nil // 空值不验证
	}
	if rule.MinLength != nil && len(value) < *rule.MinLength {
		return fmt.Errorf("%s", rule.MessageOrDefault("%s must be at least %d characters", rule.Field, *rule.MinLength))
	}
	if rule.MaxLength != nil && len(value) > *rule.MaxLength {
		return fmt.Errorf("%s", rule.MessageOrDefault("%s must be at most %d characters", rule.Field, *rule.MaxLength))
	}
	return nil
}

// validateNumberRule 验证数值格式与范围。
func validateNumberRule(rule ValidationRule, value string) error {
	if value == "" {
		return nil
	}
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("%s", rule.MessageOrDefault("%s must be a number", rule.Field))
	}
	if rule.Min != nil && num < *rule.Min {
		return fmt.Errorf("%s", rule.MessageOrDefault("%s must be at least %f", rule.Field, *rule.Min))
	}
	if rule.Max != nil && num > *rule.Max {
		return fmt.Errorf("%s", rule.MessageOrDefault("%s must be at most %f", rule.Field, *rule.Max))
	}
	return nil
}

// validateEmailRule 验证邮箱格式。
func validateEmailRule(rule ValidationRule, value string) error {
	if value == "" {
		return nil
	}
	if !strings.Contains(value, "@") || !strings.Contains(value, ".") {
		return fmt.Errorf("%s", rule.MessageOrDefault("%s must be a valid email", rule.Field))
	}
	return nil
}

// validateRegexRule 使用预编译正则验证格式。
func (v *RequestValidator) validateRegexRule(rule ValidationRule, value string) error {
	if value == "" {
		return nil
	}
	if regex, ok := v.regexs[rule.Field]; ok {
		if !regex.MatchString(value) {
			return fmt.Errorf("%s", rule.MessageOrDefault("%s format is invalid", rule.Field))
		}
	}
	return nil
}

// validateEnumRule 验证枚举取值。
func validateEnumRule(rule ValidationRule, value string) error {
	if value == "" {
		return nil
	}
	for _, val := range rule.In {
		if value == val {
			return nil
		}
	}
	return fmt.Errorf("%s", rule.MessageOrDefault("%s must be one of %v", rule.Field, rule.In))
}

// validateMinRule 验证最小值。
func validateMinRule(rule ValidationRule, value string) error {
	if value == "" {
		return nil
	}
	if rule.Min != nil {
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("%s", rule.MessageOrDefault("%s must be a number", rule.Field))
		}
		if num < *rule.Min {
			return fmt.Errorf("%s", rule.MessageOrDefault("%s must be at least %f", rule.Field, *rule.Min))
		}
	}
	return nil
}

// validateMaxRule 验证最大值。
func validateMaxRule(rule ValidationRule, value string) error {
	if value == "" {
		return nil
	}
	if rule.Max != nil {
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("%s", rule.MessageOrDefault("%s must be a number", rule.Field))
		}
		if num > *rule.Max {
			return fmt.Errorf("%s", rule.MessageOrDefault("%s must be at most %f", rule.Field, *rule.Max))
		}
	}
	return nil
}

// validateLengthRule 验证长度范围。
func validateLengthRule(rule ValidationRule, value string) error {
	if value == "" {
		return nil
	}
	if rule.MinLength != nil && len(value) < *rule.MinLength {
		return fmt.Errorf("%s", rule.MessageOrDefault("%s length must be at least %d", rule.Field, *rule.MinLength))
	}
	if rule.MaxLength != nil && len(value) > *rule.MaxLength {
		return fmt.Errorf("%s", rule.MessageOrDefault("%s length must be at most %d", rule.Field, *rule.MaxLength))
	}
	return nil
}

// MessageOrDefault 获取自定义消息或默认消息。
func (r ValidationRule) MessageOrDefault(format string, args ...any) string {
	if r.Message != "" {
		return r.Message
	}
	return fmt.Sprintf(format, args...)
}

// ValidateJSONBody 验证 JSON body。
func ValidateJSONBody(body []byte, rules []ValidationRule) *RuleValidationResult {
	var bodyFields map[string]any
	if err := json.Unmarshal(body, &bodyFields); err != nil {
		return &RuleValidationResult{
			Valid: false,
			Errors: []RuleValidationError{{
				Field:   "body",
				Message: "Invalid JSON body",
				Type:    "json",
			}},
		}
	}

	validationResult := &RuleValidationResult{Valid: true}
	for _, rule := range rules {
		valueStr := ""
		if value, exists := bodyFields[rule.Field]; exists {
			// 处理不同类型的值，保留原始类型信息
			valueStr = formatValueToString(value)
		}
		ruleValidator, err := NewRequestValidator(ValidationConfig{Rules: []ValidationRule{rule}})
		if err != nil {
			validationResult.Valid = false
			validationResult.Errors = append(validationResult.Errors, RuleValidationError{
				Field:   rule.Field,
				Message: err.Error(),
				Type:    "config",
			})
			continue
		}
		if err := ruleValidator.validateRule(rule, valueStr); err != nil {
			validationResult.Valid = false
			validationResult.Errors = append(validationResult.Errors, RuleValidationError{
				Field:   rule.Field,
				Message: err.Error(),
				Type:    rule.Type,
			})
		}
	}

	return validationResult
}

// ValidateHeaders 快速验证请求头。
func ValidateHeaders(req *http.Request, rules []ValidationRule) *RuleValidationResult {
	v, err := NewRequestValidator(ValidationConfig{
		Source:   "header",
		Rules:    rules,
		FailFast: false,
	})
	if err != nil {
		return &RuleValidationResult{
			Valid: false,
			Errors: []RuleValidationError{{
				Field:   "config",
				Message: err.Error(),
				Type:    "config",
			}},
		}
	}
	return v.Validate(req)
}

// ValidateQuery 快速验证查询参数。
func ValidateQuery(req *http.Request, rules []ValidationRule) *RuleValidationResult {
	v, err := NewRequestValidator(ValidationConfig{
		Source:   "query",
		Rules:    rules,
		FailFast: false,
	})
	if err != nil {
		return &RuleValidationResult{
			Valid: false,
			Errors: []RuleValidationError{{
				Field:   "config",
				Message: err.Error(),
				Type:    "config",
			}},
		}
	}
	return v.Validate(req)
}
