package actuator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"sort"
	"time"

	"github.com/xudefa/enhance/actuator/health"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/metrics"
)

// writeJSONResponse 写入 JSON 响应，出错时返回 500，复用公共模式
func writeJSONResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

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
	healthResult := a.healthAggregator.Aggregate(r.Context())

	jsonData, err := json.Marshal(healthResult)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statusCode := http.StatusOK
	if healthResult.Status != health.StatusUp {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(jsonData)
}

// MetricsHandler 指标 HTTP 处理器
func (a *Actuator) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	if a.metricsRegistry == nil {
		http.Error(w, "no metrics registry", http.StatusInternalServerError)
		return
	}
	m := a.metricsRegistry.Collect()
	writeJSONResponse(w, m)
}

// envPropertyItem 环境属性键值项。
type envPropertyItem struct {
	Name  string `json:"name"`
	Value any    `json:"value,omitempty"`
}

// envSourceInfo 环境属性源信息。
type envSourceInfo struct {
	Name       string            `json:"name"`
	Priority   int               `json:"priority"`
	Properties []envPropertyItem `json:"properties,omitempty"`
}

// EnvHandler 环境信息 HTTP 处理器
func (a *Actuator) EnvHandler(w http.ResponseWriter, r *http.Request) {
	if a.appContext == nil {
		http.Error(w, "no application context", http.StatusInternalServerError)
		return
	}
	env := a.appContext.Environment()
	if env == nil {
		http.Error(w, "no environment", http.StatusInternalServerError)
		return
	}

	sourceInfoList := a.buildEnvSourceInfoList(env.GetPropertySources())

	writeJSONResponse(w, sourceInfoList)
}

// buildEnvSourceInfoList 构建环境属性源信息列表。
func (a *Actuator) buildEnvSourceInfoList(sources []environment.PropertySource) []envSourceInfo {
	sourceInfoList := make([]envSourceInfo, 0, len(sources))
	for _, s := range sources {
		sourceDetail := envSourceInfo{
			Name:     s.Name(),
			Priority: int(s.Priority()),
		}
		if mp, ok := s.(*environment.MapPropertySource); ok {
			keys := mp.Keys()
			props := make([]envPropertyItem, 0, len(keys))
			for _, k := range keys {
				v, _ := mp.GetProperty(k)
				// 脱敏处理敏感值
				sanitizedValue := a.sanitizer.Sanitize(k, v)
				props = append(props, envPropertyItem{
					Name:  k,
					Value: sanitizedValue,
				})
			}
			sourceDetail.Properties = props
			sourceInfoList = append(sourceInfoList, sourceDetail)
			continue
		}
		// 对于非 MapPropertySource，简单处理
		sourceDetail.Properties = make([]envPropertyItem, 0)
		sourceInfoList = append(sourceInfoList, sourceDetail)
	}
	return sourceInfoList
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

	writeJSONResponse(w, map[string]any{"beans": beans})
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

// PprofHandlers 返回 pprof 调试端点的 HTTP 处理函数集合。
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
	if env == nil {
		http.Error(w, "no environment", http.StatusInternalServerError)
		return
	}

	appInfo := map[string]any{
		"app": map[string]any{
			"name":    env.GetString("app.name", "enhance-app"),
			"version": env.GetString("app.version", "1.0.0"),
		},
		"build": map[string]any{
			"time": env.GetString("build.time", ""),
		},
	}

	writeJSONResponse(w, appInfo)
}

// PrometheusHandler Prometheus 指标 HTTP 处理器
func (a *Actuator) PrometheusHandler(w http.ResponseWriter, r *http.Request) {
	if a.metricsRegistry == nil {
		http.Error(w, "no metrics registry", http.StatusInternalServerError)
		return
	}

	collected := a.metricsRegistry.Collect()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	if err := writePrometheus(w, collected); err != nil {
		return
	}
}

// writePrometheus 以 Prometheus 文本格式输出指标。
//
// 每个指标族只输出一次 TYPE 行，标签按键排序并对值做 Prometheus 转义。
func writePrometheus(w io.Writer, collected []metrics.Metric) error {
	type family struct {
		typ    string
		sample string
	}

	families := make(map[string]*family)
	names := make([]string, 0, len(collected))
	for _, m := range collected {
		familyEntry, ok := families[m.Name]
		if !ok {
			familyEntry = &family{typ: m.Type}
			families[m.Name] = familyEntry
			names = append(names, m.Name)
		}
		familyEntry.sample += formatPrometheusSample(m)
	}
	sort.Strings(names)

	for _, name := range names {
		familyEntry := families[name]
		if _, err := fmt.Fprintf(w, "# TYPE %s %s\n", name, familyEntry.typ); err != nil {
			return fmt.Errorf("写入指标类型行失败: %w", err)
		}
		if _, err := io.WriteString(w, familyEntry.sample); err != nil {
			return fmt.Errorf("写入指标样本失败: %w", err)
		}
	}
	return nil
}

// formatPrometheusSample 格式化单条指标样本行，标签值经过 Prometheus 转义。
func formatPrometheusSample(m metrics.Metric) string {
	labels := metrics.FormatLabels(m.Tags)
	if labels == "" {
		return fmt.Sprintf("%s %g\n", m.Name, m.Value)
	}
	return fmt.Sprintf("%s%s %g\n", m.Name, labels, m.Value)
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
