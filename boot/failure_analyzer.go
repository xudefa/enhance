// Package boot 提供应用启动器功能，用于 enhance 框架。
package boot

import (
	"fmt"
	"strings"
)

// NewSimpleFailureAnalyzer 创建简单失败分析器
//
// 参数：
//   - analyzeFn: 分析函数，接收错误返回失败报告，返回 nil 表示无法分析
func NewSimpleFailureAnalyzer(analyzeFn func(err error) *FailureReport) *SimpleFailureAnalyzer {
	return &SimpleFailureAnalyzer{analyzeFn: analyzeFn}
}

// CanAnalyze 检查分析函数是否能处理该错误
func (s *SimpleFailureAnalyzer) CanAnalyze(err error) bool {
	return s.analyzeFn(err) != nil
}

// Analyze 使用分析函数分析错误。
func (s *SimpleFailureAnalyzer) Analyze(err error) *FailureReport {
	return s.analyzeFn(err)
}

var globalAnalyzerRegistry = NewFailureAnalyzerRegistry()

// NewFailureAnalyzerRegistry 创建失败分析器注册表
func NewFailureAnalyzerRegistry() *FailureAnalyzerRegistry {
	return &FailureAnalyzerRegistry{}
}

// Register 注册失败分析器
func (r *FailureAnalyzerRegistry) Register(analyzer FailureAnalyzer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.analyzers = append(r.analyzers, analyzer)
}

// Analyze 分析错误，返回第一个匹配的失败报告
func (r *FailureAnalyzerRegistry) Analyze(err error) *FailureReport {
	r.mu.RLock()
	analyzers := make([]FailureAnalyzer, len(r.analyzers))
	copy(analyzers, r.analyzers)
	r.mu.RUnlock()

	for _, analyzer := range analyzers {
		if analyzer.CanAnalyze(err) {
			return analyzer.Analyze(err)
		}
	}
	return nil
}

// GlobalAnalyzerRegistry 返回全局失败分析器注册表
func GlobalAnalyzerRegistry() *FailureAnalyzerRegistry {
	return globalAnalyzerRegistry
}

// RegisterFailureAnalyzer 注册失败分析器到全局注册表
func RegisterFailureAnalyzer(analyzer FailureAnalyzer) {
	globalAnalyzerRegistry.Register(analyzer)
}

// formatFailure 格式化失败报告为可读字符串
func formatFailure(report *FailureReport) string {
	var result strings.Builder
	fmt.Fprintf(&result, `
====================
APPLICATION FAILED TO START
====================

描述: %s

动作: %s

原因: %s
`, report.Description, report.Action, report.Cause)
	if len(report.PossibleSolutions) > 0 {
		result.WriteString("\n可能的解决方案:\n")
		for i, sol := range report.PossibleSolutions {
			fmt.Fprintf(&result, "  %d. %s\n", i+1, sol)
		}
	}
	return result.String()
}

// ReportFailure 分析并格式化输出失败报告
//
// 如果没有匹配的分析器，返回简单的错误信息字符串。
func ReportFailure(err error) string {
	report := globalAnalyzerRegistry.Analyze(err)
	if report == nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return formatFailure(report)
}
