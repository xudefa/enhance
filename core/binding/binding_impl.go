package binding

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/xudefa/enhance/core"
)

const (
	injectTag = "inject"
	valueTag  = "value"
)

// Resolve 实现 ValueResolver 接口。
func (f ValueResolverFunc) Resolve(key string) (string, bool) {
	return f(key)
}

// defaultBinder 默认数据绑定器实现。
type defaultBinder struct {
	converter TypeConverter
}

// BindFields 将容器中的 Bean 注入到目标对象的字段中。
func (b *defaultBinder) BindFields(target any, container core.BeanGet) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return core.ErrInjectFailed
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		// 检查 inject 标签
		if tag, ok := field.Tag.Lookup(injectTag); ok {
			if err := b.injectField(val, field, tag, container); err != nil {
				return err
			}
		}
	}

	return nil
}

// injectField 注入单个字段。
func (b *defaultBinder) injectField(val reflect.Value, field reflect.StructField, tag string, container core.BeanGet) error {
	fieldType := field.Type
	fieldValue := val.FieldByName(field.Name)

	// 解析标签值（格式：beanName 或 空字符串表示按类型注入）
	beanName := strings.TrimSpace(tag)

	var bean any
	var err error

	if beanName != "" {
		bean, err = container.GetByTypeAndName(beanName, fieldType)
		if err != nil {
			return err
		}
		if bean == nil {
			return fmt.Errorf("no bean found with name '%s'", beanName)
		}

		fieldValue.Set(reflect.ValueOf(bean))
		return nil
	}

	// 按类型获取
	beans, err := container.Get(fieldType)
	if err != nil {
		return err
	}
	if len(beans) == 0 {
		return fmt.Errorf("no bean found for type %v", fieldType)
	}
	bean = beans[0]

	fieldValue.Set(reflect.ValueOf(bean))
	return nil
}

// BindValue 将配置值绑定到目标对象的字段中。
func (b *defaultBinder) BindValue(target any, resolver ValueResolver) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return core.ErrInjectFailed
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		// 检查 value 标签
		if tag := field.Tag.Get(valueTag); tag != "" {
			if err := b.bindValueField(val, field, tag, resolver); err != nil {
				return err
			}
		}
	}

	return nil
}

// bindValueField 绑定单个配置值字段。
func (b *defaultBinder) bindValueField(val reflect.Value, field reflect.StructField, tag string, resolver ValueResolver) error {
	key := strings.TrimSpace(tag)
	valueStr, ok := resolver.Resolve(key)
	if !ok {
		return fmt.Errorf("config value not found for key: %s", key)
	}

	fieldValue := val.FieldByName(field.Name)
	return b.setFieldValue(fieldValue, valueStr, field.Type)
}

// setFieldValue 设置字段值（支持类型转换）。
func (b *defaultBinder) setFieldValue(fieldValue reflect.Value, valueStr string, fieldType reflect.Type) error {
	// 使用类型转换器
	if b.converter != nil {
		converted, err := b.converter.Convert(valueStr, fieldType.String())
		if err == nil {
			// 转换器返回 (nil, nil) 时不能使用反射操作
			if converted == nil {
				return fmt.Errorf("type converter returned nil for field type %v", fieldType)
			}
			convertedVal := reflect.ValueOf(converted)
			if convertedVal.Type().AssignableTo(fieldType) {
				fieldValue.Set(convertedVal)
				return nil
			}
			return fmt.Errorf("type converter returned %v which is not assignable to field type %v", convertedVal.Type(), fieldType)
		}
	}

	// 特殊类型处理：time.Duration 的 Kind 是 Int64，
	// 必须先于 kind switch 判断，否则会落入 Int64 分支导致 "30s" 解析失败
	if fieldType == reflect.TypeOf(time.Duration(0)) {
		v, err := time.ParseDuration(valueStr)
		if err != nil {
			return err
		}
		fieldValue.Set(reflect.ValueOf(v))
		return nil
	}

	// 内置类型转换
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(valueStr)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return err
		}
		bits := fieldType.Bits()
		if bits > 0 && bits < 64 {
			min, max := int64(-1)<<(bits-1), (int64(1)<<(bits-1))-1
			if v < min || v > max {
				return fmt.Errorf("value %d overflows %s (range [%d, %d])", v, fieldType, min, max)
			}
		}
		fieldValue.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(valueStr, 10, 64)
		if err != nil {
			return err
		}
		bits := fieldType.Bits()
		if bits > 0 && bits < 64 {
			max := (uint64(1) << bits) - 1
			if v > max {
				return fmt.Errorf("value %d overflows %s (max %d)", v, fieldType, max)
			}
		}
		fieldValue.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(valueStr)
		if err != nil {
			return err
		}
		fieldValue.SetBool(v)
	default:
		return fmt.Errorf("unsupported field type: %v", fieldType)
	}

	return nil
}

// BindAll 执行完整的绑定流程（字段注入 + 配置绑定）。
func (b *defaultBinder) BindAll(target any, container core.BeanGet, resolver ValueResolver) error {
	if err := b.BindFields(target, container); err != nil {
		return err
	}
	if err := b.BindValue(target, resolver); err != nil {
		return err
	}
	return nil
}

// defaultTypeConverter 默认类型转换器实现。
type defaultTypeConverter struct{}

// Convert 将字符串值转换为目标类型。
func (c *defaultTypeConverter) Convert(value string, targetType string) (any, error) {
	switch targetType {
	case "string":
		return value, nil
	case "int":
		return strconv.Atoi(value)
	case "int64":
		return strconv.ParseInt(value, 10, 64)
	case "float64":
		return strconv.ParseFloat(value, 64)
	case "bool":
		return strconv.ParseBool(value)
	case "time.Duration":
		return time.ParseDuration(value)
	default:
		return nil, fmt.Errorf("unsupported target type: %s", targetType)
	}
}

// NewBinder 创建数据绑定器实例。
func NewBinder() Binder {
	return &defaultBinder{
		converter: NewTypeConverter(),
	}
}

// NewTypeConverter 创建类型转换器实例。
func NewTypeConverter() TypeConverter {
	return &defaultTypeConverter{}
}
