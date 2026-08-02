package failure

import (
	"fmt"
	"strings"
)

// GetSuggestions 根据失败分析结果生成修复建议列表。
//
// 根据分析结果中的 Components 和 Description 生成针对性的建议。
// 返回值为建议字符串切片，按优先级排序。
//
// 参数:
//   - analysis: 失败分析结果，不能为 nil
//
// 返回:
//   - []string: 修复建议列表
func GetSuggestions(analysis *FailureAnalysis) []string {
	if analysis == nil {
		return nil
	}

	suggestions := make([]string, 0, 4)

	switch {
	case containsComponent(analysis.Components, "network"):
		suggestions = append(suggestions,
			"使用 lsof -i :<port> 查看占用端口的进程",
			"使用 kill <pid> 终止占用进程",
			"修改配置文件中的 server.port 更换端口",
		)
	case containsComponent(analysis.Components, "filesystem") &&
		containsComponent(analysis.Components, "config"):
		suggestions = append(suggestions,
			"检查配置文件路径是否正确",
			"确认配置文件格式（JSON/YAML）无语法错误",
			"检查环境变量是否已正确设置",
		)
	case containsComponent(analysis.Components, "filesystem"):
		suggestions = append(suggestions,
			"检查文件或目录路径是否正确",
			"确认当前用户具有足够的访问权限",
			"使用 chmod 修改文件权限",
		)
	default:
		suggestions = append(suggestions, "查看错误日志获取详细信息")
	}

	return suggestions
}

// FormatFailureAnalysis 格式化失败分析结果为可读字符串。
//
// 将分析结果格式化为包含描述、建议动作和修复建议的多行字符串。
//
// 参数:
//   - analysis: 失败分析结果，不能为 nil
//
// 返回:
//   - string: 格式化后的字符串
func FormatFailureAnalysis(analysis *FailureAnalysis) string {
	if analysis == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("\n====================\n")
	sb.WriteString("APPLICATION FAILED TO START\n")
	sb.WriteString("====================\n\n")
	sb.WriteString("Description: ")
	sb.WriteString(analysis.Description)
	sb.WriteString("\n\nAction: ")
	sb.WriteString(analysis.Action)

	suggestions := GetSuggestions(analysis)
	if len(suggestions) > 0 {
		sb.WriteString("\n\nSuggestions:\n")
		for i, s := range suggestions {
			fmt.Fprintf(&sb, "  %d. %s\n", i+1, s)
		}
	}

	return sb.String()
}

// containsComponent 检查组件列表中是否包含指定组件。
func containsComponent(components []string, target string) bool {
	for _, c := range components {
		if c == target {
			return true
		}
	}
	return false
}
