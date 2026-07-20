package security

import (
	"testing"
)

// TestGenerateSecureToken_CryptographicSecurity 测试 CSRF token 的密码学安全性
func TestGenerateSecureToken_CryptographicSecurity(t *testing.T) {
	t.Parallel()
	// 测试 1: 生成的 token 应该具有足够的熵，不会重复
	tokens := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		token := generateSecureToken(32)
		if tokens[token] {
			t.Fatal("duplicate token detected - token generation is not secure")
		}
		tokens[token] = true
	}
}

// TestGenerateSecureToken_Unpredictability 测试 token 的不可预测性
func TestGenerateSecureToken_Unpredictability(t *testing.T) {
	t.Parallel()
	// 测试 2: 连续生成的 token 不应该有可预测的模式
	token1 := generateSecureToken(32)
	token2 := generateSecureToken(32)
	if token1 == token2 {
		t.Fatal("consecutive tokens should be different")
	}

	// 测试多个 token 之间的差异性
	tokens := make([]string, 100)
	for i := 0; i < 100; i++ {
		tokens[i] = generateSecureToken(32)
	}

	// 验证所有 token 都是唯一的
	seen := make(map[string]bool)
	for _, token := range tokens {
		if seen[token] {
			t.Fatal("found duplicate token in batch generation")
		}
		seen[token] = true
	}
}

// TestGenerateSecureToken_Length 测试 token 长度
func TestGenerateSecureToken_Length(t *testing.T) {
	t.Parallel()
	// 测试 3: token 长度应该符合预期
	// base64 编码后的长度约为 ceil(length/3)*4
	tests := []struct {
		inputLength  int
		minOutputLen int
	}{
		{32, 40}, // ceil(32/3)*4 = 44
		{16, 20}, // ceil(16/3)*4 = 24
		{64, 80}, // ceil(64/3)*4 = 88
	}

	for _, tt := range tests {
		token := generateSecureToken(tt.inputLength)
		if len(token) < tt.minOutputLen {
			t.Errorf("token too short for input %d: got %d, want >= %d",
				tt.inputLength, len(token), tt.minOutputLen)
		}
	}
}

// TestCsrfFilter_AddExcludePath_Validation 测试排除路径的输入验证
func TestCsrfFilter_AddExcludePath_Validation(t *testing.T) {
	t.Parallel()
	filter := NewCsrfFilter(NewCookieCsrfTokenRepository())

	// 测试空路径
	err := filter.AddExcludePath("")
	if err == nil {
		t.Error("expected error for empty path")
	}

	// 测试不以 / 开头的路径
	err = filter.AddExcludePath("api/test")
	if err == nil {
		t.Error("expected error for path not starting with '/'")
	}

	// 测试有效路径
	err = filter.AddExcludePath("/api/test")
	if err != nil {
		t.Errorf("unexpected error for valid path: %v", err)
	}

	// 测试多个路径，其中包含无效路径
	err = filter.AddExcludePath("/valid", "invalid", "/also-valid")
	if err == nil {
		t.Error("expected error when one of multiple paths is invalid")
	}
}

// TestCsrfFilter_ExcludePathValidation 测试排除路径验证后的行为
func TestCsrfFilter_ExcludePathValidation(t *testing.T) {
	t.Parallel()
	filter := NewCsrfFilter(NewCookieCsrfTokenRepository())

	// 添加有效路径
	err := filter.AddExcludePath("/api/health", "/api/status")
	if err != nil {
		t.Fatalf("failed to add valid paths: %v", err)
	}

	// 验证路径已添加
	if len(filter.excludePaths) != 2 {
		t.Errorf("expected 2 exclude paths, got %d", len(filter.excludePaths))
	}
}
