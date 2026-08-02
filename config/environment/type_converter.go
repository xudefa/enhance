package environment

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// TypeConverter 类型转换器
//
// 统一的类型转换逻辑，消除重复代码。
type TypeConverter struct{}

// NewTypeConverter 创建类型转换器
func NewTypeConverter() *TypeConverter {
	return &TypeConverter{}
}

// ConvertTo 将值转换为目标类型
func (c *TypeConverter) ConvertTo(val any, targetType reflect.Type) (reflect.Value, error) {
	if val == nil {
		return reflect.Zero(targetType), nil
	}

	rv := reflect.ValueOf(val)

	// 如果类型已经匹配，直接返回
	if rv.Type().AssignableTo(targetType) {
		return rv, nil
	}

	// 特殊处理：数值类型到 string 的转换（避免 ASCII 转换）
	if targetType.Kind() == reflect.String && isNumeric(rv.Kind()) {
		return c.toString(val)
	}

	// 数值类型之间的转换做范围检查，避免静默溢出（如 300→int8、-1→uint、NaN→int）
	if isNumeric(rv.Kind()) && isNumeric(targetType.Kind()) {
		return c.convertNumeric(val, targetType)
	}

	// 尝试标准转换
	if rv.Type().ConvertibleTo(targetType) {
		return rv.Convert(targetType), nil
	}

	// 特殊处理
	return c.specialConvert(val, targetType)
}

// convertNumeric 数值类型之间的转换，带范围检查
func (c *TypeConverter) convertNumeric(val any, targetType reflect.Type) (reflect.Value, error) {
	rv := normalizeNumericValue(val)
	var (
		result reflect.Value
		err    error
	)
	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		result, err = c.toInt(rv, targetType)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result, err = c.toUint(rv, targetType)
	case reflect.Float32, reflect.Float64:
		result, err = c.toFloat(rv, targetType)
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, targetType)
	}
	if err != nil {
		return reflect.Value{}, err
	}
	return assignToType(result, targetType), nil
}

// normalizeNumericValue 将命名数值类型（如 type MyInt int、time.Duration）转为底层基础类型
func normalizeNumericValue(val any) any {
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint()
	case reflect.Float32, reflect.Float64:
		return rv.Float()
	}
	return val
}

// assignToType 将转换结果转为目标类型（支持命名类型如 type MyInt int、time.Duration）
func assignToType(v reflect.Value, targetType reflect.Type) reflect.Value {
	if !v.Type().AssignableTo(targetType) && v.Type().ConvertibleTo(targetType) {
		return v.Convert(targetType)
	}
	return v
}

// isNumeric 检查是否为数值类型
func isNumeric(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// specialConvert 处理特殊类型转换
func (c *TypeConverter) specialConvert(val any, targetType reflect.Type) (reflect.Value, error) {
	// 处理 time.Duration 类型
	if targetType == reflect.TypeOf(time.Duration(0)) {
		return c.toDuration(val)
	}

	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return c.toInt(val, targetType)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return c.toUint(val, targetType)
	case reflect.Float32, reflect.Float64:
		return c.toFloat(val, targetType)
	case reflect.Bool:
		return c.toBool(val)
	case reflect.String:
		return c.toString(val)
	case reflect.Slice:
		return c.toSlice(val, targetType)
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, targetType)
}

// toSlice 将值转换为切片类型，支持逗号分隔字符串（如 "a,b,c"）和已有切片
func (c *TypeConverter) toSlice(val any, targetType reflect.Type) (reflect.Value, error) {
	var items []string

	switch v := val.(type) {
	case string:
		items = strings.Split(v, ",")
	case []string:
		items = v
	default:
		rv := reflect.ValueOf(val)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			items = make([]string, rv.Len())
			for i := range rv.Len() {
				items[i] = fmt.Sprintf("%v", rv.Index(i).Interface())
			}
		} else {
			items = []string{fmt.Sprintf("%v", val)}
		}
	}

	elemType := targetType.Elem()
	result := reflect.MakeSlice(targetType, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		converted, err := c.ConvertTo(trimmed, elemType)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert slice element %q to %s: %w", trimmed, elemType, err)
		}
		result = reflect.Append(result, converted)
	}
	return result, nil
}

