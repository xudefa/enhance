package security

import (
	"context"
	"net/http"
	"testing"
)

// ============================================================
// filter_chain.go 测试
// ============================================================

// TestSecurityFilterChainAdapter 测试安全过滤器链适配器
func TestCsrfFilter_NewCsrfFilter_Error(t *testing.T) {
	t.Parallel()

	_, err := NewCsrfFilter(nil)
	if err == nil {
		t.Error("expected error for nil token repository")
	}
}

// TestCsrfFilter_MustNewCsrfFilter_Panic 测试 MustNewCsrfFilter 的 panic

func TestCsrfFilter_MustNewCsrfFilter_Panic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil token repository")
		}
	}()
	MustNewCsrfFilter(nil)
}

// TestCsrfFilter_AddExcludePath_Error 测试 AddExcludePath 的错误路径

func TestCsrfFilter_AddExcludePath_Error(t *testing.T) {
	t.Parallel()

	repo := NewCookieCsrfTokenRepository()
	filter, err := NewCsrfFilter(repo)
	if err != nil {
		t.Fatalf("NewCsrfFilter error: %v", err)
	}

	t.Run("empty path", func(t *testing.T) {
		err := filter.AddExcludePath("")
		if err == nil {
			t.Error("expected error for empty path")
		}
	})

	t.Run("path without leading slash", func(t *testing.T) {
		err := filter.AddExcludePath("api/public")
		if err == nil {
			t.Error("expected error for path without leading slash")
		}
	})
}

// TestCsrfFilter_DoFilter_InvalidTypes 测试 CsrfFilter 的类型检查

func TestCsrfFilter_DoFilter_InvalidTypes(t *testing.T) {
	t.Parallel()

	repo := NewCookieCsrfTokenRepository()
	f, err := NewCsrfFilter(repo)
	if err != nil {
		t.Fatalf("NewCsrfFilter error: %v", err)
	}

	t.Run("invalid context", func(t *testing.T) {
		err := f.DoFilter("invalid", &mockSecurityRequest{}, &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid context")
		}
	})

	t.Run("invalid request", func(t *testing.T) {
		err := f.DoFilter(context.Background(), "invalid", &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid request")
		}
	})

	t.Run("invalid response", func(t *testing.T) {
		err := f.DoFilter(context.Background(), &mockSecurityRequest{}, "invalid", &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid response")
		}
	})
}

// TestCsrfFilter_DoFilter_CSRFProtection 测试 CSRF 保护

func TestCsrfFilter_DoFilter_CSRFProtection(t *testing.T) {
	t.Parallel()

	repo := NewCookieCsrfTokenRepository()
	f, err := NewCsrfFilter(repo)
	if err != nil {
		t.Fatalf("NewCsrfFilter error: %v", err)
	}

	t.Run("POST without CSRF token returns error", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/data"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}
		err := f.DoFilter(context.Background(), req, resp, chain)
		if err == nil {
			t.Error("expected error for missing CSRF token")
		}
	})

	t.Run("POST with invalid CSRF token returns error", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/data"}
		req.SetHeader("X-CSRF-Token", "invalid-token")
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}
		err := f.DoFilter(context.Background(), req, resp, chain)
		if err == nil {
			t.Error("expected error for invalid CSRF token")
		}
	})

	t.Run("X-XSRF-Token header", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/data"}
		req.SetHeader("X-XSRF-Token", "invalid-token")
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}
		err := f.DoFilter(context.Background(), req, resp, chain)
		if err == nil {
			t.Error("expected error for invalid X-XSRF-Token")
		}
	})
}

// TestCookieCsrfTokenRepository_LoadCookieValue 测试 Cookie 加载

func TestCookieCsrfTokenRepository_LoadCookieValue(t *testing.T) {
	t.Parallel()

	t.Run("cookie not found", func(t *testing.T) {
		repo := NewCookieCsrfTokenRepository()
		req := &mockSecurityRequest{method: "GET", uri: "/test"}
		_, exists := repo.loadCookieValue(req)
		if exists {
			t.Error("expected false for missing cookie")
		}
	})

	t.Run("cookie found", func(t *testing.T) {
		repo := NewCookieCsrfTokenRepository()
		req := &mockSecurityRequest{method: "GET", uri: "/test"}
		req.SetHeader("Cookie", "_csrf_token=abc123; other=val")
		val, exists := repo.loadCookieValue(req)
		if !exists {
			t.Error("expected true for existing cookie")
		}
		if val != "abc123" {
			t.Errorf("expected 'abc123', got '%s'", val)
		}
	})

	t.Run("ValidateToken with cookie", func(t *testing.T) {
		repo := NewCookieCsrfTokenRepository()
		req := &mockSecurityRequest{method: "POST", uri: "/test"}
		req.SetHeader("Cookie", "_csrf_token=valid-token-value")
		if !repo.ValidateToken(context.Background(), req, "valid-token-value") {
			t.Error("expected valid token to be validated")
		}
	})

	t.Run("ValidateToken with non-string attribute", func(t *testing.T) {
		repo := NewCookieCsrfTokenRepository()
		req := &mockSecurityRequest{method: "POST", uri: "/test"}
		req.SetAttribute("csrf.token", 12345) // non-string attribute
		if repo.ValidateToken(context.Background(), req, "token") {
			t.Error("expected false for non-string attribute")
		}
	})
}

// TestSameSiteString 测试 SameSite 字符串转换

func TestSameSiteString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode http.SameSite
		want string
	}{
		{http.SameSiteLaxMode, "Lax"},
		{http.SameSiteStrictMode, "Strict"},
		{http.SameSiteNoneMode, "None"},
		{http.SameSiteDefaultMode, "Lax"},
	}

	for _, tt := range tests {
		got := sameSiteString(tt.mode)
		if got != tt.want {
			t.Errorf("sameSiteString(%d) = %s, want %s", tt.mode, got, tt.want)
		}
	}
}

// ============================================================
// config.go 测试
// ============================================================

// TestConfigRegistry_AddRule 测试 configRegistry 的 addRule 方法
