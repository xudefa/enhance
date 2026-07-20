// Package binding 提供 HTTP 参数绑定功能。
package binding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// Binder 表单绑定器。
type Binder struct {
	tagName string
}

// Option 配置绑定器。
type Option func(*Binder)

// WithTagName 设置结构体标签名称(默认: "form")。
func WithTagName(tagName string) Option {
	return func(b *Binder) {
		b.tagName = tagName
	}
}

// NewBinder 创建一个新的表单绑定器。
func NewBinder(opts ...Option) *Binder {
	b := &Binder{
		tagName: "form",
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Bind 将 HTTP 请求表单数据绑定到结构体。
func (b *Binder) Bind(req *http.Request, target any) error {
	if err := req.ParseForm(); err != nil {
		return fmt.Errorf("parse form failed: %w", err)
	}

	return b.bindForm(req.Form, target)
}

// BindQuery 将 HTTP 请求查询参数绑定到结构体。
func (b *Binder) BindQuery(req *http.Request, target any) error {
	return b.bindForm(req.URL.Query(), target)
}

// BindJSON 将 JSON 请求体绑定到结构体。
func (b *Binder) BindJSON(req *http.Request, target any) error {
	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	return json.NewDecoder(req.Body).Decode(target)
}

func (b *Binder) bindForm(values map[string][]string, target any) error {
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

		parts := strings.Split(tag, ",")
		fieldName := parts[0]

		formValues := values[fieldName]
		if len(formValues) == 0 {
			required := false
			for _, part := range parts[1:] {
				if part == "required" {
					required = true
					break
				}
			}
			if required {
				return &Error{
					Field:   fieldName,
					Message: "required field missing",
				}
			}
			continue
		}

		if err := b.setFieldValue(field, formValues[0]); err != nil {
			return &Error{
				Field:   fieldName,
				Message: err.Error(),
			}
		}
	}

	return nil
}

func (b *Binder) setFieldValue(field reflect.Value, value string) error {
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

// Error 表示绑定错误。
type Error struct {
	Field   string
	Message string
}

func (e *Error) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("binding error: field %q: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("binding error: %s", e.Message)
}

// 哨兵错误。
var (
	ErrNotPointer = &Error{Message: "target must be a pointer"}
	ErrNotStruct  = &Error{Message: "target must be a struct"}
)
