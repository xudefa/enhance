package failure

import (
	"errors"
	"os"
	"strings"
)

// DefaultFailureAnalyzer 默认失败分析器。
//
// 分析常见的应用启动错误，包括端口占用、权限拒绝和文件未找到。
// 每种错误类型都提供详细的描述和可操作的建议。
type DefaultFailureAnalyzer struct{}

// NewDefaultFailureAnalyzer 创建默认失败分析器。
//
// 返回值:
//   - FailureAnalyzer: 默认失败分析器实例
func NewDefaultFailureAnalyzer() FailureAnalyzer {
	return &DefaultFailureAnalyzer{}
}

// Supports 判断是否支持分析该错误。
//
// 支持以下错误类型：
//   - 端口占用（address already in use）
//   - 权限拒绝（permission denied）
//   - 文件未找到（file not found / no such file）
func (a *DefaultFailureAnalyzer) Supports(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "address already in use") ||
		strings.Contains(msg, "permission denied") || errors.Is(err, os.ErrPermission) ||
		strings.Contains(msg, "file not found") ||
		strings.Contains(msg, "no such file")
}

// Analyze 分析错误并返回失败分析结果。
//
// 根据错误类型生成对应的 FailureAnalysis，包含错误描述、建议动作和相关组件。
func (a *DefaultFailureAnalyzer) Analyze(err error) *FailureAnalysis {
	if err == nil {
		return nil
	}

	msg := err.Error()

	switch {
	case strings.Contains(msg, "address already in use"):
		return a.analyzePortInUse(err)
	case strings.Contains(msg, "permission denied") || errors.Is(err, os.ErrPermission):
		return a.analyzePermissionDenied(err)
	case strings.Contains(msg, "file not found") || strings.Contains(msg, "no such file"):
		return a.analyzeFileNotFound(err)
	}

	return nil
}

// analyzePortInUse 分析端口占用错误。
func (a *DefaultFailureAnalyzer) analyzePortInUse(err error) *FailureAnalysis {
	return &FailureAnalysis{
		Description: "服务器端口已被占用，无法启动",
		Action:      "检查端口是否被其他进程占用，或更换端口",
		Exception:   err,
		Components:  []string{"server", "network"},
	}
}

// analyzePermissionDenied 分析权限拒绝错误。
func (a *DefaultFailureAnalyzer) analyzePermissionDenied(err error) *FailureAnalysis {
	return &FailureAnalysis{
		Description: "权限不足，无法访问所需资源",
		Action:      "检查文件或目录的访问权限，或以适当权限运行应用",
		Exception:   err,
		Components:  []string{"filesystem"},
	}
}

// analyzeFileNotFound 分析文件未找到错误。
func (a *DefaultFailureAnalyzer) analyzeFileNotFound(err error) *FailureAnalysis {
	return &FailureAnalysis{
		Description: "配置文件或资源文件未找到",
		Action:      "检查文件路径是否正确，确保文件存在",
		Exception:   err,
		Components:  []string{"filesystem", "config"},
	}
}
