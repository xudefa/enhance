// Package email 提供邮件发送功能，用于 enhance 框架。
package email

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

// smtpSender 邮件发送器实现。
type smtpSender struct {
	host     string
	port     int
	username string
	password string
	auth     smtp.Auth
}

// WithHost 设置 SMTP 主机地址。
func WithHost(host string) SenderOption {
	return func(s Sender) {
		if impl, ok := s.(*smtpSender); ok {
			impl.host = host
		}
	}
}

// WithPort 设置 SMTP 端口。
func WithPort(port int) SenderOption {
	return func(s Sender) {
		if impl, ok := s.(*smtpSender); ok {
			impl.port = port
		}
	}
}

// WithAuth 设置认证信息。
func WithAuth(username, password string) SenderOption {
	return func(s Sender) {
		if impl, ok := s.(*smtpSender); ok {
			impl.username = username
			impl.password = password
			impl.auth = smtp.PlainAuth("", username, password, impl.host)
		}
	}
}

// WithPasswordFromEnv 从环境变量读取密码。
func WithPasswordFromEnv() SenderOption {
	return func(s Sender) {
		if impl, ok := s.(*smtpSender); ok {
			impl.password = os.Getenv(EnvPasswordKey)
			impl.auth = smtp.PlainAuth("", impl.username, impl.password, impl.host)
		}
	}
}

// NewSender 创建邮件发送器。
func NewSender(opts ...SenderOption) Sender {
	s := &smtpSender{
		host:     DefaultSMTPHost,
		port:     DefaultSMTPPort,
		username: DefaultFrom,
	}

	s.password = os.Getenv(EnvPasswordKey)

	for _, opt := range opts {
		opt(s)
	}

	if s.password != "" && s.auth == nil {
		s.auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	return s
}

// Send 发送邮件。
func (s *smtpSender) Send(ctx context.Context, msg *Message) error {
	if s.host == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	if msg.From == "" {
		msg.From = s.username
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	body := s.buildMessage(msg)

	err := smtp.SendMail(addr, s.auth, msg.From, msg.To, []byte(body))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// Close 关闭发送器。
func (s *smtpSender) Close() error {
	return nil
}

func (s *smtpSender) buildMessage(msg *Message) string {
	headers := make(map[string]string)

	if msg.From != "" {
		headers["From"] = msg.From
	}
	if len(msg.To) > 0 {
		headers["To"] = joinAddresses(msg.To)
	}
	if msg.Subject != "" {
		headers["Subject"] = msg.Subject
	}
	headers["MIME-Version"] = "1.0"

	for k, v := range msg.Headers {
		headers[k] = v
	}

	var builder strings.Builder
	for k, v := range headers {
		fmt.Fprintf(&builder, "%s: %s\r\n", k, v)
	}

	builder.WriteString("\r\n")

	if msg.HTML != "" {
		builder.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		builder.WriteString(msg.HTML)
		return builder.String()
	}
	builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	builder.WriteString(msg.Body)

	return builder.String()
}

func joinAddresses(addresses []string) string {
	return strings.Join(addresses, ", ")
}
