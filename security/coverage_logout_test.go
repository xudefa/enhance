package security

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/security/filter"
)

// ============================================================
// filter_chain.go 测试
// ============================================================

// TestSecurityFilterChainAdapter 测试安全过滤器链适配器
func TestLogoutFilter_AddLogoutHandler(t *testing.T) {
	t.Parallel()

	filter, err := NewLogoutFilter("/logout", nil)
	if err != nil {
		t.Fatalf("NewLogoutFilter error: %v", err)
	}

	handler := &mockLogoutHandler{}
	filter.AddLogoutHandler(handler)
	if len(filter.handlers) != 1 {
		t.Errorf("expected 1 handler, got %d", len(filter.handlers))
	}
}

// TestLogoutFilter_DoFilter_InvalidTypes 测试登出过滤器的类型检查

func TestLogoutFilter_DoFilter_InvalidTypes(t *testing.T) {
	t.Parallel()

	f, err := NewLogoutFilter("/logout", nil)
	if err != nil {
		t.Fatalf("NewLogoutFilter error: %v", err)
	}

	t.Run("invalid context", func(t *testing.T) {
		err := f.DoFilter("invalid", &mockSecurityRequest{}, &mockSecurityResponse{}, filter.NewDefaultFilterChain())
		if err == nil {
			t.Error("expected error for invalid context")
		}
	})

	t.Run("invalid request", func(t *testing.T) {
		err := f.DoFilter(context.Background(), "invalid", &mockSecurityResponse{}, filter.NewDefaultFilterChain())
		if err == nil {
			t.Error("expected error for invalid request")
		}
	})

	t.Run("invalid response", func(t *testing.T) {
		err := f.DoFilter(context.Background(), &mockSecurityRequest{}, "invalid", filter.NewDefaultFilterChain())
		if err == nil {
			t.Error("expected error for invalid response")
		}
	})
}

// TestLogoutFilter_DoFilter 测试登出过滤器

func TestLogoutFilter_DoFilter(t *testing.T) {
	t.Parallel()

	t.Run("no logout handlers", func(t *testing.T) {
		ctx := context.Background()
		req := &mockSecurityRequest{method: "POST", uri: "/logout"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		f, err := NewLogoutFilter("/logout", nil)
		if err != nil {
			t.Fatalf("NewLogoutFilter error: %v", err)
		}

		err = f.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("non-logout URL passes through", func(t *testing.T) {
		ctx := context.Background()
		req := &mockSecurityRequest{method: "POST", uri: "/api/test"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		f, err := NewLogoutFilter("/logout", nil)
		if err != nil {
			t.Fatalf("NewLogoutFilter error: %v", err)
		}

		err = f.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !chain.called {
			t.Error("expected chain to be called for non-logout URL")
		}
	})

	t.Run("GET method not allowed for logout", func(t *testing.T) {
		ctx := context.Background()
		req := &mockSecurityRequest{method: "GET", uri: "/logout"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		f, err := NewLogoutFilter("/logout", nil)
		if err != nil {
			t.Fatalf("NewLogoutFilter error: %v", err)
		}

		err = f.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !chain.called {
			t.Error("expected chain to be called for GET logout")
		}
	})
}

// TestLogoutFilter_WithSuccessHandler 测试带成功处理器的登出过滤器

func TestLogoutFilter_WithSuccessHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	req := &mockSecurityRequest{method: "POST", uri: "/logout"}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	successHandler := &mockLogoutSuccessHandler{}
	f, err := NewLogoutFilter("/logout", nil)
	if err != nil {
		t.Fatalf("NewLogoutFilter error: %v", err)
	}
	f.SetSuccessHandler(successHandler)

	err = f.DoFilter(ctx, req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !successHandler.called {
		t.Error("expected success handler to be called")
	}
}

// TestLogoutFilter_MustNew 测试 MustNewLogoutFilter

func TestLogoutFilter_MustNew(t *testing.T) {
	t.Parallel()

	t.Run("valid creation", func(t *testing.T) {
		f := MustNewLogoutFilter("/logout", nil)
		if f == nil {
			t.Fatal("expected non-nil filter")
		}
	})

	t.Run("panic on empty url", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for empty url")
			}
		}()
		MustNewLogoutFilter("", nil)
	})
}

// TestSimpleLogoutSuccessHandler 测试简单登出成功处理器

func TestSimpleLogoutSuccessHandler(t *testing.T) {
	t.Parallel()

	handler := NewSimpleLogoutSuccessHandler("/goodbye")
	resp := &mockSecurityResponse{headers: make(map[string]string)}
	handler.OnLogoutSuccess(context.Background(), &mockSecurityRequest{}, resp, nil)

	if resp.statusCode != 302 {
		t.Errorf("expected status 302, got %d", resp.statusCode)
	}
	if resp.headers["Location"] != "/goodbye" {
		t.Errorf("expected Location '/goodbye', got '%s'", resp.headers["Location"])
	}
}

// TestDefaultLogoutSuccessHandler 测试默认登出成功处理器

func TestDefaultLogoutSuccessHandler(t *testing.T) {
	t.Parallel()

	handler := NewDefaultLogoutSuccessHandler("/login?logout")
	resp := &mockSecurityResponse{headers: make(map[string]string)}
	handler.OnLogoutSuccess(context.Background(), &mockSecurityRequest{}, resp, nil)

	if resp.statusCode != 302 {
		t.Errorf("expected status 302, got %d", resp.statusCode)
	}
	if resp.headers["Location"] != "/login?logout" {
		t.Errorf("expected Location '/login?logout', got '%s'", resp.headers["Location"])
	}
}

// ============================================================
// filter.go 测试
// ============================================================

// TestFilterSecurityInterceptor_Setters 测试 FilterSecurityInterceptor 的 setter 方法
