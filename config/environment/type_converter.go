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
		convertedVal reflect.Value
		err          error
	)
	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		convertedVal, err = c.toInt(rv, targetType)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		convertedVal, err = c.toUint(rv, targetType)
	case reflect.Float32, reflect.Float64:
		convertedVal, err = c.toFloat(rv, targetType)
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, targetType)
	}
	if err != nil {
		return reflect.Value{}, fmt.Errorf("转换数值为 %s 失败: %w", targetType, err)
	}
	return assignToType(convertedVal, targetType), nil
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

	switch raw := val.(type) {
	case string:
		items = strings.Split(raw, ",")
	case []string:
		items = raw
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
	resultSlice := reflect.MakeSlice(targetType, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		converted, err := c.ConvertTo(trimmed, elemType)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert slice element %q to %s: %w", trimmed, elemType, err)
		}
		resultSlice = reflect.Append(resultSlice, converted)
	}
	return resultSlice, nil
}

func (c *TypeConverter) toInt(val any, targetType reflect.Type) (reflect.Value, error) {
	// 从源值解析 int64
	intVal, err := parseIntSource(val)
	if err != nil {
		return reflect.Value{}, err
	}

	// 根据目标类型做范围检查，避免静默溢出
	return convertToIntType(intVal, targetType)
}

// parseIntSource 将源值解析为 int64。
func parseIntSource(val any) (int64, error) {
	switch raw := val.(type) {
	case int:
		return int64(raw), nil
	case int8:
		return int64(raw), nil
	case int16:
		return int64(raw), nil
	case int32:
		return int64(raw), nil
	case int64:
		return raw, nil
	case uint64:
		if raw > math.MaxInt64 {
			return 0, fmt.Errorf("uint64 value %d overflows int64", raw)
		}
		return int64(raw), nil
	case float64:
		if math.IsNaN(raw) || math.IsInf(raw, 0) {
			return 0, fmt.Errorf("cannot convert %v to int", raw)
		}
		if raw > float64(math.MaxInt64) || raw < float64(math.MinInt64) {
			return 0, fmt.Errorf("float64 value %v overflows int64", raw)
		}
		return int64(raw), nil
	case float32:
		floatVal := float64(raw)
		if math.IsNaN(floatVal) || math.IsInf(floatVal, 0) {
			return 0, fmt.Errorf("cannot convert %v to int", raw)
		}
		if floatVal > float64(math.MaxInt64) || floatVal < float64(math.MinInt64) {
			return 0, fmt.Errorf("float32 value %v overflows int64", raw)
		}
		return int64(raw), nil
	case string:
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string %q to int: %w", raw, err)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", val)
	}
}

// convertToIntType 将 int64 按目标类型范围检查并转换为目标类型。
func convertToIntType(intVal int64, targetType reflect.Type) (reflect.Value, error) {
	switch targetType.Kind() {
	case reflect.Int:
		if int64(int(intVal)) != intVal {
			return reflect.Value{}, fmt.Errorf("int64 value %d overflows int", intVal)
		}
		return reflect.ValueOf(int(intVal)), nil
	case reflect.Int8:
		if intVal > math.MaxInt8 || intVal < math.MinInt8 {
			return reflect.Value{}, fmt.Errorf("int64 value %d overflows int8", intVal)
		}
		return reflect.ValueOf(int8(intVal)), nil
	case reflect.Int16:
		if intVal > math.MaxInt16 || intVal < math.MinInt16 {
			return reflect.Value{}, fmt.Errorf("int64 value %d overflows int16", intVal)
		}
		return reflect.ValueOf(int16(intVal)), nil
	case reflect.Int32:
		if intVal > math.MaxInt32 || intVal < math.MinInt32 {
			return reflect.Value{}, fmt.Errorf("int64 value %d overflows int32", intVal)
		}
		return reflect.ValueOf(int32(intVal)), nil
	case reflect.Int64:
		return reflect.ValueOf(intVal), nil
	}

	return reflect.Value{}, fmt.Errorf("unsupported int type: %s", targetType)
}

func (c *TypeConverter) toUint(val any, targetType reflect.Type) (reflect.Value, error) {
	// 从源值解析 uint64
	uintVal, err := parseUintSource(val)
	if err != nil {
		return reflect.Value{}, err
	}

	// 根据目标类型做范围检查，避免静默溢出
	return convertToUintType(uintVal, targetType)
}

// parseUintSource 将源值解析为 uint64。
func parseUintSource(val any) (uint64, error) {
	switch raw := val.(type) {
	case uint:
		return uint64(raw), nil
	case uint8:
		return uint64(raw), nil
	case uint16:
		return uint64(raw), nil
	case uint32:
		return uint64(raw), nil
	case uint64:
		return raw, nil
	case float64:
		return parseFloatToUint(raw, "float64", raw)
	case float32:
		return parseFloatToUint(float64(raw), "float32", raw)
	case int:
		return parseSignedToUint(int64(raw), "int")
	case int8:
		return parseSignedToUint(int64(raw), "int8")
	case int16:
		return parseSignedToUint(int64(raw), "int16")
	case int32:
		return parseSignedToUint(int64(raw), "int32")
	case int64:
		return parseSignedToUint(raw, "int64")
	case string:
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string %q to uint: %w", raw, err)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to uint", val)
	}
}

// parseFloatToUint 将浮点数转换为 uint64，超出范围或非有限值时返回错误。
func parseFloatToUint(floatVal float64, typeName string, display any) (uint64, error) {
	if math.IsNaN(floatVal) || math.IsInf(floatVal, 0) {
		return 0, fmt.Errorf("cannot convert %v to uint", display)
	}
	if floatVal < 0 || floatVal > float64(math.MaxUint64) {
		return 0, fmt.Errorf("%s value %v overflows uint64", typeName, display)
	}
	return uint64(floatVal), nil
}

