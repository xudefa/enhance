// Package prometheus 提供 Prometheus 监控指标自动配置。
//
// Prometheus 是最流行的云原生监控系统。
//
// 功能特性：
//   - 自动配置 Prometheus 指标收集
//   - 支持 OpenMetrics 格式
//   - 自定义指标创建（Counter/Gauge/Histogram）
//   - 独立 HTTP 服务器暴露指标
//
// 配置示例：
//
//	{
//	  "prometheus": {
//	    "enabled": true,
//	    "host": "0.0.0.0",
//	    "port": 9090,
//	    "metrics_path": "/metrics"
//	  }
//	}
//
// 使用示例：
//
//	prom := core.MustGetBean[*prometheus.PrometheusAutoConfiguration](app.Container())
//	counter := prom.NewCounter("requests_total", "Total requests")
//	counter.Inc()
package prometheus
