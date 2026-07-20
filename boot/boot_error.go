// Package boot 提供应用启动器功能，用于 enhance 框架。
package boot

import (
	"fmt"
	"strings"
)

// Error 实现 error 接口
func (e *BootError) Error() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "boot failed during %s: %v", e.Phase, e.Original)

	if e.Analyzed != "" {
		fmt.Fprintf(&sb, "\n\nAnalysis: %s", e.Analyzed)
	}

	if len(e.Suggestions) > 0 {
		fmt.Fprintf(&sb, "\n\nSuggestions:")
		for _, s := range e.Suggestions {
			fmt.Fprintf(&sb, "\n  - %s", s)
		}
	}

	return sb.String()
}

// Unwrap 实现 errors.Unwrap 接口
func (e *BootError) Unwrap() error {
	return e.Original
}

// NewBootError 创建结构化启动错误
func NewBootError(phase string, err error) *BootError {
	return &BootError{
		Phase:    phase,
		Original: err,
	}
}

// WithAnalysis 添加分析结果
func (e *BootError) WithAnalysis(analysis string) *BootError {
	e.Analyzed = analysis
	return e
}

// WithSuggestions 添加修复建议
func (e *BootError) WithSuggestions(suggestions ...string) *BootError {
	e.Suggestions = suggestions
	return e
}
