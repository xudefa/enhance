package boot

import (
	"errors"
	"strings"

	"github.com/xudefa/enhance/core"
)

// CircularDependencyAnalyzer 循环依赖错误分析器
type CircularDependencyAnalyzer struct{}

// NewCircularDependencyAnalyzer 创建循环依赖错误分析器
func NewCircularDependencyAnalyzer() *CircularDependencyAnalyzer {
	return &CircularDependencyAnalyzer{}
}

// CanAnalyze 检查是否能分析该错误
func (a *CircularDependencyAnalyzer) CanAnalyze(err error) bool {
	return errors.Is(err, core.ErrCircularDependency) ||
		strings.Contains(err.Error(), "circular dependency")
}

// Analyze 分析错误并返回失败报告
func (a *CircularDependencyAnalyzer) Analyze(err error) *FailureReport {
	return &FailureReport{
		Headline:    "Circular Dependency Detected",
		Description: "检测到循环依赖，Bean A 依赖 Bean B，Bean B 又依赖 Bean A",
		Action:      "使用懒加载注入（lazy）或重新设计依赖关系",
		Cause:       err.Error(),
		PossibleSolutions: []string{
			"在一方使用懒加载注入：`inject:\"name,lazy\"`",
			"使用 Provider 模式延迟获取依赖",
			"重新设计架构，避免循环依赖",
			"使用事件驱动替代直接依赖",
			"引入中间层解耦",
		},
	}
}
