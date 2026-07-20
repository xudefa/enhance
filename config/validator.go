package config

import (
	"fmt"
	"regexp"
)

// DefaultValidator 默认验证器实现。
type DefaultValidator struct {
	rules []ValidationRule
}

// Error 实现 error 接口。
func (e ValidationError) Error() string {
	return fmt.Sprintf("field %s: %s", e.Field, e.Message)
}

// Error 实现 error 接口。
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	msg := "validation errors:"
	for _, err := range e {
		msg += "\n  - " + err.Field + ": " + err.Message
	}
	return msg
}

// NewValidator 创建默认验证器
func NewValidator() *DefaultValidator {
	return &DefaultValidator{
		rules: make([]ValidationRule, 0),
	}
}

// Validate 验证配置数据
func (v *DefaultValidator) Validate(data map[string]any) error {
	var errors ValidationErrors

	for _, rule := range v.rules {
		value, exists := data[rule.Field]
		if !exists {
			errors = append(errors, ValidationError{
				Field:   rule.Field,
				Message: "field not found",
			})
			continue
		}

		if err := rule.Check(value); err != nil {
			errors = append(errors, ValidationError{
				Field:   rule.Field,
				Message: err.Error(),
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// AddRequired 添加必填字段验证
func (v *DefaultValidator) AddRequired(fields ...string) {
	for _, field := range fields {
		v.rules = append(v.rules, ValidationRule{
			Field: field,
			Check: func(value any) error {
				if value == nil || value == "" {
					return fmt.Errorf("required field is empty")
				}
				return nil
			},
		})
	}
}

// AddMin 添加最小值验证
func (v *DefaultValidator) AddMin(field string, min int) {
	v.rules = append(v.rules, ValidationRule{
		Field: field,
		Check: func(value any) error {
			switch val := value.(type) {
			case int:
				if val < min {
					return fmt.Errorf("value %d below minimum %d", val, min)
				}
			case float64:
				if val < float64(min) {
					return fmt.Errorf("value %f below minimum %d", val, min)
				}
			case string:
				if len(val) < min {
					return fmt.Errorf("length %d below minimum %d", len(val), min)
				}
			}
			return nil
		},
	})
}

// AddMax 添加最大值验证
func (v *DefaultValidator) AddMax(field string, max int) {
	v.rules = append(v.rules, ValidationRule{
		Field: field,
		Check: func(value any) error {
			switch val := value.(type) {
			case int:
				if val > max {
					return fmt.Errorf("value %d above maximum %d", val, max)
				}
			case float64:
				if val > float64(max) {
					return fmt.Errorf("value %f above maximum %d", val, max)
				}
			case string:
				if len(val) > max {
					return fmt.Errorf("length %d above maximum %d", len(val), max)
				}
			}
			return nil
		},
	})
}

// AddRegex 添加正则表达式验证
func (v *DefaultValidator) AddRegex(field, pattern string) {
	v.rules = append(v.rules, ValidationRule{
		Field: field,
		Check: func(value any) error {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("field is not a string")
			}
			matched, err := regexp.MatchString(pattern, str)
			if err != nil {
				return fmt.Errorf("invalid regex pattern: %w", err)
			}
			if !matched {
				return fmt.Errorf("value does not match pattern %s", pattern)
			}
			return nil
		},
	})
}

// AddEnum 添加枚举值验证
func (v *DefaultValidator) AddEnum(field string, values ...any) {
	v.rules = append(v.rules, ValidationRule{
		Field: field,
		Check: func(value any) error {
			for _, allowed := range values {
				if value == allowed {
					return nil
				}
			}
			return fmt.Errorf("value %v not in enum %v", value, values)
		},
	})
}

// AddCustomRule 添加自定义验证规则
func (v *DefaultValidator) AddCustomRule(field string, fn func(any) error) {
	v.rules = append(v.rules, ValidationRule{
		Field: field,
		Check: fn,
	})
}
