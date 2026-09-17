package security

import (
	"testing"

	"github.com/xudefa/enhance/log"
)

func TestSecurityConfig_WithFilterBefore(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	customFilter := &mockSecurityFilter{}
	authFilter := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", authManager, log.Build(), WithDefaultSuccessURL("/dashboard"), WithFailureURL("/login?error"))

	cfg := NewSecurityConfig(
		WithAuthenticationManager(authManager),
		WithFilterBefore(customFilter, authFilter),
	)

	if len(cfg.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(cfg.Filters))
	}
	if cfg.Filters[0].Before == nil {
		t.Error("expected Before filter to be set")
	}
}

func TestSecurityConfig_WithFilterAfter(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	customFilter := &mockSecurityFilter{}
	authFilter := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", authManager, log.Build(), WithDefaultSuccessURL("/dashboard"), WithFailureURL("/login?error"))

	cfg := NewSecurityConfig(
		WithAuthenticationManager(authManager),
		WithFilterAfter(customFilter, authFilter),
	)

	if len(cfg.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(cfg.Filters))
	}
	if cfg.Filters[0].After == nil {
		t.Error("expected After filter to be set")
	}
}

func TestInsertFilterBefore(t *testing.T) {
	t.Parallel()

	filter1 := &mockSecurityFilter{order: 1}
	filter2 := &mockSecurityFilter{order: 2}
	filter3 := &mockSecurityFilter{order: 3}
	filters := []SecurityFilter{filter1, filter2, filter3}

	newFilter := &mockSecurityFilter{order: 0}
	filtersAfterInsert := insertFilterBefore(filters, newFilter, filter2)

	if len(filtersAfterInsert) != 4 {
		t.Fatalf("expected 4 filters, got %d", len(filtersAfterInsert))
	}
	// 验证newFilter在filter2之前
	if filtersAfterInsert[1] != newFilter {
		t.Error("expected newFilter to be before filter2")
	}
	if filtersAfterInsert[2] != filter2 {
		t.Error("expected filter2 to be at index 2")
	}
}

func TestInsertFilterBefore_NotFound(t *testing.T) {
	t.Parallel()

	filter1 := &mockSecurityFilter{order: 1}
	filter2 := &mockSecurityFilter{order: 2}
	filters := []SecurityFilter{filter1, filter2}

	newFilter := &mockSecurityFilter{order: 0}
	notFound := &mockSecurityFilter{order: 99}
	filtersAfterInsert := insertFilterBefore(filters, newFilter, notFound)

	if len(filtersAfterInsert) != 3 {
		t.Fatalf("expected 3 filters, got %d", len(filtersAfterInsert))
	}
	// 验证newFilter被添加到末尾
	if filtersAfterInsert[2] != newFilter {
		t.Error("expected newFilter to be appended at the end")
	}
}

func TestInsertFilterAfter(t *testing.T) {
	t.Parallel()

	filter1 := &mockSecurityFilter{order: 1}
	filter2 := &mockSecurityFilter{order: 2}
	filter3 := &mockSecurityFilter{order: 3}
	filters := []SecurityFilter{filter1, filter2, filter3}

	newFilter := &mockSecurityFilter{order: 0}
	filtersAfterInsert := insertFilterAfter(filters, newFilter, filter2)

	if len(filtersAfterInsert) != 4 {
		t.Fatalf("expected 4 filters, got %d", len(filtersAfterInsert))
	}
	// 验证newFilter在filter2之后
	if filtersAfterInsert[2] != newFilter {
		t.Error("expected newFilter to be after filter2")
	}
	if filtersAfterInsert[1] != filter2 {
		t.Error("expected filter2 to be at index 1")
	}
}

func TestInsertFilterAfter_NotFound(t *testing.T) {
	t.Parallel()

	filter1 := &mockSecurityFilter{order: 1}
	filter2 := &mockSecurityFilter{order: 2}
	filters := []SecurityFilter{filter1, filter2}

	newFilter := &mockSecurityFilter{order: 0}
	notFound := &mockSecurityFilter{order: 99}
	filtersAfterInsert := insertFilterAfter(filters, newFilter, notFound)

	if len(filtersAfterInsert) != 3 {
		t.Fatalf("expected 3 filters, got %d", len(filtersAfterInsert))
	}
	// 验证newFilter被添加到末尾
	if filtersAfterInsert[2] != newFilter {
		t.Error("expected newFilter to be appended at the end")
	}
}
