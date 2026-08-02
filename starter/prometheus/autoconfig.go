// Package prometheus 提供 Prometheus 监控指标自动配置。
package prometheus

import (
	"context"
	"errors"
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

var prometheusAutoConfig = &PrometheusAutoConfiguration{}

func init() {
	boot.RegisterAutoConfigWith(prometheusAutoConfig,
		boot.WithConditions(
			condition.OnProperty(PrometheusEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityMonitoringLayer)),
	)
	boot.RegisterStarter(prometheusAutoConfig)
}

// PrometheusAutoConfiguration Prometheus 监控自动配置类。
type PrometheusAutoConfiguration struct {
	logger   log.Logger
	registry *prometheus.Registry
	server   *http.Server
	config   *PrometheusConfig
	ctx      context.Context
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
		return fmt.Errorf("failed to load Prometheus config: %w", err)
	}

	c.config = cfg
	c.registry = prometheus.NewRegistry()

	// 存储应用上下文
	c.ctx = ctx.Context()

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
		return fmt.Errorf("failed to register Prometheus Registry: %w", err)
	}

	if err := ctx.Container().RegisterInstance(c.server, reflect.TypeFor[*http.Server]()); err != nil {
		return fmt.Errorf("failed to register Prometheus Server: %w", err)
	}

	c.logger.Info(ctx.Context(), "Prometheus monitoring configured",
		log.KeyValue{Key: "port", Value: cfg.Port},
		log.KeyValue{Key: "path", Value: cfg.MetricsPath},
	)

	return nil
}

// Start 启动 Prometheus 监控服务器。
func (c *PrometheusAutoConfiguration) Start(ctx boot.ApplicationContext) error {
	if c.server == nil {
		return nil
	}
	go func() {
		if err := c.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.logger.Error(context.Background(), "prometheus metrics server error",
				log.KeyValue{Key: "error", Value: err.Error()},
			)
		}
	}()
	c.logger.Info(ctx.Context(), "prometheus metrics server started",
		log.KeyValue{Key: "addr", Value: c.server.Addr},
	)
	return nil
}

// Stop 停止 Prometheus 监控服务器。
func (c *PrometheusAutoConfiguration) Stop(ctx boot.ApplicationContext) error {
	var sCtx context.Context
	var cancel context.CancelFunc
	if ctx != nil {
		sCtx, cancel = context.WithTimeout(ctx.Context(), 30*time.Second)
	} else {
		sCtx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	}
	defer cancel()
	return c.server.Shutdown(sCtx)
}

// Name 返回启动器名称。
func (c *PrometheusAutoConfiguration) Name() string {
	return "PrometheusStarter"
}

// Dependencies 返回依赖的其他启动器名称。
func (c *PrometheusAutoConfiguration) Dependencies() []string {
	return nil
}

// GetCondition 返回启动器条件。
func (c *PrometheusAutoConfiguration) GetCondition() condition.Condition {
	return condition.OnProperty(PrometheusEnabled, ConditionTrue)
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
		return nil, fmt.Errorf("failed to bind Prometheus config: %w", err)
	}

	return cfg, nil
}