func (c *TypeConverter) toInt(val any, targetType reflect.Type) (reflect.Value, error) {
	var n int64
	switch v := val.(type) {
	case int:
		n = int64(v)
	case int8:
		n = int64(v)
	case int16:
		n = int64(v)
	case int32:
		n = int64(v)
	case int64:
		n = v
	case uint64:
		if v > math.MaxInt64 {
			return reflect.Value{}, fmt.Errorf("uint64 value %d overflows int64", v)
		}
		n = int64(v)
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return reflect.Value{}, fmt.Errorf("cannot convert %v to int", v)
		}
		if v > float64(math.MaxInt64) || v < float64(math.MinInt64) {
			return reflect.Value{}, fmt.Errorf("float64 value %v overflows int64", v)
		}
		n = int64(v)
	case float32:
		f := float64(v)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return reflect.Value{}, fmt.Errorf("cannot convert %v to int", v)
		}
		if f > float64(math.MaxInt64) || f < float64(math.MinInt64) {
			return reflect.Value{}, fmt.Errorf("float32 value %v overflows int64", v)
		}
		n = int64(v)
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert string %q to int: %w", v, err)
		}
		n = parsed
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to int", val)
	}

	// 根据目标类型做范围检查，避免静默溢出
	switch targetType.Kind() {
	case reflect.Int:
		if int64(int(n)) != n {
			return reflect.Value{}, fmt.Errorf("int64 value %d overflows int", n)
		}
		return reflect.ValueOf(int(n)), nil
	case reflect.Int8:
		if n > math.MaxInt8 || n < math.MinInt8 {
			return reflect.Value{}, fmt.Errorf("int64 value %d overflows int8", n)
		}
		return reflect.ValueOf(int8(n)), nil
	case reflect.Int16:
		if n > math.MaxInt16 || n < math.MinInt16 {
			return reflect.Value{}, fmt.Errorf("int64 value %d overflows int16", n)
		}
		return reflect.ValueOf(int16(n)), nil
	case reflect.Int32:
		if n > math.MaxInt32 || n < math.MinInt32 {
			return reflect.Value{}, fmt.Errorf("int64 value %d overflows int32", n)
		}
		return reflect.ValueOf(int32(n)), nil
	case reflect.Int64:
		return reflect.ValueOf(n), nil
	}

	return reflect.Value{}, fmt.Errorf("unsupported int type: %s", targetType)
}

func (c *TypeConverter) toUint(val any, targetType reflect.Type) (reflect.Value, error) {
	var n uint64
	switch v := val.(type) {
	case uint:
		n = uint64(v)
	case uint8:
		n = uint64(v)
	case uint16:
		n = uint64(v)
	case uint32:
		n = uint64(v)
	case uint64:
		n = v
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return reflect.Value{}, fmt.Errorf("cannot convert %v to uint", v)
		}
		if v < 0 || v > float64(math.MaxUint64) {
			return reflect.Value{}, fmt.Errorf("float64 value %v overflows uint64", v)
		}
		n = uint64(v)
	case float32:
		f := float64(v)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return reflect.Value{}, fmt.Errorf("cannot convert %v to uint", v)
		}
		if f < 0 || f > float64(math.MaxUint64) {
			return reflect.Value{}, fmt.Errorf("float32 value %v overflows uint64", v)
		}
		n = uint64(v)
	case int:
		if v < 0 {
			return reflect.Value{}, fmt.Errorf("int value %d overflows uint64", v)
		}
		n = uint64(v)
	case int8:
		if v < 0 {
			return reflect.Value{}, fmt.Errorf("int8 value %d overflows uint64", v)
		}
		n = uint64(v)
	case int16:
		if v < 0 {
			return reflect.Value{}, fmt.Errorf("int16 value %d overflows uint64", v)
		}
		n = uint64(v)
	case int32:
		if v < 0 {
			return reflect.Value{}, fmt.Errorf("int32 value %d overflows uint64", v)
		}
		n = uint64(v)
	case int64:
		if v < 0 {
			return reflect.Value{}, fmt.Errorf("int64 value %d overflows uint64", v)
		}
		n = uint64(v)
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert string %q to uint: %w", v, err)
		}
		n = parsed
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to uint", val)
	}

	// 根据目标类型做范围检查，避免静默溢出
	switch targetType.Kind() {
	case reflect.Uint:
		if uint64(uint(n)) != n {
			return reflect.Value{}, fmt.Errorf("uint64 value %d overflows uint", n)
		}
		return reflect.ValueOf(uint(n)), nil
	case reflect.Uint8:
		if n > math.MaxUint8 {
			return reflect.Value{}, fmt.Errorf("uint64 value %d overflows uint8", n)
		}
		return reflect.ValueOf(uint8(n)), nil
	case reflect.Uint16:
		if n > math.MaxUint16 {
			return reflect.Value{}, fmt.Errorf("uint64 value %d overflows uint16", n)
		}
		return reflect.ValueOf(uint16(n)), nil
	case reflect.Uint32:
		if n > math.MaxUint32 {
			return reflect.Value{}, fmt.Errorf("uint64 value %d overflows uint32", n)
		}
		return reflect.ValueOf(uint32(n)), nil
	case reflect.Uint64:
		return reflect.ValueOf(n), nil
	}

	return reflect.Value{}, fmt.Errorf("unsupported uint type: %s", targetType)
}

