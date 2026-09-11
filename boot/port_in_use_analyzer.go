package boot

import "strings"

// PortInUseAnalyzer 端口占用错误分析器
type PortInUseAnalyzer struct{}

// NewPortInUseAnalyzer 创建端口占用错误分析器
func NewPortInUseAnalyzer() *PortInUseAnalyzer {
	return &PortInUseAnalyzer{}
}

// CanAnalyze 检查是否能分析该错误
func (a *PortInUseAnalyzer) CanAnalyze(err error) bool {
	return strings.Contains(err.Error(), "address already in use") ||
		(strings.Contains(err.Error(), "port") && strings.Contains(err.Error(), "bind"))
}

// Analyze 分析错误并返回失败报告
func (a *PortInUseAnalyzer) Analyze(err error) *FailureReport {
	return &FailureReport{
		Headline:    "Port Already in Use",
		Description: "服务器端口已被占用，无法启动",
		Action:      "检查端口是否被其他进程占用，或更换端口",
		Cause:       err.Error(),
		PossibleSolutions: []string{
			"使用 `lsof -i :<port>` 查看占用端口的进程",
			"使用 `kill <pid>` 终止占用进程",
			"修改配置文件中的 server.port 更换端口",
			"设置环境变量 SERVER_PORT 覆盖配置",
		},
	}
}
