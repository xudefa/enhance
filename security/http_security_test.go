package security

import (
	"testing"
)

func TestNewHttpSecurity(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity()
	if http == nil {
		t.Fatal("expected non-nil HttpSecurity")
	}
}

func TestHttpSecurity_FormLogin(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity()
	result := http.FormLogin("/login", "/dashboard")

	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHttpSecurity_FormLogin_Defaults(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity()
	http.FormLogin("")

	h := http.(*httpSecurity)
	if h.loginProcessingUrl != "/login" {
		t.Errorf("expected default loginProcessingUrl /login, got %s", h.loginProcessingUrl)
	}
	if h.defaultSuccessUrl != "/" {
		t.Errorf("expected default success url /, got %s", h.defaultSuccessUrl)
	}
	if h.failureUrl != "/login?error" {
		t.Errorf("expected default failure url /login?error, got %s", h.failureUrl)
	}
}

func TestHttpSecurity_HttpBasic(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity()
	result := http.HttpBasic()

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	h := http.(*httpSecurity)
	if !h.httpBasicEnabled {
		t.Error("expected httpBasicEnabled to be true")
	}
	if h.realmName != "Secured Area" {
		t.Errorf("expected realmName 'Secured Area', got %s", h.realmName)
	}
}

func TestHttpSecurity_Logout(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity()
	result := http.Logout("/logout")

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	h := http.(*httpSecurity)
	if h.logoutUrl != "/logout" {
		t.Errorf("expected logoutUrl /logout, got %s", h.logoutUrl)
	}
}

func TestHttpSecurity_Logout_DefaultUrl(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity()
	http.Logout("")

	h := http.(*httpSecurity)
	if h.logoutUrl != "/logout" {
		t.Errorf("expected default logoutUrl /logout, got %s", h.logoutUrl)
	}
}

func TestHttpSecurity_Anonymous(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity()
	result := http.Anonymous()

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	h := http.(*httpSecurity)
	if h.anonymousFilter == nil {
		t.Error("expected anonymousFilter to be set")
	}
}

func TestHttpSecurity_Csrf(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity()
	result := http.Csrf()

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	h := http.(*httpSecurity)
	if !h.csrfEnabled {
		t.Error("expected csrfEnabled to be true")
	}
}

func TestHttpSecurity_Build_NoAuthManager(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity()
	_, err := http.Build()
	if err == nil {
		t.Error("expected error when AuthenticationManager is nil")
	}
}

func TestHttpSecurity_Build_WithAuthManager(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity()
	mgr := NewProviderManager()
	http.AuthenticationManager(mgr)

	chain, err := http.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain == nil {
		t.Fatal("expected non-nil chain")
	}
}

func TestHttpSecurity_AuthorizeRequests(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity()
	http.AuthorizeRequests(func(authz AuthorizeRequests) {
		authz.AnyRequest().Authenticated()
	})

	h := http.(*httpSecurity)
	if len(h.authorizeRules) == 0 {
		t.Error("expected authorize rules to be set")
	}
}

func TestHttpSecurity_AuthenticationManager(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity()
	mgr := NewProviderManager()
	http.AuthenticationManager(mgr)

	h := http.(*httpSecurity)
	if h.authenticationManager == nil {
		t.Error("expected authenticationManager to be set")
	}
}

func TestHttpSecurity_ExceptionHandling_Setter(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity()
	handler := NewHttp403ForbiddenAccessDeniedHandler()
	entryPoint := NewHttp401UnauthorizedEntryPoint()
	http.ExceptionHandling(handler, entryPoint)

	h := http.(*httpSecurity)
	if h.exceptionTranslationFilter == nil {
		t.Error("expected exceptionTranslationFilter to be set")
	}
}

func TestDefaultSecurityFilterChain_DoFilter_Noop(t *testing.T) {
	t.Parallel()

	chain := &DefaultSecurityFilterChain{}
	err := chain.DoFilter(nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaultSecurityFilterChain_MatchesAlwaysTrue(t *testing.T) {
	t.Parallel()

	chain := &DefaultSecurityFilterChain{}
	if !chain.Matches("anything") {
		t.Error("expected Matches to return true")
	}
}

func TestDefaultSecurityFilterChain_GetFiltersEmpty(t *testing.T) {
	t.Parallel()

	chain := &DefaultSecurityFilterChain{}
	if chain.GetFilters() != nil {
		t.Error("expected nil filters")
	}
}

func TestNewWebSecurity(t *testing.T) {
	t.Parallel()

	ws := NewWebSecurity()
	if ws == nil {
		t.Fatal("expected non-nil WebSecurity")
	}
}

func TestWebSecurity_Build(t *testing.T) {
	t.Parallel()

	ws := NewWebSecurity()
	_, err := ws.Build()
	if err == nil {
		t.Error("expected error from WebSecurity.Build()")
	}
}

func TestExpressionInterceptUrlRegistry_PermitAll(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity().(*httpSecurity)
	reg := &expressionInterceptUrlRegistry{
		httpSecurity: http,
		patterns:     []string{"/public/**"},
	}
	result := reg.PermitAll()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(http.authorizeRules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(http.authorizeRules))
	}
	if http.authorizeRules[0].attrs[0] != "permitAll" {
		t.Errorf("expected permitAll, got %s", http.authorizeRules[0].attrs[0])
	}
}

func TestExpressionInterceptUrlRegistry_DenyAll(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity().(*httpSecurity)
	reg := &expressionInterceptUrlRegistry{
		httpSecurity: http,
		patterns:     []string{"/admin/**"},
	}
	reg.DenyAll()
	if len(http.authorizeRules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(http.authorizeRules))
	}
	if http.authorizeRules[0].attrs[0] != "denyAll" {
		t.Errorf("expected denyAll, got %s", http.authorizeRules[0].attrs[0])
	}
}

func TestExpressionInterceptUrlRegistry_HasRole(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity().(*httpSecurity)
	reg := &expressionInterceptUrlRegistry{
		httpSecurity: http,
		patterns:     []string{"/admin/**"},
	}
	reg.HasRole("ADMIN")
	if len(http.authorizeRules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(http.authorizeRules))
	}
	if http.authorizeRules[0].attrs[0] != "hasRole('ADMIN')" {
		t.Errorf("expected hasRole('ADMIN'), got %s", http.authorizeRules[0].attrs[0])
	}
}

