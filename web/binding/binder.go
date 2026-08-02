// Package binding 提供 HTTP 参数绑定功能。
package binding

import (
	"encoding/json"
	"fmt"
	"io"
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
	defer func() { _ = req.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(req.Body, 32<<20))
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	return json.Unmarshal(body, target)
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

		required := false
		for _, part := range parts[1:] {
			if part == "required" {
				required = true
				break
			}
		}

		formValues := values[fieldName]
		// 参数缺失；required 字段的显式空值（?name=）也视为缺失
		missing := len(formValues) == 0 || (required && len(formValues) == 1 && formValues[0] == "")
		if missing {
			if required {
				return &Error{
					Field:   fieldName,
					Message: "required field missing",
				}
			}
			continue
		}

		if err := b.setFieldValues(field, formValues); err != nil {
			return &Error{
				Field:   fieldName,
				Message: err.Error(),
			}
		}
	}

	return nil
}

// setFieldValues 绑定字段值：切片字段绑定全部值，标量字段使用第一个值。
func (b *Binder) setFieldValues(field reflect.Value, values []string) error {
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
			return err
		}
	}
	field.Set(slice)
	return nil
}

func (b *Binder) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, field.Type().Bits())
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

// Error 返回绑定错误的描述信息。
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
