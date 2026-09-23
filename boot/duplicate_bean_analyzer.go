package boot

import (
	"errors"
	"strings"

	"github.com/xudefa/enhance/core"
)

// DuplicateBeanAnalyzer 重复 Bean 错误分析器
type DuplicateBeanAnalyzer struct{}

// NewDuplicateBeanAnalyzer 创建重复 Bean 错误分析器
func NewDuplicateBeanAnalyzer() *DuplicateBeanAnalyzer {
	return &DuplicateBeanAnalyzer{}
}

// CanAnalyze 检查是否能分析该错误
func (a *DuplicateBeanAnalyzer) CanAnalyze(err error) bool {
	return errors.Is(err, core.ErrBeanAlreadyExists) ||
		strings.Contains(err.Error(), "bean already exists:")
}

// Analyze 分析错误并返回失败报告
func (a *DuplicateBeanAnalyzer) Analyze(err error) *FailureReport {
	return &FailureReport{
		Headline:    "Duplicate Bean Definition",
		Description: "Bean 定义重复，同一名称或类型已被注册",
		Action:      "检查是否重复注册了相同的 Bean，或使用 Named 区分多个实例",
		Cause:       err.Error(),
		PossibleSolutions: []string{
			"检查是否在多处注册了相同名称的 Bean",
			"使用 Named(\"name\") 为多个实例指定不同名称",
			"使用 Primary() 标记主要实例",
			"检查自动配置是否重复注册",
		},
	}
}
