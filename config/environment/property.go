package environment

import (
	"fmt"
	"reflect"
	"strconv"
)

// getRawProperty 从所有配置源中获取属性原始值（不解析占位符）
func (e *Environment) getRawProperty(key string) (any, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for i := len(e.sources) - 1; i >= 0; i-- {
		if val, ok := e.sources[i].GetProperty(key); ok {
			return val, true
		}
	}
	return nil, false
}

// GetProperty 从所有配置源中获取属性值
//
// 遍历配置源（从高优先级到低优先级），返回第一个匹配的值。
// 如果值是字符串且包含 ${...} 占位符，自动递归解析。
// 高优先级的配置源会覆盖低优先级的同名属性。
func (e *Environment) GetProperty(key string) (any, bool) {
	val, ok := e.getRawProperty(key)
	if !ok {
		return nil, false
	}
	if s, ok := val.(string); ok {
		return e.resolvePlaceholders(s, make(map[string]bool)), true
	}
	return val, true
}

// GetString 获取字符串类型的属性值.
//
// 参数:
//   - key: 属性键名
//   - defaultVal: 属性不存在时返回的默认值
//
// 返回:
//   - string: 属性值,不存在时返回 defaultVal
func (e *Environment) GetString(key, defaultVal string) string {
	if val, ok := e.GetProperty(key); ok {
		converted, err := globalTypeConverter.ConvertTo(val, reflect.TypeOf(""))
		if err == nil {
			return converted.String()
		}
	}
	return defaultVal
}

// GetInt 获取整数类型的属性值.
//
// 支持 int、float64 和字符串形式的整数值.
//
// 参数:
//   - key: 属性键名
//   - defaultVal: 属性不存在时返回的默认值
//
// 返回:
//   - int: 属性值,不存在时返回 defaultVal
func (e *Environment) GetInt(key string, defaultVal int) int {
	if val, ok := e.GetProperty(key); ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if n, err := strconv.Atoi(v); err == nil {
				return n
			}
		}
	}
	return defaultVal
}

// GetBool 获取布尔类型的属性值.
//
// 支持 bool 和字符串 "true"/"false" 形式.
//
// 参数:
//   - key: 属性键名
//   - defaultVal: 属性不存在时返回的默认值
//
// 返回:
//   - bool: 属性值,不存在时返回 defaultVal
func (e *Environment) GetBool(key string, defaultVal bool) bool {
	if val, ok := e.GetProperty(key); ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				return b
			}
		}
	}
	return defaultVal
}

// ContainsProperty 检查属性是否存在.
//
// 参数:
//   - key: 属性键名
//
// 返回:
//   - bool: 是否存在
func (e *Environment) ContainsProperty(key string) bool {
	_, ok := e.GetProperty(key)
	return ok
}

// GetRequiredProperty 获取必需属性，不存在时返回错误
func (e *Environment) GetRequiredProperty(key string) (any, error) {
	val, ok := e.GetProperty(key)
	if !ok {
		return nil, fmt.Errorf("required property not found: %s", key)
	}
	return val, nil
}

// GetFloat64 获取 float64 类型属性
func (e *Environment) GetFloat64(key string, defaultVal float64) float64 {
	if val, ok := e.GetProperty(key); ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return defaultVal
}

// IsPropertyEmpty 检查属性是否为空（不存在或空字符串）
func (e *Environment) IsPropertyEmpty(key string) bool {
	val, ok := e.GetProperty(key)
	if !ok {
		return true
	}
	s, ok := val.(string)
	if !ok {
		return false
	}
	return s == ""
}
