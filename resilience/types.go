// Package resilience 提供弹性容错功能，用于 enhance 框架。
package resilience

// String 返回状态的字符串表示。
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}
