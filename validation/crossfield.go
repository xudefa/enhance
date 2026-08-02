package validation

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// interfaceOrNil 安全地获取 reflect.Value 的接口值，避免因未导出字段导致 panic。
func interfaceOrNil(v reflect.Value) any {
	if v.CanInterface() {
		return v.Interface()
	}
	return nil
}

// validateCrossField 验证跨字段规则
func (v *TagValidator) validateCrossField(field reflect.Value, rule, fieldName string, obj any) error {
	// 解析规则：fieldmatch=OtherField, fieldne=OtherField, fieldgt=OtherField, etc.
	parts := strings.SplitN(rule, "=", 2)
	if len(parts) < 2 {
		return ValidationError{
			Field:   fieldName,
			Message: "跨字段验证规则格式错误",
			Value:   interfaceOrNil(field),
		}
	}

	ruleName := parts[0]
	otherFieldName := parts[1]

	// 提取验证类型（去掉 "field" 前缀）
	if !strings.HasPrefix(ruleName, "field") {
		return ValidationError{
			Field:   fieldName,
			Message: "跨字段验证规则格式错误",
			Value:   interfaceOrNil(field),
		}
	}

	validationType := strings.TrimPrefix(ruleName, "field")
	if validationType == "" || validationType == "match" {
		validationType = "eq" // field 或 fieldmatch 都表示相等验证
	}

	// 验证另一个字段是否存在
	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}
	_, ok := objType.FieldByName(otherFieldName)
	if !ok {
		return ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("字段 %s 不存在", otherFieldName),
			Value:   interfaceOrNil(field),
		}
	}

	objVal := reflect.ValueOf(obj)
	if objVal.Kind() == reflect.Ptr {
		objVal = objVal.Elem()
	}
	otherValue := objVal.FieldByName(otherFieldName)
	if !otherValue.IsValid() {
		return ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("字段 %s 的值无效", otherFieldName),
			Value:   interfaceOrNil(field),
		}
	}

	switch validationType {
	case "eq":
		if !v.fieldsEqual(field, otherValue) {
			return ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("字段必须与 %s 相等", otherFieldName),
				Value:   interfaceOrNil(field),
			}
		}
	case "ne":
		if v.fieldsEqual(field, otherValue) {
			return ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("字段必须与 %s 不相等", otherFieldName),
				Value:   interfaceOrNil(field),
			}
		}
	case "gt":
		if !v.fieldGreaterThan(field, otherValue) {
			return ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("字段必须大于 %s", otherFieldName),
				Value:   interfaceOrNil(field),
			}
		}
	case "gte":
		if !v.fieldGreaterThanOrEqual(field, otherValue) {
			return ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("字段必须大于或等于 %s", otherFieldName),
				Value:   interfaceOrNil(field),
			}
		}
	case "lt":
		if !v.fieldLessThan(field, otherValue) {
			return ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("字段必须小于 %s", otherFieldName),
				Value:   interfaceOrNil(field),
			}
		}
	case "lte":
		if !v.fieldLessThanOrEqual(field, otherValue) {
			return ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("字段必须小于或等于 %s", otherFieldName),
				Value:   interfaceOrNil(field),
			}
		}
	}

	return nil
}

// validateWhenCondition 验证条件依赖规则
func (v *TagValidator) validateWhenCondition(field reflect.Value, rule, fieldName string, obj any) []ValidationError {
	// 解析规则：when=condition:rules
	parts := strings.SplitN(rule, "=", 2)
	if len(parts) < 2 {
		return nil
	}

	conditionParts := strings.SplitN(parts[1], ":", 2)
	if len(conditionParts) < 2 {
		return nil
	}

	condition := conditionParts[0]
	rules := conditionParts[1]

	// 检查条件是否满足
	if !v.evaluateCondition(condition, obj) {
		return nil
	}

	// 条件满足，验证规则（将 ; 转换为 , 以兼容 validateField）
	rules = strings.ReplaceAll(rules, ";", ",")
	return v.validateField(field, rules, fieldName, obj)
}

// evaluateCondition 评估条件表达式
func (v *TagValidator) evaluateCondition(condition string, obj any) bool {
	// 解析条件：field==value, field!=value, field<value, field>value, field<=value, field>=value

	// 检查 == 操作符
	if idx := strings.Index(condition, "=="); idx != -1 {
		fieldName := condition[:idx]
		expectedValue := condition[idx+2:]
		actualValue := v.getFieldValue(obj, fieldName)
		return fmt.Sprintf("%v", actualValue) == expectedValue
	}

	// 检查 != 操作符
	if idx := strings.Index(condition, "!="); idx != -1 {
		fieldName := condition[:idx]
		expectedValue := condition[idx+2:]
		actualValue := v.getFieldValue(obj, fieldName)
		return fmt.Sprintf("%v", actualValue) != expectedValue
	}

	// 检查 <= 操作符
	if idx := strings.Index(condition, "<="); idx != -1 {
		fieldName := condition[:idx]
		expectedValue := condition[idx+2:]
		return v.compareFieldValue(obj, fieldName, expectedValue, "<=")
	}

	// 检查 >= 操作符
	if idx := strings.Index(condition, ">="); idx != -1 {
		fieldName := condition[:idx]
		expectedValue := condition[idx+2:]
		return v.compareFieldValue(obj, fieldName, expectedValue, ">=")
	}

	// 检查 < 操作符
	if idx := strings.Index(condition, "<"); idx != -1 {
		fieldName := condition[:idx]
		expectedValue := condition[idx+1:]
		return v.compareFieldValue(obj, fieldName, expectedValue, "<")
	}

	// 检查 > 操作符
	if idx := strings.Index(condition, ">"); idx != -1 {
		fieldName := condition[:idx]
		expectedValue := condition[idx+1:]
		return v.compareFieldValue(obj, fieldName, expectedValue, ">")
	}

	return false
}

