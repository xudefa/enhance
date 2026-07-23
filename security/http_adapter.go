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

func (a *HttpRequestAdapter) GetMethod() string {
	return a.request.Method
}

func (a *HttpRequestAdapter) GetURI() string {
	return a.request.URL.Path
}

func (a *HttpRequestAdapter) GetHeader(key string) string {
	return a.request.Header.Get(key)
}

func (a *HttpRequestAdapter) SetAttribute(key string, value any) {
	ctx := context.WithValue(a.request.Context(), contextKey(key), value)
	*a.request = *a.request.WithContext(ctx)
}

func (a *HttpRequestAdapter) GetAttribute(key string) (any, bool) {
	value := a.request.Context().Value(contextKey(key))
	return value, value != nil
}

// HttpResponseAdapter HTTP响应适配器
type HttpResponseAdapter struct {
	responseWriter http.ResponseWriter
	statusCode     int
	written        bool
}

func NewHttpResponseAdapter(responseWriter http.ResponseWriter) *HttpResponseAdapter {
	return &HttpResponseAdapter{responseWriter: responseWriter, statusCode: 200}
}

func (a *HttpResponseAdapter) SetStatusCode(code int) {
	a.statusCode = code
	a.responseWriter.WriteHeader(code)
	a.written = true
}

func (a *HttpResponseAdapter) StatusCode() int {
	return a.statusCode
}

func (a *HttpResponseAdapter) SetHeader(key, value string) {
	a.responseWriter.Header().Set(key, value)
}

func (a *HttpResponseAdapter) Write(data []byte) error {
	_, err := a.responseWriter.Write(data)
	a.written = true
	return err
}

func (a *HttpResponseAdapter) headersWritten() bool {
	return a.written
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
			response.SetStatusCode(500)
			if writeErr := response.Write([]byte(err.Error())); writeErr != nil {
				fmt.Printf("[enhance] failed to write error response: %v\n", writeErr)
			}
		}
		return
	}

	if response.StatusCode() >= 400 {
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
	ctxVal, _ := ctx.(context.Context)
	req, _ := request.(SecurityRequest)
	resp, _ := response.(SecurityResponse)
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

	SetAuthentication(authenticated)
	ctx = ContextWithAuthentication(ctx, authenticated)

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
