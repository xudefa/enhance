package security

import (
	"testing"
)

// ============================================================
// filter_chain.go 测试
// ============================================================

// TestSecurityFilterChainAdapter 测试安全过滤器链适配器
func TestStandardPasswordEncoder(t *testing.T) {
	t.Parallel()

	encoder := NewStandardPasswordEncoder("my-secret-key")

	password := "testPassword123"
	encoded := encoder.Encode(password)

	if encoded == password {
		t.Error("encoded password should differ from raw password")
	}

	if !encoder.Matches(password, encoded) {
		t.Error("should match correct password")
	}

	if encoder.Matches("wrongPassword", encoded) {
		t.Error("should not match wrong password")
	}

	// Same secret produces same hash
	encoder2 := NewStandardPasswordEncoder("my-secret-key")
	if encoder2.Encode(password) != encoded {
		t.Error("same secret should produce same hash")
	}

	// Different secret produces different hash
	encoder3 := NewStandardPasswordEncoder("different-secret")
	if encoder3.Encode(password) == encoded {
		t.Error("different secret should produce different hash")
	}
}

// TestMustNewDelegatingPasswordEncoder 测试 MustNewDelegatingPasswordEncoder

func TestMustNewDelegatingPasswordEncoder(t *testing.T) {
	t.Parallel()

	t.Run("valid creation", func(t *testing.T) {
		encoder := MustNewDelegatingPasswordEncoder("sha256", map[string]PasswordEncoder{
			"sha256": NewSha256PasswordEncoder(),
		})
		if encoder == nil {
			t.Fatal("expected non-nil encoder")
		}
	})

	t.Run("panic on invalid id", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for invalid id")
			}
		}()
		MustNewDelegatingPasswordEncoder("unknown", map[string]PasswordEncoder{
			"sha256": NewSha256PasswordEncoder(),
		})
	})
}

// TestDelegatingPasswordEncoder_EncodeWithMissingEncoder 测试 Encode 时编码器丢失的情况

func TestDelegatingPasswordEncoder_EncodeWithMissingEncoder(t *testing.T) {
	t.Parallel()

	// 创建一个内部编码器映射为空的委托编码器（绕过构造检查）
	encoder := &DelegatingPasswordEncoder{
		idForEncode:      "missing",
		passwordEncoders: map[string]PasswordEncoder{},
	}

	encodedPassword := encoder.Encode("password")
	if encodedPassword != "" {
		t.Errorf("expected empty string, got '%s'", encodedPassword)
	}
}

// ============================================================
// rate_limit.go 测试
// ============================================================

// TestSlidingWindowRateLimiter_Close 测试关闭限流器