func (c *TypeConverter) toFloat(val any, targetType reflect.Type) (reflect.Value, error) {
	var f float64
	switch v := val.(type) {
	case float64:
		f = v
	case float32:
		f = float64(v)
	case int:
		f = float64(v)
	case int8:
		f = float64(v)
	case int16:
		f = float64(v)
	case int32:
		f = float64(v)
	case int64:
		f = float64(v)
	case uint:
		f = float64(v)
	case uint8:
		f = float64(v)
	case uint16:
		f = float64(v)
	case uint32:
		f = float64(v)
	case uint64:
		f = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert string %q to float: %w", v, err)
		}
		f = parsed
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to float", val)
	}

	if targetType.Kind() == reflect.Float32 {
		return reflect.ValueOf(float32(f)), nil
	}
	return reflect.ValueOf(f), nil
}

func (c *TypeConverter) toBool(val any) (reflect.Value, error) {
	switch v := val.(type) {
	case bool:
		return reflect.ValueOf(v), nil
	case string:
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert string %q to bool: %w", v, err)
		}
		return reflect.ValueOf(parsed), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to bool", val)
}

func (c *TypeConverter) toString(val any) (reflect.Value, error) {
	switch v := val.(type) {
	case string:
		return reflect.ValueOf(v), nil
	case int:
		return reflect.ValueOf(strconv.Itoa(v)), nil
	case int8:
		return reflect.ValueOf(strconv.FormatInt(int64(v), 10)), nil
	case int16:
		return reflect.ValueOf(strconv.FormatInt(int64(v), 10)), nil
	case int32:
		return reflect.ValueOf(strconv.FormatInt(int64(v), 10)), nil
	case int64:
		return reflect.ValueOf(strconv.FormatInt(v, 10)), nil
	case uint:
		return reflect.ValueOf(strconv.FormatUint(uint64(v), 10)), nil
	case uint8:
		return reflect.ValueOf(strconv.FormatUint(uint64(v), 10)), nil
	case uint16:
		return reflect.ValueOf(strconv.FormatUint(uint64(v), 10)), nil
	case uint32:
		return reflect.ValueOf(strconv.FormatUint(uint64(v), 10)), nil
	case uint64:
		return reflect.ValueOf(strconv.FormatUint(v, 10)), nil
	case float64:
		return reflect.ValueOf(strconv.FormatFloat(v, 'f', -1, 64)), nil
	case float32:
		return reflect.ValueOf(strconv.FormatFloat(float64(v), 'f', -1, 32)), nil
	case bool:
		return reflect.ValueOf(strconv.FormatBool(v)), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to string", val)
}

// toDuration 转换为 time.Duration 类型
func (c *TypeConverter) toDuration(val any) (reflect.Value, error) {
	switch v := val.(type) {
	case time.Duration:
		return reflect.ValueOf(v), nil
	case int64:
		return reflect.ValueOf(time.Duration(v)), nil
	case int:
		return reflect.ValueOf(time.Duration(v)), nil
	case uint64:
		return reflect.ValueOf(time.Duration(v)), nil
	case float64:
		return reflect.ValueOf(time.Duration(v)), nil
	case string:
		// 尝试解析为时间间隔字符串，如 "30s", "5m", "1h30m"
		duration, err := time.ParseDuration(v)
		if err != nil {
			// 如果解析失败，尝试解析为纳秒数
			if ns, parseErr := strconv.ParseInt(v, 10, 64); parseErr == nil {
				return reflect.ValueOf(time.Duration(ns)), nil
			}
			return reflect.Value{}, fmt.Errorf("cannot convert string %q to time.Duration: %w", v, err)
		}
		return reflect.ValueOf(duration), nil
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to time.Duration", val)
	}
}
