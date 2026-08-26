package server

import (
	"context"
	"net/http"

	"github.com/xudefa/enhance/web/core"
)

// ScopeContextKey 是 RequestScope 在 context 中的键
type ScopeContextKey struct{}

// RequestScopeMiddleware 请求级别作用域中间件
//
// 为每个 HTTP 请求创建独立的 RequestScope，并将其注入到 context 中。
// 请求结束时自动清理作用域中的所有 Bean 实例。
//
// 使用方式：
//
//	// 注册 RequestScope
//	core.RegisterScope("request", core.NewRequestScope())
//
//	// 在路由中使用中间件
//	handler := net.RequestScopeMiddleware(router)
//	http.ListenAndServe(":8080", handler)
//
// 在 Handler 中获取 RequestScope：
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    scope := net.GetRequestScope(r.Context())
//	    if scope != nil {
//	        user := scope.Get("user", func() any {
//	            return loadUserFromDB(r)
//	        }).(*User)
//	    }
//	}
func RequestScopeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := NewRequestScope()
		ctx := context.WithValue(r.Context(), ScopeContextKey{}, scope)
		defer scope.Clear()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestScope 从 context 中获取 RequestScope
//
// 如果 context 中没有 RequestScope，返回 nil。
func GetRequestScope(ctx context.Context) *RequestScope {
	if scope, ok := ctx.Value(ScopeContextKey{}).(*RequestScope); ok {
		return scope
	}
	return nil
}

// MustGetRequestScope 从 context 中获取 RequestScope，不存在时 panic
func MustGetRequestScope(ctx context.Context) *RequestScope {
	scope := GetRequestScope(ctx)
	if scope == nil {
		panic("RequestScope not found in context, make sure RequestScopeMiddleware is applied")
	}
	return scope
}

// RequestScopeMiddlewareFunc 返回 MiddlewareFunc 版本的请求级别作用域中间件
//
// 适用于使用 enhance 中间件体系的场景。
func RequestScopeMiddlewareFunc() core.MiddlewareFunc {
	return func(ctx core.Context) {
		scope := NewRequestScope()
		newCtx := context.WithValue(ctx.Context(), ScopeContextKey{}, scope)
		ctx.SetContext(newCtx)
		defer scope.Clear()
		ctx.Next()
	}
}
