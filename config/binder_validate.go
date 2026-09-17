package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Validate 根据 validate 标签验证结构体字段
//
// 支持的验证规则：
//   - required: 字段不能为空
//   - min=N: 数值字段最小值，或字符串最小长度
//   - max=N: 数值字段最大值，或字符串最大长度
//   - enum=a,b,c: 枚举值限制
//
// 参数：
//   - target: 目标结构体指针
//
// 返回：
//   - error: 验证错误
func Validate(target any) error {
	// 检查 nil 值，避免反射操作 panic
	if target == nil {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	return validateStruct(targetValue.Elem(), "")
}

// validateStruct 递归验证结构体
func validateStruct(v reflect.Value, prefix string) error {
	structType := v.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldVal := v.Field(i)
		fieldName := prefix + field.Name

		// 处理嵌套结构体
		if field.Type.Kind() == reflect.Struct {
			if err := validateStruct(fieldVal, fieldName+"."); err != nil {
				return err
			}
			continue
		}

		// 验证字段
		if err := validateField(field, fieldVal, fieldName); err != nil {
			return err
		}
	}

	return nil
}

// validateField 验证单个字段
func validateField(field reflect.StructField, fieldVal reflect.Value, fieldName string) error {
	validateTag := field.Tag.Get("validate")
	if validateTag == "" {
		return nil
	}

	rules := strings.Split(validateTag, ",")
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if err := applyRule(rule, field, fieldVal, fieldName); err != nil {
			return err
		}
	}

	return nil
}

// applyRule 应用单个验证规则
func applyRule(rule string, field reflect.StructField, fieldVal reflect.Value, fieldName string) error {
	switch {
	case rule == "required":
		return validateRequired(fieldVal, fieldName)
	case strings.HasPrefix(rule, "min="):
		limitStr := strings.TrimPrefix(rule, "min=")
		return validateMin(fieldVal, limitStr, fieldName)
	case strings.HasPrefix(rule, "max="):
		limitStr := strings.TrimPrefix(rule, "max=")
		return validateMax(fieldVal, limitStr, fieldName)
	case strings.HasPrefix(rule, "enum="):
		vals := strings.TrimPrefix(rule, "enum=")
		return validateEnum(fieldVal, vals, fieldName)
	}
	return nil
}

// validateRequired 验证必填
func validateRequired(fieldVal reflect.Value, fieldName string) error {
	switch fieldVal.Kind() {
	case reflect.String:
		if fieldVal.String() == "" {
			return ValidationError{Field: fieldName, Message: "required field is empty"}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if fieldVal.Int() == 0 {
			return ValidationError{Field: fieldName, Message: "required field is zero"}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if fieldVal.Uint() == 0 {
			return ValidationError{Field: fieldName, Message: "required field is zero"}
		}
	case reflect.Float32, reflect.Float64:
		if fieldVal.Float() == 0 {
			return ValidationError{Field: fieldName, Message: "required field is zero"}
		}
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
		if fieldVal.IsNil() {
			return ValidationError{Field: fieldName, Message: "required field is nil"}
		}
	case reflect.Struct:
		if fieldVal.IsZero() {
			return ValidationError{Field: fieldName, Message: "required field is zero value"}
		}
	}
	return nil
}

// validateMin 验证最小值
func validateMin(fieldVal reflect.Value, minStr string, fieldName string) error {
	switch fieldVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		min, err := strconv.ParseInt(minStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid min value %q: %w", minStr, err)
		}
		if fieldVal.Int() < min {
			return ValidationError{Field: fieldName, Message: fmt.Sprintf("value %d below minimum %d", fieldVal.Int(), min)}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		min, err := strconv.ParseUint(minStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid min value %q: %w", minStr, err)
		}
		if fieldVal.Uint() < min {
			return ValidationError{Field: fieldName, Message: fmt.Sprintf("value %d below minimum %d", fieldVal.Uint(), min)}
		}
	case reflect.Float32, reflect.Float64:
		min, err := strconv.ParseFloat(minStr, 64)
		if err != nil {
			return fmt.Errorf("invalid min value %q: %w", minStr, err)
		}
		if fieldVal.Float() < min {
			return ValidationError{Field: fieldName, Message: fmt.Sprintf("value %f below minimum %f", fieldVal.Float(), min)}
		}
	case reflect.String:
		min, err := strconv.Atoi(minStr)
		if err != nil {
			return fmt.Errorf("invalid min length %q: %w", minStr, err)
		}
		if len(fieldVal.String()) < min {
			return ValidationError{Field: fieldName, Message: fmt.Sprintf("length %d below minimum %d", len(fieldVal.String()), min)}
		}
	}
	return nil
}

// validateMax 验证最大值
func validateMax(fieldVal reflect.Value, maxStr string, fieldName string) error {
	switch fieldVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		max, err := strconv.ParseInt(maxStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid max value %q: %w", maxStr, err)
		}
		if fieldVal.Int() > max {
			return ValidationError{Field: fieldName, Message: fmt.Sprintf("value %d above maximum %d", fieldVal.Int(), max)}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		max, err := strconv.ParseUint(maxStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid max value %q: %w", maxStr, err)
		}
		if fieldVal.Uint() > max {
			return ValidationError{Field: fieldName, Message: fmt.Sprintf("value %d above maximum %d", fieldVal.Uint(), max)}
		}
	case reflect.Float32, reflect.Float64:
		max, err := strconv.ParseFloat(maxStr, 64)
		if err != nil {
			return fmt.Errorf("invalid max value %q: %w", maxStr, err)
		}
		if fieldVal.Float() > max {
			return ValidationError{Field: fieldName, Message: fmt.Sprintf("value %f above maximum %f", fieldVal.Float(), max)}
		}
	case reflect.String:
		max, err := strconv.Atoi(maxStr)
		if err != nil {
			return fmt.Errorf("invalid max length %q: %w", maxStr, err)
		}
		if len(fieldVal.String()) > max {
			return ValidationError{Field: fieldName, Message: fmt.Sprintf("length %d above maximum %d", len(fieldVal.String()), max)}
		}
	}
	return nil
}

// validateEnum 验证枚举值
func validateEnum(fieldVal reflect.Value, enumStr string, fieldName string) error {
	allowed := strings.Split(enumStr, "|")
	strVal := fmt.Sprintf("%v", fieldVal.Interface())

	for _, allowedVal := range allowed {
		if strings.TrimSpace(allowedVal) == strVal {
			return nil
		}
	}

	return ValidationError{Field: fieldName, Message: fmt.Sprintf("value %q not in enum [%s]", strVal, enumStr)}
}
