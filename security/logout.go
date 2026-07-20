// Package security 提供安全功能支持，用于 enhance 框架。
package security

import (
	"context"
	"fmt"
)

// LogoutFilter 登出过滤器
// 职责：处理登出请求，清除认证信息并执行登出处理器
// 执行流程：
// 1. 检查请求方法和URL是否匹配登出配置
// 2. 执行所有LogoutHandler（清除安全上下文、清除Cookie等）
// 3. 清除全局认证信息
// 4. 调用登出成功处理器（重定向或返回JSON）
type LogoutFilter struct {
	logoutUrl      string
	handlers       []LogoutHandler
	successHandler LogoutSuccessHandler
	filterChain    SecurityFilterChain
	httpMethods    []string
}

// NewLogoutFilter 创建登出过滤器。
func NewLogoutFilter(logoutUrl string, handlers []LogoutHandler) *LogoutFilter {
	return &LogoutFilter{
		logoutUrl:   logoutUrl,
		handlers:    handlers,
		httpMethods: []string{"POST", "GET"},
	}
}

// AddLogoutHandler 添加登出处理器。
func (f *LogoutFilter) AddLogoutHandler(handler LogoutHandler) {
	f.handlers = append(f.handlers, handler)
}

// SetSuccessHandler 设置登出成功处理器。
func (f *LogoutFilter) SetSuccessHandler(handler LogoutSuccessHandler) {
	f.successHandler = handler
}

// DoFilter 处理登出请求。
func (f *LogoutFilter) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
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

// DefaultLogoutSuccessHandler 默认登出成功处理器
// 职责：登出成功后重定向到指定URL（默认/login?logout）
type DefaultLogoutSuccessHandler struct {
	defaultTargetUrl string
}

// NewDefaultLogoutSuccessHandler 创建默认登出成功处理器。
func NewDefaultLogoutSuccessHandler(defaultTargetUrl string) *DefaultLogoutSuccessHandler {
	return &DefaultLogoutSuccessHandler{
		defaultTargetUrl: defaultTargetUrl,
	}
}

// OnLogoutSuccess 处理登出成功。
func (h *DefaultLogoutSuccessHandler) OnLogoutSuccess(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	response.SetStatusCode(302)
	response.SetHeader("Location", h.defaultTargetUrl)
}

// SimpleLogoutSuccessHandler 简单登出成功处理器。
type SimpleLogoutSuccessHandler struct {
	targetUrl string
}

// NewSimpleLogoutSuccessHandler 创建简单登出成功处理器。
func NewSimpleLogoutSuccessHandler(targetUrl string) *SimpleLogoutSuccessHandler {
	return &SimpleLogoutSuccessHandler{
		targetUrl: targetUrl,
	}
}

// OnLogoutSuccess 处理登出成功，重定向到指定 URL。
func (h *SimpleLogoutSuccessHandler) OnLogoutSuccess(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	response.SetStatusCode(302)
	response.SetHeader("Location", h.targetUrl)
}

// SecurityContextLogoutHandler 安全上下文登出处理器
// 职责：登出时清除全局认证信息
type SecurityContextLogoutHandler struct{}

// NewSecurityContextLogoutHandler 创建安全上下文登出处理器。
func NewSecurityContextLogoutHandler() *SecurityContextLogoutHandler {
	return &SecurityContextLogoutHandler{}
}

// Logout 清除安全上下文。
func (h *SecurityContextLogoutHandler) Logout(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	ClearAuthentication()
}

// CookieClearingLogoutHandler Cookie清除登出处理器
// 职责：登出时清除指定的Cookie（如会话Cookie、CSRF Token等）
type CookieClearingLogoutHandler struct {
	cookieNames []string
}

// NewCookieClearingLogoutHandler 创建 Cookie 清除登出处理器。
func NewCookieClearingLogoutHandler(cookieNames ...string) *CookieClearingLogoutHandler {
	return &CookieClearingLogoutHandler{
		cookieNames: cookieNames,
	}
}

// Logout 清除指定 Cookie。
func (h *CookieClearingLogoutHandler) Logout(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	for _, name := range h.cookieNames {
		response.SetHeader("Set-Cookie", fmt.Sprintf("%s=; Path=/; Max-Age=0", name))
	}
}
