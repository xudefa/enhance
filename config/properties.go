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
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	return b.bindStruct(v, b.prefix)
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
				return err
			}
			continue
		}

		// 获取 enhance 标签
		tag := field.Tag.Get("enhance")
		if tag == "" {
			// 没有标签，尝试递归处理嵌套结构体
			if fieldValue.Kind() == reflect.Struct {
				if err := b.bindStruct(fieldValue, prefix); err != nil {
					return err
				}
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
		if val, err := time.ParseDuration(defaultVal); err == nil {
			v.Set(reflect.ValueOf(val))
		}
		return nil
	}

	// 根据字段类型转换默认值
	switch v.Kind() {
	case reflect.String:
		v.SetString(defaultVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, err := strconv.ParseInt(defaultVal, 10, 64); err == nil {
			v.SetInt(val)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val, err := strconv.ParseUint(defaultVal, 10, 64); err == nil {
			v.SetUint(val)
		}
	case reflect.Float32, reflect.Float64:
		if val, err := strconv.ParseFloat(defaultVal, 64); err == nil {
			v.SetFloat(val)
		}
	case reflect.Bool:
		if val, err := strconv.ParseBool(defaultVal); err == nil {
			v.SetBool(val)
		}
	case reflect.Slice:
		// 逗号分隔的字符串转切片
		parts := strings.Split(defaultVal, ",")
		slice := reflect.MakeSlice(targetType, len(parts), len(parts))
		for i, part := range parts {
			elem := slice.Index(i)
			if err := b.setValue(elem, strings.TrimSpace(part)); err != nil {
				return err
			}
		}
		v.Set(slice)
	}
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

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch val := value.(type) {
		case int:
			v.SetInt(int64(val))
		case int64:
			v.SetInt(val)
		case float64:
			v.SetInt(int64(val))
		case string:
			if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
				v.SetInt(parsed)
			}
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch val := value.(type) {
		case uint:
			v.SetUint(uint64(val))
		case uint64:
			v.SetUint(val)
		case float64:
			v.SetUint(uint64(val))
		case string:
			if parsed, err := strconv.ParseUint(val, 10, 64); err == nil {
				v.SetUint(parsed)
			}
		}

	case reflect.Float32, reflect.Float64:
		switch val := value.(type) {
		case float64:
			v.SetFloat(val)
		case int:
			v.SetFloat(float64(val))
		case string:
			if parsed, err := strconv.ParseFloat(val, 64); err == nil {
				v.SetFloat(parsed)
			}
		}

	case reflect.Bool:
		switch val := value.(type) {
		case bool:
			v.SetBool(val)
		case string:
			if parsed, err := strconv.ParseBool(val); err == nil {
				v.SetBool(parsed)
			}
		}

	case reflect.Struct:
		if v.Type() == reflect.TypeOf(time.Duration(0)) {
			switch val := value.(type) {
			case time.Duration:
				v.Set(reflect.ValueOf(val))
			case string:
				if parsed, err := time.ParseDuration(val); err == nil {
					v.Set(reflect.ValueOf(parsed))
				}
			case int64:
				v.Set(reflect.ValueOf(time.Duration(val)))
			}
		}

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

// bindSlice 绑定切片字段
func (b *PropertyBinder) bindSlice(v reflect.Value, value any) error {
	switch val := value.(type) {
	case []any:
		slice := reflect.MakeSlice(v.Type(), len(val), len(val))
		for i, item := range val {
			elem := slice.Index(i)
			if err := b.setValue(elem, item); err != nil {
				return err
			}
		}
		v.Set(slice)
	case string:
		// 逗号分隔的字符串
		parts := strings.Split(val, ",")
		slice := reflect.MakeSlice(v.Type(), len(parts), len(parts))
		for i, part := range parts {
			elem := slice.Index(i)
			if err := b.setValue(elem, strings.TrimSpace(part)); err != nil {
				return err
			}
		}
		v.Set(slice)
	}
	return nil
}

// bindMap 绑定映射字段
func (b *PropertyBinder) bindMap(v reflect.Value, value any) error {
	switch val := value.(type) {
	case map[string]any:
		mapType := v.Type()
		mapVal := reflect.MakeMap(mapType)
		for k, item := range val {
			keyVal := reflect.New(mapType.Key()).Elem()
			if err := b.setValue(keyVal, k); err != nil {
				return err
			}
			elemVal := reflect.New(mapType.Elem()).Elem()
			if err := b.setValue(elemVal, item); err != nil {
				return err
			}
			mapVal.SetMapIndex(keyVal, elemVal)
		}
		v.Set(mapVal)
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
