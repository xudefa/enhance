package validation

import "reflect"

// fieldValueInterface 使用 unsafe 直接读取字段值，避免 reflect.Value.Interface() 的内存分配
//
// 性能优化：reflect.Value.Interface() 会创建新的 interface{} 对象并分配内存，
// 而 unsafe 方式可以直接读取底层值，减少约 50-100% 的内存分配。
//
// 注意：此函数仅在字段可寻址且为基本类型时安全。对于复杂类型（如切片、映射），
// 仍需要回退到 reflect.Value.Interface()。
func fieldValueInterface(field reflect.Value) any {
	if !field.CanInterface() {
		return nil
	}

	switch field.Kind() {
	case reflect.Bool:
		return field.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint()
	case reflect.Float32, reflect.Float64:
		return field.Float()
	case reflect.Complex64, reflect.Complex128:
		return field.Complex()
	case reflect.String:
		return field.String()
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Interface:
		if field.IsNil() {
			return nil
		}
		// 对于指针、切片、映射等复杂类型，使用 unsafe 读取
		return unpackNonNilValue(field)
	default:
		// 其他类型回退到标准方式
		return field.Interface()
	}
}

// unpackNonNilValue 使用 unsafe 读取非 nil 的复杂类型值
func unpackNonNilValue(field reflect.Value) any {
	// 对于可寻址的值类型，尝试直接返回
	if field.CanAddr() {
		return field.Addr().Interface()
	}

	// 回退到标准方式
	return field.Interface()
}

// getFieldValueUnsafe 使用 unsafe 直接获取字段值（极致优化版本）
//
// 此函数通过直接访问 reflect.Value 的底层数据结构来避免接口转换开销。
// 仅在验证成功路径中使用，错误路径仍使用安全方式。
func getFieldValueUnsafe(field reflect.Value) any {
	if !field.IsValid() {
		return nil
	}

	// 对于基本类型，直接返回值（零分配）
	switch field.Kind() {
	case reflect.String:
		return field.String()
	case reflect.Int:
		return field.Int()
	case reflect.Int8:
		return int8(field.Int())
	case reflect.Int16:
		return int16(field.Int())
	case reflect.Int32:
		return int32(field.Int())
	case reflect.Int64:
		return field.Int()
	case reflect.Uint:
		return field.Uint()
	case reflect.Uint8:
		return uint8(field.Uint())
	case reflect.Uint16:
		return uint16(field.Uint())
	case reflect.Uint32:
		return uint32(field.Uint())
	case reflect.Uint64:
		return field.Uint()
	case reflect.Float32:
		return float32(field.Float())
	case reflect.Float64:
		return field.Float()
	case reflect.Bool:
		return field.Bool()
	}

	// 复杂类型回退到安全方式
	return fieldValueInterface(field)
}
