package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/xudefa/enhance/config/environment"
)

// PropertyBinder 配置属性绑定器
//
// 将 Environment 中的配置值绑定到结构体字段。
// 支持嵌套结构体、切片、映射和自定义类型转换。
type PropertyBinder struct {
	env       *environment.Environment
	prefix    string
	validator Validator
}

// NewPropertyBinder 创建配置属性绑定器
func NewPropertyBinder(env *environment.Environment) *PropertyBinder {
	return &PropertyBinder{
		env: env,
	}
}

// WithPrefix 设置配置前缀
func (b *PropertyBinder) WithPrefix(prefix string) *PropertyBinder {
	b.prefix = prefix
	return b
}

// WithValidator 设置验证器
func (b *PropertyBinder) WithValidator(v Validator) *PropertyBinder {
	b.validator = v
	return b
}

// Bind 将配置绑定到目标结构体
//
// 参数:
//   - target: 目标结构体指针
//
// 返回:
//   - error: 绑定错误
func (b *PropertyBinder) Bind(target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	targetValue = targetValue.Elem()
	if targetValue.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	return b.bindStruct(targetValue, b.prefix)
}

// bindStruct 递归绑定结构体字段
func (b *PropertyBinder) bindStruct(v reflect.Value, prefix string) error {
	t := v.Type()

	for i := range v.NumField() {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 跳过未导出字段
		if !fieldValue.CanSet() {
			continue
		}

		// 处理嵌入字段
		if field.Anonymous {
			if err := b.bindStruct(fieldValue, prefix); err != nil {
				return fmt.Errorf("bind embedded struct: %w", err)
			}
			continue
		}

		// 获取 enhance 标签
		tag := field.Tag.Get("enhance")
		if tag == "" {
			// 没有标签，尝试递归处理嵌套结构体（含指针结构体）
			if err := b.bindNestedStruct(fieldValue, prefix); err != nil {
				return fmt.Errorf("bind nested struct: %w", err)
			}
			continue
		}

		// 构建完整配置键
		key := tag
		if prefix != "" {
			key = prefix + "." + tag
		}

		// 获取配置值
		if err := b.bindField(fieldValue, key, field); err != nil {
			return fmt.Errorf("failed to bind field %s: %w", field.Name, err)
		}
	}

	return nil
}

// bindNestedStruct 递归处理无标签的嵌套结构体（含指针结构体）。
func (b *PropertyBinder) bindNestedStruct(fieldValue reflect.Value, prefix string) error {
	switch fieldValue.Kind() {
	case reflect.Struct:
		return b.bindStruct(fieldValue, prefix)
	case reflect.Ptr:
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
		}
		if fieldValue.Elem().Kind() == reflect.Struct {
			return b.bindStruct(fieldValue.Elem(), prefix)
		}
	}
	return nil
}

// bindField 绑定单个字段
func (b *PropertyBinder) bindField(v reflect.Value, key string, field reflect.StructField) error {
	// 获取配置值
	value, ok := b.env.GetProperty(key)

	if !ok {
		// 配置不存在，尝试使用默认值
		defaultVal := field.Tag.Get("default")
		if defaultVal != "" {
			return b.setDefaultValue(v, defaultVal, field.Type)
		}
		return nil
	}

	// 设置值
	return b.setValue(v, value)
}

// setDefaultValue 设置默认值
func (b *PropertyBinder) setDefaultValue(v reflect.Value, defaultVal string, targetType reflect.Type) error {
	// 特殊处理 time.Duration
	if targetType == reflect.TypeOf(time.Duration(0)) {
		return setDurationValue(v, defaultVal)
	}

	// 根据字段类型转换默认值
	switch v.Kind() {
	case reflect.Slice:
		// 逗号分隔的字符串转切片
		return b.setDefaultSliceValue(v, defaultVal, targetType)
	}

	return setDefaultBasicValue(v, defaultVal, targetType)
}

// setDurationValue 解析并设置 time.Duration 字段。
func setDurationValue(v reflect.Value, defaultVal string) error {
	duration, err := time.ParseDuration(defaultVal)
	if err != nil {
		return fmt.Errorf("invalid duration value %q: %w", defaultVal, err)
	}
	v.Set(reflect.ValueOf(duration))
	return nil
}

// setDefaultBasicValue 将默认值按基本类型解析并设置字段。
func setDefaultBasicValue(v reflect.Value, defaultVal string, targetType reflect.Type) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(defaultVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(defaultVal, 10, targetType.Bits())
		if err != nil {
			return fmt.Errorf("invalid integer value %q: %w", defaultVal, err)
		}
		v.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(defaultVal, 10, targetType.Bits())
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value %q: %w", defaultVal, err)
		}
		v.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(defaultVal, 64)
		if err != nil {
			return fmt.Errorf("invalid float value %q: %w", defaultVal, err)
		}
		v.SetFloat(parsed)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(defaultVal)
		if err != nil {
			return fmt.Errorf("invalid boolean value %q: %w", defaultVal, err)
		}
		v.SetBool(parsed)
	}
	return nil
}

// setDefaultSliceValue 将逗号分隔的默认值转为切片并设置字段。
func (b *PropertyBinder) setDefaultSliceValue(v reflect.Value, defaultVal string, targetType reflect.Type) error {
	parts := strings.Split(defaultVal, ",")
	slice := reflect.MakeSlice(targetType, len(parts), len(parts))
	for i, part := range parts {
		elem := slice.Index(i)
		if err := b.setValue(elem, strings.TrimSpace(part)); err != nil {
			return fmt.Errorf("设置切片元素 %d 失败: %w", i, err)
		}
	}
	v.Set(slice)
	return nil
}

