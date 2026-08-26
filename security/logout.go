// Package security 提供安全功能支持，用于 enhance 框架。
package security

import (
	"context"
	"fmt"

	"github.com/xudefa/enhance/security/filter"
)

// LogoutFilter 登出过滤器。
//
// 处理用户登出请求，支持自定义登出 URL 和登出处理器。
type LogoutFilter struct {
	logoutUrl      string
	handlers       []LogoutHandler
	successHandler LogoutSuccessHandler
	httpMethods    []string
}

// NewLogoutFilter 创建登出过滤器。
//
// 参数:
//   - logoutUrl: 登出 URL 路径，必须以 "/" 开头
//   - handlers: 登出处理器列表，可为 nil
//
// 返回:
//   - *LogoutFilter: 登出过滤器实例
//   - error: 参数错误
func NewLogoutFilter(logoutUrl string, handlers []LogoutHandler) (*LogoutFilter, error) {
	if logoutUrl == "" {
		return nil, fmt.Errorf("logout: logoutUrl must not be empty")
	}
	return &LogoutFilter{
		logoutUrl:   logoutUrl,
		handlers:    handlers,
		httpMethods: []string{"POST", "DELETE"},
	}, nil
}

// MustNewLogoutFilter 创建登出过滤器，失败则 panic。
func MustNewLogoutFilter(logoutUrl string, handlers []LogoutHandler) *LogoutFilter {
	filter, err := NewLogoutFilter(logoutUrl, handlers)
	if err != nil {
		panic(err)
	}
	return filter
}

func (f *LogoutFilter) AddLogoutHandler(handler LogoutHandler) {
	f.handlers = append(f.handlers, handler)
}

func (f *LogoutFilter) SetSuccessHandler(handler LogoutSuccessHandler) {
	f.successHandler = handler
}

// DoFilter 实现 filter.Filter 接口
func (f *LogoutFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	ctxVal, ok := ctx.(context.Context)
	if !ok {
		return fmt.Errorf("LogoutFilter: ctx must be context.Context")
	}
	req, ok := request.(SecurityRequest)
	if !ok {
		return fmt.Errorf("LogoutFilter: request must be SecurityRequest")
	}
	resp, ok := response.(SecurityResponse)
	if !ok {
		return fmt.Errorf("LogoutFilter: response must be SecurityResponse")
	}
	return f.doFilter(ctxVal, req, resp, chain)
}

func (f *LogoutFilter) doFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {

	method := request.GetMethod()
	methodAllowed := false
	for _, m := range f.httpMethods {
		if m == method {
			methodAllowed = true
			break
		}
	}
	if !methodAllowed {
		return chain.DoFilter(ctx, request, response)
	}

	uri := request.GetURI()
	if uri != f.logoutUrl {
		return chain.DoFilter(ctx, request, response)
	}

	authentication := GetAuthenticationFromContext(ctx)

	for _, handler := range f.handlers {
		handler.Logout(ctx, request, response, authentication)
	}

	if f.successHandler == nil {
		response.SetStatusCode(200)
		response.SetHeader("Content-Type", "application/json")
		if writeErr := response.Write([]byte(`{"message":"logout success"}`)); writeErr != nil {
			fmt.Printf("[enhance] failed to write logout success response: %v\n", writeErr)
		}
		return nil
	}
	f.successHandler.OnLogoutSuccess(ctx, request, response, authentication)

	return nil
}

// Order 实现 filter.Filter 接口
func (f *LogoutFilter) Order() int { return 0 }

// DefaultLogoutSuccessHandler 默认登出成功处理器
type DefaultLogoutSuccessHandler struct {
	defaultTargetUrl string
}

func NewDefaultLogoutSuccessHandler(defaultTargetUrl string) *DefaultLogoutSuccessHandler {
	return &DefaultLogoutSuccessHandler{
		defaultTargetUrl: defaultTargetUrl,
	}
}

func (h *DefaultLogoutSuccessHandler) OnLogoutSuccess(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	response.SetStatusCode(302)
	response.SetHeader("Location", h.defaultTargetUrl)
}

// SimpleLogoutSuccessHandler 简单登出成功处理器
type SimpleLogoutSuccessHandler struct {
	targetUrl string
}

func NewSimpleLogoutSuccessHandler(targetUrl string) *SimpleLogoutSuccessHandler {
	return &SimpleLogoutSuccessHandler{
		targetUrl: targetUrl,
	}
}

func (h *SimpleLogoutSuccessHandler) OnLogoutSuccess(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	response.SetStatusCode(302)
	response.SetHeader("Location", h.targetUrl)
}

// SecurityContextLogoutHandler 安全上下文登出处理器
type SecurityContextLogoutHandler struct{}

func NewSecurityContextLogoutHandler() *SecurityContextLogoutHandler {
	return &SecurityContextLogoutHandler{}
}

func (h *SecurityContextLogoutHandler) Logout(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
}

// CookieClearingLogoutHandler Cookie清除登出处理器
type CookieClearingLogoutHandler struct {
	cookieNames []string
}

func NewCookieClearingLogoutHandler(cookieNames ...string) *CookieClearingLogoutHandler {
	return &CookieClearingLogoutHandler{
		cookieNames: cookieNames,
	}
}

func (h *CookieClearingLogoutHandler) Logout(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	for _, name := range h.cookieNames {
		response.SetHeader("Set-Cookie", fmt.Sprintf("%s=; Path=/; Max-Age=0", name))
	}
}
