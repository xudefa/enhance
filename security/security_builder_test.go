package security

import (
	"testing"
)

func TestNewSecurityBuilder(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	if b == nil {
		t.Fatal("expected non-nil builder")
	}
}

func TestSecurityBuilder_AuthenticationManager(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	mgr := NewProviderManager()
	result := b.AuthenticationManager(mgr)

	if result != b {
		t.Error("expected builder to be returned for chaining")
	}
	if b.authManager == nil {
		t.Error("expected authManager to be set")
	}
}

func TestSecurityBuilder_UserDetailsService(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	service := NewInMemoryUserDetailsService()
	result := b.UserDetailsService(service)

	if result != b {
		t.Error("expected builder to be returned for chaining")
	}
	if b.userDetailsService == nil {
		t.Error("expected userDetailsService to be set")
	}
}

func TestSecurityBuilder_PasswordEncoder(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	encoder := NewNoOpPasswordEncoder()
	result := b.PasswordEncoder(encoder)

	if result != b {
		t.Error("expected builder to be returned for chaining")
	}
	if b.passwordEncoder == nil {
		t.Error("expected passwordEncoder to be set")
	}
}

func TestSecurityBuilder_AccessDecisionManagerFunc(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	voter := NewWebExpressionVoter()
	mgr := NewAffirmativeBased(voter)
	result := b.AccessDecisionManager(mgr)

	if result != b {
		t.Error("expected builder to be returned for chaining")
	}
	if b.accessDecisionMgr == nil {
		t.Error("expected accessDecisionMgr to be set")
	}
}

func TestSecurityBuilder_AddFilter(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	f := NewAnonymousAuthenticationFilter()
	result := b.AddFilter(f)

	if result != b {
		t.Error("expected builder to be returned for chaining")
	}
	if len(b.filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(b.filters))
	}
}

func TestSecurityBuilder_AddFilterBeforeFunc(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	f1 := NewAnonymousAuthenticationFilter()
	f2 := NewBasicAuthenticationFilter(NewProviderManager())
	b.AddFilter(f1)
	b.AddFilterBefore(f2, f1)

	if len(b.filters) != 2 {
		t.Errorf("expected 2 filters, got %d", len(b.filters))
	}
}

func TestSecurityBuilder_AddFilterAfter(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	f1 := NewAnonymousAuthenticationFilter()
	f2 := NewBasicAuthenticationFilter(NewProviderManager())
	b.AddFilter(f1)
	b.AddFilterAfter(f2, f1)

	if len(b.filters) != 2 {
		t.Errorf("expected 2 filters, got %d", len(b.filters))
	}
}

func TestSecurityBuilder_EnableAnonymous(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	result := b.EnableAnonymous()

	if result != b {
		t.Error("expected builder to be returned for chaining")
	}
	if !b.anonymous {
		t.Error("expected anonymous to be true")
	}
}

func TestSecurityBuilder_EnableCsrf(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	result := b.EnableCsrf()

	if result != b {
		t.Error("expected builder to be returned for chaining")
	}
	if !b.csrf {
		t.Error("expected csrf to be true")
	}
}

func TestSecurityBuilder_EnableFormLogin(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	result := b.EnableFormLogin("/login", "/dashboard")

	if result != b {
		t.Error("expected builder to be returned for chaining")
	}
	if b.formLogin == nil {
		t.Fatal("expected formLogin to be set")
	}
	if b.formLogin.processingUrl != "/login" {
		t.Errorf("expected processingUrl /login, got %s", b.formLogin.processingUrl)
	}
	if b.formLogin.defaultSuccessUrl != "/dashboard" {
		t.Errorf("expected defaultSuccessUrl /dashboard, got %s", b.formLogin.defaultSuccessUrl)
	}
}

func TestSecurityBuilder_EnableFormLogin_NoSuccessUrl(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	b.EnableFormLogin("/login")

	if b.formLogin.defaultSuccessUrl != "" {
		t.Errorf("expected empty defaultSuccessUrl, got %s", b.formLogin.defaultSuccessUrl)
	}
}

