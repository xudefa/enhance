package security

import (
	"context"
	"fmt"
	"testing"
)

func TestNewFilterChainProxy(t *testing.T) {
	t.Parallel()

	chain := &DefaultSecurityFilterChain{}
	proxy := newFilterChainProxy(nil, chain)
	if proxy == nil {
		t.Fatal("expected non-nil proxy")
	}
	if proxy.chain != chain {
		t.Error("expected chain to be set")
	}
}

func TestFilterChainProxy_DoFilterWithChain_NilProxy(t *testing.T) {
	t.Parallel()

	var p *filterChainProxy
	err := p.doFilterWithChain(context.Background(), nil, nil, &mockFilterChain{})
	if err != nil {
		t.Fatalf("expected nil error for nil proxy, got %v", err)
	}
}

func TestFilterChainProxy_DoFilterWithChain_NilChain(t *testing.T) {
	t.Parallel()

	proxy := newFilterChainProxy(nil, &DefaultSecurityFilterChain{})
	err := proxy.doFilterWithChain(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("expected nil error for nil chain, got %v", err)
	}
}

func TestFilterChainProxy_DoFilterInternal_ExceedsIndex(t *testing.T) {
	t.Parallel()

	chain := &DefaultSecurityFilterChain{}
	proxy := newFilterChainProxy(nil, chain)

	err := proxy.doFilterInternal(context.Background(), nil, nil, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSecurityFilterChainAdapter_DoFilter_TypeErrors(t *testing.T) {
	t.Parallel()

	adapter := &securityFilterChainProxy{proxy: newFilterChainProxy(nil, &DefaultSecurityFilterChain{})}

	err := adapter.DoFilter("notContext", nil, nil)
	if err == nil {
		t.Error("expected error for non-context")
	}

	err = adapter.DoFilter(context.Background(), "notReq", nil)
	if err == nil {
		t.Error("expected error for non-request")
	}

	err = adapter.DoFilter(context.Background(), newMockSecurityRequest("GET", "/", nil), "notResp")
	if err == nil {
		t.Error("expected error for non-response")
	}
}

func TestSecurityFilterChainAdapter_Matches(t *testing.T) {
	t.Parallel()

	adapter := &securityFilterChainProxy{proxy: newFilterChainProxy(nil, &DefaultSecurityFilterChain{})}

	if !adapter.Matches(newMockSecurityRequest("GET", "/", nil)) {
		t.Error("expected Matches to return true for SecurityRequest")
	}
	if adapter.Matches("notRequest") {
		t.Error("expected Matches to return false for non-SecurityRequest")
	}
}

func TestSecurityFilterChainAdapter_GetFilters(t *testing.T) {
	t.Parallel()

	adapter := &securityFilterChainProxy{
		proxy: newFilterChainProxy([]SecurityFilter{NewAnonymousAuthenticationFilter()}, &DefaultSecurityFilterChain{}),
	}

	filters := adapter.GetFilters()
	if len(filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(filters))
	}
}

func TestDefaultSecurityFilterChain_DoFilter(t *testing.T) {
	t.Parallel()

	chain := &DefaultSecurityFilterChain{}
	err := chain.DoFilter(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaultSecurityFilterChain_Matches(t *testing.T) {
	t.Parallel()

	chain := &DefaultSecurityFilterChain{}
	if !chain.Matches("anything") {
		t.Error("expected Matches to always return true")
	}
}

func TestDefaultSecurityFilterChain_GetFilters(t *testing.T) {
	t.Parallel()

	chain := &DefaultSecurityFilterChain{}
	filters := chain.GetFilters()
	if filters != nil {
		t.Errorf("expected nil filters, got %v", filters)
	}
}

func TestFilterChainAdapter_DoFilter_TypeErrors(t *testing.T) {
	t.Parallel()

	vfc := &virtualFilterChain{
		proxy: newFilterChainProxy(nil, &DefaultSecurityFilterChain{}),
		index: 0,
	}
	adapter := &filterChainAdapter{vfc: vfc}

	err := adapter.DoFilter("notContext", nil, nil)
	if err == nil {
		t.Error("expected error for non-context")
	}

	err = adapter.DoFilter(context.Background(), "notReq", nil)
	if err == nil {
		t.Error("expected error for non-request")
	}

	err = adapter.DoFilter(context.Background(), newMockSecurityRequest("GET", "/", nil), "notResp")
	if err == nil {
		t.Error("expected error for non-response")
	}
}

func TestFilterChainAdapter_AddFilter(t *testing.T) {
	t.Parallel()

	vfc := &virtualFilterChain{
		proxy: newFilterChainProxy(nil, &DefaultSecurityFilterChain{}),
		index: 0,
	}
	adapter := &filterChainAdapter{vfc: vfc}

	adapter.AddFilter(NewAnonymousAuthenticationFilter())

	filters := adapter.GetFilters()
	if filters != nil {
		t.Error("expected nil from virtual chain GetFilters")
	}
}

func TestFilterChainAdapter_GetFilters(t *testing.T) {
	t.Parallel()

	vfc := &virtualFilterChain{
		proxy: newFilterChainProxy(nil, &DefaultSecurityFilterChain{}),
		index: 0,
	}
	adapter := &filterChainAdapter{vfc: vfc}

	if adapter.GetFilters() != nil {
		t.Error("expected nil from virtual chain GetFilters")
	}
}

func TestVirtualFilterChain_DoFilter(t *testing.T) {
	t.Parallel()

	f1 := NewAnonymousAuthenticationFilter()
	proxy := newFilterChainProxy([]SecurityFilter{f1}, &DefaultSecurityFilterChain{})
	vfc := &virtualFilterChain{proxy: proxy, index: 0}

	err := vfc.DoFilter(context.Background(), newMockSecurityRequest("GET", "/", nil), newMockSecurityResponse())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSecurityFilterChainProxy_DoFilterWithChain_CallsChain(t *testing.T) {
	t.Parallel()

	chain := &mockFilterChain{}
	proxy := newFilterChainProxy(nil, &DefaultSecurityFilterChain{})

	err := proxy.doFilterWithChain(context.Background(), newMockSecurityRequest("GET", "/", nil), newMockSecurityResponse(), chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called")
	}
}

type securityFilterChainProxy struct {
	proxy *filterChainProxy
}

func (a *securityFilterChainProxy) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
	ctxVal, ok := ctx.(context.Context)
	if !ok {
		return fmt.Errorf("securityFilterChainProxy: ctx must be context.Context")
	}
	req, ok := request.(SecurityRequest)
	if !ok {
		return fmt.Errorf("securityFilterChainProxy: request must be SecurityRequest")
	}
	resp, ok := response.(SecurityResponse)
	if !ok {
		return fmt.Errorf("securityFilterChainProxy: response must be SecurityResponse")
	}
	return a.proxy.doFilterWithChain(ctxVal, req, resp, &filterChainAdapter{vfc: &virtualFilterChain{proxy: a.proxy, index: 0}})
}

func (a *securityFilterChainProxy) Matches(request interface{}) bool {
	_, ok := request.(SecurityRequest)
	return ok
}

func (a *securityFilterChainProxy) GetFilters() []SecurityFilter {
	return a.proxy.filters
}
