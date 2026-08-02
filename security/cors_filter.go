package security

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security/filter"
)

// CorsConfig CORS配置
type CorsConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
	Log              log.Logger
}

// CorsFilter CORS过滤器
type CorsFilter struct {
	config CorsConfig
	logger log.Logger
}

func NewCorsFilter(config CorsConfig) *CorsFilter {
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{"Content-Type", "Authorization", "X-Requested-With"}
	}
	if config.MaxAge == 0 {
		config.MaxAge = 3600
	}
	if config.Log == nil {
		config.Log = log.Build()
	}
	return &CorsFilter{
		config: config,
		logger: config.Log,
	}
}

// DoFilter 实现 filter.Filter 接口
func (f *CorsFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	ctxVal, ok := ctx.(context.Context)
	if !ok {
		return fmt.Errorf("CorsFilter: ctx must be context.Context")
	}
	req, ok := request.(SecurityRequest)
	if !ok {
		return fmt.Errorf("CorsFilter: request must be SecurityRequest")
	}
	resp, ok := response.(SecurityResponse)
	if !ok {
		return fmt.Errorf("CorsFilter: response must be SecurityResponse")
	}
	return f.doFilter(ctxVal, req, resp, chain)
}

func (f *CorsFilter) doFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {
	origin := request.GetHeader("Origin")
	if origin == "" {
		return chain.DoFilter(ctx, request, response)
	}

	if f.logger != nil {
		f.logger.Debug(ctx, "处理 CORS 请求", log.KeyValue{Key: "origin", Value: origin})
	}

	if f.isOriginAllowed(origin) {
		response.SetHeader("Access-Control-Allow-Origin", origin)
		if f.config.AllowCredentials {
			response.SetHeader("Access-Control-Allow-Credentials", "true")
		}
	} else if f.logger != nil {
		f.logger.Warn(ctx, "CORS 来源被拒绝", log.KeyValue{Key: "origin", Value: origin})
	}

	if request.GetMethod() == http.MethodOptions {
		if f.logger != nil {
			f.logger.Debug(ctx, "处理 CORS 预检请求")
		}
		if len(f.config.AllowedMethods) > 0 {
			response.SetHeader("Access-Control-Allow-Methods", strings.Join(f.config.AllowedMethods, ", "))
		}
		if len(f.config.AllowedHeaders) > 0 {
			response.SetHeader("Access-Control-Allow-Headers", strings.Join(f.config.AllowedHeaders, ", "))
		}
		if len(f.config.ExposedHeaders) > 0 {
			response.SetHeader("Access-Control-Expose-Headers", strings.Join(f.config.ExposedHeaders, ", "))
		}
		if f.config.MaxAge > 0 {
			response.SetHeader("Access-Control-Max-Age", strconv.Itoa(f.config.MaxAge))
		}
		response.SetStatusCode(http.StatusNoContent)
		return nil
	}

	return chain.DoFilter(ctx, request, response)
}

// Order 实现 filter.Filter 接口
func (f *CorsFilter) Order() int { return -100 }

func (f *CorsFilter) isOriginAllowed(origin string) bool {
	if len(f.config.AllowedOrigins) == 0 {
		return false
	}
	for _, allowedOrigin := range f.config.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
		if strings.HasSuffix(allowedOrigin, "*") {
			prefix := strings.TrimSuffix(allowedOrigin, "*")
			if strings.HasPrefix(origin, prefix) {
				rest := origin[len(prefix):]
				if rest == "" || strings.HasPrefix(rest, "/") {
					return true
				}
			}
		}
	}
	return false
}
