// Package log 提供日志管理功能，用于 enhance 框架。
package log

import (
	"context"
)

// LoggerBuilder 日志器构建器
type LoggerBuilder struct {
	level      Level
	format     string
	addSource  bool
	outputPath string
	sampler    Sampler
	name       string
}

// NewLoggerBuilder 创建日志器构建器
func NewLoggerBuilder() *LoggerBuilder {
	return &LoggerBuilder{
		level:  InfoLevel,
		format: "json",
	}
}

// Level 设置日志级别
func (b *LoggerBuilder) Level(level Level) *LoggerBuilder {
	b.level = level
	return b
}

// Format 设置输出格式（json 或 text）
func (b *LoggerBuilder) Format(format string) *LoggerBuilder {
	b.format = format
	return b
}

// AddSource 设置是否添加源码位置
func (b *LoggerBuilder) AddSource(addSource bool) *LoggerBuilder {
	b.addSource = addSource
	return b
}

// OutputPath 设置日志文件输出路径
func (b *LoggerBuilder) OutputPath(path string) *LoggerBuilder {
	b.outputPath = path
	return b
}

// Sampler 设置采样策略
func (b *LoggerBuilder) Sampler(sampler Sampler) *LoggerBuilder {
	b.sampler = sampler
	return b
}

// Name 设置日志器名称
func (b *LoggerBuilder) Name(name string) *LoggerBuilder {
	b.name = name
	return b
}

// Build 构建日志器
func (b *LoggerBuilder) Build() Logger {
	opts := []Option{
		WithLevel(b.level),
		WithFormat(b.format),
		WithAddSource(b.addSource),
	}

	if b.outputPath != "" {
		opts = append(opts, WithOutputPath(b.outputPath))
	}

	var logger Logger = NewSlogLogger(opts...)

	if b.sampler != nil {
		logger = NewSampledLogger(logger, b.sampler)
	}

	if b.name != "" {
		// 使用 context.Background() 添加 logger 名称字段
		logger = logger.With(context.Background(), KeyValue{Key: "logger", Value: b.name})
	}

	return logger
}
