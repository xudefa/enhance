// Package validation 提供参数校验功能，用于 enhance 框架。
package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Error 返回单个字段校验错误的文本描述。
func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Error 将所有校验错误用分号连接后返回。
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

	errs := v.validateStructFields(rv, rt, obj)
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// validateStructFields 遍历结构体字段并执行标签验证。
func (v *TagValidator) validateStructFields(rv reflect.Value, rt reflect.Type, obj any) ValidationErrors {
	errsPtr := acquireValidationErrors()
	defer releaseValidationErrors(errsPtr)

	for i := range rv.NumField() {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue
		}

		fieldErrors := v.validateField(field, tag, resolveFieldName(fieldType), obj)
		*errsPtr = append(*errsPtr, fieldErrors...)
	}

	if len(*errsPtr) == 0 {
		return nil
	}
	errsCopy := make(ValidationErrors, len(*errsPtr))
	copy(errsCopy, *errsPtr)
	return errsCopy
}

// resolveFieldName 解析字段的展示名称（优先使用 json 标签）。
func resolveFieldName(fieldType reflect.StructField) string {
	fieldName := fieldType.Name
	jsonTag := fieldType.Tag.Get("json")
	if jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" {
			fieldName = parts[0]
		}
	}
	return fieldName
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

	if hasRequiredRule(rules) && !v.isRequiredValid(field) {
		errs = append(errs, ValidationError{
			Field:   fieldName,
			Message: "字段是必需的",
			Value:   getFieldValueUnsafe(field),
		})
	}

	return append(errs, v.evaluateFieldRules(field, rules, fieldName, obj)...)
}

// hasRequiredRule 检查规则列表中是否包含 required。
func hasRequiredRule(rules []string) bool {
	for _, rule := range rules {
		if strings.TrimSpace(rule) == "required" {
			return true
		}
	}
	return false
}

// evaluateFieldRules 遍历规则列表并逐个执行规则验证。
func (v *TagValidator) evaluateFieldRules(field reflect.Value, rules []string, fieldName string, obj any) []ValidationError {
	errs := make([]ValidationError, 0, len(rules))
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		switch {
		case strings.HasPrefix(rule, "when="):
			whenErrs := v.validateWhenCondition(field, rule, fieldName, obj)
			if len(whenErrs) > 0 {
				errs = append(errs, whenErrs...)
			}
		case strings.HasPrefix(rule, "field"):
			errs = append(errs, v.applyCrossFieldRule(field, rule, fieldName, obj)...)
		case rule == "required":
		case strings.Contains(rule, "="):
			parts := strings.SplitN(rule, "=", 2)
			errs = v.appendEvalErrors(errs, field, validationRule{name: parts[0], value: parts[1]}, fieldName)
		default:
			// 处理不含参数的规则
			switch rule {
			case "email", "url", "ip":
				errs = v.appendEvalErrors(errs, field, validationRule{name: rule}, fieldName)
			default:
				errs = v.applyCustomValidator(errs, field, rule, fieldName)
			}
		}
	}
	return errs
}

// applyCrossFieldRule 应用跨字段验证规则并归一化返回结果。
func (v *TagValidator) applyCrossFieldRule(field reflect.Value, rule, fieldName string, obj any) []ValidationError {
	err := v.validateCrossField(field, rule, fieldName, obj)
	if err == nil {
		return nil
	}
	if ve, ok := err.(ValidationError); ok {
		return []ValidationError{ve}
	}
	return []ValidationError{{
		Field:   fieldName,
		Message: err.Error(),
		Value:   getFieldValueUnsafe(field),
	}}
}

// applyCustomValidator 使用注册表中的自定义验证器验证字段。
func (v *TagValidator) applyCustomValidator(errs []ValidationError, field reflect.Value, rule, fieldName string) []ValidationError {
	if v.registry == nil {
		return errs
	}
	customValidator, ok := v.registry.GetFunc(rule)
	if !ok {
		return errs
	}
	if valid, msg := customValidator(field, ""); !valid {
		errs = append(errs, ValidationError{
			Field:   fieldName,
			Message: msg,
			Value:   getFieldValueUnsafe(field),
		})
	}
	return errs
}

// validationRule 一条验证规则（名称=值）。
type validationRule struct {
	name  string
	value string
}

