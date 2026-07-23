// Package failure 提供应用启动失败分析功能，用于 enhance 框架。
//
// 该包定义了失败分析器的接口和多种实现，在应用启动失败时提供友好的错误提示，
// 帮助开发者快速定位问题。参考 Spring Boot 的 FailureAnalyzer 设计。
//
// # 架构设计
//
//   - FailureAnalyzer: 失败分析器接口，定义错误分析和判断方法
//   - FailureAnalysis: 失败分析结果，包含错误描述、建议动作和相关组件
//   - DefaultFailureAnalyzer: 默认失败分析器，分析常见启动错误（端口占用、权限拒绝、文件未找到）
//   - GetSuggestions: 根据分析结果生成修复建议
//
// # 核心功能
//
//   - 错误分析: 在应用启动失败时分析错误类型并生成结构化报告
//   - 友好提示: 将技术错误转换为开发者可操作的建议
//   - 可扩展性: 通过 FailureAnalyzer 接口支持自定义分析器
//
// # 使用方式
//
// 使用默认分析器：
//
//	analyzer := failure.NewDefaultFailureAnalyzer()
//	if analyzer.Supports(err) {
//	    result := analyzer.Analyze(err)
//	    suggestions := failure.GetSuggestions(result)
//	}
package failure

// FailureAnalyzer 失败分析器接口。
//
// 参考 Spring Boot 的 FailureAnalyzer。
// 在应用启动失败时分析错误，生成结构化的失败分析结果。
type FailureAnalyzer interface {
	// Analyze 分析错误并返回失败分析结果。
	//
	// 参数 err 为需要分析的错误。
	// 返回 FailureAnalysis 包含错误描述、建议动作和相关组件信息。
	Analyze(err error) *FailureAnalysis

	// Supports 判断是否支持分析该错误。
	//
	// 参数 err 为需要判断的错误。
	// 返回 true 表示可以分析该错误，false 表示不支持。
	Supports(err error) bool
}

// FailureAnalysis 失败分析结果。
//
// 参考 Spring Boot 的 FailureAnalysis，在应用启动失败时提供结构化的错误信息。
// 包含错误描述、建议动作、原始异常和相关组件。
type FailureAnalysis struct {
	// Description 错误描述，说明发生了什么问题。
	Description string

	// Action 建议操作，说明开发者应该采取什么措施。
	Action string

	// Exception 原始异常，保留原始错误以便调试。
	Exception error

	// Components 相关组件列表，标识涉及的组件或模块。
	Components []string
}
