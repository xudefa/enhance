// Package i18n 提供国际化支持，用于 enhance 框架。
package i18n

import (
	"sync"
)

// ResourceBundleMessageSource 基于资源包的消息源实现。
type ResourceBundleMessageSource struct {
	bundles  sync.Map // map[string]map[string]string，使用 sync.Map 优化并发读取
	mu       sync.RWMutex
	fallback MessageSource
}

// StaticMessageSource 静态消息源，适用于测试或简单场景。
type StaticMessageSource struct {
	messages sync.Map // map[string]string，使用 sync.Map 优化并发读取
}

// DelegatingMessageSource 委托消息源，支持父子层级结构。
type DelegatingMessageSource struct {
	mu       sync.RWMutex
	parent   MessageSource
	children []MessageSource
}

// String 返回区域的字符串表示，格式为 "language_country"。
func (l Locale) String() string {
	if l.Country != "" {
		return l.Language + "_" + l.Country
	}
	return l.Language
}

// NewResourceBundleMessageSource 创建一个新的资源包消息源。
func NewResourceBundleMessageSource() *ResourceBundleMessageSource {
	return &ResourceBundleMessageSource{}
}

// AddResourceBundle 为指定区域添加消息映射表。
func (m *ResourceBundleMessageSource) AddResourceBundle(locale Locale, messages map[string]string) {
	m.bundles.Store(locale.String(), messages)
}

// SetFallback 设置回退消息源，当当前消息源找不到消息时会委托给 fallback。
func (m *ResourceBundleMessageSource) SetFallback(fallback MessageSource) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fallback = fallback
}

// GetMessage 使用默认区域获取消息。
func (m *ResourceBundleMessageSource) GetMessage(code string, args ...any) string {
	return m.GetMessageWithLocale(code, DefaultLocale, args...)
}

// GetMessageWithLocale 使用指定区域获取消息。
func (m *ResourceBundleMessageSource) GetMessageWithLocale(code string, locale Locale, args ...any) string {
	if value, ok := m.bundles.Load(locale.String()); ok {
		if bundle, ok := value.(map[string]string); ok {
			if msg, ok := bundle[code]; ok {
				return formatMessage(msg, args)
			}
		}
	}

	if locale.Country != "" {
		langOnly := locale.Language
		if value, ok := m.bundles.Load(langOnly); ok {
			if bundle, ok := value.(map[string]string); ok {
				if msg, ok := bundle[code]; ok {
					return formatMessage(msg, args)
				}
			}
		}
	}

	m.mu.RLock()
	fallback := m.fallback
	m.mu.RUnlock()

	if fallback != nil {
		return fallback.GetMessageWithLocale(code, locale, args...)
	}

	return code
}

// NewStaticMessageSource 创建一个新的静态消息源。
func NewStaticMessageSource() *StaticMessageSource {
	return &StaticMessageSource{}
}

// AddMessage 添加一条消息。
func (m *StaticMessageSource) AddMessage(code, message string) {
	m.messages.Store(code, message)
}

// AddMessages 批量添加消息。
func (m *StaticMessageSource) AddMessages(messages map[string]string) {
	for code, msg := range messages {
		m.messages.Store(code, msg)
	}
}

// GetMessage 获取消息，如果找不到则返回消息代码本身。
func (m *StaticMessageSource) GetMessage(code string, args ...any) string {
	if value, ok := m.messages.Load(code); ok {
		if msg, ok := value.(string); ok {
			return formatMessage(msg, args)
		}
	}
	return code
}

// GetMessageWithLocale 获取指定区域的消息（静态源忽略区域参数）。
func (m *StaticMessageSource) GetMessageWithLocale(code string, locale Locale, args ...any) string {
	return m.GetMessage(code, args...)
}

// NewDelegatingMessageSource 创建一个新的委托消息源。
func NewDelegatingMessageSource() *DelegatingMessageSource {
	return &DelegatingMessageSource{}
}

// SetParent 设置父消息源。
func (m *DelegatingMessageSource) SetParent(parent MessageSource) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.parent = parent
}

// AddChild 添加子消息源。
func (m *DelegatingMessageSource) AddChild(child MessageSource) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.children = append(m.children, child)
}

// GetMessage 使用默认区域获取消息。
func (m *DelegatingMessageSource) GetMessage(code string, args ...any) string {
	return m.GetMessageWithLocale(code, DefaultLocale, args...)
}

// GetMessageWithLocale 使用指定区域获取消息。
func (m *DelegatingMessageSource) GetMessageWithLocale(code string, locale Locale, args ...any) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, child := range m.children {
		msg := child.GetMessageWithLocale(code, locale, args...)
		if msg != code {
			return msg
		}
	}

	if m.parent != nil {
		return m.parent.GetMessageWithLocale(code, locale, args...)
	}

	return code
}
