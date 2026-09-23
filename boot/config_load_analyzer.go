package boot

import (
	"errors"
	"strings"
)

// ConfigLoadAnalyzer 配置加载错误分析器
type ConfigLoadAnalyzer struct{}

// NewConfigLoadAnalyzer 创建配置加载错误分析器
func NewConfigLoadAnalyzer() *ConfigLoadAnalyzer {
	return &ConfigLoadAnalyzer{}
}

// CanAnalyze 检查是否能分析该错误
func (a *ConfigLoadAnalyzer) CanAnalyze(err error) bool {
	return errors.Is(err, ErrPropertyNotFound) ||
		errors.Is(err, ErrTypeConversion) ||
		strings.Contains(err.Error(), "config") ||
		strings.Contains(err.Error(), "property")
}

// Analyze 分析错误并返回失败报告
func (a *ConfigLoadAnalyzer) Analyze(err error) *FailureReport {
	return &FailureReport{
		Headline:    "Configuration Load Failed",
		Description: "配置加载失败，无法获取所需的配置项",
		Action:      "检查配置文件是否存在，配置项名称是否正确",
		Cause:       err.Error(),
		PossibleSolutions: []string{
			"检查 application.json/yaml 配置文件是否存在",
			"确认配置项名称拼写正确（注意大小写）",
			"检查环境变量是否已正确设置",
			"使用 environment.GetRequiredProperty() 获取必填配置",
			"检查 Profile 配置是否正确激活",
		},
	}
}
