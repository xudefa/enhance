package actuator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/xudefa/enhance/actuator/health"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/metrics"
)

// Actuator 运维端点管理器
//
// 提供多种运维端点，包括健康检查、指标收集、环境信息、Bean 列表等。
// 支持多种 HTTP 框架集成，如标准库 http、Gin、Hertz 等。
type Actuator struct {
	healthAggregator *health.Aggregator    // 健康检查聚合器
	metricsRegistry  metrics.MeterRegistry // 指标注册表
	appContext       AppContext            // 应用上下文
	sanitizer        *Sanitizer            // 敏感信息检测器
}

// New 创建 Actuator 实例
func New(ctx AppContext) *Actuator {
	return &Actuator{
		healthAggregator: health.NewAggregator(),
		metricsRegistry:  metrics.NewSimpleRegistry(),
		appContext:       ctx,
		sanitizer:        NewSanitizer(),
	}
}

// HealthHandler 健康检查 HTTP 处理器
//
// 返回聚合后的健康状态信息，包含所有健康指标的详细状态。
// 响应格式：
//
//	{
//	  "status": "UP",
//	  "details": {
//	    "database": {
//	      "status": "UP",
//	      "detail": {}
//	    }
//	  },
//	  "timestamp": "2024-01-01T00:00:00Z"
//	}
func (a *Actuator) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if a.healthAggregator == nil {
		http.Error(w, "no health aggregator", http.StatusInternalServerError)
		return
	}
	h := a.healthAggregator.Aggregate(r.Context())
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(h); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// MetricsHandler 指标 HTTP 处理器
func (a *Actuator) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	if a.metricsRegistry == nil {
		http.Error(w, "no metrics registry", http.StatusInternalServerError)
		return
	}
	m := a.metricsRegistry.Collect()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// EnvHandler 环境信息 HTTP 处理器
