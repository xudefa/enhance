package log

import (
	"context"
	"math/rand/v2"
	"os"
)

// NewRandomSampler 创建随机采样器
func NewRandomSampler(rate float64) *RandomSampler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	return &RandomSampler{rate: rate}
}

// ShouldSample 判断是否采样
func (s *RandomSampler) ShouldSample() bool {
	return randomFloat() < s.rate
}

// NewThresholdSampler 创建阈值采样器
func NewThresholdSampler(threshold int64) *ThresholdSampler {
	if threshold <= 0 {
		threshold = 1
	}
	return &ThresholdSampler{threshold: threshold}
}

// ShouldSample 判断是否采样
func (s *ThresholdSampler) ShouldSample() bool {
	return s.counter.Add(1)%s.threshold == 0
}

// NewSampledLogger 创建带采样的日志器
func NewSampledLogger(logger Logger, sampler Sampler) *SampledLogger {
	return &SampledLogger{
		logger:  logger,
		sampler: sampler,
	}
}

// Debug 记录调试日志
func (l *SampledLogger) Debug(ctx context.Context, msg string, keys ...KeyValue) {
	if l.sampler.ShouldSample() {
		l.logger.Debug(ctx, msg, keys...)
	}
}

// Info 记录信息日志
func (l *SampledLogger) Info(ctx context.Context, msg string, keys ...KeyValue) {
	if l.sampler.ShouldSample() {
		l.logger.Info(ctx, msg, keys...)
	}
}

// Warn 记录警告日志
func (l *SampledLogger) Warn(ctx context.Context, msg string, keys ...KeyValue) {
	if l.sampler.ShouldSample() {
		l.logger.Warn(ctx, msg, keys...)
	}
}

// Error 记录错误日志（错误日志不采样，全部记录）
func (l *SampledLogger) Error(ctx context.Context, msg string, keys ...KeyValue) {
	l.logger.Error(ctx, msg, keys...)
}

// DPanic 记录致命错误日志并 panic
func (l *SampledLogger) DPanic(ctx context.Context, msg string, keys ...KeyValue) {
	if fl, ok := l.logger.(LoggerFatal); ok {
		fl.DPanic(ctx, msg, keys...)
		return
	}
	panic(msg)
}

// Panic 记录日志并 panic
func (l *SampledLogger) Panic(ctx context.Context, msg string, keys ...KeyValue) {
	if fl, ok := l.logger.(LoggerFatal); ok {
		fl.Panic(ctx, msg, keys...)
		return
	}
	panic(msg)
}

// Fatal 记录致命级别日志
func (l *SampledLogger) Fatal(ctx context.Context, msg string, keys ...KeyValue) {
	if fl, ok := l.logger.(LoggerFatal); ok {
		fl.Fatal(ctx, msg, keys...)
		return
	}
	os.Exit(1)
}

// Sync 同步日志缓冲区
func (l *SampledLogger) Sync() error {
	return l.logger.Sync()
}

// With 返回带有额外字段的日志记录器
func (l *SampledLogger) With(ctx context.Context, keys ...KeyValue) Logger {
	return &SampledLogger{
		logger:  l.logger.With(ctx, keys...),
		sampler: l.sampler,
	}
}

var _ Logger = (*SampledLogger)(nil)

// randomFloat 生成 0.0-1.0 之间的随机数。
func randomFloat() float64 {
	return rand.Float64()
}
