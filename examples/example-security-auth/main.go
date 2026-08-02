// Package main demonstrates the enhance security framework:
// security filter chain setup, username/password authentication,
// role-based authorization, and access control.
package main

import (
	"context"
	"fmt"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security"
	"github.com/xudefa/enhance/security/authorization"
)

// SimpleLogger implements log.Logger for demo purposes.
type SimpleLogger struct{}

func (l *SimpleLogger) Debug(_ context.Context, msg string, _ ...log.KeyValue) {
	fmt.Printf("  [DEBUG] %s\n", msg)
}
func (l *SimpleLogger) Info(_ context.Context, msg string, _ ...log.KeyValue) {
	fmt.Printf("  [INFO] %s\n", msg)
}
func (l *SimpleLogger) Warn(_ context.Context, msg string, _ ...log.KeyValue) {
	fmt.Printf("  [WARN] %s\n", msg)
}
func (l *SimpleLogger) Error(_ context.Context, msg string, _ ...log.KeyValue) {
	fmt.Printf("  [ERROR] %s\n", msg)
}
func (l *SimpleLogger) Sync() error                              { return nil }
func (l *SimpleLogger) With(_ context.Context, _ ...log.KeyValue) log.Logger { return l }

func main() {
	fmt.Println("=== enhance Security Auth Example ===")
	fmt.Println()

	ctx := context.Background()
	logger := &SimpleLogger{}

	// ---- 1. Set up UserDetailsService with in-memory users ----
	fmt.Println("--- 1. Creating Users ---")
	userDetailsService := security.NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN", "ROLE_USER"})
	userDetailsService.CreateUser("user", "user123", []string{"ROLE_USER"})
	userDetailsService.CreateUser("manager", "mgr123", []string{"ROLE_MANAGER", "ROLE_USER"})
	fmt.Printf("  Created %d users\n", userDetailsService.UserCount())

	// ---- 2. Create PasswordEncoder ----
	encoder := security.NewNoOpPasswordEncoder()

	// ---- 3. Create AuthenticationManager with DaoAuthenticationProvider ----
	fmt.Println()
	fmt.Println("--- 2. Setting up Authentication Manager ---")
	provider := security.NewDaoAuthenticationProvider(userDetailsService, encoder, logger)
	authManager := security.NewProviderManager(provider)

	// ---- 4. Authenticate users ----
	fmt.Println()
	fmt.Println("--- 3. Authenticating Users ---")

	// Successful login
	token := security.NewUsernamePasswordAuthenticationToken("admin", "admin123")
	result, err := authManager.Authenticate(ctx, token)
	if err != nil {
		fmt.Printf("  admin auth failed: %v\n", err)
	} else {
		fmt.Printf("  admin authenticated: %v, authorities: %v\n",
			result.Authenticated(), result.Authorities())
	}

	// Successful login
	token = security.NewUsernamePasswordAuthenticationToken("user", "user123")
	result, err = authManager.Authenticate(ctx, token)
	if err != nil {
		fmt.Printf("  user auth failed: %v\n", err)
	} else {
		fmt.Printf("  user authenticated: %v, authorities: %v\n",
			result.Authenticated(), result.Authorities())
	}

	// Failed login (wrong password)
	token = security.NewUsernamePasswordAuthenticationToken("admin", "wrongpassword")
	_, err = authManager.Authenticate(ctx, token)
	if err != nil {
		fmt.Printf("  admin wrong password: %v (expected)\n", err)
	}

	// Failed login (non-existent user)
	token = security.NewUsernamePasswordAuthenticationToken("unknown", "pass")
	_, err = authManager.Authenticate(ctx, token)
	if err != nil {
		fmt.Printf("  unknown user: %v (expected)\n", err)
	}

	// ---- 5. Role-based authorization ----
	fmt.Println()
	fmt.Println("--- 4. Role-Based Authorization ---")

	voter := authorization.NewWebExpressionVoter()
	decisionManager := authorization.NewAffirmativeBased(voter)

	adminAuth := security.NewAuthenticatedUsernamePasswordAuthenticationToken(
		"admin", []string{"ROLE_ADMIN", "ROLE_USER"})
	userAuth := security.NewAuthenticatedUsernamePasswordAuthenticationToken(
		"user", []string{"ROLE_USER"})
	managerAuth := security.NewAuthenticatedUsernamePasswordAuthenticationToken(
		"manager", []string{"ROLE_MANAGER", "ROLE_USER"})

	// Test admin access to /api/admin
	fmt.Println("  Resource: /api/admin")
	testDecision(ctx, decisionManager, adminAuth, "/api/admin",
		[]string{"hasRole('ADMIN')"})
	testDecision(ctx, decisionManager, userAuth, "/api/admin",
		[]string{"hasRole('ADMIN')"})
	testDecision(ctx, decisionManager, managerAuth, "/api/admin",
		[]string{"hasRole('ADMIN')"})

	// Test access to /api/users (any authenticated user)
	fmt.Println("  Resource: /api/users")
	testDecision(ctx, decisionManager, adminAuth, "/api/users",
		[]string{"authenticated"})
	testDecision(ctx, decisionManager, userAuth, "/api/users",
		[]string{"authenticated"})

	// Test hasAnyRole
	fmt.Println("  Resource: /api/reports (any of ADMIN, MANAGER)")
	testDecision(ctx, decisionManager, adminAuth, "/api/reports",
		[]string{"hasAnyRole('ADMIN','MANAGER')"})
	testDecision(ctx, decisionManager, userAuth, "/api/reports",
		[]string{"hasAnyRole('ADMIN','MANAGER')"})
	testDecision(ctx, decisionManager, managerAuth, "/api/reports",
		[]string{"hasAnyRole('ADMIN','MANAGER')"})

	// Test denyAll
	fmt.Println("  Resource: /api/secret")
	testDecision(ctx, decisionManager, adminAuth, "/api/secret",
		[]string{"denyAll"})

	// Test permitAll
	fmt.Println("  Resource: /public")
	testDecision(ctx, decisionManager, userAuth, "/public",
		[]string{"permitAll"})

	// ---- 6. SecurityBuilder demo ----
	fmt.Println()
	fmt.Println("--- 5. SecurityBuilder Configuration ---")
	secConfig := security.NewSecurityBuilder().
		UserDetailsService(userDetailsService).
		PasswordEncoder(encoder).
		AuthenticationManager(authManager).
		EnableCsrf().
		EnableAnonymous().
		EnableFormLogin("/login", "/dashboard").
		EnableLogout("/logout").
		EnableHttpBasic().
		Build()
	fmt.Printf("  Security config built: %T\n", secConfig)

	// ---- 7. Unanimous-based decision manager ----
	fmt.Println()
	fmt.Println("--- 6. Unanimous-Based Decision Manager ---")
	unanimousMgr := authorization.NewUnanimousBased(voter)

	fmt.Println("  Admin (has ROLE_ADMIN):")
	testDecision(ctx, unanimousMgr, adminAuth, "/resource",
		[]string{"hasRole('ADMIN')"})

	fmt.Println("  User (has ROLE_USER only):")
	testDecision(ctx, unanimousMgr, userAuth, "/resource",
		[]string{"hasRole('ADMIN')"})

	fmt.Println("  Manager (has ROLE_MANAGER):")
	testDecision(ctx, unanimousMgr, managerAuth, "/resource",
		[]string{"hasRole('ADMIN')"})

	fmt.Println()
	fmt.Println("=== Example completed successfully ===")
}

// testDecision checks authorization and prints the result.
func testDecision(ctx context.Context, mgr authorization.AccessDecisionManager,
	auth security.Authentication, resource string, attributes []string) {
	err := mgr.Decide(ctx, auth, resource, attributes)
	status := "GRANTED"
	if err != nil {
		status = fmt.Sprintf("DENIED (%v)", err)
	}
	fmt.Printf("    %s -> %s %v: %s\n",
		auth.Principal(), resource, attributes, status)
}
