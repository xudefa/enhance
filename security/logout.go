// Package security 提供安全功能支持，用于 enhance 框架。
package security

import (
	"context"
	"fmt"

	"github.com/xudefa/enhance/security/filter"
)

// LogoutFilter 登出过滤器
type LogoutFilter struct {
	logoutUrl      string
	handlers       []LogoutHandler
	successHandler LogoutSuccessHandler
	filterChain    filter.FilterChain
	httpMethods    []string
}

func NewLogoutFilter(logoutUrl string, handlers []LogoutHandler) *LogoutFilter {
	return &LogoutFilter{
		logoutUrl:   logoutUrl,
		handlers:    handlers,
		httpMethods: []string{"POST", "GET"},
	}
}

func (f *LogoutFilter) AddLogoutHandler(handler LogoutHandler) {
	f.handlers = append(f.handlers, handler)
}

func (f *LogoutFilter) SetSuccessHandler(handler LogoutSuccessHandler) {
	f.successHandler = handler
}

// DoFilter 实现 filter.Filter 接口
func (f *LogoutFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	ctxVal, _ := ctx.(context.Context)
	req, _ := request.(SecurityRequest)
	resp, _ := response.(SecurityResponse)
	return f.doFilter(ctxVal, req, resp, chain)
}

func (f *LogoutFilter) doFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {
	f.filterChain = chain

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

	authentication := GetAuthentication()

	for _, handler := range f.handlers {
		handler.Logout(ctx, request, response, authentication)
	}

	ClearAuthentication()

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
	ClearAuthentication()
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
