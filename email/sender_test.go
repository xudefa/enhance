package email

import (
	"context"
	"os"
	"testing"
)

// getImpl 获取 smtpSender 实现（仅用于测试）。
func getImpl(s Sender) *smtpSender {
	return s.(*smtpSender)
}

func TestNewSender_Defaults(t *testing.T) {
	t.Parallel()
	sender := NewSender()
	impl := getImpl(sender)

	if impl.host != DefaultSMTPHost {
		t.Errorf("期望主机地址 %s, 得到 %s", DefaultSMTPHost, impl.host)
	}

	if impl.port != DefaultSMTPPort {
		t.Errorf("期望端口 %d, 得到 %d", DefaultSMTPPort, impl.port)
	}

	if impl.username != DefaultFrom {
		t.Errorf("期望用户名 %s, 得到 %s", DefaultFrom, impl.username)
	}
}

func TestNewSender_WithOptions(t *testing.T) {
	t.Parallel()
	sender := NewSender(
		WithHost("smtp.example.com"),
		WithPort(587),
		WithAuth("user@example.com", "password123"),
	)
	impl := getImpl(sender)

	if impl.host != "smtp.example.com" {
		t.Errorf("期望主机地址 smtp.example.com, 得到 %s", impl.host)
	}

	if impl.port != 587 {
		t.Errorf("期望端口 587, 得到 %d", impl.port)
	}

	if impl.username != "user@example.com" {
		t.Errorf("期望用户名 user@example.com, 得到 %s", impl.username)
	}

	if impl.password != "password123" {
		t.Errorf("期望密码 password123, 得到 %s", impl.password)
	}
}

func TestNewSender_PasswordFromEnv(t *testing.T) {
	t.Parallel()
	_ = os.Setenv(EnvPasswordKey, "env_password_123")
	defer func() { _ = os.Unsetenv(EnvPasswordKey) }()

	sender := NewSender()
	impl := getImpl(sender)

	if impl.password != "env_password_123" {
		t.Errorf("期望环境变量密码 env_password_123, 得到 %s", impl.password)
	}
}

func TestNewSender_WithPasswordFromEnv(t *testing.T) {
	t.Parallel()
	_ = os.Setenv(EnvPasswordKey, "initial_password")
	sender := NewSender(
		WithHost("smtp.test.com"),
		WithAuth("user@test.com", "override_password"),
		WithPasswordFromEnv(),
	)
	impl := getImpl(sender)

	if impl.password != "initial_password" {
		t.Errorf("期望环境变量密码 initial_password, 得到 %s", impl.password)
	}

	_ = os.Unsetenv(EnvPasswordKey)
}

func TestSmtpSender_Close(t *testing.T) {
	t.Parallel()
	sender := NewSender()

	err := sender.Close()
	if err != nil {
		t.Errorf("期望 Close 返回 nil, 得到 %v", err)
	}
}

func TestSmtpSender_Send_EmptyHost(t *testing.T) {
	t.Parallel()
	sender := &smtpSender{
		host: "",
	}

	ctx := context.Background()
	msg := &Message{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
	}

	err := sender.Send(ctx, msg)
	if err == nil {
		t.Error("期望空主机返回错误")
	}

	if err.Error() != "SMTP host not configured" {
		t.Errorf("expected 'SMTP host not configured' error, got %v", err)
	}
}

func TestSmtpSender_Send_DefaultFrom(t *testing.T) {
	t.Parallel()
	sender := &smtpSender{
		host:     "smtp.test.com",
		port:     25,
		username: "default@test.com",
	}

	msg := &Message{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
	}

	ctx := context.Background()
	err := sender.Send(ctx, msg)

	if err == nil {
		t.Log("发送成功（SMTP 服务器可用时）")
	} else {
		t.Logf("发送失败（SMTP 服务器不可用时）: %v", err)
	}

	if msg.From != "" {
		t.Errorf("期望 Send 不修改调用方的 Message.From, 得到 %s", msg.From)
	}
}

func TestBuildMessage_PlainText(t *testing.T) {
	t.Parallel()
	sender := NewSender()
	impl := getImpl(sender)

	msg := &Message{
		From:    "sender@test.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		Body:    "Test Body",
	}

	result := impl.buildMessage(msg, msg.From)

	if !containsLine(result, "From: sender@test.com") {
		t.Error("期望消息中包含 From 头")
	}

	if !containsLine(result, "To: recipient@example.com") {
		t.Error("期望消息中包含 To 头")
	}

	if !containsLine(result, "Subject: Test Subject") {
		t.Error("期望消息中包含 Subject 头")
	}

	if !containsLine(result, "MIME-Version: 1.0") {
		t.Error("期望消息中包含 MIME-Version 头")
	}

	if !containsLine(result, "Content-Type: text/plain; charset=UTF-8") {
		t.Error("期望文本内容为 text/plain 类型")
	}

	if !containsSubstring(result, "Test Body") {
		t.Error("期望消息中包含正文内容")
	}
}

