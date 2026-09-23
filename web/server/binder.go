// Package server 提供 HTTP 服务器和客户端实现。
//
// 本包包含 HTTP 请求的表单绑定功能，
// 仅使用 Go 标准库（无第三方依赖）。
package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// defaultMaxJSONBodySize 默认 JSON 请求体大小限制（32 MB，与 web/core/route_registry.go 保持一致）。
const defaultMaxJSONBodySize = 32 << 20

// FormBinder 将 HTTP 表单数据绑定到 Go 结构体。
type FormBinder struct {
	tagName     string
	maxBodySize int64
}

// BinderOption 配置表单绑定器。
type BinderOption func(*FormBinder)

// WithTagName 设置结构体标签名（默认 "form"）。
func WithTagName(tagName string) BinderOption {
	return func(b *FormBinder) {
		b.tagName = tagName
	}
}

// WithMaxBodySize 设置 JSON 请求体大小限制（默认 32 MB）。
func WithMaxBodySize(size int64) BinderOption {
	return func(b *FormBinder) {
		b.maxBodySize = size
	}
}

// NewFormBinder 创建一个新的表单绑定器。
func NewFormBinder(opts ...BinderOption) *FormBinder {
	binder := &FormBinder{
		tagName:     "form",
		maxBodySize: defaultMaxJSONBodySize,
	}
	for _, opt := range opts {
		opt(binder)
	}
	return binder
}

// Bind 将 HTTP 请求表单数据绑定到结构体。
func (b *FormBinder) Bind(req *http.Request, target any) error {
	if err := req.ParseForm(); err != nil {
		return fmt.Errorf("parse form failed: %w", err)
	}

	return b.bindForm(req.Form, target)
}

// BindQuery 将 HTTP 请求查询参数绑定到结构体。
func (b *FormBinder) BindQuery(req *http.Request, target any) error {
	return b.bindForm(req.URL.Query(), target)
}

// BindJSON 将 JSON 请求体绑定到结构体。
func (b *FormBinder) BindJSON(req *http.Request, target any) error {
	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		return ErrNotPointer
	}
	defer func() { _ = req.Body.Close() }()

	// 限制请求体大小，防止无界读取
	r := http.MaxBytesReader(nil, req.Body, b.maxBodySize)
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("failed to decode JSON body: %w", err)
	}

	// 拒绝第一个 JSON 值之后的尾部内容
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("request body contains trailing data after JSON value")
		}
		return fmt.Errorf("failed to decode JSON body: %w", err)
	}

	return nil
}

func (b *FormBinder) bindForm(values map[string][]string, target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	targetValue = targetValue.Elem()
	if targetValue.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	typ := targetValue.Type()
	for i := range targetValue.NumField() {
		field := targetValue.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		tag := fieldType.Tag.Get(b.tagName)
		fieldName, required := parseFormTag(tag)
		if fieldName == "" {
			continue
		}

		formValues := values[fieldName]
		// 参数缺失；required 字段的显式空值（?name=）也视为缺失
		missing := len(formValues) == 0 || (required && len(formValues) == 1 && formValues[0] == "")
		if missing {
			if required {
				return &BindingError{
					Field:   fieldName,
					Message: "required field missing",
				}
			}
			continue
		}

		if err := b.setFieldValues(field, formValues); err != nil {
			return &BindingError{
				Field:   fieldName,
				Message: err.Error(),
			}
		}
	}

	return nil
}

// parseFormTag 解析表单标签，返回字段名与是否必填；忽略空标签和 "-" 标签。
func parseFormTag(tag string) (fieldName string, required bool) {
	if tag == "" || tag == "-" {
		return "", false
	}
	parts := strings.Split(tag, ",")
	fieldName = parts[0]
	for _, part := range parts[1:] {
		if part == "required" {
			return fieldName, true
		}
	}
	return fieldName, false
}

// setFieldValues 绑定字段值：切片字段绑定全部值，标量字段使用第一个值。
func (b *FormBinder) setFieldValues(field reflect.Value, values []string) error {
	if field.Kind() != reflect.Slice {
		return b.setFieldValue(field, values[0])
	}

	// 支持重复参数（?tags=a&tags=b）以及逗号分隔（tags=a,b）
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, strings.Split(v, ",")...)
	}

	slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))
	for i, part := range parts {
		if err := b.setFieldValue(slice.Index(i), part); err != nil {
			return fmt.Errorf("set slice element: %w", err)
		}
	}
	field.Set(slice)
	return nil
}

func (b *FormBinder) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(value, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("parse int value: %w", err)
		}
		field.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(value, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("parse uint value: %w", err)
		}
		field.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("parse float value: %w", err)
		}
		field.SetFloat(parsed)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("parse bool value: %w", err)
		}
		field.SetBool(parsed)
	case reflect.Ptr:
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return b.setFieldValue(field.Elem(), value)
	}

	return nil
}

// BindingError 表示表单绑定错误。
type BindingError struct {
	Field   string
	Message string
}

// Error 返回 BindingError 的错误描述信息。
func (e *BindingError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("binding error: field %q: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("binding error: %s", e.Message)
}

// 表单绑定的哨兵错误。
var (
	ErrNotPointer = &BindingError{Message: "target must be a pointer"}
	ErrNotStruct  = &BindingError{Message: "target must be a struct"}
)
