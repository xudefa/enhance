// Package email 提供邮件发送功能，用于 enhance 框架。
//
// 该模块提供统一的 SMTP 邮件发送接口，使用 Go 原生 net/smtp 库。
// 支持纯文本和 HTML 格式邮件发送。
//
// # 架构设计
//
//   - Sender: 邮件发送器接口，封装 SMTP 协议
//   - Message: 邮件消息结构，包含收件人、主题、正文等
//   - Attachment: 邮件附件结构
//   - SenderOption: 发送器配置选项函数
//
// # 核心功能
//
//   - SMTP 发送: 支持标准 SMTP 协议发送邮件
//   - 多收件人: 支持发送给多个收件人
//   - HTML 邮件: 支持 HTML 格式邮件正文
//   - 环境变量: 支持从环境变量读取 SMTP 密码
//
// # 使用方式
//
// 创建发送器：
//
//	sender := email.NewSender()
//
// 发送邮件：
//
//	err := sender.Send(ctx, &email.Message{
//	    To:      []string{"recipient@example.com"},
//	    Subject: "测试",
//	    Body:    "你好！",
//	})
//
// # 配置属性
//
//   - email.smtp.host: SMTP 服务器地址（默认 localhost）
//   - email.smtp.port: SMTP 端口（默认 25）
//   - email.from: 发件人地址（必须显式设置）
//   - email.password: SMTP 密码（从环境变量 ENHANCE_EMAIL_PASSWORD 读取）
//
// # 配置示例
//
// 环境变量：
//
//	export EMAIL_SMTP_HOST=localhost
//	export EMAIL_SMTP_PORT=25
//	export EMAIL_FROM=sender@example.com
//	export ENHANCE_EMAIL_PASSWORD=your_password
package email

import (
	"context"
)

// 默认配置常量。
const (
	// DefaultSMTPHost 默认 SMTP 主机地址。
	DefaultSMTPHost = "localhost"
	// DefaultSMTPPort 默认 SMTP 端口。
	DefaultSMTPPort = 25
	// DefaultFrom 默认发件人地址（必须显式设置）。
	DefaultFrom = ""
	// EnvPasswordKey 环境变量密码键名。
	EnvPasswordKey = "ENHANCE_EMAIL_PASSWORD"
)

// Message 邮件消息结构。
type Message struct {
	From        string            // 发件人地址
	To          []string          // 收件人列表
	Subject     string            // 邮件主题
	Body        string            // 纯文本正文
	HTML        string            // HTML 正文
	Attachments []Attachment      // 附件列表
	Headers     map[string]string // 自定义头部
}

// Attachment 邮件附件结构。
type Attachment struct {
	Filename    string // 文件名
	ContentType string // 内容类型
	Data        []byte // 文件数据
}

// Sender 邮件发送器接口。
type Sender interface {
	// Send 发送邮件。
	Send(ctx context.Context, msg *Message) error

	// Close 关闭发送器并释放资源。
	Close() error
}

// SenderOption 发送器配置选项函数。
type SenderOption func(sender Sender)