func TestBuildMessage_HTML(t *testing.T) {
	t.Parallel()
	sender := NewSender()
	impl := getImpl(sender)

	msg := &Message{
		From:    "sender@test.com",
		To:      []string{"recipient@example.com"},
		Subject: "HTML Test",
		HTML:    "<h1>Hello</h1>",
	}

	result := impl.buildMessage(msg, msg.From)

	if !containsLine(result, "Content-Type: text/html; charset=UTF-8") {
		t.Error("期望 HTML内容为 text/html 类型")
	}

	if !containsSubstring(result, "<h1>Hello</h1>") {
		t.Error("期望消息中包含 HTML 内容")
	}
}

func TestBuildMessage_MultipleRecipients(t *testing.T) {
	t.Parallel()
	sender := NewSender()
	impl := getImpl(sender)

	msg := &Message{
		From:    "sender@test.com",
		To:      []string{"user1@example.com", "user2@example.com"},
		Subject: "Test",
		Body:    "Body",
	}

	result := impl.buildMessage(msg, msg.From)

	if !containsLine(result, "To: user1@example.com, user2@example.com") {
		t.Errorf("期望 To 头包含多个收件人, 得到:\n%s", result)
	}
}

func TestBuildMessage_CustomHeaders(t *testing.T) {
	t.Parallel()
	sender := NewSender()
	impl := getImpl(sender)

	msg := &Message{
		From:    "sender@test.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Body",
		Headers: map[string]string{
			"X-Custom-Header": "custom-value",
		},
	}

	result := impl.buildMessage(msg, msg.From)

	if !containsLine(result, "X-Custom-Header: custom-value") {
		t.Error("期望消息中包含自定义头")
	}
}

func TestJoinAddresses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		addresses []string
		expected  string
	}{
		{
			name:      "单个地址",
			addresses: []string{"user@example.com"},
			expected:  "user@example.com",
		},
		{
			name:      "多个地址",
			addresses: []string{"user1@example.com", "user2@example.com", "user3@example.com"},
			expected:  "user1@example.com, user2@example.com, user3@example.com",
		},
		{
			name:      "空地址列表",
			addresses: []string{},
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinAddresses(tt.addresses)
			if result != tt.expected {
				t.Errorf("期望 %s, 得到 %s", tt.expected, result)
			}
		})
	}
}

func TestMessage_StructFields(t *testing.T) {
	t.Parallel()
	msg := &Message{
		From:    "sender@test.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Body",
		HTML:    "<html></html>",
		Attachments: []Attachment{
			{
				Filename:    "test.txt",
				ContentType: "text/plain",
				Data:        []byte("test data"),
			},
		},
		Headers: map[string]string{
			"X-Priority": "1",
		},
	}

	if msg.From != "sender@test.com" {
		t.Errorf("期望发件人 sender@test.com, 得到 %s", msg.From)
	}

	if len(msg.To) != 1 || msg.To[0] != "recipient@example.com" {
		t.Errorf("期望收件人 [recipient@example.com], 得到 %v", msg.To)
	}

	if len(msg.Attachments) != 1 {
		t.Errorf("期望 1 个附件, 得到 %d", len(msg.Attachments))
	}

	if msg.Attachments[0].Filename != "test.txt" {
		t.Errorf("期望附件文件名 test.txt, 得到 %s", msg.Attachments[0].Filename)
	}
}

func TestAttachment_StructFields(t *testing.T) {
	t.Parallel()
	attachment := Attachment{
		Filename:    "report.pdf",
		ContentType: "application/pdf",
		Data:        []byte{0x25, 0x50, 0x44, 0x46},
	}

	if attachment.Filename != "report.pdf" {
		t.Errorf("期望文件名 report.pdf, 得到 %s", attachment.Filename)
	}

	if attachment.ContentType != "application/pdf" {
		t.Errorf("期望内容类型 application/pdf, 得到 %s", attachment.ContentType)
	}

	if len(attachment.Data) != 4 {
		t.Errorf("期望数据长度 4, 得到 %d", len(attachment.Data))
	}
}

func containsLine(s, line string) bool {
	lines := splitLines(s)
	for _, l := range lines {
		if l == line {
			return true
		}
	}
	return false
}

func containsSubstring(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && searchSubstring(s, substr))
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
