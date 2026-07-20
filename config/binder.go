package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TypeConverter 类型转换函数类型
type TypeConverter func(string) (any, error)

var converters sync.Map // map[reflect.Type]TypeConverter

func init() {
	// 注册内置转换器
	RegisterConverter(time.Duration(0), parseDuration)
	RegisterConverter([]string(nil), parseStringList)
	RegisterConverter(map[string]string(nil), parseStringMap)
}

// RegisterConverter 注册自定义类型转换器
func RegisterConverter(target any, fn TypeConverter) {
	converters.Store(reflect.TypeOf(target), fn)
}

// GetConverter 获取指定类型的转换器
func GetConverter(target reflect.Type) (TypeConverter, bool) {
	fn, ok := converters.Load(target)
	if !ok {
		return nil, false
	}
	return fn.(TypeConverter), true
}

func parseDuration(s string) (any, error) {
	return time.ParseDuration(s)
}

func parseStringList(s string) (any, error) {
	s = strings.Trim(s, "[]")
	if s == "" {
		return []string{}, nil
	}
	parts := strings.Split(s, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts, nil
}

func parseStringMap(s string) (any, error) {
	result := make(map[string]string)
	s = strings.Trim(s, "{}")
	if s == "" {
		return result, nil
	}
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return result, nil
}

// Bind 将配置数据绑定到结构体
//
// 支持以下特性：
//   - env 标签：从扁平配置键映射到结构体字段
//   - 复杂类型自动转换：Duration、[]string、map[string]string 等
//   - 嵌套结构体：自动递归绑定嵌套字段
//   - validate 标签：集成验证规则（required、min、max 等）
//
// 参数：
//   - cfg: 配置数据源（Config 接口或 map[string]any）
//   - target: 目标结构体指针
//
// 返回：
//   - error: 绑定或验证错误
func Bind(cfg any, target any) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	var data map[string]any
	switch c := cfg.(type) {
	case Config:
		data = c.GetAll()
	case map[string]any:
		data = c
	default:
		return fmt.Errorf("cfg must be Config or map[string]any")
	}

	return bindStruct(data, v.Elem())
}

// bindStruct 递归绑定结构体字段
func bindStruct(data map[string]any, v reflect.Value) error {
	t := v.Type()

	for i := range t.NumField() {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// 跳过未导出字段
		if !field.IsExported() {
			continue
		}

		// 处理嵌套结构体（但有转换器的结构体类型优先使用转换器）
		if field.Type.Kind() == reflect.Struct {
			// 检查是否有注册的转换器，有则当作普通字段处理
			if _, ok := GetConverter(field.Type); ok {
				if err := bindField(data, field, fieldVal); err != nil {
					return err
				}
				continue
			}
			if err := bindNestedStruct(data, field, fieldVal); err != nil {
				return err
			}
			continue
		}

		// 处理普通字段
		if err := bindField(data, field, fieldVal); err != nil {
			return err
		}
	}

	return nil
}

// bindNestedStruct 处理嵌套结构体字段
func bindNestedStruct(data map[string]any, field reflect.StructField, fieldVal reflect.Value) error {
	// 检查是否有 env 标签
	envTag := field.Tag.Get("env")
	if envTag == "" {
		// 无 env 标签，尝试直接递归绑定（扁平结构体）
		return bindStruct(data, fieldVal)
	}

	// 有 env 标签，从配置中提取对应前缀的子配置
	prefix := envTag + "."
	subData := extractSubMap(data, prefix)
	if len(subData) == 0 {
		return nil
	}

	return bindStruct(subData, fieldVal)
}

// extractSubMap 从扁平配置中提取指定前缀的子配置
func extractSubMap(data map[string]any, prefix string) map[string]any {
	result := make(map[string]any)
	for key, value := range data {
		if strings.HasPrefix(key, prefix) {
			subKey := strings.TrimPrefix(key, prefix)
			result[subKey] = value
		}
	}
	return result
}

// bindField 绑定单个字段
func bindField(data map[string]any, field reflect.StructField, fieldVal reflect.Value) error {
	envTag := field.Tag.Get("env")
	if envTag == "" {
		return nil
	}

	value, exists := data[envTag]
	if !exists {
		// 检查是否有默认值
		defaultTag := field.Tag.Get("default")
		if defaultTag != "" {
			return setFieldValue(fieldVal, field.Type, defaultTag)
		}
		return nil
	}

	// 如果值已经是目标类型，直接设置
	if reflect.TypeOf(value) == field.Type {
		fieldVal.Set(reflect.ValueOf(value))
		return nil
	}

	// 尝试字符串转换
	strVal, ok := value.(string)
	if !ok {
		// 尝试 fmt.Sprintf 转换
		strVal = fmt.Sprintf("%v", value)
	}

	return setFieldValue(fieldVal, field.Type, strVal)
}

// setFieldValue 设置字段值，支持类型转换
func setFieldValue(fieldVal reflect.Value, targetType reflect.Type, strVal string) error {
	// 检查是否有注册的转换器
	if converter, ok := GetConverter(targetType); ok {
		converted, err := converter(strVal)
		if err != nil {
			return fmt.Errorf("failed to convert %q to %s: %w", strVal, targetType, err)
		}
		convertedVal := reflect.ValueOf(converted)
		// 如果转换器返回的是接口类型，需要提取底层值
		if convertedVal.Kind() == reflect.Interface {
			convertedVal = convertedVal.Elem()
		}
		if !convertedVal.Type().AssignableTo(fieldVal.Type()) {
			return fmt.Errorf("converted value of type %s is not assignable to field type %s", convertedVal.Type(), fieldVal.Type())
		}
		fieldVal.Set(convertedVal)
		return nil
	}

	// 基本类型转换
	switch targetType.Kind() {
	case reflect.String:
		fieldVal.SetString(strVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse int %q: %w", strVal, err)
		}
		fieldVal.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(strVal, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse uint %q: %w", strVal, err)
		}
		fieldVal.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			return fmt.Errorf("failed to parse float %q: %w", strVal, err)
		}
		fieldVal.SetFloat(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(strVal)
		if err != nil {
			return fmt.Errorf("failed to parse bool %q: %w", strVal, err)
		}
		fieldVal.SetBool(v)
	default:
		return fmt.Errorf("unsupported type %s for field", targetType)
	}

	return nil
}

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
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	return validateStruct(v.Elem(), "")
}

// validateStruct 递归验证结构体
func validateStruct(v reflect.Value, prefix string) error {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
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
		val := strings.TrimPrefix(rule, "min=")
		return validateMin(fieldVal, val, fieldName)
	case strings.HasPrefix(rule, "max="):
		val := strings.TrimPrefix(rule, "max=")
		return validateMax(fieldVal, val, fieldName)
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
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
		if fieldVal.IsNil() {
			return ValidationError{Field: fieldName, Message: "required field is nil"}
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