func (a *Actuator) EnvHandler(w http.ResponseWriter, r *http.Request) {
	if a.appContext == nil {
		http.Error(w, "no application context", http.StatusInternalServerError)
		return
	}
	env := a.appContext.Environment()
	sources := env.GetPropertySources()

	type propertyItem struct {
		Name  string `json:"name"`
		Value any    `json:"value,omitempty"`
	}

	type sourceInfo struct {
		Name       string         `json:"name"`
		Priority   int            `json:"priority"`
		Properties []propertyItem `json:"properties,omitempty"`
	}

	result := make([]sourceInfo, 0, len(sources))
	for _, s := range sources {
		info := sourceInfo{
			Name:     s.Name(),
			Priority: int(s.Priority()),
		}
		if mp, ok := s.(*environment.MapPropertySource); ok {
			props := make([]propertyItem, 0)
			for _, k := range mp.Keys() {
				v, _ := mp.GetProperty(k)
				// 脱敏处理敏感值
				sanitizedValue := a.sanitizer.Sanitize(k, v)
				props = append(props, propertyItem{
					Name:  k,
					Value: sanitizedValue,
				})
			}
			info.Properties = props
			result = append(result, info)
			continue
		}
		// 对于非 MapPropertySource，简单处理
		info.Properties = make([]propertyItem, 0)
		result = append(result, info)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// BeansHandler Bean 列表 HTTP 处理器
func (a *Actuator) BeansHandler(w http.ResponseWriter, r *http.Request) {
	if a.appContext == nil {
		http.Error(w, "no application context", http.StatusInternalServerError)
		return
	}

	beanDefs := a.appContext.Container().ListBeans()
	beans := make([]map[string]string, 0, len(beanDefs))
	for id, def := range beanDefs {
		typeName := ""
		if def.Type != nil {
			typeName = def.Type.String()
		}
		beans = append(beans, map[string]string{
			"id":   id,
			"type": typeName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"beans": beans,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// SetHealthAggregator 设置健康检查聚合器
func (a *Actuator) SetHealthAggregator(agg *health.Aggregator) {
	a.healthAggregator = agg
}

// SetMetricsRegistry 设置指标注册表
func (a *Actuator) SetMetricsRegistry(reg metrics.MeterRegistry) {
	a.metricsRegistry = reg
}

// MetricsRegistry 获取指标注册表
func (a *Actuator) MetricsRegistry() metrics.MeterRegistry {
	return a.metricsRegistry
}

// RegisterRoutes 注册 Actuator 路由
//
// 使用 RouteRegistrar 接口解耦路由注册逻辑，
// 支持不同的 HTTP 框架实现。
func (a *Actuator) RegisterRoutes(registrar RouteRegistrar, config RouteConfig) {
	base := config.BasePath

	registrar.Handle(base+"/health", http.HandlerFunc(a.HealthHandler))
	registrar.Handle(base+"/metrics", http.HandlerFunc(a.MetricsHandler))
	registrar.Handle(base+"/env", http.HandlerFunc(a.EnvHandler))
	registrar.Handle(base+"/beans", http.HandlerFunc(a.BeansHandler))
	registrar.Handle(base+"/info", http.HandlerFunc(a.InfoHandler))
	registrar.Handle("/metrics", http.HandlerFunc(a.PrometheusHandler))

	if config.ExposeDebug {
		a.RegisterDebugRoutes(registrar)
	}
}

// RegisterDebugRoutes 注册调试路由
func (a *Actuator) RegisterDebugRoutes(registrar RouteRegistrar) {
	handlers := a.PprofHandlers()
	for path, handler := range handlers {
		registrar.Handle(path, http.HandlerFunc(handler))
	}
}

// PprofHandlers returns handlers for pprof endpoints
func (a *Actuator) PprofHandlers() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"/debug/pprof/":        pprof.Index,
		"/debug/pprof/cmdline": pprof.Cmdline,
		"/debug/pprof/profile": pprof.Profile,
		"/debug/pprof/symbol":  pprof.Symbol,
		"/debug/pprof/trace":   pprof.Trace,
	}
}

// InfoHandler 应用信息 HTTP 处理器
func (a *Actuator) InfoHandler(w http.ResponseWriter, r *http.Request) {
	if a.appContext == nil {
		http.Error(w, "no application context", http.StatusInternalServerError)
		return
	}

	env := a.appContext.Environment()

	info := map[string]any{
		"app": map[string]any{
			"name":    env.GetString("app.name", "enhance-app"),
			"version": env.GetString("app.version", "1.0.0"),
		},
		"build": map[string]any{
			"time": env.GetString("build.time", ""),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// PrometheusHandler Prometheus 指标 HTTP 处理器
func (a *Actuator) PrometheusHandler(w http.ResponseWriter, r *http.Request) {
	if a.metricsRegistry == nil {
		http.Error(w, "no metrics registry", http.StatusInternalServerError)
		return
	}

	metrics := a.metricsRegistry.Collect()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	for _, metric := range metrics {
		if len(metric.Tags) > 0 {
			tags := make([]string, 0, len(metric.Tags))
			for k, v := range metric.Tags {
				tags = append(tags, fmt.Sprintf(`%s="%s"`, k, v))
			}
			tagStr := strings.Join(tags, ",")
			if _, err := fmt.Fprintf(w, "%s{%s} %v\n", metric.Name, tagStr, metric.Value); err != nil {
				return
			}
			continue
		}
		if _, err := fmt.Fprintf(w, "%s %v\n", metric.Name, metric.Value); err != nil {
			return
		}
	}
}

// NewDatabaseHealthIndicator 创建数据库健康指示器
func NewDatabaseHealthIndicator(checkFunc func(context.Context) error) health.Indicator {
	return health.NewIndicatorBuilder().
		Name("database").
		CheckFunc(checkFunc).
		Timeout(5*time.Second).
		Detail("type", "database").
		Build()
}

// NewRedisHealthIndicator 创建Redis健康指示器
func NewRedisHealthIndicator(checkFunc func(context.Context) error) health.Indicator {
	return health.NewIndicatorBuilder().
		Name("redis").
		CheckFunc(checkFunc).
		Timeout(5*time.Second).
		Detail("type", "redis").
		Build()
}