// setValue 设置字段值
func (b *PropertyBinder) setValue(v reflect.Value, value any) error {
	if value == nil {
		return nil
	}

	srcVal := reflect.ValueOf(value)

	// 直接类型匹配
	if srcVal.Type().AssignableTo(v.Type()) {
		v.Set(srcVal)
		return nil
	}

	// 类型转换
	switch v.Kind() {
	case reflect.String:
		v.SetString(fmt.Sprintf("%v", value))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return setNumericValue(v, value)

	case reflect.Float32, reflect.Float64:
		return setFloatValue(v, value)

	case reflect.Bool:
		return setBoolValue(v, value)

	case reflect.Struct:
		return setDurationFieldValue(v, value)

	case reflect.Slice:
		return b.bindSlice(v, value)

	case reflect.Map:
		return b.bindMap(v, value)

	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return b.setValue(v.Elem(), value)
	}

	return nil
}

// setNumericValue 使用类型转换器设置整数/无符号整数字段。
func setNumericValue(v reflect.Value, value any) error {
	converted, err := environment.NewTypeConverter().ConvertTo(value, v.Type())
	if err != nil {
		return fmt.Errorf("cannot convert %T to %s: %w", value, v.Type(), err)
	}
	v.Set(converted)
	return nil
}

// setFloatValue 设置浮点字段，支持 float64/int/string 源值。
func setFloatValue(v reflect.Value, value any) error {
	switch raw := value.(type) {
	case float64:
		v.SetFloat(raw)
	case int:
		v.SetFloat(float64(raw))
	case string:
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %q to %s: %w", raw, v.Type(), err)
		}
		v.SetFloat(parsed)
	default:
		return fmt.Errorf("cannot convert %T to %s", value, v.Type())
	}
	return nil
}

// setBoolValue 设置布尔字段，支持 bool/string 源值。
func setBoolValue(v reflect.Value, value any) error {
	switch raw := value.(type) {
	case bool:
		v.SetBool(raw)
	case string:
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("cannot convert %q to %s: %w", raw, v.Type(), err)
		}
		v.SetBool(parsed)
	default:
		return fmt.Errorf("cannot convert %T to %s", value, v.Type())
	}
	return nil
}

// setDurationFieldValue 设置 time.Duration 结构体字段。
// 非 time.Duration 结构体保持原有跳过语义，直接返回 nil。
func setDurationFieldValue(v reflect.Value, value any) error {
	if v.Type() != reflect.TypeOf(time.Duration(0)) {
		return nil
	}
	switch duration := value.(type) {
	case time.Duration:
		v.Set(reflect.ValueOf(duration))
	case string:
		parsed, err := time.ParseDuration(duration)
		if err != nil {
			return fmt.Errorf("cannot convert %q to %s: %w", duration, v.Type(), err)
		}
		v.Set(reflect.ValueOf(parsed))
	case int64:
		v.Set(reflect.ValueOf(time.Duration(duration)))
	default:
		return fmt.Errorf("cannot convert %T to %s", value, v.Type())
	}
	return nil
}

// bindSlice 绑定切片字段
func (b *PropertyBinder) bindSlice(v reflect.Value, value any) error {
	switch raw := value.(type) {
	case []any:
		slice := reflect.MakeSlice(v.Type(), len(raw), len(raw))
		for i, item := range raw {
			elem := slice.Index(i)
			if err := b.setValue(elem, item); err != nil {
				return fmt.Errorf("设置切片元素 %d 失败: %w", i, err)
			}
		}
		v.Set(slice)
	case string:
		// 逗号分隔的字符串
		parts := strings.Split(raw, ",")
		slice := reflect.MakeSlice(v.Type(), len(parts), len(parts))
		for i, part := range parts {
			elem := slice.Index(i)
			if err := b.setValue(elem, strings.TrimSpace(part)); err != nil {
				return fmt.Errorf("设置切片元素 %d 失败: %w", i, err)
			}
		}
		v.Set(slice)
	default:
		return fmt.Errorf("cannot convert %T to %s", value, v.Type())
	}
	return nil
}

// bindMap 绑定映射字段
func (b *PropertyBinder) bindMap(v reflect.Value, value any) error {
	switch raw := value.(type) {
	case map[string]any:
		mapType := v.Type()
		mapVal := reflect.MakeMap(mapType)
		for k, item := range raw {
			keyVal := reflect.New(mapType.Key()).Elem()
			if err := b.setValue(keyVal, k); err != nil {
				return fmt.Errorf("设置映射键 %v 失败: %w", k, err)
			}
			elemVal := reflect.New(mapType.Elem()).Elem()
			if err := b.setValue(elemVal, item); err != nil {
				return fmt.Errorf("设置映射值 %v 失败: %w", k, err)
			}
			mapVal.SetMapIndex(keyVal, elemVal)
		}
		v.Set(mapVal)
	default:
		return fmt.Errorf("cannot convert %T to %s", value, v.Type())
	}
	return nil
}

// BindProperties 便捷函数：将配置绑定到目标结构体
func BindProperties(target any, env *environment.Environment, opts ...BindOption) error {
	binder := NewPropertyBinder(env)
	for _, opt := range opts {
		opt(binder)
	}
	return binder.Bind(target)
}

// BindOption 绑定选项
type BindOption func(*PropertyBinder)

// WithBindPrefix 设置绑定前缀
func WithBindPrefix(prefix string) BindOption {
	return func(b *PropertyBinder) {
		b.prefix = prefix
	}
}

// WithBindValidator 设置验证器
func WithBindValidator(v Validator) BindOption {
	return func(b *PropertyBinder) {
		b.validator = v
	}
}
