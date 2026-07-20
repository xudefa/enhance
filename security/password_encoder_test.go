package security

import (
	"strings"
	"testing"
)

// TestSha256PasswordEncoder 测试 SHA256 密码编码器
func TestSha256PasswordEncoder(t *testing.T) {
	t.Parallel()
	encoder := NewSha256PasswordEncoder()

	password := "mySecurePassword123!"
	encoded := encoder.Encode(password)

	// 测试编码后的密码不等于原始密码
	if encoded == password {
		t.Error("encoded password should be different from raw password")
	}

	// 测试匹配正确密码
	if !encoder.Matches(password, encoded) {
		t.Error("should match correct password")
	}

	// 测试不匹配错误密码
	if encoder.Matches("wrongPassword", encoded) {
		t.Error("should not match wrong password")
	}

	// 测试相同密码生成相同的 hash（SHA256 无盐值）
	encoded2 := encoder.Encode(password)
	if encoded != encoded2 {
		t.Error("SHA256 should produce same hash for same password")
	}
}

// TestDelegatingPasswordEncoder_ExtractId 测试密码格式解析
func TestDelegatingPasswordEncoder_ExtractId(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input   string
		wantId  string
		wantPwd string
		wantErr bool
	}{
		{"{bcrypt}$2a$12$abc123", "bcrypt", "$2a$12$abc123", false},
		{"{sha256}abc123", "sha256", "abc123", false},
		{"{noop}password", "noop", "password", false},
		{"invalid", "", "", true},
		{"{noclose", "", "", true},
		{"}id{password", "", "", true},
		{"{id}", "id", "", false},
		{"{}", "", "", true}, // 空 ID 应该报错
	}

	encoder := &DelegatingPasswordEncoder{}

	for _, tt := range tests {
		id, pwd, err := encoder.extractId(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("extractId(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if id != tt.wantId || pwd != tt.wantPwd {
			t.Errorf("extractId(%q) = (%q, %q), want (%q, %q)",
				tt.input, id, pwd, tt.wantId, tt.wantPwd)
		}
	}
}

// TestDelegatingPasswordEncoder_FullCycle 测试完整的编码和验证流程
func TestDelegatingPasswordEncoder_FullCycle(t *testing.T) {
	t.Parallel()
	sha256Encoder := NewSha256PasswordEncoder()
	noopEncoder := NewNoOpPasswordEncoder()

	encoders := map[string]PasswordEncoder{
		"sha256": sha256Encoder,
		"noop":   noopEncoder,
	}

	encoder := NewDelegatingPasswordEncoder("sha256", encoders)

	password := "myPassword"
	encoded := encoder.Encode(password)

	// 验证格式正确
	if !strings.HasPrefix(encoded, "{sha256}") {
		t.Errorf("encoded password should start with {sha256}, got: %s", encoded[:10])
	}

	// 测试匹配
	if !encoder.Matches(password, encoded) {
		t.Error("should match correct password")
	}

	if encoder.Matches("wrongPassword", encoded) {
		t.Error("should not match wrong password")
	}
}

// TestDelegatingPasswordEncoder_UnknownId 测试未知编码器 ID
func TestDelegatingPasswordEncoder_UnknownId(t *testing.T) {
	t.Parallel()
	sha256Encoder := NewSha256PasswordEncoder()

	encoders := map[string]PasswordEncoder{
		"sha256": sha256Encoder,
	}

	encoder := NewDelegatingPasswordEncoder("sha256", encoders)

	// 使用未知的编码器 ID 编码的密码
	encoded := "{unknown}somehash"

	// 应该返回 false 而不是 panic
	if encoder.Matches("password", encoded) {
		t.Error("should not match unknown encoder id")
	}
}

// TestDelegatingPasswordEncoder_InvalidFormat 测试无效格式
func TestDelegatingPasswordEncoder_InvalidFormat(t *testing.T) {
	t.Parallel()
	sha256Encoder := NewSha256PasswordEncoder()

	encoders := map[string]PasswordEncoder{
		"sha256": sha256Encoder,
	}

	encoder := NewDelegatingPasswordEncoder("sha256", encoders)

	invalidFormats := []string{
		"",
		"no-braces",
		"{missing-close",
		"}missing-open{",
	}

	for _, encoded := range invalidFormats {
		if encoder.Matches("password", encoded) {
			t.Errorf("should not match invalid format: %s", encoded)
		}
	}
}
