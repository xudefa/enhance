// Package main 演示 enhance 安全框架：
// 安全过滤器链设置、用户名/密码认证、基于角色的授权和访问控制。
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

// Debug 输出调试级别日志。
func (l *SimpleLogger) Debug(_ context.Context, msg string, _ ...log.KeyValue) {
	fmt.Printf("  [DEBUG] %s\n", msg)
}

// Info 输出信息级别日志。
func (l *SimpleLogger) Info(_ context.Context, msg string, _ ...log.KeyValue) {
	fmt.Printf("  [INFO] %s\n", msg)
}

// Warn 输出警告级别日志。
func (l *SimpleLogger) Warn(_ context.Context, msg string, _ ...log.KeyValue) {
	fmt.Printf("  [WARN] %s\n", msg)
}

// Error 输出错误级别日志。
func (l *SimpleLogger) Error(_ context.Context, msg string, _ ...log.KeyValue) {
	fmt.Printf("  [ERROR] %s\n", msg)
}

// Sync 刷新日志缓冲区。
func (l *SimpleLogger) Sync() error { return nil }

// With 返回带有附加字段的日志实例。
func (l *SimpleLogger) With(_ context.Context, _ ...log.KeyValue) log.Logger { return l }

func main() {
	fmt.Println("=== enhance Security Auth Example ===")
	fmt.Println()

	ctx := context.Background()
	logger := &SimpleLogger{}

	userDetailsService, encoder := setupUsersAndEncoder()
	authManager := setupAuthManager(userDetailsService, encoder, logger)
	voter := authorization.NewWebExpressionVoter()

	authenticateUsers(ctx, authManager)
	demoRoleAuthorization(ctx, voter)
	demoSecurityBuilder(userDetailsService, encoder, authManager)
	demoUnanimousDecision(ctx, voter)

	fmt.Println()
	fmt.Println("=== Example completed successfully ===")
}

// setupUsersAndEncoder 创建内存用户与密码编码器。
func setupUsersAndEncoder() (*security.InMemoryUserDetailsService, *security.NoOpPasswordEncoder) {
	// ---- 1. Set up UserDetailsService with in-memory users ----
	fmt.Println("--- 1. Creating Users ---")
	userDetailsService := security.NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN", "ROLE_USER"})
	userDetailsService.CreateUser("user", "user123", []string{"ROLE_USER"})
	userDetailsService.CreateUser("manager", "mgr123", []string{"ROLE_MANAGER", "ROLE_USER"})
	fmt.Printf("  Created %d users\n", userDetailsService.UserCount())

	// ---- 2. Create PasswordEncoder ----
	return userDetailsService, security.NewNoOpPasswordEncoder()
}

// setupAuthManager 创建基于 Dao 的认证管理器。
func setupAuthManager(userDetailsService *security.InMemoryUserDetailsService, encoder *security.NoOpPasswordEncoder, logger *SimpleLogger) security.AuthenticationManager {
	// ---- 3. Create AuthenticationManager with DaoAuthenticationProvider ----
	fmt.Println()
	fmt.Println("--- 2. Setting up Authentication Manager ---")
	provider := security.NewDaoAuthenticationProvider(userDetailsService, encoder, logger)
	return security.NewProviderManager(provider)
}

