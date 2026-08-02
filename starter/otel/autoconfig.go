// Package otel provides OpenTelemetry tracing auto-configuration.
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

// OtelAutoConfiguration OpenTelemetry auto-configuration.
type OtelAutoConfiguration struct {
	logger         log.Logger
	tracerProvider *sdktrace.TracerProvider
}

// Configure configures OpenTelemetry tracing.
func (c *OtelAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load OpenTelemetry config: %w", err)
	}

	otelCtx, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	exporter, err := otlptracegrpc.New(otelCtx,
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	res, err := resource.New(otelCtx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
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
		return fmt.Errorf("failed to register TracerProvider: %w", err)
	}

	c.logger.Info(ctx.Context(), "OpenTelemetry tracing enabled",
		log.KeyValue{Key: "endpoint", Value: cfg.Endpoint},
		log.KeyValue{Key: "service_name", Value: cfg.ServiceName},
	)

	return nil
}

// Shutdown closes OpenTelemetry.
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

// OtelConfig OpenTelemetry config.
type OtelConfig struct {
	Enabled        bool    `json:"enabled" mapstructure:"enabled"`
	Endpoint       string  `json:"endpoint" mapstructure:"endpoint"`
	ServiceName    string  `json:"service_name" mapstructure:"service_name"`
	ServiceVersion string  `json:"service_version" mapstructure:"service_version"`
	SamplingRate   float64 `json:"sampling_rate" mapstructure:"sampling_rate"`
}

// Configuration constants.
const (
	OtelEnabled           = "otel.enabled"
	DefaultOtelEndpoint   = "localhost:4317"
	DefaultServiceName    = "enhance-app"
	DefaultServiceVersion = "1.0.0"
	DefaultSamplingRate   = 1.0
	ConditionTrue         = "true"
)

// loadConfig loads OpenTelemetry config from Environment.
func (c *OtelAutoConfiguration) loadConfig(env *environment.Environment) (*OtelConfig, error) {
	cfg := &OtelConfig{
		Endpoint:       DefaultOtelEndpoint,
		ServiceName:    DefaultServiceName,
		ServiceVersion: DefaultServiceVersion,
		SamplingRate:   DefaultSamplingRate,
	}

	if err := env.BindPrefix("otel", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind OpenTelemetry config: %w", err)
	}

	return cfg, nil
}