func TestExpressionInterceptUrlRegistry_HasAnyRole(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity().(*httpSecurity)
	reg := &expressionInterceptUrlRegistry{
		httpSecurity: http,
		patterns:     []string{"/api/**"},
	}
	reg.HasAnyRole("ADMIN", "USER")
	if len(http.authorizeRules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(http.authorizeRules))
	}
	if http.authorizeRules[0].attrs[0] != "hasAnyRole('ADMIN','USER')" {
		t.Errorf("expected hasAnyRole('ADMIN','USER'), got %s", http.authorizeRules[0].attrs[0])
	}
}

func TestExpressionInterceptUrlRegistry_HasAuthority(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity().(*httpSecurity)
	reg := &expressionInterceptUrlRegistry{
		httpSecurity: http,
		patterns:     []string{"/data/**"},
	}
	reg.HasAuthority("read")
	if http.authorizeRules[0].attrs[0] != "hasAuthority('read')" {
		t.Errorf("expected hasAuthority('read'), got %s", http.authorizeRules[0].attrs[0])
	}
}

func TestExpressionInterceptUrlRegistry_HasAnyAuthority(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity().(*httpSecurity)
	reg := &expressionInterceptUrlRegistry{
		httpSecurity: http,
		patterns:     []string{"/data/**"},
	}
	reg.HasAnyAuthority("read", "write")
	if http.authorizeRules[0].attrs[0] != "hasAnyAuthority('read','write')" {
		t.Errorf("expected hasAnyAuthority('read','write'), got %s", http.authorizeRules[0].attrs[0])
	}
}

func TestHttpSecurity_AntMatchers(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity().(*httpSecurity)
	authorizer := &httpSecurityAuthorizer{httpSecurity: http}
	reg := authorizer.AntMatchers("/api/**")
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestHttpSecurity_AnyRequest(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity().(*httpSecurity)
	authorizer := &httpSecurityAuthorizer{httpSecurity: http}
	reg := authorizer.AnyRequest()
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestExpressionInterceptUrlRegistry_EmptyPatterns(t *testing.T) {
	t.Parallel()

	http := NewHttpSecurity().(*httpSecurity)
	reg := &expressionInterceptUrlRegistry{
		httpSecurity: http,
		patterns:     []string{},
	}
	result := reg.addRule([]string{"permitAll"})
	if len(http.authorizeRules) != 0 {
		t.Error("expected no rules for empty patterns")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}
