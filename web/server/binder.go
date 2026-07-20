// Package server 提供 HTTP 服务器和客户端实现。
//
// 本包包含 HTTP 请求的表单绑定功能，
// 仅使用 Go 标准库（无第三方依赖）。
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// FormBinder 将 HTTP 表单数据绑定到 Go 结构体。
type FormBinder struct {
	tagName string
}

// BinderOption 配置表单绑定器。
type BinderOption func(*FormBinder)

// WithTagName sets the struct tag name (default: "form").
func WithTagName(tagName string) BinderOption {
	return func(b *FormBinder) {
		b.tagName = tagName
	}
}

// NewFormBinder 创建一个新的表单绑定器。
func NewFormBinder(opts ...BinderOption) *FormBinder {
	b := &FormBinder{
		tagName: "form",
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
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

	return json.NewDecoder(req.Body).Decode(target)
}

func (b *FormBinder) bindForm(values map[string][]string, target any) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	typ := val.Type()
	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		tag := fieldType.Tag.Get(b.tagName)
		if tag == "" || tag == "-" {
			continue
		}

		// 解析标签选项
		parts := strings.Split(tag, ",")
		fieldName := parts[0]

		formValues := values[fieldName]
		if len(formValues) == 0 {
			// 检查字段是否必填
			required := false
			for _, part := range parts[1:] {
				if part == "required" {
					required = true
					break
				}
			}
			if required {
				return &BindingError{
					Field:   fieldName,
					Message: "required field missing",
				}
			}
			continue
		}

		if err := b.setFieldValue(field, formValues[0]); err != nil {
			return &BindingError{
				Field:   fieldName,
				Message: err.Error(),
			}
		}
	}

	return nil
}

func (b *FormBinder) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(v)
	case reflect.Slice:
		// 处理切片字段
		if field.Type().Elem().Kind() == reflect.String {
			field.Set(reflect.ValueOf(strings.Split(value, ",")))
		}
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
