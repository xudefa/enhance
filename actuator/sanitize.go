package actuator

import (
	"encoding/base64"
	"strings"
	"sync/atomic"
)

const redactedValue = "***REDACTED***"

// strategyList 包装器，使 slice 可用于 atomic.Value 的 CAS 操作
type strategyList struct {
	list []SanitizeStrategy
}

// Sanitizer 敏感信息检测器
//
// 使用策略模式管理多种敏感信息检测规则，
// 支持自定义检测策略。
// 使用 atomic.Value 优化高频读取场景的性能。
type Sanitizer struct {
	keywords   atomic.Value // []string
	strategies atomic.Value // *strategyList
}

// NewSanitizer 创建敏感信息检测器
func NewSanitizer() *Sanitizer {
	s := &Sanitizer{}
	s.keywords.Store(defaultKeywords())
	s.strategies.Store(&strategyList{list: []SanitizeStrategy{}})
	return s
}

// AddStrategy 添加自定义检测策略
func (s *Sanitizer) AddStrategy(strategy SanitizeStrategy) {
	// 使用 CAS 无锁算法添加策略
	for {
		old, _ := s.strategies.Load().(*strategyList)
		newList := make([]SanitizeStrategy, len(old.list)+1)
		copy(newList, old.list)
		newList[len(old.list)] = strategy
		newStrategyList := &strategyList{list: newList}
		if s.strategies.CompareAndSwap(old, newStrategyList) {
			return
		}
	}
}

// Sanitize 掩盖敏感值
func (s *Sanitizer) Sanitize(key string, value any) any {
	if value == nil {
		return value
	}

	// 检查自定义策略（无锁读取）
	sl, _ := s.strategies.Load().(*strategyList)
	strategies := sl.list
	for _, strategy := range strategies {
		if strategy.IsSensitive(key, value) {
			return redactedValue
		}
	}

	// 检查关键词（无锁读取）
	keyLower := strings.ToLower(key)
	keywords, _ := s.keywords.Load().([]string)
	for _, keyword := range keywords {
		if strings.Contains(keyLower, keyword) {
			return redactedValue
		}
	}

	// 检查值本身
	if str, ok := value.(string); ok {
		if looksLikeSensitiveData(str) {
			return redactedValue
		}
	}

	return value
}

// defaultKeywords 返回默认敏感关键词列表
func defaultKeywords() []string {
	return []string{
		"password", "secret", "token", "key", "auth",
		"credential", "private", "api_key", "access_token",
		"client_secret", "oauth", "bearer", "jwt",
	}
}

// looksLikeSensitiveData 检查字符串是否看起来像敏感数据
func looksLikeSensitiveData(value string) bool {
	if len(value) > 32 {
		if strings.HasPrefix(value, "-----BEGIN") && strings.Contains(value, "PRIVATE KEY") {
			return true
		}
		if isRandomLookingString(value) {
			return true
		}
		if isTokenFormat(value) {
			return true
		}
	}
	return false
}

// isRandomLookingString 检查字符串是否看起来像随机字符串
func isRandomLookingString(s string) bool {
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial && len(s) > 16
}

// isTokenFormat 检查字符串是否为令牌格式
func isTokenFormat(s string) bool {
	// JWT 令牌格式检测
	if strings.Count(s, ".") >= 2 {
		parts := strings.Split(s, ".")
		if len(parts) == 3 {
			for _, part := range parts {
				if !isValidBase64(part) {
					return false
				}
			}
			return true
		}
	}
	return false
}

// isValidBase64 检查字符串是否为有效的 Base64 编码
func isValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
