package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// DefaultBinder 默认绑定器，支持从JSON、表单和查询参数绑定数据
type DefaultBinder struct {
	Validator Validator // 验证器实例
}

// NewDefaultBinder 创建默认绑定器实例
func NewDefaultBinder(validator Validator) *DefaultBinder {
	if validator == nil {
		validator = NewTagValidator()
	}
	return &DefaultBinder{
		Validator: validator,
	}
}

// Bind 将请求参数绑定到目标对象，根据请求的内容类型选择适当的绑定方式
func (b *DefaultBinder) Bind(req *http.Request, obj any) error {
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("bind target must be a non-nil pointer")
	}

	rv = rv.Elem()

	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("bind target must be a struct pointer")
	}

	// 根据内容类型选择绑定方法
	contentType := req.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return b.bindJSON(req, obj)
	}

	if strings.Contains(contentType, "application/x-www-form-urlencoded") || strings.Contains(contentType, "multipart/form-data") {
		return b.bindForm(req, obj)
	}

	// 默认尝试从URL查询参数绑定
	return b.bindQuery(req.URL.Query(), obj)
}

// bindJSON 从JSON请求体绑定数据到对象
func (b *DefaultBinder) bindJSON(req *http.Request, obj any) error {
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields() // 严格模式，不允许未知字段
	err := decoder.Decode(obj)
	if err != nil {
		return err
	}

	// 验证绑定的对象
	return b.Validator.Validate(obj)
}

// bindForm 从表单数据绑定到对象
func (b *DefaultBinder) bindForm(req *http.Request, obj any) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	return b.bindQuery(req.Form, obj)
}

// bindQuery 从查询参数绑定数据到对象
func (b *DefaultBinder) bindQuery(values map[string][]string, obj any) error {
	return bindFieldsFromValues(values, obj, b.Validator)
}

// BindAndValidate 用于绑定和验证的便捷函数，自动创建绑定器并执行绑定和验证
func BindAndValidate(req *http.Request, obj any) error {
	binder := NewDefaultBinder(nil)
	return binder.Bind(req, obj)
}

// setFieldValue 设置字段值，支持多种基本类型转换
func setFieldValue(field reflect.Value, value string) error {
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}

	return nil
}

// bindFieldsFromValues 从键值对绑定字段值的通用方法
func bindFieldsFromValues(values map[string][]string, obj any, validator Validator) error {
	rv := reflect.ValueOf(obj).Elem()

	for i := range rv.NumField() {
		field := rv.Field(i)
		fieldType := rv.Type().Field(i)

		if !field.CanSet() {
			continue
		}

		fieldName := extractFieldName(fieldType)
		valuesList, exists := values[fieldName]
		if !exists || len(valuesList) == 0 {
			continue
		}

		if err := setFieldValue(field, valuesList[0]); err != nil {
			return fmt.Errorf("failed to bind field %s: %w", fieldName, err)
		}
	}

	return validator.Validate(obj)
}

// extractFieldName 从字段类型提取绑定名称
func extractFieldName(fieldType reflect.StructField) string {
	formTag := fieldType.Tag.Get("form")
	jsonTag := fieldType.Tag.Get("json")

	if formTag != "" && formTag != "-" {
		return formTag
	}

	if jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" && parts[0] != "-" {
			return parts[0]
		}
	}

	return fieldType.Name
}

// JSONBinder 专门用于JSON绑定的绑定器
type JSONBinder struct {
	Validator Validator
}

// NewJSONBinder 创建JSON绑定器实例
func NewJSONBinder(validator Validator) *JSONBinder {
	if validator == nil {
		validator = NewTagValidator()
	}
	return &JSONBinder{
		Validator: validator,
	}
}

// BindJSON 仅从JSON请求体绑定数据
func (j *JSONBinder) BindJSON(req *http.Request, obj any) error {
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields() // 严格模式，不允许未知字段
	err := decoder.Decode(obj)
	if err != nil {
		return err
	}

	// 验证绑定的对象
	return j.Validator.Validate(obj)
}

// FormBinder 专门用于表单绑定的绑定器
type FormBinder struct {
	Validator Validator
}

// NewFormBinder 创建表单绑定器实例
func NewFormBinder(validator Validator) *FormBinder {
	if validator == nil {
		validator = NewTagValidator()
	}
	return &FormBinder{
		Validator: validator,
	}
}

// BindForm 仅从表单数据绑定
func (f *FormBinder) BindForm(req *http.Request, obj any) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	return bindFieldsFromValues(req.Form, obj, f.Validator)
}

// QueryBinder 专门用于查询参数绑定的绑定器
type QueryBinder struct {
	Validator Validator
}

// NewQueryBinder 创建查询参数绑定器实例
func NewQueryBinder(validator Validator) *QueryBinder {
	if validator == nil {
		validator = NewTagValidator()
	}
	return &QueryBinder{
		Validator: validator,
	}
}

// BindQuery 仅从查询参数绑定
func (q *QueryBinder) BindQuery(req *http.Request, obj any) error {
	return q.bindQuery(req.URL.Query(), obj)
}

// bindQuery 内部方法，从查询参数绑定数据到对象
func (q *QueryBinder) bindQuery(values map[string][]string, obj any) error {
	return bindFieldsFromValues(values, obj, q.Validator)
}
