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
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return core.ErrInjectFailed
	}

	value = value.Elem()
	typ := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		// 检查 inject 标签
		if tag, ok := field.Tag.Lookup(injectTag); ok {
			if err := b.injectField(value, field, tag, container); err != nil {
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
			return fmt.Errorf("注入字段 %s（名称 %s）失败: %w", field.Name, beanName, err)
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
		return fmt.Errorf("注入字段 %s（类型 %v）失败: %w", field.Name, fieldType, err)
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
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return core.ErrInjectFailed
	}

	value = value.Elem()
	typ := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		// 检查 value 标签
		if tag := field.Tag.Get(valueTag); tag != "" {
			if err := b.bindValueField(value, field, tag, resolver); err != nil {
				return fmt.Errorf("绑定配置值字段 %s 失败: %w", field.Name, err)
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
	if handled, err := b.applyCustomConverter(fieldValue, valueStr, fieldType); handled {
		return err
	}

	// 特殊类型处理：time.Duration 的 Kind 是 Int64，
	// 必须先于 kind switch 判断，否则会落入 Int64 分支导致 "30s" 解析失败
	if fieldType == reflect.TypeOf(time.Duration(0)) {
		duration, err := time.ParseDuration(valueStr)
		if err != nil {
			return fmt.Errorf("解析时长 %q 失败: %w", valueStr, err)
		}
		fieldValue.Set(reflect.ValueOf(duration))
		return nil
	}

	// 内置类型转换
	return setBuiltinTypeValue(fieldValue, valueStr, fieldType)
}

// applyCustomConverter 使用注册的类型转换器转换并设置字段值。
// 返回是否已由转换器处理；转换器未注册或转换失败返回 false，回退到内置类型转换。
func (b *defaultBinder) applyCustomConverter(fieldValue reflect.Value, valueStr string, fieldType reflect.Type) (bool, error) {
	if b.converter == nil {
		return false, nil
	}
	converted, err := b.converter.Convert(valueStr, fieldType.String())
	if err != nil {
		// 转换器失败时回退到内置类型转换
		return false, nil
	}
	// 转换器返回 (nil, nil) 时不能使用反射操作
	if converted == nil {
		return true, fmt.Errorf("type converter returned nil for field type %v", fieldType)
	}
	convertedVal := reflect.ValueOf(converted)
	if convertedVal.Type().AssignableTo(fieldType) {
		fieldValue.Set(convertedVal)
		return true, nil
	}
	return true, fmt.Errorf("type converter returned %v which is not assignable to field type %v", convertedVal.Type(), fieldType)
}

// setBuiltinTypeValue 按内置类型转换并设置字段值。
func setBuiltinTypeValue(fieldValue reflect.Value, valueStr string, fieldType reflect.Type) error {
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(valueStr)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsedInt, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return fmt.Errorf("解析整数 %q 失败: %w", valueStr, err)
		}
		bits := fieldType.Bits()
		if bits > 0 && bits < 64 {
			min, max := int64(-1)<<(bits-1), (int64(1)<<(bits-1))-1
			if parsedInt < min || parsedInt > max {
				return fmt.Errorf("value %d overflows %s (range [%d, %d])", parsedInt, fieldType, min, max)
			}
		}
		fieldValue.SetInt(parsedInt)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsedUint, err := strconv.ParseUint(valueStr, 10, 64)
		if err != nil {
			return fmt.Errorf("解析无符号整数 %q 失败: %w", valueStr, err)
		}
		bits := fieldType.Bits()
		if bits > 0 && bits < 64 {
			max := (uint64(1) << bits) - 1
			if parsedUint > max {
				return fmt.Errorf("value %d overflows %s (max %d)", parsedUint, fieldType, max)
			}
		}
		fieldValue.SetUint(parsedUint)
	case reflect.Float32, reflect.Float64:
		parsedFloat, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return fmt.Errorf("解析浮点数 %q 失败: %w", valueStr, err)
		}
		fieldValue.SetFloat(parsedFloat)
	case reflect.Bool:
		parsedBool, err := strconv.ParseBool(valueStr)
		if err != nil {
			return fmt.Errorf("解析布尔值 %q 失败: %w", valueStr, err)
		}
		fieldValue.SetBool(parsedBool)
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