// appendEvalErrors 按规则名评估验证规则并追加错误。
func (v *TagValidator) appendEvalErrors(errs []ValidationError, field reflect.Value, rule validationRule, fieldName string) []ValidationError {
	var ruleErrs []ValidationError
	switch rule.name {
	case "min", "max", "gt", "gte", "lt", "lte":
		ruleErrs = v.validateComparisonRules(field, rule.name, rule.value, fieldName)
	default:
		ruleErrs = v.validateFormatRules(field, rule.name, rule.value, fieldName)
	}
	return append(errs, ruleErrs...)
}

// validateComparisonRules 验证数值比较类规则（min/max/gt/gte/lt/lte）。
func (v *TagValidator) validateComparisonRules(field reflect.Value, ruleName, ruleValue, fieldName string) []ValidationError {
	var errs []ValidationError
	switch ruleName {
	case "min", "max", "gt", "gte", "lt", "lte":
		errs = append(errs, v.checkComparisonRule(field, ruleName, ruleValue, fieldName)...)
	}
	return errs
}

// checkComparisonRule 校验单条数值比较规则并返回对应的错误。
func (v *TagValidator) checkComparisonRule(field reflect.Value, ruleName, ruleValue, fieldName string) []ValidationError {
	valid := false
	switch ruleName {
	case "min":
		valid = v.isMinValid(field, ruleValue)
	case "max":
		valid = v.isMaxValid(field, ruleValue)
	case "gt":
		valid = v.isGtValid(field, ruleValue)
	case "gte":
		valid = v.isGteValid(field, ruleValue)
	case "lt":
		valid = v.isLtValid(field, ruleValue)
	case "lte":
		valid = v.isLteValid(field, ruleValue)
	}
	if valid {
		return nil
	}

	threshold, _ := strconv.Atoi(ruleValue)
	message := fmt.Sprintf("字段值必须大于或等于 %d", threshold)
	switch ruleName {
	case "max":
		message = fmt.Sprintf("字段值必须小于或等于 %d", threshold)
	case "gt":
		message = fmt.Sprintf("字段值必须大于 %d", threshold)
	case "gte":
		message = fmt.Sprintf("字段值必须大于或等于 %d", threshold)
	case "lt":
		message = fmt.Sprintf("字段值必须小于 %d", threshold)
	case "lte":
		message = fmt.Sprintf("字段值必须小于或等于 %d", threshold)
	}

	return []ValidationError{
		{
			Field:   fieldName,
			Message: message,
			Value:   getFieldValueUnsafe(field),
		},
	}
}

// validateFormatRules 验证格式类规则（len/regexp/email/url/ip/oneof）。
func (v *TagValidator) validateFormatRules(field reflect.Value, ruleName, ruleValue, fieldName string) []ValidationError {
	var errs []ValidationError
	switch ruleName {
	case "len", "email", "regexp", "url", "ip", "oneof":
		errs = append(errs, v.checkFormatRule(field, ruleName, ruleValue, fieldName)...)
	}
	return errs
}

// checkFormatRule 校验单条格式规则并返回对应的错误。
func (v *TagValidator) checkFormatRule(field reflect.Value, ruleName, ruleValue, fieldName string) []ValidationError {
	if v.isFormatValid(field, ruleName, ruleValue) {
		return nil
	}
	msg := v.formatRuleMessage(ruleName, ruleValue)
	return []ValidationError{
		{Field: fieldName, Message: msg, Value: getFieldValueUnsafe(field)},
	}
}

// isFormatValid 根据规则名称判定格式是否合法。
func (v *TagValidator) isFormatValid(field reflect.Value, ruleName, ruleValue string) bool {
	switch ruleName {
	case "len":
		return v.isLenValid(field, ruleValue)
	case "email":
		return v.isEmailValid(field)
	case "regexp":
		return v.isRegexpValid(field, ruleValue)
	case "url":
		return v.isURLValid(field)
	case "ip":
		return v.isIPValid(field)
	case "oneof":
		return v.isOneOfValid(field, ruleValue)
	default:
		return true
	}
}

// formatRuleMessage 根据规则名称生成中文校验失败消息。
func (v *TagValidator) formatRuleMessage(ruleName, ruleValue string) string {
	switch ruleName {
	case "len":
		length, _ := strconv.Atoi(ruleValue)
		return fmt.Sprintf("字段长度必须为 %d", length)
	case "email":
		return "字段必须是有效的邮箱地址"
	case "regexp":
		return fmt.Sprintf("字段不匹配正则表达式: %s", ruleValue)
	case "url":
		return "字段必须是有效的URL地址"
	case "ip":
		return "字段必须是有效的IP地址"
	case "oneof":
		return fmt.Sprintf("字段值必须是以下选项之一: %s", ruleValue)
	default:
		return "字段格式不正确"
	}
}
