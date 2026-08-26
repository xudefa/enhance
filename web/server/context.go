package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/xudefa/enhance/web/core"
)

// DefaultContext 默认的 HTTP 上下文实现
type DefaultContext struct {
	writer  http.ResponseWriter
	request *http.Request
	params  map[string]string
	index   int
	mw      []core.MiddlewareFunc
	handler core.HandlerFunc
	aborted bool
}

// NewContext 创建新的 HTTP 上下文
func NewContext(w http.ResponseWriter, r *http.Request) *DefaultContext {
	return &DefaultContext{
		writer:  w,
		request: r,
		params:  make(map[string]string),
		index:   -1,
	}
}

// WithParams 设置路径参数
func (c *DefaultContext) WithParams(params map[string]string) *DefaultContext {
	c.params = params
	return c
}

// WithMiddleware 设置中间件链
func (c *DefaultContext) WithMiddleware(mw []core.MiddlewareFunc, handler core.HandlerFunc) *DefaultContext {
	c.mw = mw
	c.handler = handler
	return c
}

// RequestMethod 返回请求方法
func (c *DefaultContext) RequestMethod() string {
	return c.request.Method
}

// RequestURI 返回请求 URI
func (c *DefaultContext) RequestURI() string {
	return c.request.RequestURI
}

// PathParam 获取路径参数
func (c *DefaultContext) PathParam(name string) string {
	return c.params[name]
}

// Query 获取查询参数
func (c *DefaultContext) Query(name string) string {
	return c.request.URL.Query().Get(name)
}

// QueryDefault 获取查询参数（带默认值）
func (c *DefaultContext) QueryDefault(name, defaultVal string) string {
	val := c.request.URL.Query().Get(name)
	if val == "" {
		return defaultVal
	}
	return val
}

// Header 获取请求头
func (c *DefaultContext) Header(key string) string {
	return c.request.Header.Get(key)
}

// BindJSON 解析 JSON 请求体
func (c *DefaultContext) BindJSON(target any) error {
	defer func() { _ = c.request.Body.Close() }()
	body, err := io.ReadAll(http.MaxBytesReader(c.writer, c.request.Body, 10<<20)) // 10 MB
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}
	return nil
}

// SetStatusCode 设置响应状态码
func (c *DefaultContext) SetStatusCode(code int) {
	c.writer.WriteHeader(code)
}

// SetHeader 设置响应头
func (c *DefaultContext) SetHeader(key, value string) {
	c.writer.Header().Set(key, value)
}

// JSON 返回 JSON 响应
func (c *DefaultContext) JSON(code int, data any) error {
	c.writer.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(data)
	if err != nil {
		// 序列化失败时直接返回 500，避免客户端收到空的 200 响应
		c.writer.WriteHeader(http.StatusInternalServerError)
		_, _ = c.writer.Write([]byte(`{"error": "json marshal failed"}`))
		return err
	}
	c.writer.WriteHeader(code)
	_, _ = c.writer.Write(body)
	return nil
}

// String 返回字符串响应
func (c *DefaultContext) String(code int, format string, args ...any) {
	c.writer.Header().Set("Content-Type", "text/plain")
	c.writer.WriteHeader(code)
	_, _ = fmt.Fprintf(c.writer, format, args...)
}

// AbortWithStatus 中止请求
func (c *DefaultContext) AbortWithStatus(code int) {
	c.aborted = true
	c.writer.WriteHeader(code)
}

// AbortWithStatusJSON 中止请求并返回 JSON
func (c *DefaultContext) AbortWithStatusJSON(code int, body any) {
	c.aborted = true
	c.writer.Header().Set("Content-Type", "application/json")
	c.writer.WriteHeader(code)
	_ = json.NewEncoder(c.writer).Encode(body)
}

// Next 执行下一个中间件
func (c *DefaultContext) Next() {
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

// IsAborted 判断是否已中止
func (c *DefaultContext) IsAborted() bool {
	return c.aborted
}

// Context 获取请求上下文
func (c *DefaultContext) Context() context.Context {
	return c.request.Context()
}

// Request 获取底层 HTTP 请求
func (c *DefaultContext) Request() *http.Request {
	return c.request
}

// SetContext 设置请求上下文
func (c *DefaultContext) SetContext(ctx context.Context) {
	c.request = c.request.WithContext(ctx)
}
