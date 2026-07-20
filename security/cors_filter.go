package security

import (
	"context"
	"net/http"
	"strings"

	"github.com/xudefa/enhance/log"
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

// NewCorsFilter 创建CORS过滤器
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
	return &CorsFilter{
		config: config,
		logger: config.Log,
	}
}

// DoFilter 处理CORS请求
func (f *CorsFilter) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	origin := request.GetHeader("Origin")
	if origin == "" {
		return chain.DoFilter(ctx, request, response)
	}

	f.logger.Debug(ctx, "处理 CORS 请求", log.KeyValue{Key: "origin", Value: origin})

	if f.isOriginAllowed(origin) {
		response.SetHeader("Access-Control-Allow-Origin", origin)
		if f.config.AllowCredentials {
			response.SetHeader("Access-Control-Allow-Credentials", "true")
		}
	} else {
		f.logger.Warn(ctx, "CORS 来源被拒绝", log.KeyValue{Key: "origin", Value: origin})
	}

	if request.GetMethod() == http.MethodOptions {
		f.logger.Debug(ctx, "处理 CORS 预检请求")
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
			response.SetHeader("Access-Control-Max-Age", string(rune(f.config.MaxAge)))
		}
		response.SetStatusCode(http.StatusNoContent)
		return nil
	}

	return chain.DoFilter(ctx, request, response)
}

func (f *CorsFilter) isOriginAllowed(origin string) bool {
	if len(f.config.AllowedOrigins) == 0 {
		return true
	}
	for _, allowedOrigin := range f.config.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
		if strings.HasSuffix(allowedOrigin, "*") {
			prefix := strings.TrimSuffix(allowedOrigin, "*")
			if strings.HasPrefix(origin, prefix) {
				return true
			}
		}
	}
	return false
}
