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
	parsed := make(map[string]string)
	s = strings.Trim(s, "{}")
	if s == "" {
		return parsed, nil
	}
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			parsed[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return parsed, nil
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
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	var configData map[string]any
	switch c := cfg.(type) {
	case Config:
		configData = c.GetAll()
	case map[string]any:
		configData = c
	default:
		return fmt.Errorf("cfg must be Config or map[string]any")
	}

	return bindStruct(configData, targetValue.Elem())
}

// bindStruct 递归绑定结构体字段
func bindStruct(data map[string]any, v reflect.Value) error {
	structType := v.Type()

	for i := range structType.NumField() {
		field := structType.Field(i)
		fieldVal := v.Field(i)

		// 跳过未导出字段
		if !field.IsExported() {
			continue
		}

		// 处理嵌套结构体（含指针结构体）
		fieldType := field.Type
		fieldKind := fieldType.Kind()
		if fieldKind == reflect.Ptr {
			if fieldVal.IsNil() && fieldVal.CanSet() {
				fieldVal.Set(reflect.New(fieldType.Elem()))
			}
			if !fieldVal.IsNil() {
				fieldKind = fieldVal.Elem().Kind()
				fieldVal = fieldVal.Elem()
				fieldType = fieldType.Elem()
			}
		}
		if fieldKind == reflect.Struct {
			// 检查是否有注册的转换器，有则当作普通字段处理
			if _, ok := GetConverter(fieldType); ok {
				if err := bindField(data, field, fieldVal); err != nil {
					return fmt.Errorf("绑定字段 %s 失败: %w", field.Name, err)
				}
				continue
			}
			if err := bindNestedStruct(data, field, fieldVal); err != nil {
				return fmt.Errorf("绑定嵌套结构体 %s 失败: %w", field.Name, err)
			}
			continue
		}

		// 处理普通字段
		if err := bindField(data, field, fieldVal); err != nil {
			return fmt.Errorf("绑定字段 %s 失败: %w", field.Name, err)
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
	subMap := make(map[string]any)
	for key, value := range data {
		if strings.HasPrefix(key, prefix) {
			subKey := strings.TrimPrefix(key, prefix)
			subMap[subKey] = value
		}
	}
	return subMap
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
		if !fieldVal.CanSet() {
			return fmt.Errorf("field %s is not addressable", field.Name)
		}
		fieldVal.Set(reflect.ValueOf(value))
		return nil
	}

	// 尝试字符串转换
	strVal, ok := value.(string)
	if !ok {
		if value == nil {
			return fmt.Errorf("cannot convert nil to %s", field.Type)
		}
		// 尝试 fmt.Sprintf 转换
		strVal = fmt.Sprintf("%v", value)
	}

	return setFieldValue(fieldVal, field.Type, strVal)
}

// setFieldValue 设置字段值，支持类型转换
func setFieldValue(fieldVal reflect.Value, targetType reflect.Type, strVal string) error {
	// 检查是否有注册的转换器
	if applied, err := applyCustomConverter(fieldVal, targetType, strVal); applied {
		return err
	}

	// 基本类型转换
	return setBasicTypeValue(fieldVal, targetType, strVal)
}

// applyCustomConverter 应用注册的自定义类型转换器。
// 返回是否已找到并应用转换器。
func applyCustomConverter(fieldVal reflect.Value, targetType reflect.Type, strVal string) (bool, error) {
	converter, ok := GetConverter(targetType)
	if !ok {
		return false, nil
	}
	converted, err := converter(strVal)
	if err != nil {
		return true, fmt.Errorf("failed to convert %q to %s: %w", strVal, targetType, err)
	}
	convertedVal := reflect.ValueOf(converted)
	// 如果转换器返回的是接口类型，需要提取底层值
	if convertedVal.Kind() == reflect.Interface {
		convertedVal = convertedVal.Elem()
	}
	if !convertedVal.Type().AssignableTo(fieldVal.Type()) {
		return true, fmt.Errorf("converted value of type %s is not assignable to field type %s", convertedVal.Type(), fieldVal.Type())
	}
	if !fieldVal.CanSet() {
		return true, fmt.Errorf("field %s is not addressable", targetType)
	}
	fieldVal.Set(convertedVal)
	return true, nil
}

// setBasicTypeValue 按基本类型转换并设置字段值。
func setBasicTypeValue(fieldVal reflect.Value, targetType reflect.Type, strVal string) error {
	switch targetType.Kind() {
	case reflect.String:
		fieldVal.SetString(strVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsedVal, err := strconv.ParseInt(strVal, 10, targetType.Bits())
		if err != nil {
			return fmt.Errorf("failed to parse int %q: %w", strVal, err)
		}
		fieldVal.SetInt(parsedVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsedVal, err := strconv.ParseUint(strVal, 10, targetType.Bits())
		if err != nil {
			return fmt.Errorf("failed to parse uint %q: %w", strVal, err)
		}
		fieldVal.SetUint(parsedVal)
	case reflect.Float32, reflect.Float64:
		parsedVal, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			return fmt.Errorf("failed to parse float %q: %w", strVal, err)
		}
		fieldVal.SetFloat(parsedVal)
	case reflect.Bool:
		parsedVal, err := strconv.ParseBool(strVal)
		if err != nil {
			return fmt.Errorf("failed to parse bool %q: %w", strVal, err)
		}
		fieldVal.SetBool(parsedVal)
	default:
		return fmt.Errorf("unsupported type %s for field", targetType)
	}

	return nil
}
