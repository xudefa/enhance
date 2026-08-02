// Package email 提供邮件发送功能，用于 enhance 框架。
package email

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"os"
	"strings"
)

// smtpSender 邮件发送器实现。
type smtpSender struct {
	host     string
	port     int
	username string
	password string
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
		}
	}
}

// WithPasswordFromEnv 从环境变量读取密码。
func WithPasswordFromEnv() SenderOption {
	return func(s Sender) {
		if impl, ok := s.(*smtpSender); ok {
			impl.password = os.Getenv(EnvPasswordKey)
		}
	}
}

// NewSender 创建邮件发送器。
// 注意：密码在构造时从环境变量读取一次并缓存，后续环境变量变更不会自动更新。
// 如需动态读取密码，请使用 WithPasswordFromEnv 选项或 WithAuth 选项。
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

	return s
}

// buildAuth 构建 SMTP 认证信息。
//
// 延迟到发送时构建，确保使用最终的 host 值，避免选项应用顺序导致的认证配置不一致。
func (s *smtpSender) buildAuth() smtp.Auth {
	if s.username == "" || s.password == "" {
		return nil
	}
	return smtp.PlainAuth("", s.username, s.password, s.host)
}

// Send 发送邮件。
func (s *smtpSender) Send(ctx context.Context, msg *Message) error {
	if msg == nil {
		return fmt.Errorf("message is nil")
	}
	if s.host == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	from := msg.From
	if from == "" {
		from = s.username
	}
	if from == "" {
		return fmt.Errorf("sender address is required: set Message.From or configure a default sender")
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	body := s.buildMessage(msg, from)

	err := smtp.SendMail(addr, s.buildAuth(), from, msg.To, []byte(body))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// Close 关闭发送器。
func (s *smtpSender) Close() error {
	return nil
}

func (s *smtpSender) buildMessage(msg *Message, from string) string {
	headers := make(map[string]string)

	if from != "" {
		headers["From"] = from
	}
	if len(msg.To) > 0 {
		headers["To"] = joinAddresses(msg.To)
	}
	if msg.Subject != "" {
		subject := strings.NewReplacer("\r", "", "\n", "").Replace(msg.Subject)
		headers["Subject"] = subject
	}
	headers["MIME-Version"] = "1.0"

	for k, v := range msg.Headers {
		headers[k] = v
	}

	if len(msg.Attachments) > 0 {
		return s.buildMultipartMessage(msg, headers)
	}

	if msg.HTML != "" {
		headers["Content-Type"] = "text/html; charset=UTF-8"
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
	}

	var builder strings.Builder
	for k, v := range headers {
		fmt.Fprintf(&builder, "%s: %s\r\n", k, v)
	}
	builder.WriteString("\r\n")

	if msg.HTML != "" {
		builder.WriteString(msg.HTML)
	} else {
		builder.WriteString(msg.Body)
	}

	return builder.String()
}

func (s *smtpSender) buildMultipartMessage(msg *Message, headers map[string]string) string {
	var headersBuf strings.Builder
	var bodyBuf strings.Builder

	mp := multipart.NewWriter(&bodyBuf)

	contentType := "text/plain; charset=UTF-8"
	if msg.HTML != "" {
		contentType = "text/html; charset=UTF-8"
	}

	headers["Content-Type"] = fmt.Sprintf(`multipart/mixed; boundary="%s"`, mp.Boundary())

	for k, v := range headers {
		fmt.Fprintf(&headersBuf, "%s: %s\r\n", k, v)
	}
	headersBuf.WriteString("\r\n")

	bodyContent := msg.Body
	if msg.HTML != "" {
		bodyContent = msg.HTML
	}
	part, err := mp.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {contentType},
		"Content-Transfer-Encoding": {"7bit"},
	})
	if err == nil {
		_, _ = part.Write([]byte(bodyContent))
	}

	for _, a := range msg.Attachments {
		encoded := base64.StdEncoding.EncodeToString(a.Data)
		part, err := mp.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {a.ContentType},
			"Content-Disposition":       {fmt.Sprintf(`attachment; filename="%s"`, a.Filename)},
			"Content-Transfer-Encoding": {"base64"},
		})
		if err != nil {
			continue
		}
		_, _ = part.Write([]byte(encoded))
	}

	_ = mp.Close()

	headersBuf.WriteString(bodyBuf.String())
	return headersBuf.String()
}

func joinAddresses(addresses []string) string {
	return strings.Join(addresses, ", ")
}
