package security

import (
	"context"
	"testing"
)

// benchAuthManager 测试用认证管理器
type benchAuthManager struct{}

func (m *benchAuthManager) Authenticate(ctx context.Context, auth Authentication) (Authentication, error) {
	return auth, nil
}

// benchUserDetails 测试用用户详情
type benchUserDetails struct {
	username string
	password string
	roles    []string
}

func (u *benchUserDetails) Username() string            { return u.username }
func (u *benchUserDetails) Password() string            { return u.password }
func (u *benchUserDetails) Authorities() []string       { return u.roles }
func (u *benchUserDetails) Enabled() bool               { return true }
func (u *benchUserDetails) AccountNonExpired() bool     { return true }
func (u *benchUserDetails) AccountNonLocked() bool      { return true }
func (u *benchUserDetails) CredentialsNonExpired() bool { return true }

// benchUserDetailsService 测试用用户详情服务
type benchUserDetailsService struct{}

func (s *benchUserDetailsService) LoadUserByUsername(ctx context.Context, username string) (UserDetails, error) {
	return &benchUserDetails{
		username: username,
		password: "password",
		roles:    []string{"USER", "ADMIN"},
	}, nil
}

// BenchmarkSecurityFilterChain_Build 测试过滤器链构建性能
func BenchmarkSecurityFilterChain_Build(b *testing.B) {
	authManager := &benchAuthManager{}
	userDetailsService := &benchUserDetailsService{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewSecurityBuilder().
			AuthenticationManager(authManager).
			UserDetailsService(userDetailsService).
			Build()
	}
}

// BenchmarkSecurityFilterChain_Configure 测试配置应用性能
func BenchmarkSecurityFilterChain_Configure(b *testing.B) {
	config := NewSecurityBuilder().
		AuthenticationManager(&benchAuthManager{}).
		UserDetailsService(&benchUserDetailsService{}).
		Build()

	http := NewHttpSecurity()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Configure(http)
	}
}

// BenchmarkPasswordEncoder_NoOp 测试无操作密码编码器性能
func BenchmarkPasswordEncoder_NoOp(b *testing.B) {
	encoder := NewNoOpPasswordEncoder()
	password := "test-password-123"

	b.Run("Encode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = encoder.Encode(password)
		}
	})

	b.Run("Matches", func(b *testing.B) {
		hashed := encoder.Encode(password)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = encoder.Matches(password, hashed)
		}
	})
}

// BenchmarkSecurityBuilder_DifferentConfigs 测试不同配置的性能
func BenchmarkSecurityBuilder_DifferentConfigs(b *testing.B) {
	authManager := &benchAuthManager{}
	userDetailsService := &benchUserDetailsService{}

	b.Run("Simple", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewSecurityBuilder().
				AuthenticationManager(authManager).
				UserDetailsService(userDetailsService).
				Build()
		}
	})

	b.Run("WithCSRF", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewSecurityBuilder().
				AuthenticationManager(authManager).
				UserDetailsService(userDetailsService).
				EnableCsrf().
				Build()
		}
	})

	b.Run("WithFormLogin", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewSecurityBuilder().
				AuthenticationManager(authManager).
				UserDetailsService(userDetailsService).
				EnableFormLogin("/login", "/home").
				Build()
		}
	})

	b.Run("WithHttpBasic", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewSecurityBuilder().
				AuthenticationManager(authManager).
				UserDetailsService(userDetailsService).
				EnableHttpBasic().
				Build()
		}
	})

	b.Run("WithLogout", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewSecurityBuilder().
				AuthenticationManager(authManager).
				UserDetailsService(userDetailsService).
				EnableLogout("/logout").
				Build()
		}
	})

	b.Run("Full-Config", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewSecurityBuilder().
				AuthenticationManager(authManager).
				UserDetailsService(userDetailsService).
				EnableCsrf().
				EnableFormLogin("/login", "/home").
				EnableHttpBasic().
				EnableLogout("/logout").
				EnableAnonymous().
				Build()
		}
	})
}