// parseSignedToUint 将有符号整数转换为 uint64，负值时返回错误。
func parseSignedToUint(intVal int64, typeName string) (uint64, error) {
	if intVal < 0 {
		return 0, fmt.Errorf("%s value %d overflows uint64", typeName, intVal)
	}
	return uint64(intVal), nil
}

// convertToUintType 将 uint64 按目标类型范围检查并转换为目标类型。
func convertToUintType(uintVal uint64, targetType reflect.Type) (reflect.Value, error) {
	switch targetType.Kind() {
	case reflect.Uint:
		if uint64(uint(uintVal)) != uintVal {
			return reflect.Value{}, fmt.Errorf("uint64 value %d overflows uint", uintVal)
		}
		return reflect.ValueOf(uint(uintVal)), nil
	case reflect.Uint8:
		if uintVal > math.MaxUint8 {
			return reflect.Value{}, fmt.Errorf("uint64 value %d overflows uint8", uintVal)
		}
		return reflect.ValueOf(uint8(uintVal)), nil
	case reflect.Uint16:
		if uintVal > math.MaxUint16 {
			return reflect.Value{}, fmt.Errorf("uint64 value %d overflows uint16", uintVal)
		}
		return reflect.ValueOf(uint16(uintVal)), nil
	case reflect.Uint32:
		if uintVal > math.MaxUint32 {
			return reflect.Value{}, fmt.Errorf("uint64 value %d overflows uint32", uintVal)
		}
		return reflect.ValueOf(uint32(uintVal)), nil
	case reflect.Uint64:
		return reflect.ValueOf(uintVal), nil
	}

	return reflect.Value{}, fmt.Errorf("unsupported uint type: %s", targetType)
}

func (c *TypeConverter) toFloat(val any, targetType reflect.Type) (reflect.Value, error) {
	var floatVal float64
	switch raw := val.(type) {
	case float64:
		floatVal = raw
	case float32:
		floatVal = float64(raw)
	case int:
		floatVal = float64(raw)
	case int8:
		floatVal = float64(raw)
	case int16:
		floatVal = float64(raw)
	case int32:
		floatVal = float64(raw)
	case int64:
		floatVal = float64(raw)
	case uint:
		floatVal = float64(raw)
	case uint8:
		floatVal = float64(raw)
	case uint16:
		floatVal = float64(raw)
	case uint32:
		floatVal = float64(raw)
	case uint64:
		floatVal = float64(raw)
	case string:
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert string %q to float: %w", raw, err)
		}
		floatVal = parsed
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to float", val)
	}

	if targetType.Kind() == reflect.Float32 {
		return reflect.ValueOf(float32(floatVal)), nil
	}
	return reflect.ValueOf(floatVal), nil
}

func (c *TypeConverter) toBool(val any) (reflect.Value, error) {
	switch raw := val.(type) {
	case bool:
		return reflect.ValueOf(raw), nil
	case string:
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert string %q to bool: %w", raw, err)
		}
		return reflect.ValueOf(parsed), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to bool", val)
}

func (c *TypeConverter) toString(val any) (reflect.Value, error) {
	switch raw := val.(type) {
	case string:
		return reflect.ValueOf(raw), nil
	case int:
		return reflect.ValueOf(strconv.Itoa(raw)), nil
	case int8:
		return reflect.ValueOf(strconv.FormatInt(int64(raw), 10)), nil
	case int16:
		return reflect.ValueOf(strconv.FormatInt(int64(raw), 10)), nil
	case int32:
		return reflect.ValueOf(strconv.FormatInt(int64(raw), 10)), nil
	case int64:
		return reflect.ValueOf(strconv.FormatInt(raw, 10)), nil
	case uint:
		return reflect.ValueOf(strconv.FormatUint(uint64(raw), 10)), nil
	case uint8:
		return reflect.ValueOf(strconv.FormatUint(uint64(raw), 10)), nil
	case uint16:
		return reflect.ValueOf(strconv.FormatUint(uint64(raw), 10)), nil
	case uint32:
		return reflect.ValueOf(strconv.FormatUint(uint64(raw), 10)), nil
	case uint64:
		return reflect.ValueOf(strconv.FormatUint(raw, 10)), nil
	case float64:
		return reflect.ValueOf(strconv.FormatFloat(raw, 'f', -1, 64)), nil
	case float32:
		return reflect.ValueOf(strconv.FormatFloat(float64(raw), 'f', -1, 32)), nil
	case bool:
		return reflect.ValueOf(strconv.FormatBool(raw)), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to string", val)
}

// toDuration 转换为 time.Duration 类型
func (c *TypeConverter) toDuration(val any) (reflect.Value, error) {
	switch raw := val.(type) {
	case time.Duration:
		return reflect.ValueOf(raw), nil
	case int64:
		return reflect.ValueOf(time.Duration(raw)), nil
	case int:
		return reflect.ValueOf(time.Duration(raw)), nil
	case uint64:
		return reflect.ValueOf(time.Duration(raw)), nil
	case float64:
		return reflect.ValueOf(time.Duration(raw)), nil
	case string:
		duration, err := time.ParseDuration(raw)
		if err != nil {
			if ns, parseErr := strconv.ParseInt(raw, 10, 64); parseErr == nil {
				return reflect.ValueOf(time.Duration(ns)), nil
			}
			return reflect.Value{}, fmt.Errorf("cannot convert string %q to time.Duration: %w", raw, err)
		}
		return reflect.ValueOf(duration), nil
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to time.Duration", val)
	}
}
