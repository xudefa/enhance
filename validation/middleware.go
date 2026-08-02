// Package validation 提供参数校验功能，用于 enhance 框架。
package validation

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
)

// ToJSON 将错误响应序列化为 JSON 字节。
func (e *ErrorResponse) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// DefaultErrorHandler 默认错误处理器。
//
// 支持多种上下文类型的错误处理:
//
//  1. http.ResponseWriter: 设置 400 状态码并写入 JSON 错误响应
//  2. ResponseWriter (自定义接口): 统一错误响应
//  3. 其他类型: 仅记录错误信息，不做响应处理
func DefaultErrorHandler(c any, err error) {
	if c == nil || err == nil {
		return
	}

	switch ctx := c.(type) {
	case http.ResponseWriter:
		ctx.Header().Set("Content-Type", "application/json")
		ctx.WriteHeader(http.StatusBadRequest)
		resp := &ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
		if data, jsonErr := resp.ToJSON(); jsonErr == nil {
			_, _ = ctx.Write(data)
		}

	case ResponseWriter:
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetHeader("Content-Type", "application/json")
		resp := &ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
		if data, jsonErr := resp.ToJSON(); jsonErr == nil {
			_ = ctx.Write(data)
		}

	case *http.Request:
		_ = ctx

	default:
		handler := findErrorHandlerMethod(c)
		if handler != nil {
			handler(c, err)
		}
	}
}

func findErrorHandlerMethod(c any) func(any, error) {
	if c == nil {
		return nil
	}
	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Ptr { //go inline: Constant reflect.Ptr should be inlined
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	method := v.MethodByName("OnValidationError")
	if method.IsValid() && method.Type().NumIn() == 1 && method.Type().NumOut() == 0 {
		return func(c any, err error) {
			method.Call([]reflect.Value{reflect.ValueOf(err)})
		}
	}

	return nil
}

// ValidateMiddleware 通用验证中间件函数类型。
type ValidateMiddleware func(c any, obj any, config *MiddlewareConfig) error

// NewValidateMiddleware 创建新的验证中间件。
func NewValidateMiddleware(config *MiddlewareConfig) ValidateMiddleware {
	return func(c any, obj any, cfg *MiddlewareConfig) error {
		if cfg.Validator == nil {
			return errors.New("validator cannot be nil")
		}

		if shouldSkipPath(c, cfg.SkipPaths) {
			return nil
		}

		var err error
		if groupedValidator, ok := cfg.Validator.(*GroupedTagValidator); ok && len(cfg.Groups) > 0 {
			err = groupedValidator.ValidateWithGroups(obj, cfg.Groups...)
			if err != nil {
				if cfg.ErrorHandler != nil {
					cfg.ErrorHandler(c, err)
				}
				return err
			}
			return nil
		}

		err = cfg.Validator.Validate(obj)
		if err != nil {
			if cfg.ErrorHandler != nil {
				cfg.ErrorHandler(c, err)
			}
			return err
		}

		return nil
	}
}

// shouldSkipPath 检查是否应该跳过验证
//
// 参数:
//   - c: 上下文对象，可以是 http.Request 或包含 Path() 方法的结构
//   - skipPaths: 需要跳过的路径列表
//
// 返回值:
//   - true: 跳过验证
//   - false: 需要验证
func shouldSkipPath(c any, skipPaths []string) bool {
	if len(skipPaths) == 0 || c == nil {
		return false
	}

	var currentPath string

	switch ctx := c.(type) {
	case *http.Request:
		currentPath = ctx.URL.Path
	case http.Request:
		currentPath = ctx.URL.Path
	default:
		pathValue := getPathFromContext(c)
		if pathValue != "" {
			currentPath = pathValue
		}
	}

	if currentPath == "" {
		return false
	}

	for _, skipPath := range skipPaths {
		if matchPath(currentPath, skipPath) {
			return true
		}
	}

	return false
}

func getPathFromContext(c any) string {
	if c == nil {
		return ""
	}

	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ""
	}

	pathMethod := v.MethodByName("Path")
	if pathMethod.IsValid() && pathMethod.Type().NumOut() == 1 {
		results := pathMethod.Call(nil)
		if len(results) > 0 {
			if path, ok := results[0].Interface().(string); ok {
				return path
			}
		}
	}

	uriMethod := v.MethodByName("RequestURI")
	if uriMethod.IsValid() && uriMethod.Type().NumOut() == 1 {
		results := uriMethod.Call(nil)
		if len(results) > 0 {
			if uri, ok := results[0].Interface().(string); ok {
				return uri
			}
		}
	}

	return ""
}

func matchPath(currentPath, pattern string) bool {
	if currentPath == pattern {
		return true
	}

	if len(pattern) > 1 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		if len(currentPath) >= len(prefix) && currentPath[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}
