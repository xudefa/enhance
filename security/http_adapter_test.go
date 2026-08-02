package security

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security/filter"
)

type adapterTestChain struct {
	doFilter func(ctx interface{}, request interface{}, response interface{}) error
}

func (c *adapterTestChain) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
	if c.doFilter != nil {
		return c.doFilter(ctx, request, response)
	}
	return nil
}

func (c *adapterTestChain) Matches(request interface{}) bool { return true }

func (c *adapterTestChain) GetFilters() []filter.Filter { return nil }

// TestSecurityFilterChainHandler_ServeHTTP 验证安全响应产生后不再调用 nextHandler（H2）。
func TestSecurityFilterChainHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	newHandler := func(chainFunc func(ctx interface{}, request interface{}, response interface{}) error) (*SecurityFilterChainHandler, *bool) {
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})
		return NewSecurityFilterChainHandler(&adapterTestChain{doFilter: chainFunc}, next), &nextCalled
	}

	tests := []struct {
		name           string
		chainFunc      func(ctx interface{}, request interface{}, response interface{}) error
		wantStatus     int
		wantBody       string
		wantNextCalled bool
	}{
		{
			name: "redirect response terminates request",
			chainFunc: func(ctx interface{}, request interface{}, response interface{}) error {
				resp := response.(SecurityResponse)
				resp.SetHeader("Location", "/login")
				resp.SetStatusCode(http.StatusFound)
				return nil
			},
			wantStatus:     http.StatusFound,
			wantNextCalled: false,
		},
		{
			name: "unauthorized response terminates request",
			chainFunc: func(ctx interface{}, request interface{}, response interface{}) error {
				resp := response.(SecurityResponse)
				resp.SetStatusCode(http.StatusUnauthorized)
				_ = resp.Write([]byte("unauthorized"))
				return nil
			},
			wantStatus:     http.StatusUnauthorized,
			wantBody:       "unauthorized",
			wantNextCalled: false,
		},
		{
			name: "chain error writes 500",
			chainFunc: func(ctx interface{}, request interface{}, response interface{}) error {
				return errors.New("boom")
			},
			wantStatus:     http.StatusInternalServerError,
			wantBody:       "boom",
			wantNextCalled: false,
		},
		{
			name: "pass through calls next handler",
			chainFunc: func(ctx interface{}, request interface{}, response interface{}) error {
				return nil
			},
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler, nextCalled := newHandler(tt.chainFunc)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api", nil))

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
			if *nextCalled != tt.wantNextCalled {
				t.Errorf("nextHandler called = %v, want %v", *nextCalled, tt.wantNextCalled)
			}
		})
	}
}

// TestHttpResponseAdapter_HeaderAfterStatusCode 验证 SetStatusCode 之后设置的响应头仍生效（H3）。
func TestHttpResponseAdapter_HeaderAfterStatusCode(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	adapter := NewHttpResponseAdapter(rec)

	adapter.SetStatusCode(http.StatusUnauthorized)
	adapter.SetHeader("WWW-Authenticate", `Basic realm="test"`)
	_ = adapter.Write([]byte("Authentication required"))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if got := rec.Header().Get("WWW-Authenticate"); got != `Basic realm="test"` {
		t.Errorf("WWW-Authenticate = %q, want %q", got, `Basic realm="test"`)
	}
	if rec.Body.String() != "Authentication required" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "Authentication required")
	}
}

// TestHttpResponseAdapter_StatusCodeWithoutBody 验证仅设置状态码也会被提交到响应。
func TestHttpResponseAdapter_StatusCodeWithoutBody(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	adapter := NewHttpResponseAdapter(rec)

	adapter.SetStatusCode(http.StatusNoContent)
	adapter.flush()

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if !adapter.headersWritten() {
		t.Error("expected headers to be marked as written")
	}
}

// TestBasicAuthenticationFilter_SetsRequestAttribute 验证 Basic 认证结果写入请求属性（H1）。
func TestBasicAuthenticationFilter_SetsRequestAttribute(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	encoded := base64.StdEncoding.EncodeToString([]byte("admin:admin123"))
	req := &mockSecurityRequest{method: "GET", uri: "/api/test"}
	req.SetHeader("Authorization", "Basic "+encoded)
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	filter := NewBasicAuthenticationFilter(authManager)
	if err := filter.DoFilter(ctx, req, resp, chain); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	authVal, exists := req.GetAttribute("security.currentAuthentication")
	if !exists || authVal == nil {
		t.Fatal("expected authentication to be set in request attribute")
	}
	if auth, ok := authVal.(Authentication); ok {
		if got := extractPrincipalName(auth); got != "admin" {
			t.Errorf("principal = %q, want %q", got, "admin")
		}
	}
}
