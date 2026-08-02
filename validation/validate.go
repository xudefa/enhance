package validation

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ValidateStruct 用于验证结构体的便捷函数。
func ValidateStruct(obj any) error {
	validator := NewTagValidator()
	return validator.Validate(obj)
}

// Validate 验证单个值是否符合规则。
func Validate(value any, rules string) error {
	ruleList := strings.Split(rules, ",")
	for _, rule := range ruleList {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		if err := validateSingleRule(value, rule); err != nil {
			return err
		}
	}

	return nil
}

// validateSingleRule 验证单个规则
func validateSingleRule(value any, rule string) error {
	if strings.Contains(rule, "=") {
		return validateRuleWithParam(value, rule)
	}
	return validateRuleWithoutParam(value, rule)
}

// validateRuleWithParam 验证带参数的规则
func validateRuleWithParam(value any, rule string) error {
	parts := strings.SplitN(rule, "=", 2)
	ruleName := parts[0]
	ruleValue := parts[1]

	var isValid bool
	switch ruleName {
	case "min":
		isValid = isMinValidForValue(value, ruleValue)
	case "max":
		isValid = isMaxValidForValue(value, ruleValue)
	case "len":
		isValid = isLenValidForValue(value, ruleValue)
	case "email":
		isValid = isEmailValidForValue(value)
	case "regexp":
		isValid = isRegexpValidForValue(value, ruleValue)
	case "gt":
		isValid = isGtValidForValue(value, ruleValue)
	case "gte":
		isValid = isGteValidForValue(value, ruleValue)
	case "lt":
		isValid = isLtValidForValue(value, ruleValue)
	case "lte":
		isValid = isLteValidForValue(value, ruleValue)
	case "url":
		isValid = isURLValidForValue(value)
	case "ip":
		isValid = isIPValidForValue(value)
	case "oneof":
		isValid = isOneOfValidForValue(value, ruleValue)
	default:
		return nil
	}

	if !isValid {
		return fmt.Errorf("validation failed, rule: %s", rule)
	}
	return nil
}

// validateRuleWithoutParam 验证不带参数的规则
func validateRuleWithoutParam(value any, rule string) error {
	var isValid bool
	switch rule {
	case "required":
		isValid = isRequiredValidForValue(value)
	case "email":
		isValid = isEmailValidForValue(value)
	case "url":
		isValid = isURLValidForValue(value)
	case "ip":
		isValid = isIPValidForValue(value)
	default:
		return nil
	}

	if !isValid {
		return fmt.Errorf("validation failed, rule: %s", rule)
	}
	return nil
}

// isRequiredValidForValue 验证值是否必需（非零值）
func isRequiredValidForValue(value any) bool {
	if value == nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.String() != ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Bool:
		return v.Bool()
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return !v.IsNil()
	default:
		return v.IsValid() && value != reflect.Zero(v.Type()).Interface()
	}
}

// isMinValidForValue 验证最小值
func isMinValidForValue(value any, minStr string) bool {
	min, err := strconv.Atoi(minStr)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String())) >= min
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() >= int64(min)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() >= uint64(min)
	case reflect.Float32, reflect.Float64:
		return v.Float() >= float64(min)
	default:
		return false
	}
}

// isMaxValidForValue 验证最大值
func isMaxValidForValue(value any, maxStr string) bool {
	max, err := strconv.Atoi(maxStr)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String())) <= max
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() <= int64(max)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() <= uint64(max)
	case reflect.Float32, reflect.Float64:
		return v.Float() <= float64(max)
	default:
		return false
	}
}

// isLenValidForValue 验证长度
func isLenValidForValue(value any, lenStr string) bool {
	length, err := strconv.Atoi(lenStr)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String())) == length
	case reflect.Array, reflect.Slice:
		return v.Len() == length
	default:
		return false
	}
}

// isEmailValidForValue 验证邮箱格式
func isEmailValidForValue(value any) bool {
	if value == nil {
		return false
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.String {
		return false
	}
	email := v.String()
	emailRegex := compileRegex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isRegexpValidForValue 验证正则表达式匹配
func isRegexpValidForValue(value any, pattern string) bool {
	if value == nil {
		return false
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.String {
		return false
	}
	re := compileRegex(pattern)
	if re == nil {
		return false
	}
	return re.MatchString(v.String())
}

// isGtValidForValue 验证大于指定值
func isGtValidForValue(value any, valueStr string) bool {
	val, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String())) > val
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() > int64(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() > uint64(val)
	case reflect.Float32, reflect.Float64:
		return v.Float() > float64(val)
	default:
		return false
	}
}

// isGteValidForValue 验证大于等于指定值
func isGteValidForValue(value any, valueStr string) bool {
	val, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String())) >= val
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() >= int64(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() >= uint64(val)
	case reflect.Float32, reflect.Float64:
		return v.Float() >= float64(val)
	default:
		return false
	}
}

// isLtValidForValue 验证小于指定值
func isLtValidForValue(value any, valueStr string) bool {
	val, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String())) < val
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() < int64(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() < uint64(val)
	case reflect.Float32, reflect.Float64:
		return v.Float() < float64(val)
	default:
		return false
	}
}

// isLteValidForValue 验证小于等于指定值
func isLteValidForValue(value any, valueStr string) bool {
	val, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String())) <= val
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() <= int64(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() <= uint64(val)
	case reflect.Float32, reflect.Float64:
		return v.Float() <= float64(val)
	default:
		return false
	}
}

// isURLValidForValue 验证URL格式
func isURLValidForValue(value any) bool {
	if value == nil {
		return false
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.String {
		return false
	}
	url := v.String()
	urlRegex := compileRegex(`^https?:\/\/(?:[-\w.])+(?:\:[0-9]+)?(?:\/(?:[\w\/_.])*(?:\?(?:[\w&=%.])*)?(?:\#(?:[\w.])*)?)?$`)
	return urlRegex.MatchString(url)
}

// isIPValidForValue 验证IP地址格式
func isIPValidForValue(value any) bool {
	if value == nil {
		return false
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.String {
		return false
	}
	ip := v.String()
	ipRegex := compileRegex(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	return ipRegex.MatchString(ip)
}

// isOneOfValidForValue 验证值是否在指定的选项中
func isOneOfValidForValue(value any, optionsStr string) bool {
	if value == nil {
		return false
	}

	options := strings.Split(optionsStr, " ")

	for i, opt := range options {
		options[i] = strings.TrimSpace(opt)
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		strVal := v.String()
		for _, opt := range options {
			if strVal == opt {
				return true
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal := v.Int()
		for _, opt := range options {
			if optNum, err := strconv.ParseInt(opt, 10, 64); err == nil && intVal == optNum {
				return true
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal := v.Uint()
		for _, opt := range options {
			if optNum, err := strconv.ParseUint(opt, 10, 64); err == nil && uintVal == optNum {
				return true
			}
		}
	case reflect.Float32, reflect.Float64:
		floatVal := v.Float()
		for _, opt := range options {
			if optNum, err := strconv.ParseFloat(opt, 64); err == nil && floatVal == optNum {
				return true
			}
		}
	}

	return false
}
