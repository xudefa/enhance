package webtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

// WebTestClient Web 测试客户端。
type WebTestClient struct {
	handler http.Handler
	baseURL string
	headers map[string]string
}

// NewWebTestClient 创建 Web 测试客户端。
func NewWebTestClient(handler http.Handler) *WebTestClient {
	return &WebTestClient{
		handler: handler,
		headers: make(map[string]string),
	}
}

// BaseURL 设置基础 URL。
func (c *WebTestClient) BaseURL(url string) *WebTestClient {
	c.baseURL = url
	return c
}

// Header 设置默认请求头。
func (c *WebTestClient) Header(name, value string) *WebTestClient {
	c.headers[name] = value
	return c
}

// Get 发起 GET 请求。
func (c *WebTestClient) Get(path string) *RequestSpec {
	return c.newRequest(http.MethodGet, path)
}

// Post 发起 POST 请求。
func (c *WebTestClient) Post(path string) *RequestSpec {
	return c.newRequest(http.MethodPost, path)
}

// Put 发起 PUT 请求。
func (c *WebTestClient) Put(path string) *RequestSpec {
	return c.newRequest(http.MethodPut, path)
}

// Delete 发起 DELETE 请求。
func (c *WebTestClient) Delete(path string) *RequestSpec {
	return c.newRequest(http.MethodDelete, path)
}

// Patch 发起 PATCH 请求。
func (c *WebTestClient) Patch(path string) *RequestSpec {
	return c.newRequest(http.MethodPatch, path)
}

// newRequest 创建请求规范。
func (c *WebTestClient) newRequest(method, path string) *RequestSpec {
	return &RequestSpec{
		client:  c,
		method:  method,
		path:    path,
		headers: make(map[string]string),
	}
}

// RequestSpec 请求规范，用于链式构建 HTTP 请求。
type RequestSpec struct {
	client      *WebTestClient
	method      string
	path        string
	headers     map[string]string
	body        io.Reader
	contentType string
}

// Header 设置请求头。
func (r *RequestSpec) Header(name, value string) *RequestSpec {
	r.headers[name] = value
	return r
}

// ContentType 设置 Content-Type。
func (r *RequestSpec) ContentType(contentType string) *RequestSpec {
	r.contentType = contentType
	r.headers["Content-Type"] = contentType
	return r
}

// Body 设置请求体。
func (r *RequestSpec) Body(body io.Reader) *RequestSpec {
	r.body = body
	return r
}

// JSON 设置 JSON 请求体。
func (r *RequestSpec) JSON(data any) *RequestSpec {
	body, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	r.body = bytes.NewReader(body)
	r.contentType = "application/json"
	r.headers["Content-Type"] = "application/json"
	return r
}

// Exchange 发起请求并返回响应规范。
func (r *RequestSpec) Exchange() *ResponseSpec {
	url := r.client.baseURL + r.path
	req := httptest.NewRequest(r.method, url, r.body)

	for name, value := range r.client.headers {
		req.Header.Set(name, value)
	}

	for name, value := range r.headers {
		req.Header.Set(name, value)
	}

	recorder := httptest.NewRecorder()
	r.client.handler.ServeHTTP(recorder, req)

	return &ResponseSpec{
		recorder: recorder,
	}
}

// ResponseSpec 响应规范，用于链式断言 HTTP 响应。
type ResponseSpec struct {
	recorder *httptest.ResponseRecorder
}

// Status 断言状态码。
func (r *ResponseSpec) Status(expected int) *ResponseSpec {
	if r.recorder.Code != expected {
		panic(fmt.Sprintf("expected status %d, got %d", expected, r.recorder.Code))
	}
	return r
}

// StatusIsOk 断言 200 状态码。
func (r *ResponseSpec) StatusIsOk() *ResponseSpec {
	return r.Status(http.StatusOK)
}

// StatusIsCreated 断言 201 状态码。
func (r *ResponseSpec) StatusIsCreated() *ResponseSpec {
	return r.Status(http.StatusCreated)
}

// StatusIsNoContent 断言 204 状态码。
func (r *ResponseSpec) StatusIsNoContent() *ResponseSpec {
	return r.Status(http.StatusNoContent)
}

// StatusIsBadRequest 断言 400 状态码。
func (r *ResponseSpec) StatusIsBadRequest() *ResponseSpec {
	return r.Status(http.StatusBadRequest)
}

// StatusIsUnauthorized 断言 401 状态码。
func (r *ResponseSpec) StatusIsUnauthorized() *ResponseSpec {
	return r.Status(http.StatusUnauthorized)
}

// StatusIsForbidden 断言 403 状态码。
func (r *ResponseSpec) StatusIsForbidden() *ResponseSpec {
	return r.Status(http.StatusForbidden)
}

// StatusIsNotFound 断言 404 状态码。
func (r *ResponseSpec) StatusIsNotFound() *ResponseSpec {
	return r.Status(http.StatusNotFound)
}

// Header 断言响应头。
func (r *ResponseSpec) Header(name, expected string) *ResponseSpec {
	actual := r.recorder.Header().Get(name)
	if actual != expected {
		panic(fmt.Sprintf("expected header %s=%s, got %s", name, expected, actual))
	}
	return r
}

// Body 获取响应体字符串。
func (r *ResponseSpec) Body() string {
	return r.recorder.Body.String()
}

// JSONBody 解析 JSON 响应体到目标对象。
func (r *ResponseSpec) JSONBody(target any) *ResponseSpec {
	if err := json.Unmarshal(r.recorder.Body.Bytes(), target); err != nil {
		panic(fmt.Sprintf("failed to unmarshal JSON response: %v", err))
	}
	return r
}

// BodyContains 断言响应体包含指定子串。
func (r *ResponseSpec) BodyContains(substring string) *ResponseSpec {
	body := r.recorder.Body.String()
	if !bytes.Contains(r.recorder.Body.Bytes(), []byte(substring)) {
		panic(fmt.Sprintf("expected body to contain %q, got %q", substring, body))
	}
	return r
}

// BodyEquals 断言响应体等于指定字符串。
func (r *ResponseSpec) BodyEquals(expected string) *ResponseSpec {
	actual := r.recorder.Body.String()
	if actual != expected {
		panic(fmt.Sprintf("expected body %q, got %q", expected, actual))
	}
	return r
}

// Print 打印响应信息（调试用）。
func (r *ResponseSpec) Print() *ResponseSpec {
	fmt.Printf("Status: %d\n", r.recorder.Code)
	fmt.Printf("Headers: %v\n", r.recorder.Header())
	fmt.Printf("Body: %s\n", r.recorder.Body.String())
	return r
}

// Recorder 获取底层 ResponseRecorder。
func (r *ResponseSpec) Recorder() *httptest.ResponseRecorder {
	return r.recorder
}

// CreateWebTestClient 创建 Web 测试客户端（便捷函数）。
func CreateWebTestClient(handler http.Handler) *WebTestClient {
	return NewWebTestClient(handler)
}
