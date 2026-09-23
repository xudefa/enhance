package validation

import (
	"reflect"
	"strconv"
	"strings"
)

// isRequiredValid 验证字段是否必需（非零值）
func (v *TagValidator) isRequiredValid(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.String:
		return field.String() != ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return field.Float() != 0
	case reflect.Bool:
		return field.Bool()
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return !field.IsNil()
	default:
		return field.IsValid() && field.Interface() != reflect.Zero(field.Type()).Interface()
	}
}

// isMinValid 验证最小值
func (v *TagValidator) isMinValid(field reflect.Value, minStr string) bool {
	min, err := strconv.Atoi(minStr)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return len([]rune(field.String())) >= min
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() >= int64(min)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() >= uint64(min)
	case reflect.Float32, reflect.Float64:
		return field.Float() >= float64(min)
	default:
		return false
	}
}

// isMaxValid 验证最大值
func (v *TagValidator) isMaxValid(field reflect.Value, maxStr string) bool {
	max, err := strconv.Atoi(maxStr)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return len([]rune(field.String())) <= max
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() <= int64(max)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() <= uint64(max)
	case reflect.Float32, reflect.Float64:
		return field.Float() <= float64(max)
	default:
		return false
	}
}

// isLenValid 验证长度
func (v *TagValidator) isLenValid(field reflect.Value, lenStr string) bool {
	length, err := strconv.Atoi(lenStr)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return len([]rune(field.String())) == length
	case reflect.Array, reflect.Slice:
		return field.Len() == length
	default:
		return false
	}
}

// isEmailValid 验证邮箱格式
func (v *TagValidator) isEmailValid(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	email := field.String()
	emailRegex := compileRegex(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isRegexpValid 验证正则表达式匹配
func (v *TagValidator) isRegexpValid(field reflect.Value, pattern string) bool {
	if field.Kind() != reflect.String {
		return false
	}
	re := compileRegex(pattern)
	if re == nil {
		return false
	}
	return re.MatchString(field.String())
}

// isGtValid 验证大于指定值
func (v *TagValidator) isGtValid(field reflect.Value, valueStr string) bool {
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return len([]rune(field.String())) > value
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() > int64(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() > uint64(value)
	case reflect.Float32, reflect.Float64:
		return field.Float() > float64(value)
	default:
		return false
	}
}

// isGteValid 验证大于等于指定值
func (v *TagValidator) isGteValid(field reflect.Value, valueStr string) bool {
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return len([]rune(field.String())) >= value
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() >= int64(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() >= uint64(value)
	case reflect.Float32, reflect.Float64:
		return field.Float() >= float64(value)
	default:
		return false
	}
}

// isLtValid 验证小于指定值
func (v *TagValidator) isLtValid(field reflect.Value, valueStr string) bool {
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return len([]rune(field.String())) < value
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() < int64(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() < uint64(value)
	case reflect.Float32, reflect.Float64:
		return field.Float() < float64(value)
	default:
		return false
	}
}

// isLteValid 验证小于等于指定值
func (v *TagValidator) isLteValid(field reflect.Value, valueStr string) bool {
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return len([]rune(field.String())) <= value
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() <= int64(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() <= uint64(value)
	case reflect.Float32, reflect.Float64:
		return field.Float() <= float64(value)
	default:
		return false
	}
}

// isURLValid 验证URL格式
func (v *TagValidator) isURLValid(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	url := field.String()
	urlRegex := compileRegex(`^https?:\/\/(?:[-\w.])+(?:\:[0-9]+)?(?:\/(?:[\w\/_.])*(?:\?(?:[\w&=%.])*)?(?:\#(?:[\w.])*)?)?$`)
	return urlRegex.MatchString(url)
}

// isIPValid 验证IP地址格式
func (v *TagValidator) isIPValid(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	ip := field.String()
	ipRegex := compileRegex(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	return ipRegex.MatchString(ip)
}

// isOneOfValid 验证值是否在指定的选项中
func (v *TagValidator) isOneOfValid(field reflect.Value, optionsStr string) bool {
	options := strings.Split(optionsStr, " ")

	for i, opt := range options {
		options[i] = strings.TrimSpace(opt)
	}

	optionSet := make(map[string]struct{}, len(options))
	for _, opt := range options {
		optionSet[opt] = struct{}{}
	}

	switch field.Kind() {
	case reflect.String:
		_, ok := optionSet[field.String()]
		return ok
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		key := strconv.FormatInt(field.Int(), 10)
		_, ok := optionSet[key]
		return ok
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		key := strconv.FormatUint(field.Uint(), 10)
		_, ok := optionSet[key]
		return ok
	case reflect.Float32, reflect.Float64:
		key := strconv.FormatFloat(field.Float(), 'f', -1, 64)
		_, ok := optionSet[key]
		return ok
	}

	return false
}
