package security

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/xudefa/enhance/security/filter"
)

type contextKey string

// HttpRequestAdapter HTTP请求适配器
type HttpRequestAdapter struct {
	request *http.Request
}

func NewHttpRequestAdapter(request *http.Request) *HttpRequestAdapter {
	return &HttpRequestAdapter{request: request}
}

// GetMethod 返回 HTTP 请求方法（GET、POST 等）。
func (a *HttpRequestAdapter) GetMethod() string {
	return a.request.Method
}

// GetURI 返回 HTTP 请求的 URI 路径。
func (a *HttpRequestAdapter) GetURI() string {
	return a.request.URL.Path
}

// GetHeader 返回 HTTP 请求头的指定键对应的值。
func (a *HttpRequestAdapter) GetHeader(key string) string {
	return a.request.Header.Get(key)
}

// RemoteAddress 返回直连对端的地址。
func (a *HttpRequestAdapter) RemoteAddress() string {
	return a.request.RemoteAddr
}

// SetAttribute 设置请求上下文的属性键值对。
func (a *HttpRequestAdapter) SetAttribute(key string, value any) {
	ctx := context.WithValue(a.request.Context(), contextKey(key), value)
	*a.request = *a.request.WithContext(ctx)
}

// GetAttribute 获取请求上下文的指定属性值。
func (a *HttpRequestAdapter) GetAttribute(key string) (any, bool) {
	value := a.request.Context().Value(contextKey(key))
	return value, value != nil
}

// HttpResponseAdapter HTTP响应适配器
type HttpResponseAdapter struct {
	responseWriter http.ResponseWriter
	statusCode     int
	statusSet      bool
	committed      bool
	written        bool
}

func NewHttpResponseAdapter(responseWriter http.ResponseWriter) *HttpResponseAdapter {
	return &HttpResponseAdapter{responseWriter: responseWriter, statusCode: http.StatusOK}
}

// SetStatusCode 设置 HTTP 响应状态码。
//
// WriteHeader 延迟到首次写入响应体时提交，确保 SetStatusCode 之后设置的
// 响应头仍能生效（net/http 中 WriteHeader 之后再修改 Header 无效）。
func (a *HttpResponseAdapter) SetStatusCode(code int) {
	a.statusCode = code
	a.statusSet = true
}

// StatusCode 返回已设置的 HTTP 响应状态码。
func (a *HttpResponseAdapter) StatusCode() int {
	return a.statusCode
}

// SetHeader 设置 HTTP 响应头的指定键值对。
func (a *HttpResponseAdapter) SetHeader(key, value string) {
	a.responseWriter.Header().Set(key, value)
}

// Write 写入 HTTP 响应体数据。
func (a *HttpResponseAdapter) Write(data []byte) error {
	a.flush()
	_, err := a.responseWriter.Write(data)
	a.written = true
	return err
}

// flush 将已设置的状态码与响应头提交到底层 ResponseWriter。
func (a *HttpResponseAdapter) flush() {
	if a.statusSet && !a.committed {
		a.responseWriter.WriteHeader(a.statusCode)
		a.committed = true
	}
}

func (a *HttpResponseAdapter) headersWritten() bool {
	return a.committed || a.written
}

// SecurityFilterChainHandler 安全过滤器链处理器
type SecurityFilterChainHandler struct {
	securityFilterChain SecurityFilterChain
	nextHandler         http.Handler
}

func NewSecurityFilterChainHandler(securityFilterChain SecurityFilterChain, nextHandler http.Handler) *SecurityFilterChainHandler {
	return &SecurityFilterChainHandler{
		securityFilterChain: securityFilterChain,
		nextHandler:         nextHandler,
	}
}

// ServeHTTP 实现http.Handler接口
func (h *SecurityFilterChainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := NewHttpRequestAdapter(r)
	response := NewHttpResponseAdapter(w)

	ctx := r.Context()
	err := h.securityFilterChain.DoFilter(ctx, request, response)
	if err != nil {
		if !response.headersWritten() {
			response.SetStatusCode(http.StatusInternalServerError)
			if writeErr := response.Write([]byte(err.Error())); writeErr != nil {
				fmt.Printf("[enhance] failed to write error response: %v\n", writeErr)
			}
		}
		response.flush()
		return
	}

	response.flush()
	if response.headersWritten() {
		return
	}

	if h.nextHandler != nil {
		if authVal, ok := request.GetAttribute("security.currentAuthentication"); ok {
			if auth, ok := authVal.(Authentication); ok {
				r = r.WithContext(ContextWithAuthentication(r.Context(), auth))
			}
		}
		h.nextHandler.ServeHTTP(w, r)
	}
}

// SetNextHandler 设置下一个 HTTP 处理器（用于过滤器链接力）。
func (h *SecurityFilterChainHandler) SetNextHandler(handler http.Handler) {
	h.nextHandler = handler
}

// BasicAuthenticationFilter Basic认证过滤器
type BasicAuthenticationFilter struct {
	authenticationManager AuthenticationManager
}

func NewBasicAuthenticationFilter(authenticationManager AuthenticationManager) *BasicAuthenticationFilter {
	return &BasicAuthenticationFilter{
		authenticationManager: authenticationManager,
	}
}

// DoFilter 处理Basic认证
func (f *BasicAuthenticationFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	ctxVal, ok := ctx.(context.Context)
	if !ok {
		return fmt.Errorf("BasicAuthenticationFilter: ctx must be context.Context")
	}
	req, ok := request.(SecurityRequest)
	if !ok {
		return fmt.Errorf("BasicAuthenticationFilter: request must be SecurityRequest")
	}
	resp, ok := response.(SecurityResponse)
	if !ok {
		return fmt.Errorf("BasicAuthenticationFilter: response must be SecurityResponse")
	}
	return f.doFilter(ctxVal, req, resp, chain)
}

func (f *BasicAuthenticationFilter) doFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {
	authHeader := request.GetHeader("Authorization")
	if authHeader == "" {
		return chain.DoFilter(ctx, request, response)
	}

	const prefix = "Basic "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		return chain.DoFilter(ctx, request, response)
	}

	username, password, err := f.extractBasicAuth(authHeader)
	if err != nil {
		return err
	}

	authToken := NewUsernamePasswordAuthenticationToken(username, password)
	authenticated, err := f.authenticationManager.Authenticate(ctx, authToken)
	if err != nil {
		return err
	}

	ctx = ContextWithAuthentication(ctx, authenticated)
	request.SetAttribute("security.currentAuthentication", authenticated)

	return chain.DoFilter(ctx, request, response)
}

// Order 实现 filter.Filter 接口
func (f *BasicAuthenticationFilter) Order() int { return 0 }

func (f *BasicAuthenticationFilter) extractBasicAuth(authHeader string) (string, string, error) {
	const prefix = "Basic "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		return "", "", ErrBadCredentials
	}

	encoded := authHeader[len(prefix):]
	decoded, err := f.decodeBase64(encoded)
	if err != nil {
		return "", "", ErrBadCredentials
	}

	credentials := string(decoded)
	sepIndex := -1
	for i, c := range credentials {
		if c == ':' {
			sepIndex = i
			break
		}
	}

	if sepIndex == -1 {
		return "", "", ErrBadCredentials
	}

	username := credentials[:sepIndex]
	password := credentials[sepIndex+1:]

	return username, password, nil
}

func (f *BasicAuthenticationFilter) decodeBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
