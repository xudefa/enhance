package environment

import (
	"fmt"
	"reflect"
	"time"
)

// BindConfig 泛型配置绑定
//
// 将环境配置绑定到指定类型的结构体并返回。
// 支持 config/mapstructure/value/env 标签。
//
// 示例:
//
//	type ServerConfig struct {
//	    Port int    `config:"server.port" default:"8080"`
//	    Host string `config:"server.host" default:"localhost"`
//	}
//
//	cfg, err := environment.BindConfig[ServerConfig](env)
func BindConfig[T any](env *Environment) (T, error) {
	var zero T
	target := new(T)
	if err := env.Bind(target); err != nil {
		return zero, err
	}
	return *target, nil
}

// BindConfigPrefix 泛型配置绑定（带前缀）
//
// 将指定前缀的配置绑定到结构体。
//
// 示例:
//
//	cfg, err := environment.BindConfigPrefix[ServerConfig](env, "server")
func BindConfigPrefix[T any](env *Environment, prefix string) (T, error) {
	var zero T
	target := new(T)
	if err := env.BindPrefix(prefix, target); err != nil {
		return zero, err
	}
	return *target, nil
}

// BindConfigRequired 泛型配置绑定（必填）
//
// 绑定配置，如果任何带 required 标签的字段缺失则返回错误。
func BindConfigRequired[T any](env *Environment) (T, error) {
	var zero T
	target := new(T)
	if err := env.Bind(target); err != nil {
		return zero, err
	}
	if errs := validateRequired(reflect.ValueOf(target).Elem()); len(errs) > 0 {
		return zero, fmt.Errorf("required config missing: %v", errs)
	}
	return *target, nil
}

// validateRequired 检查带 required 标签的字段是否都有值
func validateRequired(v reflect.Value) []string {
	var missing []string
	t := v.Type()
	for i := range v.NumField() {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// 始终递归检查嵌套结构体（time.Time 等特殊类型除外）
		if fieldVal.Kind() == reflect.Struct && fieldVal.CanAddr() && !isTimeType(fieldVal.Type()) {
			missing = append(missing, validateRequired(fieldVal)...)
		}

		if field.Tag.Get("required") == "true" || field.Tag.Get("validate") == "required" {
			if isEmptyValue(fieldVal) {
				missing = append(missing, field.Name)
			}
		}
	}
	return missing
}

// isTimeType 检查是否为 time.Time 类型
func isTimeType(t reflect.Type) bool {
	timeType := reflect.TypeOf(time.Time{})
	return t == timeType
}

// isEmptyValue 检查值是否为空
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	case reflect.Struct:
		if !v.CanInterface() {
			return false
		}
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
	return false
}

// MustBindConfig 泛型配置绑定（失败时 panic）
//
// 适用于启动时配置，配置错误应该直接终止程序。
func MustBindConfig[T any](env *Environment) T {
	cfg, err := BindConfig[T](env)
	if err != nil {
		panic(fmt.Sprintf("config bind failed: %v", err))
	}
	return cfg
}

// MustBindConfigPrefix 泛型配置绑定带前缀（失败时 panic）
func MustBindConfigPrefix[T any](env *Environment, prefix string) T {
	cfg, err := BindConfigPrefix[T](env, prefix)
	if err != nil {
		panic(fmt.Sprintf("config bind failed for prefix %s: %v", prefix, err))
	}
	return cfg
}