func TestSecurityBuilder_EnableHttpBasic(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	result := b.EnableHttpBasic()

	if result != b {
		t.Error("expected builder to be returned for chaining")
	}
	if !b.httpBasic {
		t.Error("expected httpBasic to be true")
	}
}

func TestSecurityBuilder_EnableLogout(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	result := b.EnableLogout("/logout")

	if result != b {
		t.Error("expected builder to be returned for chaining")
	}
	if b.logoutConfig == nil {
		t.Fatal("expected logoutConfig to be set")
	}
	if b.logoutConfig.url != "/logout" {
		t.Errorf("expected url /logout, got %s", b.logoutConfig.url)
	}
}

func TestSecurityBuilder_EnableLogout_WithHandler(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	handler := &mockLogoutSuccessHandler{}
	b.EnableLogout("/logout", handler)

	if b.logoutConfig.successHandler == nil {
		t.Error("expected successHandler to be set")
	}
}

func TestSecurityBuilder_Build(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	config := b.Build()

	if config == nil {
		t.Fatal("expected non-nil configurer")
	}
}

func TestBuiltSecurityConfig_Configure(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	mgr := NewProviderManager()
	service := NewInMemoryUserDetailsService()
	encoder := NewNoOpPasswordEncoder()
	voter := NewWebExpressionVoter()
	adm := NewAffirmativeBased(voter)

	b.AuthenticationManager(mgr)
	b.UserDetailsService(service)
	b.PasswordEncoder(encoder)
	b.AccessDecisionManager(adm)
	b.EnableAnonymous()
	b.EnableCsrf()
	b.EnableFormLogin("/login")
	b.EnableHttpBasic()
	b.EnableLogout("/logout")

	config := b.Build()
	http := NewHttpSecurity()

	err := config.Configure(http)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	h := http.(*httpSecurity)
	if h.authenticationManager == nil {
		t.Error("expected authenticationManager to be set")
	}
	if h.userDetailsService == nil {
		t.Error("expected userDetailsService to be set")
	}
	if h.passwordEncoder == nil {
		t.Error("expected passwordEncoder to be set")
	}
	if h.accessDecisionManager == nil {
		t.Error("expected accessDecisionManager to be set")
	}
	if h.anonymousFilter == nil {
		t.Error("expected anonymousFilter to be set")
	}
	if !h.csrfEnabled {
		t.Error("expected csrfEnabled to be true")
	}
	if !h.formLoginEnabled {
		t.Error("expected formLoginEnabled to be true")
	}
	if !h.httpBasicEnabled {
		t.Error("expected httpBasicEnabled to be true")
	}
	if h.logoutUrl == "" {
		t.Error("expected logoutUrl to be set")
	}
}

func TestBuiltSecurityConfig_Configure_FormLoginDefaultUrl(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	b.EnableFormLogin("/login")

	config := b.Build()
	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	h := http.(*httpSecurity)
	if h.loginProcessingUrl != "/login" {
		t.Errorf("expected /login, got %s", h.loginProcessingUrl)
	}
}

func TestBuiltSecurityConfig_Configure_LogoutWithHandler(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	handler := &mockLogoutSuccessHandler{}
	b.EnableLogout("/logout", handler)

	config := b.Build()
	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	h := http.(*httpSecurity)
	if h.logoutSuccessHandler == nil {
		t.Error("expected successHandler to be set")
	}
}

func TestSecurityBuilder_Chaining(t *testing.T) {
	t.Parallel()

	b := NewSecurityBuilder()
	result := b.
		AuthenticationManager(NewProviderManager()).
		UserDetailsService(NewInMemoryUserDetailsService()).
		PasswordEncoder(NewNoOpPasswordEncoder()).
		AddFilter(NewAnonymousAuthenticationFilter()).
		EnableAnonymous().
		EnableCsrf().
		EnableFormLogin("/login").
		EnableHttpBasic().
		EnableLogout("/logout")

	if result != b {
		t.Error("expected chaining to return same builder")
	}
	if b.filters == nil || len(b.filters) == 0 {
		t.Error("expected filters to be added")
	}
}
