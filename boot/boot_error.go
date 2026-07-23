// Package boot 提供应用启动器功能，用于 enhance 框架。
package boot

import (
	"fmt"
	"strings"
)

// 错误码常量，用于程序化错误处理。
const (
	ErrCodeConfigLoad    = "BOOT_CONFIG_LOAD"    // 配置加载失败
	ErrCodeConfigCenter  = "BOOT_CONFIG_CENTER"  // 配置中心加载失败
	ErrCodeAutoConfig    = "BOOT_AUTO_CONFIG"    // 自动配置执行失败
	ErrCodeModuleInstall = "BOOT_MODULE_INSTALL" // 模块安装失败
	ErrCodeStarterConfig = "BOOT_STARTER_CONFIG" // Starter 配置失败
	ErrCodeStarterStart  = "BOOT_STARTER_START"  // Starter 启动失败
	ErrCodeLifecycle     = "BOOT_LIFECYCLE"      // 生命周期阶段错误
	ErrCodeUnknown       = "BOOT_UNKNOWN"        // 未知错误
)

// bootError BootError 接口的具体实现。
//
// 包含错误码、阶段信息、原始错误、分析结果和修复建议，便于调试和错误处理。
type bootError struct {
	code        string   // 错误码
	message     string   // 错误消息
	phase       string   // 错误发生的阶段
	original    error    // 原始错误
	analyzed    string   // FailureAnalyzer 分析结果
	suggestions []string // 修复建议
}

// Code 返回错误码，用于程序化错误处理。
func (e *bootError) Code() string {
	return e.code
}

// Message 返回人类可读的错误消息。
func (e *bootError) Message() string {
	return e.message
}

// Cause 返回原始错误，用于错误链追踪。
func (e *bootError) Cause() error {
	return e.original
}

// Error 实现 error 接口。
func (e *bootError) Error() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "[%s] boot failed during %s: %v", e.code, e.phase, e.original)

	if e.analyzed != "" {
		fmt.Fprintf(&sb, "\n\nAnalysis: %s", e.analyzed)
	}

	if len(e.suggestions) > 0 {
		fmt.Fprintf(&sb, "\n\nSuggestions:")
		for _, s := range e.suggestions {
			fmt.Fprintf(&sb, "\n  - %s", s)
		}
	}

	return sb.String()
}

// Unwrap 实现 errors.Unwrap 接口，支持 errors.Is/As。
func (e *bootError) Unwrap() error {
	return e.original
}

// Phase 返回错误发生的阶段。
func (e *bootError) Phase() string {
	return e.phase
}

// Analyzed 返回 FailureAnalyzer 分析结果。
func (e *bootError) Analyzed() string {
	return e.analyzed
}

// Suggestions 返回修复建议列表（返回副本）。
func (e *bootError) Suggestions() []string {
	if len(e.suggestions) == 0 {
		return nil
	}
	result := make([]string, len(e.suggestions))
	copy(result, e.suggestions)
	return result
}

// NewBootErr 创建结构化启动错误（推荐使用）。
//
// 参数:
//   - code: 错误码，用于程序化错误处理（如 ErrCodeConfigLoad）
//   - phase: 错误发生的阶段（如 "初始化"、"启动"、"停止"）
//   - err: 原始错误
//
// 返回值:
//   - BootError: 结构化错误接口，支持 Error/Unwrap/Code/Message/Cause
//
// 示例：
//
//	return boot.NewBootErr(boot.ErrCodeConfigLoad, "初始化", err)
func NewBootErr(code, phase string, err error) BootError {
	return &bootError{
		code:     code,
		message:  err.Error(),
		phase:    phase,
		original: err,
	}
}

// NewBootErrf 创建带格式化消息的结构化启动错误。
//
// 参数:
//   - code: 错误码，用于程序化错误处理（如 ErrCodeAutoConfig）
//   - phase: 错误发生的阶段（如 "初始化"、"启动"、"停止"）
//   - format: 格式化消息模板
//   - args: 格式化参数
//
// 可选地传入一个底层错误作为最后参数，支持 errors.Is/As 错误链追踪：
//
//	return boot.NewBootErrf(boot.ErrCodeAutoConfig, "初始化", "自动配置 %T 失败: %v", config, rootErr)
func NewBootErrf(code, phase, format string, args ...any) BootError {
	// 支持可选的底层错误参数：如果最后一个参数是 error，则作为 Cause
	var original error
	if len(args) > 0 {
		if err, ok := args[len(args)-1].(error); ok {
			original = err
			args = args[:len(args)-1]
		}
	}

	return &bootError{
		code:     code,
		message:  fmt.Sprintf(format, args...),
		phase:    phase,
		original: original,
	}
}