// compareFieldValue 比较字段值与期望值
func (v *TagValidator) compareFieldValue(obj any, fieldName, expectedValue, operator string) bool {
	actualValue := v.getFieldValue(obj, fieldName)
	if actualValue == nil {
		return false
	}

	rv := reflect.ValueOf(actualValue)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		expected, err := strconv.ParseInt(expectedValue, 10, 64)
		if err != nil {
			return false
		}
		actual := rv.Int()
		switch operator {
		case "<":
			return actual < expected
		case "<=":
			return actual <= expected
		case ">":
			return actual > expected
		case ">=":
			return actual >= expected
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		expected, err := strconv.ParseUint(expectedValue, 10, 64)
		if err != nil {
			return false
		}
		actual := rv.Uint()
		switch operator {
		case "<":
			return actual < expected
		case "<=":
			return actual <= expected
		case ">":
			return actual > expected
		case ">=":
			return actual >= expected
		}
	case reflect.Float32, reflect.Float64:
		expected, err := strconv.ParseFloat(expectedValue, 64)
		if err != nil {
			return false
		}
		actual := rv.Float()
		switch operator {
		case "<":
			return actual < expected
		case "<=":
			return actual <= expected
		case ">":
			return actual > expected
		case ">=":
			return actual >= expected
		}
	case reflect.String:
		actual := rv.String()
		switch operator {
		case "<":
			return len([]rune(actual)) < len([]rune(expectedValue))
		case "<=":
			return len([]rune(actual)) <= len([]rune(expectedValue))
		case ">":
			return len([]rune(actual)) > len([]rune(expectedValue))
		case ">=":
			return len([]rune(actual)) >= len([]rune(expectedValue))
		}
	}

	return false
}

// getFieldValue 获取字段值
func (v *TagValidator) getFieldValue(obj any, fieldName string) any {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	field := rv.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}

	return interfaceOrNil(field)
}

// fieldsEqual 比较两个字段是否相等
func (v *TagValidator) fieldsEqual(f1, f2 reflect.Value) bool {
	if f1.Kind() != f2.Kind() {
		return false
	}

	switch f1.Kind() {
	case reflect.String:
		return f1.String() == f2.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return f1.Int() == f2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return f1.Uint() == f2.Uint()
	case reflect.Float32, reflect.Float64:
		return f1.Float() == f2.Float()
	case reflect.Bool:
		return f1.Bool() == f2.Bool()
	default:
		return f1.Interface() == f2.Interface()
	}
}

// fieldGreaterThan 比较字段大小
func (v *TagValidator) fieldGreaterThan(f1, f2 reflect.Value) bool {
	switch f1.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return f1.Int() > f2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return f1.Uint() > f2.Uint()
	case reflect.Float32, reflect.Float64:
		return f1.Float() > f2.Float()
	case reflect.String:
		return len([]rune(f1.String())) > len([]rune(f2.String()))
	default:
		return false
	}
}

// fieldGreaterThanOrEqual 比较字段大小
func (v *TagValidator) fieldGreaterThanOrEqual(f1, f2 reflect.Value) bool {
	switch f1.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return f1.Int() >= f2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return f1.Uint() >= f2.Uint()
	case reflect.Float32, reflect.Float64:
		return f1.Float() >= f2.Float()
	case reflect.String:
		return len([]rune(f1.String())) >= len([]rune(f2.String()))
	default:
		return false
	}
}

// fieldLessThan 比较字段大小
func (v *TagValidator) fieldLessThan(f1, f2 reflect.Value) bool {
	switch f1.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return f1.Int() < f2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return f1.Uint() < f2.Uint()
	case reflect.Float32, reflect.Float64:
		return f1.Float() < f2.Float()
	case reflect.String:
		return len([]rune(f1.String())) < len([]rune(f2.String()))
	default:
		return false
	}
}

// fieldLessThanOrEqual 比较字段大小
func (v *TagValidator) fieldLessThanOrEqual(f1, f2 reflect.Value) bool {
	switch f1.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return f1.Int() <= f2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return f1.Uint() <= f2.Uint()
	case reflect.Float32, reflect.Float64:
		return f1.Float() <= f2.Float()
	case reflect.String:
		return len([]rune(f1.String())) <= len([]rune(f2.String()))
	default:
		return false
	}
}
