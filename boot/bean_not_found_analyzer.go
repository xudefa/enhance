package boot

import (
	"errors"
	"strings"

	"github.com/xudefa/enhance/core"
)

// BeanNotFoundAnalyzer Bean 未找到错误分析器
type BeanNotFoundAnalyzer struct{}

// NewBeanNotFoundAnalyzer 创建 Bean 未找到错误分析器
func NewBeanNotFoundAnalyzer() *BeanNotFoundAnalyzer {
	return &BeanNotFoundAnalyzer{}
}

// CanAnalyze 检查是否能分析该错误
func (a *BeanNotFoundAnalyzer) CanAnalyze(err error) bool {
	return errors.Is(err, core.ErrBeanNotFound) ||
		strings.Contains(err.Error(), "bean not found")
}

// Analyze 分析错误并返回失败报告
func (a *BeanNotFoundAnalyzer) Analyze(err error) *FailureReport {
	return &FailureReport{
		Headline:    "Bean Not Found",
		Description: "无法找到所需的 Bean 实例",
		Action:      "检查 Bean 是否已正确注册，或检查条件装配是否满足",
		Cause:       err.Error(),
		PossibleSolutions: []string{
			"确认 Bean 已通过 Register/Constructor/Factory 注册",
			"检查是否使用了 @Component 标签且包扫描已启用",
			"检查条件装配（OnProperty/OnBean/OnClass）是否满足",
			"确认 Bean 名称拼写正确",
			"检查是否存在循环依赖导致 Bean 未创建",
		},
	}
}
