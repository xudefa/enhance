package security

import (
	"context"
	"testing"
)

// TestHttpSecurity_ExceptionHandling 测试异常处理配置

func TestHttpSecurity_ExceptionHandling(t *testing.T) {
	t.Parallel()

	httpSec := NewHttpSecurity().(*httpSecurity)
	handler := NewHttp403ForbiddenAccessDeniedHandler()
	entryPoint := NewHttp401UnauthorizedEntryPoint()

	securityBuilder := httpSec.ExceptionHandling(handler, entryPoint)
	if securityBuilder == nil {
		t.Fatal("expected non-nil result")
	}

	if httpSec.exceptionTranslationFilter == nil {
		t.Error("expected exceptionTranslationFilter to be set")
	}
}

// TestWebSecurity 测试 WebSecurity 入口

func TestWebSecurity(t *testing.T) {
	t.Parallel()

	ws := NewWebSecurity()
	if ws == nil {
		t.Fatal("expected non-nil WebSecurity")
	}

	hs := ws.HttpSecurity()
	if hs == nil {
		t.Fatal("expected non-nil httpSecurity")
	}
	if hs.filters == nil {
		t.Error("expected filters to be initialized")
	}

	_, err := ws.Build()
	if err == nil {
		t.Error("expected error from WebSecurity.Build")
	}
}

// TestHttp403ForbiddenEntryPoint 测试 403 入口点

func TestHttp403ForbiddenEntryPoint(t *testing.T) {
	t.Parallel()

	entryPoint := NewHttp403ForbiddenEntryPoint()
	resp := &mockSecurityResponse{}
	err := entryPoint.Commence(context.Background(), &mockSecurityRequest{}, resp, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.statusCode != 403 {
		t.Errorf("expected status 403, got %d", resp.statusCode)
	}
}

// TestHttp401UnauthorizedEntryPoint 测试 401 入口点

func TestHttp401UnauthorizedEntryPoint(t *testing.T) {
	t.Parallel()

	entryPoint := NewHttp401UnauthorizedEntryPoint()
	resp := &mockSecurityResponse{}
	err := entryPoint.Commence(context.Background(), &mockSecurityRequest{}, resp, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.statusCode != 401 {
		t.Errorf("expected status 401, got %d", resp.statusCode)
	}
}

// TestHttp403ForbiddenAccessDeniedHandler 测试 403 访问拒绝处理器

func TestHttp403ForbiddenAccessDeniedHandler(t *testing.T) {
	t.Parallel()

	handler := NewHttp403ForbiddenAccessDeniedHandler()
	resp := &mockSecurityResponse{}
	err := handler.Handle(context.Background(), &mockSecurityRequest{}, resp, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.statusCode != 403 {
		t.Errorf("expected status 403, got %d", resp.statusCode)
	}
}
