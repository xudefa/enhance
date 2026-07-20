// Package i18n 提供国际化支持，用于 enhance 框架。
//
// 该模块提供多语言消息管理、消息格式化、语言环境切换等功能。
// 参考 Spring 的 MessageSource 设计。
//
// # 架构设计
//
//   - MessageSource: 消息源接口，提供消息获取
//   - Locale: 语言环境，表示特定的语言和地区
//
// # 核心功能
//
//   - 多语言消息: 支持多种语言的消息配置
//   - 消息格式化: 支持参数替换和消息格式化
//   - 语言切换: 支持运行时切换语言环境
//   - 回退机制: 支持消息缺失时的回退策略
//
// # 使用方式
//
// 创建消息源：
//
//	source := i18n.NewResourceBundleMessageSource()
//	source.AddResourceBundle(i18n.Locale{Language: "zh", Country: "CN"}, map[string]string{
//	    "user.welcome": "欢迎使用 enhance 框架",
//	})
//
// 获取消息：
//
//	msg := source.GetMessage("user.welcome")
//	// 输出: 欢迎使用 enhance 框架
//
// 带参数的消息：
//
//	msg := source.GetMessage("user.greeting", "张三")
//	// 输出: 你好，张三！
//
// # 回退机制
//
// 消息查找按以下顺序回退：
//  1. 精确匹配区域（如 "zh_CN"）
//  2. 语言匹配（如 "zh"）
//  3. 回退到父消息源（如果设置了 fallback）
//  4. 返回消息代码本身
package i18n

import (
	"fmt"
)

// Locale 表示语言和区域设置。
//
// Language 为语言代码（如 "en", "zh"），
// Country 为国家/地区代码（如 "US", "CN"），
// Variant 为变体标识（可选，如 "Hans" 简体、"Hant" 繁体）。
type Locale struct {
	Language string // 语言代码，如 "en", "zh", "ja"
	Country  string // 国家/地区代码，如 "US", "CN", "JP"
	Variant  string // 变体标识，如 "Hans", "Hant"
}

// DefaultLocale 是默认区域设置（英语）。
//
// 当未指定区域时，使用此区域获取消息。
var DefaultLocale = Locale{Language: "en"}

// MessageSource 消息源接口，用于获取国际化消息。
//
// 支持消息模板和参数格式化。
// 实现应支持多区域和回退机制。
type MessageSource interface {
	// GetMessage 使用默认区域获取消息，args 用于格式化消息模板。
	GetMessage(code string, args ...any) string

	// GetMessageWithLocale 使用指定区域获取消息。
	GetMessageWithLocale(code string, locale Locale, args ...any) string
}

// formatMessage 格式化消息模板。
func formatMessage(msg string, args []any) string {
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}
