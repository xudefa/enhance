package binding

import (
	"fmt"
	"net/http"

	"github.com/xudefa/enhance/web/server"
)

// Binder 表单绑定器。
// 委托给 server.FormBinder 实现，消除代码重复。
type Binder struct {
	inner *server.FormBinder
}

// Option 配置绑定器。
type Option func(*Binder)

// WithTagName 设置结构体标签名称(默认: "form")。
func WithTagName(tagName string) Option {
	return func(b *Binder) {
		b.inner = server.NewFormBinder(server.WithTagName(tagName))
	}
}

// NewBinder 创建一个新的表单绑定器。
func NewBinder(opts ...Option) *Binder {
	binder := &Binder{
		inner: server.NewFormBinder(),
	}
	for _, opt := range opts {
		opt(binder)
	}
	return binder
}

// Bind 将 HTTP 请求表单数据绑定到结构体。
func (b *Binder) Bind(req *http.Request, target any) error {
	return convertError(b.inner.Bind(req, target))
}

// BindQuery 将 HTTP 请求查询参数绑定到结构体。
func (b *Binder) BindQuery(req *http.Request, target any) error {
	return convertError(b.inner.BindQuery(req, target))
}

// BindJSON 将 JSON 请求体绑定到结构体。
func (b *Binder) BindJSON(req *http.Request, target any) error {
	return convertError(b.inner.BindJSON(req, target))
}

// convertError 将 server.BindingError 转换为 binding.Error，保持 API 兼容。
func convertError(err error) error {
	if err == nil {
		return nil
	}
	if serverErr, ok := err.(*server.BindingError); ok {
		return &Error{Field: serverErr.Field, Message: serverErr.Message}
	}
	return err
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

var (
	ErrNotPointer = &Error{Message: "target must be a pointer"}
	ErrNotStruct  = &Error{Message: "target must be a struct"}
)
