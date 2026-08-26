package spel

import "reflect"

// equals 比较两个值是否相等（替代 reflect.DeepEqual）。
// 使用类型安全的比较，避免 DeepEqual 在 nil/empty slice、不同数值类型上的不一致行为。
func equals(left, right any) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}

	switch l := left.(type) {
	case bool:
		r, ok := right.(bool)
		return ok && l == r
	case string:
		r, ok := right.(string)
		return ok && l == r
	case int:
		switch r := right.(type) {
		case int:
			return l == r
		case int8:
			return l == int(r)
		case int16:
			return l == int(r)
		case int32:
			return l == int(r)
		case int64:
			return int64(l) == r
		case uint:
			return l >= 0 && uint(l) == r
		case uint8:
			return l >= 0 && uint8(l) == r
		case uint16:
			return l >= 0 && uint16(l) == r
		case uint32:
			return l >= 0 && uint32(l) == r
		case uint64:
			return l >= 0 && uint64(l) == r
		case float32:
			return float64(l) == float64(r)
		case float64:
			return float64(l) == r
		}
	case int8:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return equals(int64(l), toInt64(r))
		case float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case int16:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return equals(int64(l), toInt64(r))
		case float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case int32:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return equals(int64(l), toInt64(r))
		case float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case int64:
		switch r := right.(type) {
		case int:
			return l == int64(r)
		case int8:
			return l == int64(r)
		case int16:
			return l == int64(r)
		case int32:
			return l == int64(r)
		case int64:
			return l == r
		case uint:
			return l >= 0 && uint64(l) == uint64(r)
		case uint8:
			return l >= 0 && uint64(l) == uint64(r)
		case uint16:
			return l >= 0 && uint64(l) == uint64(r)
		case uint32:
			return l >= 0 && uint64(l) == uint64(r)
		case uint64:
			return l >= 0 && uint64(l) == r
		case float32:
			return float64(l) == float64(r)
		case float64:
			return float64(l) == r
		}
	case uint:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return equals(uint64(l), toUint64(r))
		case float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case uint8:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return equals(uint64(l), toUint64(r))
		case float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case uint16:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return equals(uint64(l), toUint64(r))
		case float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case uint32:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return equals(uint64(l), toUint64(r))
		case float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case uint64:
		switch r := right.(type) {
		case int:
			return int64(l) >= 0 && l == uint64(r)
		case int8:
			return l == uint64(r)
		case int16:
			return l == uint64(r)
		case int32:
			return l == uint64(r)
		case int64:
			return r >= 0 && l == uint64(r)
		case uint:
			return l == uint64(r)
		case uint8:
			return l == uint64(r)
		case uint16:
			return l == uint64(r)
		case uint32:
			return l == uint64(r)
		case uint64:
			return l == r
		case float32:
			return float64(l) == float64(r)
		case float64:
			return float64(l) == r
		}
	case float32:
		switch r := right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return equals(float64(l), toFloat64Value(r))
		}
	case float64:
		switch r := right.(type) {
		case int:
			return l == float64(r)
		case int8:
			return l == float64(r)
		case int16:
			return l == float64(r)
		case int32:
			return l == float64(r)
		case int64:
			return l == float64(r)
		case uint:
			return l == float64(r)
		case uint8:
			return l == float64(r)
		case uint16:
			return l == float64(r)
		case uint32:
			return l == float64(r)
		case uint64:
			return l == float64(r)
		case float32:
			return l == float64(r)
		case float64:
			return l == r
		}
	default:
		// 对于无法直接比较的类型，使用 reflect.DeepEqual 作为兜底
		return reflect.DeepEqual(left, right)
	}
	return false
}

// toInt64 attempts to convert v to int64.
func toInt64(v any) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int8:
		return int64(val)
	case int16:
		return int64(val)
	case int32:
		return int64(val)
	case int64:
		return val
	}
	return 0
}

// toUint64 attempts to convert v to uint64.
func toUint64(v any) uint64 {
	switch val := v.(type) {
	case uint:
		return uint64(val)
	case uint8:
		return uint64(val)
	case uint16:
		return uint64(val)
	case uint32:
		return uint64(val)
	case uint64:
		return val
	}
	return 0
}

// toFloat64Value attempts to convert v to float64.
func toFloat64Value(v any) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	}
	return 0
}
