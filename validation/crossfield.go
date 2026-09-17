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
		return crossRuleFormatError(fieldName, field)
	}

	ruleName := parts[0]
	otherFieldName := parts[1]

	// 提取验证类型（去掉 "field" 前缀）
	if !strings.HasPrefix(ruleName, "field") {
		return crossRuleFormatError(fieldName, field)
	}

	validationType := strings.TrimPrefix(ruleName, "field")
	if validationType == "" || validationType == "match" {
		validationType = "eq" // field 或 fieldmatch 都表示相等验证
	}

	// 验证另一个字段是否存在并获取其值
	otherValue, err := v.resolveCrossFieldValue(obj, otherFieldName)
	if err != nil {
		return ValidationError{
			Field:   fieldName,
			Message: err.Error(),
			Value:   interfaceOrNil(field),
		}
	}

	return v.applyCrossValidation(field, otherValue, otherFieldName, validationType, fieldName)
}

// crossRuleFormatError 构造跨字段规则格式错误。
func crossRuleFormatError(fieldName string, field reflect.Value) error {
	return ValidationError{
		Field:   fieldName,
		Message: "跨字段验证规则格式错误",
		Value:   interfaceOrNil(field),
	}
}

// resolveCrossFieldValue 定位并获取对端字段的反射值。
func (v *TagValidator) resolveCrossFieldValue(obj any, otherFieldName string) (reflect.Value, error) {
	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}
	_, ok := objType.FieldByName(otherFieldName)
	if !ok {
		return reflect.Value{}, fmt.Errorf("字段 %s 不存在", otherFieldName)
	}

	objVal := reflect.ValueOf(obj)
	if objVal.Kind() == reflect.Ptr {
		objVal = objVal.Elem()
	}
	otherValue := objVal.FieldByName(otherFieldName)
	if !otherValue.IsValid() {
		return reflect.Value{}, fmt.Errorf("字段 %s 的值无效", otherFieldName)
	}
	return otherValue, nil
}

// crossMismatchError 构造跨字段比较失败错误。
func crossMismatchError(fieldName string, field reflect.Value, message string) error {
	return ValidationError{
		Field:   fieldName,
		Message: message,
		Value:   interfaceOrNil(field),
	}
}

// applyCrossValidation 根据比较类型执行跨字段比较。
func (v *TagValidator) applyCrossValidation(field, otherValue reflect.Value, otherFieldName, validationType, fieldName string) error {
	switch validationType {
	case "eq":
		if !v.fieldsEqual(field, otherValue) {
			return crossMismatchError(fieldName, field, fmt.Sprintf("字段必须与 %s 相等", otherFieldName))
		}
	case "ne":
		if v.fieldsEqual(field, otherValue) {
			return crossMismatchError(fieldName, field, fmt.Sprintf("字段必须与 %s 不相等", otherFieldName))
		}
	case "gt":
		if !v.fieldGreaterThan(field, otherValue) {
			return crossMismatchError(fieldName, field, fmt.Sprintf("字段必须大于 %s", otherFieldName))
		}
	case "gte":
		if !v.fieldGreaterThanOrEqual(field, otherValue) {
			return crossMismatchError(fieldName, field, fmt.Sprintf("字段必须大于或等于 %s", otherFieldName))
		}
	case "lt":
		if !v.fieldLessThan(field, otherValue) {
			return crossMismatchError(fieldName, field, fmt.Sprintf("字段必须小于 %s", otherFieldName))
		}
	case "lte":
		if !v.fieldLessThanOrEqual(field, otherValue) {
			return crossMismatchError(fieldName, field, fmt.Sprintf("字段必须小于或等于 %s", otherFieldName))
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
		return compareIntFieldValue(rv.Int(), expectedValue, operator)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return compareUintFieldValue(rv.Uint(), expectedValue, operator)
	case reflect.Float32, reflect.Float64:
		return compareFloatFieldValue(rv.Float(), expectedValue, operator)
	case reflect.String:
		return compareStringFieldValue(rv.String(), expectedValue, operator)
	}

	return false
}

// compareIntFieldValue 比较有符号整数字段值与期望值。
func compareIntFieldValue(actual int64, expectedValue, operator string) bool {
	expected, err := strconv.ParseInt(expectedValue, 10, 64)
	if err != nil {
		return false
	}
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
	return false
}

// compareUintFieldValue 比较无符号整数字段值与期望值。
func compareUintFieldValue(actual uint64, expectedValue, operator string) bool {
	expected, err := strconv.ParseUint(expectedValue, 10, 64)
	if err != nil {
		return false
	}
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
	return false
}

// compareFloatFieldValue 比较浮点字段值与期望值。
func compareFloatFieldValue(actual float64, expectedValue, operator string) bool {
	expected, err := strconv.ParseFloat(expectedValue, 64)
	if err != nil {
		return false
	}
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
	return false
}

// compareStringFieldValue 按 rune 长度比较字符串字段值与期望值。
func compareStringFieldValue(actual, expectedValue, operator string) bool {
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
