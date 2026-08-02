package tracing

import "errors"

// tracing 包错误定义。
var (
	// ErrExporterNotSet 导出器未设置错误。
	ErrExporterNotSet = errors.New("exporter not set")

	// ErrInvalidSamplingRate 无效的采样率错误。
	ErrInvalidSamplingRate = errors.New("invalid sampling rate: must be between 0.0 and 1.0")
)
