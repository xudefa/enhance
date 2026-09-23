package main

// Category 审计问题类别。
type Category string

const (
	// CategoryBadNames 命名语义化问题。
	CategoryBadNames Category = "bad_names"
	// CategoryMagicNumbers 魔法数字/魔法字符串问题。
	CategoryMagicNumbers Category = "magic_numbers"
	// CategoryErrorContext 错误信息上下文问题。
	CategoryErrorContext Category = "error_context"
	// CategoryDocumentation 注释与文档问题。
	CategoryDocumentation Category = "documentation"
	// CategoryMaintainability 可维护性结构问题。
	CategoryMaintainability Category = "maintainability"
)

// Severity 严重程度：决定是否需要人工决策。
type Severity string

const (
	// SeverityMechanical 机械可改：局部修改即可，可安全批量处理。
	SeverityMechanical Severity = "Mechanical"
	// SeverityBreaking 需决策：可能影响调用方或需拆包，必须人工确认。
	SeverityBreaking Severity = "Breaking"
)

// Finding 表示单条审计发现。
type Finding struct {
	File       string   `json:"file"`
	Line       int      `json:"line"`
	Category   Category `json:"category"`
	Severity   Severity `json:"severity"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion"`
}