// authenticateUsers 演示成功与失败的登录认证流程。
func authenticateUsers(ctx context.Context, authManager security.AuthenticationManager) {
	// ---- 4. Authenticate users ----
	fmt.Println()
	fmt.Println("--- 3. Authenticating Users ---")

	// Successful login
	token := security.NewUsernamePasswordAuthenticationToken("admin", "admin123")
	authResult, err := authManager.Authenticate(ctx, token)
	if err != nil {
		fmt.Printf("  admin auth failed: %v\n", err)
	} else {
		fmt.Printf("  admin authenticated: %v, authorities: %v\n",
			authResult.Authenticated(), authResult.Authorities())
	}

	// Successful login
	token = security.NewUsernamePasswordAuthenticationToken("user", "user123")
	authResult, err = authManager.Authenticate(ctx, token)
	if err != nil {
		fmt.Printf("  user auth failed: %v\n", err)
	} else {
		fmt.Printf("  user authenticated: %v, authorities: %v\n",
			authResult.Authenticated(), authResult.Authorities())
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
}

// demoRoleAuthorization 演示基于角色的授权决策。
func demoRoleAuthorization(ctx context.Context, voter authorization.AccessDecisionVoter) {
	// ---- 5. Role-based authorization ----
	fmt.Println()
	fmt.Println("--- 4. Role-Based Authorization ---")

	decisionManager := authorization.NewAffirmativeBased(voter)

	adminAuth := security.NewAuthenticatedUsernamePasswordAuthenticationToken(
		"admin", []string{"ROLE_ADMIN", "ROLE_USER"})
	userAuth := security.NewAuthenticatedUsernamePasswordAuthenticationToken(
		"user", []string{"ROLE_USER"})
	managerAuth := security.NewAuthenticatedUsernamePasswordAuthenticationToken(
		"manager", []string{"ROLE_MANAGER", "ROLE_USER"})

	// Test admin access to /api/admin
	fmt.Println("  Resource: /api/admin")
	testDecision(testDecisionArgs{ctx, decisionManager, adminAuth, "/api/admin",
		[]string{"hasRole('ADMIN')"}})
	testDecision(testDecisionArgs{ctx, decisionManager, userAuth, "/api/admin",
		[]string{"hasRole('ADMIN')"}})
	testDecision(testDecisionArgs{ctx, decisionManager, managerAuth, "/api/admin",
		[]string{"hasRole('ADMIN')"}})

	// Test access to /api/users (any authenticated user)
	fmt.Println("  Resource: /api/users")
	testDecision(testDecisionArgs{ctx, decisionManager, adminAuth, "/api/users",
		[]string{"authenticated"}})
	testDecision(testDecisionArgs{ctx, decisionManager, userAuth, "/api/users",
		[]string{"authenticated"}})

	// Test hasAnyRole
	fmt.Println("  Resource: /api/reports (any of ADMIN, MANAGER)")
	testDecision(testDecisionArgs{ctx, decisionManager, adminAuth, "/api/reports",
		[]string{"hasAnyRole('ADMIN','MANAGER')"}})
	testDecision(testDecisionArgs{ctx, decisionManager, userAuth, "/api/reports",
		[]string{"hasAnyRole('ADMIN','MANAGER')"}})
	testDecision(testDecisionArgs{ctx, decisionManager, managerAuth, "/api/reports",
		[]string{"hasAnyRole('ADMIN','MANAGER')"}})

	// Test denyAll
	fmt.Println("  Resource: /api/secret")
	testDecision(testDecisionArgs{ctx, decisionManager, adminAuth, "/api/secret",
		[]string{"denyAll"}})

	// Test permitAll
	fmt.Println("  Resource: /public")
	testDecision(testDecisionArgs{ctx, decisionManager, userAuth, "/public",
		[]string{"permitAll"}})
}

// demoSecurityBuilder 演示安全构建器的配置。
func demoSecurityBuilder(userDetailsService *security.InMemoryUserDetailsService, encoder *security.NoOpPasswordEncoder, authManager security.AuthenticationManager) {
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
}

// demoUnanimousDecision 演示一票否决式授权决策管理器。
func demoUnanimousDecision(ctx context.Context, voter authorization.AccessDecisionVoter) {
	// ---- 7. Unanimous-based decision manager ----
	fmt.Println()
	fmt.Println("--- 6. Unanimous-Based Decision Manager ---")
	unanimousMgr := authorization.NewUnanimousBased(voter)

	adminAuth := security.NewAuthenticatedUsernamePasswordAuthenticationToken(
		"admin", []string{"ROLE_ADMIN", "ROLE_USER"})
	userAuth := security.NewAuthenticatedUsernamePasswordAuthenticationToken(
		"user", []string{"ROLE_USER"})
	managerAuth := security.NewAuthenticatedUsernamePasswordAuthenticationToken(
		"manager", []string{"ROLE_MANAGER", "ROLE_USER"})

	fmt.Println("  Admin (has ROLE_ADMIN):")
	testDecision(testDecisionArgs{ctx, unanimousMgr, adminAuth, "/resource",
		[]string{"hasRole('ADMIN')"}})

	fmt.Println("  User (has ROLE_USER only):")
	testDecision(testDecisionArgs{ctx, unanimousMgr, userAuth, "/resource",
		[]string{"hasRole('ADMIN')"}})

	fmt.Println("  Manager (has ROLE_MANAGER):")
	testDecision(testDecisionArgs{ctx, unanimousMgr, managerAuth, "/resource",
		[]string{"hasRole('ADMIN')"}})
}

// testDecisionArgs 授权决策测试参数。
type testDecisionArgs struct {
	ctx        context.Context
	mgr        authorization.AccessDecisionManager
	auth       security.Authentication
	resource   string
	attributes []string
}

// testDecision checks authorization and prints the result.
func testDecision(args testDecisionArgs) {
	err := args.mgr.Decide(args.ctx, args.auth, args.resource, args.attributes)
	status := "GRANTED"
	if err != nil {
		status = fmt.Sprintf("DENIED (%v)", err)
	}
	fmt.Printf("    %s -> %s %v: %s\n",
		args.auth.Principal(), args.resource, args.attributes, status)
}
