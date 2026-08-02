// Package stdlib 提供基于标准库 net/http 的引擎实现。
package stdlib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/xudefa/enhance/web/core"
)

// Context 标准库 HTTP 上下文实现。
type Context struct {
	writer  http.ResponseWriter
	request *http.Request
	params  map[string]string
	index   int
	mw      []core.MiddlewareFunc
	handler core.HandlerFunc
	aborted bool
	code    int
}

// NewContext 创建新的 HTTP 上下文。
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		writer:  w,
		request: r,
		params:  make(map[string]string),
		index:   -1,
	}
}

// WithParams 设置路径参数。
func (c *Context) WithParams(params map[string]string) *Context {
	c.params = params
	return c
}

// WithMiddleware 设置中间件链。
func (c *Context) WithMiddleware(mw []core.MiddlewareFunc, handler core.HandlerFunc) *Context {
	c.mw = mw
	c.handler = handler
	return c
}

// RequestMethod 返回请求方法。
func (c *Context) RequestMethod() string {
	return c.request.Method
}

// RequestURI 返回请求 URI。
func (c *Context) RequestURI() string {
	return c.request.RequestURI
}

// PathParam 获取路径参数。
func (c *Context) PathParam(name string) string {
	return c.params[name]
}

// Query 获取查询参数。
func (c *Context) Query(name string) string {
	return c.request.URL.Query().Get(name)
}

// QueryDefault 获取查询参数（带默认值）。
func (c *Context) QueryDefault(name, defaultVal string) string {
	val := c.request.URL.Query().Get(name)
	if val == "" {
		return defaultVal
	}
	return val
}

// Header 获取请求头。
func (c *Context) Header(key string) string {
	return c.request.Header.Get(key)
}

// BindJSON 解析 JSON 请求体。
func (c *Context) BindJSON(target any) error {
	defer func() { _ = c.request.Body.Close() }()
	c.request.Body = http.MaxBytesReader(c.writer, c.request.Body, 32<<20)
	body, err := io.ReadAll(c.request.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}
	return nil
}

// SetStatusCode 设置响应状态码。
func (c *Context) SetStatusCode(code int) {
	c.code = code
	c.writer.WriteHeader(code)
}

// SetHeader 设置响应头。
func (c *Context) SetHeader(key, value string) {
	c.writer.Header().Set(key, value)
}

// JSON 返回 JSON 响应。
func (c *Context) JSON(code int, data any) error {
	c.writer.Header().Set("Content-Type", "application/json")
	c.writer.WriteHeader(code)
	return json.NewEncoder(c.writer).Encode(data)
}

// String 返回字符串响应。
func (c *Context) String(code int, format string, args ...any) {
	c.writer.Header().Set("Content-Type", "text/plain")
	c.writer.WriteHeader(code)
	_, _ = fmt.Fprintf(c.writer, format, args...)
}

// AbortWithStatus 中止请求。
func (c *Context) AbortWithStatus(code int) {
	c.aborted = true
	c.writer.WriteHeader(code)
}

// AbortWithStatusJSON 中止请求并返回 JSON。
func (c *Context) AbortWithStatusJSON(code int, body any) {
	c.aborted = true
	c.writer.Header().Set("Content-Type", "application/json")
	c.writer.WriteHeader(code)
	_ = json.NewEncoder(c.writer).Encode(body)
}

// Next 执行下一个中间件。
func (c *Context) Next() {
	c.index++
	for c.index < len(c.mw) {
		if c.aborted {
			return
		}
		c.mw[c.index](c)
		c.index++
	}

	if c.index == len(c.mw) && !c.aborted && c.handler != nil {
		c.handler(c)
	}
}

// IsAborted 判断是否已中止。
func (c *Context) IsAborted() bool {
	return c.aborted
}

// Context 获取请求上下文。
func (c *Context) Context() context.Context {
	return c.request.Context()
}

// SetContext 设置请求上下文。
func (c *Context) SetContext(ctx context.Context) {
	c.request = c.request.WithContext(ctx)
}

// Request 获取底层 HTTP 请求。
func (c *Context) Request() *http.Request {
	return c.request
}
