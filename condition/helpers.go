package condition

import "strconv"

// valAsString 将任意值转换为字符串
//
// 支持 string、bool、int、float64 类型，其他类型返回空字符串。
func valAsString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return ""
}

// listKeys 尝试列举 PropertySource 的键
func listKeys(src any) []string {
	if kl, ok := src.(keyLister); ok {
		return kl.Keys()
	}
	return nil
}
