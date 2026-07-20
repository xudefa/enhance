// Package otel 提供 OpenTelemetry 链路追踪自动配置。
package otel

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&OtelAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(OtelEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityMonitoringLayer)),
	)
}

// OtelAutoConfiguration OpenTelemetry 自动配置类。
type OtelAutoConfiguration struct {
	logger         log.Logger
	tracerProvider *sdktrace.TracerProvider
}

// Configure 配置 OpenTelemetry 链路追踪。
func (c *OtelAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("加载 OpenTelemetry 配置失败: %w", err)
	}

	if !cfg.Enabled {
		c.logger.Info(context.Background(), "OpenTelemetry 未启用，跳过配置")
		return nil
	}

	otelCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exporter, err := otlptracegrpc.New(otelCtx,
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("创建 OTLP 导出器失败: %w", err)
	}

	res, err := resource.New(otelCtx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return fmt.Errorf("创建资源失败: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SamplingRate)),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	c.tracerProvider = tracerProvider

	if err := ctx.Container().RegisterInstance(tracerProvider, reflect.TypeFor[*sdktrace.TracerProvider]()); err != nil {
		return fmt.Errorf("注册 TracerProvider 失败: %w", err)
	}

	c.logger.Info(context.Background(), "OpenTelemetry 链路追踪已启用",
		log.KeyValue{Key: "endpoint", Value: cfg.Endpoint},
		log.KeyValue{Key: "service_name", Value: cfg.ServiceName},
	)

	return nil
}

// Shutdown 关闭 OpenTelemetry。
func (c *OtelAutoConfiguration) Shutdown(ctx context.Context) error {
	if c.tracerProvider != nil {
		return c.tracerProvider.Shutdown(ctx)
	}
	return nil
}

// GetTracer 获取 Tracer 实例。
func (c *OtelAutoConfiguration) GetTracer(name string) *sdktrace.TracerProvider {
	return c.tracerProvider
}

// OtelConfig OpenTelemetry 配置。
type OtelConfig struct {
	Enabled        bool    `json:"enabled" mapstructure:"enabled"`
	Endpoint       string  `json:"endpoint" mapstructure:"endpoint"`
	ServiceName    string  `json:"service_name" mapstructure:"service_name"`
	ServiceVersion string  `json:"service_version" mapstructure:"service_version"`
	SamplingRate   float64 `json:"sampling_rate" mapstructure:"sampling_rate"`
}

// 配置常量。
const (
	OtelEnabled           = "otel.enabled"
	DefaultOtelEndpoint   = "localhost:4317"
	DefaultServiceName    = "enhance-app"
	DefaultServiceVersion = "1.0.0"
	DefaultSamplingRate   = 1.0
	ConditionTrue         = "true"
)

// loadConfig 从 Environment 加载 OpenTelemetry 配置。
func (c *OtelAutoConfiguration) loadConfig(env *environment.Environment) (*OtelConfig, error) {
	cfg := &OtelConfig{
		Endpoint:       DefaultOtelEndpoint,
		ServiceName:    DefaultServiceName,
		ServiceVersion: DefaultServiceVersion,
		SamplingRate:   DefaultSamplingRate,
	}

	if err := env.BindPrefix("otel", cfg); err != nil {
		return nil, fmt.Errorf("绑定 OpenTelemetry 配置失败: %w", err)
	}

	return cfg, nil
}
