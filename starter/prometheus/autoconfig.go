// Package prometheus 提供 Prometheus 监控指标自动配置。
package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&PrometheusAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(PrometheusEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityMonitoringLayer)),
	)
}

// PrometheusAutoConfiguration Prometheus 监控自动配置类。
type PrometheusAutoConfiguration struct {
	logger   log.Logger
	registry *prometheus.Registry
	server   *http.Server
	config   *PrometheusConfig
}

// Configure 配置 Prometheus 监控。
func (c *PrometheusAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("加载 Prometheus 配置失败: %w", err)
	}

	c.config = cfg
	c.registry = prometheus.NewRegistry()

	handler := promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: cfg.EnableOpenMetrics,
	})

	mux := http.NewServeMux()
	mux.Handle(cfg.MetricsPath, handler)

	c.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := ctx.Container().RegisterInstance(c.registry, reflect.TypeFor[*prometheus.Registry]()); err != nil {
		return fmt.Errorf("注册 Prometheus Registry 失败: %w", err)
	}

	if err := ctx.Container().RegisterInstance(c.server, reflect.TypeFor[*http.Server]()); err != nil {
		return fmt.Errorf("注册 Prometheus Server 失败: %w", err)
	}

	c.logger.Info(context.Background(), "Prometheus 监控已配置",
		log.KeyValue{Key: "port", Value: cfg.Port},
		log.KeyValue{Key: "path", Value: cfg.MetricsPath},
	)

	return nil
}

// Start 启动 Prometheus 监控服务器。
func (c *PrometheusAutoConfiguration) Start() error {
	c.logger.Info(context.Background(), "Prometheus 监控服务器启动中",
		log.KeyValue{Key: "addr", Value: c.server.Addr},
	)
	return c.server.ListenAndServe()
}

// Stop 停止 Prometheus 监控服务器。
func (c *PrometheusAutoConfiguration) Stop(ctx context.Context) error {
	return c.server.Shutdown(ctx)
}

// GetRegistry 获取 Prometheus Registry 实例。
func (c *PrometheusAutoConfiguration) GetRegistry() *prometheus.Registry {
	return c.registry
}

// NewCounter 创建 Counter 指标。
func (c *PrometheusAutoConfiguration) NewCounter(name, help string) prometheus.Counter {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	})
	c.registry.MustRegister(counter)
	return counter
}

// NewGauge 创建 Gauge 指标。
func (c *PrometheusAutoConfiguration) NewGauge(name, help string) prometheus.Gauge {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})
	c.registry.MustRegister(gauge)
	return gauge
}

// NewHistogram 创建 Histogram 指标。
func (c *PrometheusAutoConfiguration) NewHistogram(name, help string, buckets []float64) prometheus.Histogram {
	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	})
	c.registry.MustRegister(histogram)
	return histogram
}

// PrometheusConfig Prometheus 监控配置。
type PrometheusConfig struct {
	Enabled           bool   `json:"enabled" mapstructure:"enabled"`
	Host              string `json:"host" mapstructure:"host"`
	Port              int    `json:"port" mapstructure:"port"`
	MetricsPath       string `json:"metrics_path" mapstructure:"metrics_path"`
	EnableOpenMetrics bool   `json:"enable_open_metrics" mapstructure:"enable_open_metrics"`
}

// 配置常量。
const (
	PrometheusEnabled        = "prometheus.enabled"
	DefaultPrometheusHost    = "0.0.0.0"
	DefaultPrometheusPort    = 9090
	DefaultMetricsPath       = "/metrics"
	DefaultEnableOpenMetrics = false
	ConditionTrue            = "true"
)

// loadConfig 从 Environment 加载 Prometheus 配置。
func (c *PrometheusAutoConfiguration) loadConfig(env *environment.Environment) (*PrometheusConfig, error) {
	cfg := &PrometheusConfig{
		Host:              DefaultPrometheusHost,
		Port:              DefaultPrometheusPort,
		MetricsPath:       DefaultMetricsPath,
		EnableOpenMetrics: DefaultEnableOpenMetrics,
	}

	if err := env.BindPrefix("prometheus", cfg); err != nil {
		return nil, fmt.Errorf("绑定 Prometheus 配置失败: %w", err)
	}

	return cfg, nil
}
