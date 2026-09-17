package security

import (
	"context"
	"sync"
	"testing"
)

func TestSecurityContextComprehensive(t *testing.T) {
	t.Parallel()

	t.Run("SetAuthentication and Authentication", func(t *testing.T) {
		t.Parallel()
		ctx := NewSecurityContext()

		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user1", []string{"ROLE_USER"})
		ctx.SetAuthentication(auth)

		got := ctx.Authentication()
		if got == nil {
			t.Fatal("expected non-nil authentication")
		}
		if extractPrincipalName(got) != "user1" {
			t.Errorf("expected principal 'user1', got '%s'", extractPrincipalName(got))
		}
	})

	t.Run("ClearAuthentication", func(t *testing.T) {
		t.Parallel()
		ctx := NewSecurityContext()

		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user1", []string{"ROLE_USER"})
		ctx.SetAuthentication(auth)

		ctx.ClearAuthentication()
		got := ctx.Authentication()
		if got != nil {
			t.Errorf("expected nil after clear, got %v", got)
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		t.Parallel()
		ctx := NewSecurityContext()

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER"})
				ctx.SetAuthentication(auth)
				_ = ctx.Authentication()
			}(i)
		}
		wg.Wait()
	})
}

func TestContextWithAuthentication(t *testing.T) {
	t.Parallel()

	t.Run("stores and retrieves authentication", func(t *testing.T) {
		t.Parallel()
		baseCtx := context.Background()
		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})

		ctx := ContextWithAuthentication(baseCtx, auth)
		got := GetAuthenticationFromContext(ctx)

		if got == nil {
			t.Fatal("expected non-nil authentication")
		}
		if extractPrincipalName(got) != "admin" {
			t.Errorf("expected principal 'admin', got '%s'", extractPrincipalName(got))
		}
	})

	t.Run("returns nil when no authentication", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		got := GetAuthenticationFromContext(ctx)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})
}

func TestUsernamePasswordAuthentication(t *testing.T) {
	t.Parallel()

	t.Run("NewUsernamePasswordAuthenticationToken", func(t *testing.T) {
		t.Parallel()
		auth := NewUsernamePasswordAuthenticationToken("user", "pass")
		if auth.Principal() != "user" {
			t.Errorf("expected principal 'user', got '%v'", auth.Principal())
		}
		if auth.Credentials() != "pass" {
			t.Errorf("expected credentials 'pass', got '%v'", auth.Credentials())
		}
		if auth.Authenticated() {
			t.Error("expected not authenticated")
		}
	})

	t.Run("NewAuthenticatedUsernamePasswordAuthenticationToken", func(t *testing.T) {
		t.Parallel()
		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})
		if !auth.Authenticated() {
			t.Error("expected authenticated")
		}
		if len(auth.Authorities()) != 1 {
			t.Errorf("expected 1 authority, got %d", len(auth.Authorities()))
		}
	})

	t.Run("SetAuthenticated and SetAuthorities", func(t *testing.T) {
		t.Parallel()
		auth := NewUsernamePasswordAuthenticationToken("user", "pass")
		auth.SetAuthenticated(true)
		if !auth.Authenticated() {
			t.Error("expected authenticated after SetAuthenticated(true)")
		}
		auth.SetAuthorities([]string{"ROLE_USER", "ROLE_ADMIN"})
		if len(auth.Authorities()) != 2 {
			t.Errorf("expected 2 authorities, got %d", len(auth.Authorities()))
		}
	})

	t.Run("Name returns principal string", func(t *testing.T) {
		t.Parallel()
		auth := NewUsernamePasswordAuthenticationToken("user", "pass")
		if auth.Name() != "user" {
			t.Errorf("expected name 'user', got '%s'", auth.Name())
		}
	})
}

func TestPasswordEncoderInterface(t *testing.T) {
	t.Parallel()

	t.Run("NoOpPasswordEncoder roundtrip", func(t *testing.T) {
		t.Parallel()
		enc := NewNoOpPasswordEncoder()
		encoded := enc.Encode("mypassword")
		if encoded != "mypassword" {
			t.Errorf("NoOp should return same string, got '%s'", encoded)
		}
		if !enc.Matches("mypassword", "mypassword") {
			t.Error("expected match")
		}
		if enc.Matches("mypassword", "wrongpassword") {
			t.Error("expected no match")
		}
	})
}

func TestSecurityRequestResponseMock(t *testing.T) {
	t.Parallel()

	t.Run("mock request methods", func(t *testing.T) {
		t.Parallel()
		req := &mockSecurityRequest{method: "POST", uri: "/test"}
		req.SetHeader("X-Custom", "value123")
		req.SetAttribute("key", "val")

		if req.GetMethod() != "POST" {
			t.Errorf("expected POST, got %s", req.GetMethod())
		}
		if req.GetHeader("X-Custom") != "value123" {
			t.Errorf("expected value123, got %s", req.GetHeader("X-Custom"))
		}
		if req.GetHeader("X-Missing") != "" {
			t.Error("expected empty string for missing header")
		}

		attributeValue, ok := req.GetAttribute("key")
		if !ok || attributeValue != "val" {
			t.Errorf("expected attribute 'val', got %v (exists=%v)", attributeValue, ok)
		}
	})

	t.Run("mock response methods", func(t *testing.T) {
		t.Parallel()
		resp := &mockSecurityResponse{headers: map[string]string{}}
		resp.SetStatusCode(404)
		resp.SetHeader("X-Test", "hello")

		if resp.statusCode != 404 {
			t.Errorf("expected 404, got %d", resp.statusCode)
		}
		if resp.headers["X-Test"] != "hello" {
			t.Errorf("expected 'hello', got '%s'", resp.headers["X-Test"])
		}
	})
}

func TestFilterChainProxy(t *testing.T) {
	t.Parallel()

	t.Run("SecurityFilterChainAdapter adapts filterChainProxy", func(t *testing.T) {
		t.Parallel()
		chain := newFilterChainProxy([]SecurityFilter{}, &DefaultSecurityFilterChain{})
		adapter := securityFilterChainAdapter{proxy: chain}

		filters := adapter.GetFilters()
		if len(filters) != 0 {
			t.Errorf("expected empty filters, got %d", len(filters))
		}
	})
}
