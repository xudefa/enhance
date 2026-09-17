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

	switch leftVal := left.(type) {
	case bool:
		rightVal, ok := right.(bool)
		return ok && leftVal == rightVal
	case string:
		rightVal, ok := right.(string)
		return ok && leftVal == rightVal
	case int:
		return equalsInt(leftVal, right)
	case int8:
		return equalsSignedInt(int64(leftVal), right)
	case int16:
		return equalsSignedInt(int64(leftVal), right)
	case int32:
		return equalsSignedInt(int64(leftVal), right)
	case int64:
		return equalsInt64(leftVal, right)
	case uint:
		return equalsUnsignedInt(uint64(leftVal), right)
	case uint8:
		return equalsUnsignedInt(uint64(leftVal), right)
	case uint16:
		return equalsUnsignedInt(uint64(leftVal), right)
	case uint32:
		return equalsUnsignedInt(uint64(leftVal), right)
	case uint64:
		return equalsUint64(leftVal, right)
	case float32:
		return equalsFloat(float64(leftVal), right)
	case float64:
		return equalsFloat(leftVal, right)
	default:
		// 对于无法直接比较的类型，使用 reflect.DeepEqual 作为兜底
		return reflect.DeepEqual(left, right)
	}
}

// equalsInt 比较 int 与任意数值类型。
func equalsInt(leftVal int, right any) bool {
	switch rightVal := right.(type) {
	case int:
		return leftVal == rightVal
	case int8:
		return leftVal == int(rightVal)
	case int16:
		return leftVal == int(rightVal)
	case int32:
		return leftVal == int(rightVal)
	case int64:
		return int64(leftVal) == rightVal
	case uint:
		return leftVal >= 0 && uint(leftVal) == rightVal
	case uint8:
		return leftVal >= 0 && uint8(leftVal) == rightVal
	case uint16:
		return leftVal >= 0 && uint16(leftVal) == rightVal
	case uint32:
		return leftVal >= 0 && uint32(leftVal) == rightVal
	case uint64:
		return leftVal >= 0 && uint64(leftVal) == rightVal
	case float32:
		return float64(leftVal) == float64(rightVal)
	case float64:
		return float64(leftVal) == rightVal
	}
	return false
}

// equalsInt64 比较 int64 与任意数值类型。
func equalsInt64(leftVal int64, right any) bool {
	switch rightVal := right.(type) {
	case int:
		return leftVal == int64(rightVal)
	case int8:
		return leftVal == int64(rightVal)
	case int16:
		return leftVal == int64(rightVal)
	case int32:
		return leftVal == int64(rightVal)
	case int64:
		return leftVal == rightVal
	case uint:
		return leftVal >= 0 && uint64(leftVal) == uint64(rightVal)
	case uint8:
		return leftVal >= 0 && uint64(leftVal) == uint64(rightVal)
	case uint16:
		return leftVal >= 0 && uint64(leftVal) == uint64(rightVal)
	case uint32:
		return leftVal >= 0 && uint64(leftVal) == uint64(rightVal)
	case uint64:
		return leftVal >= 0 && uint64(leftVal) == rightVal
	case float32:
		return float64(leftVal) == float64(rightVal)
	case float64:
		return float64(leftVal) == rightVal
	}
	return false
}

// equalsSignedInt 比较有符号整数与任意数值类型。
func equalsSignedInt(leftVal int64, right any) bool {
	switch rightVal := right.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return equals(leftVal, toInt64(rightVal))
	case float32, float64:
		return equals(float64(leftVal), toFloat64Value(rightVal))
	}
	return false
}

// equalsUnsignedInt 比较无符号整数与任意数值类型。
func equalsUnsignedInt(leftVal uint64, right any) bool {
	switch rightVal := right.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return equals(leftVal, toUint64(rightVal))
	case float32, float64:
		return equals(float64(leftVal), toFloat64Value(rightVal))
	}
	return false
}

// equalsUint64 比较 uint64 与任意数值类型，处理有符号数溢出情况。
func equalsUint64(leftVal uint64, right any) bool {
	switch rightVal := right.(type) {
	case int:
		return int64(leftVal) >= 0 && leftVal == uint64(rightVal)
	case int8:
		return leftVal == uint64(rightVal)
	case int16:
		return leftVal == uint64(rightVal)
	case int32:
		return leftVal == uint64(rightVal)
	case int64:
		return rightVal >= 0 && leftVal == uint64(rightVal)
	case uint:
		return leftVal == uint64(rightVal)
	case uint8:
		return leftVal == uint64(rightVal)
	case uint16:
		return leftVal == uint64(rightVal)
	case uint32:
		return leftVal == uint64(rightVal)
	case uint64:
		return leftVal == rightVal
	case float32:
		return float64(leftVal) == float64(rightVal)
	case float64:
		return float64(leftVal) == rightVal
	}
	return false
}

// equalsFloat 比较浮点数与任意数值类型。
func equalsFloat(leftVal float64, right any) bool {
	switch rightVal := right.(type) {
	case int:
		return leftVal == float64(rightVal)
	case int8:
		return leftVal == float64(rightVal)
	case int16:
		return leftVal == float64(rightVal)
	case int32:
		return leftVal == float64(rightVal)
	case int64:
		return leftVal == float64(rightVal)
	case uint:
		return leftVal == float64(rightVal)
	case uint8:
		return leftVal == float64(rightVal)
	case uint16:
		return leftVal == float64(rightVal)
	case uint32:
		return leftVal == float64(rightVal)
	case uint64:
		return leftVal == float64(rightVal)
	case float32:
		return leftVal == float64(rightVal)
	case float64:
		return leftVal == rightVal
	}
	return false
}

// toInt64 attempts to convert v to int64.
func toInt64(v any) int64 {
	switch typed := v.(type) {
	case int:
		return int64(typed)
	case int8:
		return int64(typed)
	case int16:
		return int64(typed)
	case int32:
		return int64(typed)
	case int64:
		return typed
	}
	return 0
}

// toUint64 attempts to convert v to uint64.
func toUint64(v any) uint64 {
	switch typed := v.(type) {
	case uint:
		return uint64(typed)
	case uint8:
		return uint64(typed)
	case uint16:
		return uint64(typed)
	case uint32:
		return uint64(typed)
	case uint64:
		return typed
	}
	return 0
}

// toFloat64Value attempts to convert v to float64.
func toFloat64Value(v any) float64 {
	switch typed := v.(type) {
	case int:
		return float64(typed)
	case int8:
		return float64(typed)
	case int16:
		return float64(typed)
	case int32:
		return float64(typed)
	case int64:
		return float64(typed)
	case uint:
		return float64(typed)
	case uint8:
		return float64(typed)
	case uint16:
		return float64(typed)
	case uint32:
		return float64(typed)
	case uint64:
		return float64(typed)
	case float32:
		return float64(typed)
	case float64:
		return typed
	}
	return 0
}
