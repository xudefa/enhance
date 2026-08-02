// Package validation 提供参数校验功能，用于 enhance 框架。
package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e ValidationErrors) Error() string {
	msgs := make([]string, 0, len(e))
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// NewTagValidator 创建新的标签验证器实例
func NewTagValidator() *TagValidator {
	return &TagValidator{
		registry: NewValidatorRegistry(),
	}
}

// NewTagValidatorWithRegistry 创建带有注册表的标签验证器实例
func NewTagValidatorWithRegistry(registry *ValidatorRegistry) *TagValidator {
	return &TagValidator{
		registry: registry,
	}
}

// Validate 验证对象，对结构体的字段进行验证
func (v *TagValidator) Validate(obj any) error {
	if obj == nil {
		return nil
	}

	rv := reflect.ValueOf(obj)
	rt := reflect.TypeOf(obj)

	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
		rt = rt.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return errors.New("validation: only struct types are supported")
	}

	errsPtr := acquireValidationErrors()
	defer releaseValidationErrors(errsPtr)

	for i := range rv.NumField() {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue
		}

		fieldName := fieldType.Name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		fieldErrors := v.validateField(field, tag, fieldName, obj)
		*errsPtr = append(*errsPtr, fieldErrors...)
	}

	if len(*errsPtr) > 0 {
		errsCopy := make(ValidationErrors, len(*errsPtr))
		copy(errsCopy, *errsPtr)
		return errsCopy
	}
	return nil
}

// splitRules 分割验证规则，保留 regexp 规则中的逗号
//
// 正则表达式可能包含逗号（如 regexp=^\d{1,3}$），不能直接按逗号分割，
// 否则正则表达式会被拆坏。约定：regexp= 规则必须放在标签最后，
// 其值将占用剩余所有内容。
func splitRules(tag string) []string {
	parts := strings.Split(tag, ",")
	rules := make([]string, 0, len(parts))
	for i, part := range parts {
		rule := strings.TrimSpace(part)
		if strings.HasPrefix(rule, "regexp=") && i < len(parts)-1 {
			rule += "," + strings.Join(parts[i+1:], ",")
			rules = append(rules, rule)
			break
		}
		rules = append(rules, rule)
	}
	return rules
}

// validateField 验证单个字段，解析验证规则并执行验证
func (v *TagValidator) validateField(field reflect.Value, tag, fieldName string, obj any) []ValidationError {
	rules := splitRules(tag)
	errs := make([]ValidationError, 0, len(rules))

	isRequired := false
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "required" {
			isRequired = true
			break
		}
	}

	isEmpty := !v.isRequiredValid(field)

	if isRequired && isEmpty {
		errs = append(errs, ValidationError{
			Field:   fieldName,
			Message: "字段是必需的",
			Value:   getFieldValueUnsafe(field),
		})
	}

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		if strings.HasPrefix(rule, "when=") {
			whenErrs := v.validateWhenCondition(field, rule, fieldName, obj)
			if len(whenErrs) > 0 {
				errs = append(errs, whenErrs...)
			}
			continue
		}

		if strings.HasPrefix(rule, "field") {
			err := v.validateCrossField(field, rule, fieldName, obj)
			if err != nil {
				if ve, ok := err.(ValidationError); ok {
					errs = append(errs, ve)
				} else {
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: err.Error(),
						Value:   getFieldValueUnsafe(field),
					})
				}
			}
			continue
		}

		if rule == "required" {
			continue
		}

		if strings.Contains(rule, "=") {
			parts := strings.SplitN(rule, "=", 2)
			ruleName := parts[0]
			ruleValue := parts[1]

			switch ruleName {
			case "min":
				if !v.isMinValid(field, ruleValue) {
					min, _ := strconv.Atoi(ruleValue)
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段值必须大于或等于 %d", min),
						Value:   getFieldValueUnsafe(field),
					})
				}
			case "max":
				if !v.isMaxValid(field, ruleValue) {
					max, _ := strconv.Atoi(ruleValue)
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段值必须小于或等于 %d", max),
						Value:   getFieldValueUnsafe(field),
					})
				}
			case "len":
				if !v.isLenValid(field, ruleValue) {
					length, _ := strconv.Atoi(ruleValue)
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段长度必须为 %d", length),
						Value:   getFieldValueUnsafe(field),
					})
				}
			case "email":
				if !v.isEmailValid(field) {
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: "字段必须是有效的邮箱地址",
						Value:   getFieldValueUnsafe(field),
					})
				}
			case "regexp":
				if !v.isRegexpValid(field, ruleValue) {
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段不匹配正则表达式: %s", ruleValue),
						Value:   getFieldValueUnsafe(field),
					})
				}
			case "gt":
				if !v.isGtValid(field, ruleValue) {
					gt, _ := strconv.Atoi(ruleValue)
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段值必须大于 %d", gt),
						Value:   getFieldValueUnsafe(field),
					})
				}
			case "gte":
				if !v.isGteValid(field, ruleValue) {
					gte, _ := strconv.Atoi(ruleValue)
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段值必须大于或等于 %d", gte),
						Value:   getFieldValueUnsafe(field),
					})
				}
			case "lt":
				if !v.isLtValid(field, ruleValue) {
					lt, _ := strconv.Atoi(ruleValue)
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段值必须小于 %d", lt),
						Value:   getFieldValueUnsafe(field),
					})
				}
			case "lte":
				if !v.isLteValid(field, ruleValue) {
					lte, _ := strconv.Atoi(ruleValue)
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段值必须小于或等于 %d", lte),
						Value:   getFieldValueUnsafe(field),
					})
				}
			case "url":
				if !v.isURLValid(field) {
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: "字段必须是有效的URL地址",
						Value:   getFieldValueUnsafe(field),
					})
				}
			case "ip":
				if !v.isIPValid(field) {
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: "字段必须是有效的IP地址",
						Value:   getFieldValueUnsafe(field),
					})
				}
			case "oneof":
				if !v.isOneOfValid(field, ruleValue) {
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段值必须是以下选项之一: %s", ruleValue),
						Value:   getFieldValueUnsafe(field),
					})
				}
			}
			continue
		}

		// 处理不含参数的规则
		switch rule {
		case "email":
			if !v.isEmailValid(field) {
				errs = append(errs, ValidationError{
					Field:   fieldName,
					Message: "字段必须是有效的邮箱地址",
					Value:   getFieldValueUnsafe(field),
				})
			}
		case "url":
			if !v.isURLValid(field) {
				errs = append(errs, ValidationError{
					Field:   fieldName,
					Message: "字段必须是有效的URL地址",
					Value:   getFieldValueUnsafe(field),
				})
			}
		case "ip":
			if !v.isIPValid(field) {
				errs = append(errs, ValidationError{
					Field:   fieldName,
					Message: "字段必须是有效的IP地址",
					Value:   getFieldValueUnsafe(field),
				})
			}
		default:
			if v.registry != nil {
				if customValidator, ok := v.registry.GetFunc(rule); ok {
					if valid, msg := customValidator(field, ""); !valid {
						errs = append(errs, ValidationError{
							Field:   fieldName,
							Message: msg,
							Value:   getFieldValueUnsafe(field),
						})
					}
				}
			}
		}
	}

	return errs
}
