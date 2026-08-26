package security

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security/authentication"
)

// mockUserDetailsService 模拟用户详情服务
type mockUserDetailsService struct {
	users map[string]authentication.UserDetails
}

func (m *mockUserDetailsService) LoadUserByUsername(ctx context.Context, username string) (authentication.UserDetails, error) {
	if user, ok := m.users[username]; ok {
		return user, nil
	}
	return nil, ErrUserNotFound
}

// mockPasswordEncoder 模拟密码编码器
type mockPasswordEncoder struct {
	matches map[string]string // encoded -> raw
}

func (m *mockPasswordEncoder) Encode(rawPassword string) string {
	return "encoded_" + rawPassword
}

func (m *mockPasswordEncoder) Matches(rawPassword, encodedPassword string) bool {
	return m.matches[encodedPassword] == rawPassword
}

// mockUserDetails 模拟用户详情
type mockUserDetails struct {
	username string
	password string
}

func (m *mockUserDetails) Username() string            { return m.username }
func (m *mockUserDetails) Password() string            { return m.password }
func (m *mockUserDetails) Authorities() []string       { return []string{"ROLE_USER"} }
func (m *mockUserDetails) Enabled() bool               { return true }
func (m *mockUserDetails) AccountNonExpired() bool     { return true }
func (m *mockUserDetails) AccountNonLocked() bool      { return true }
func (m *mockUserDetails) CredentialsNonExpired() bool { return true }

// mockLogger 模拟日志器
type mockLogger struct{}

func (m *mockLogger) Debug(ctx context.Context, msg string, kvs ...log.KeyValue) {}
func (m *mockLogger) Info(ctx context.Context, msg string, kvs ...log.KeyValue)  {}
func (m *mockLogger) Warn(ctx context.Context, msg string, kvs ...log.KeyValue)  {}
func (m *mockLogger) Error(ctx context.Context, msg string, kvs ...log.KeyValue) {}

func TestProviderManager_AddProvider(t *testing.T) {
	t.Parallel()

	pm := NewProviderManager()
	provider := &mockAuthProvider{}
	pm.AddProvider(provider)

	// 验证提供者已添加
	if len(pm.providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(pm.providers))
	}
}

func TestDaoAuthenticationProvider_Authenticate_Success(t *testing.T) {
	t.Parallel()

	userDetails := &mockUserDetails{
		username: "testuser",
		password: "password123", // 存储的密码
	}

	userDetailsService := &mockUserDetailsService{
		users: map[string]authentication.UserDetails{
			"testuser": userDetails,
		},
	}

	// 密码编码器：直接比较（模拟明文存储）
	passwordEncoder := &mockPasswordEncoder{
		matches: map[string]string{
			"password123": "password123", // encoded -> raw
		},
	}

	logger := &mockLogger{}
	provider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, logger)

	token := NewUsernamePasswordAuthenticationToken("testuser", "password123")
	result, err := provider.Authenticate(context.Background(), token)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil authentication result")
	}
	// Principal 返回的是 UserDetails 对象
	if ud, ok := result.Principal().(authentication.UserDetails); ok {
		if ud.Username() != "testuser" {
			t.Errorf("expected principal username 'testuser', got %s", ud.Username())
		}
	} else {
		t.Errorf("expected principal to be UserDetails, got %T", result.Principal())
	}
}

func TestDaoAuthenticationProvider_Authenticate_UserNotFound(t *testing.T) {
	t.Parallel()

	userDetailsService := &mockUserDetailsService{
		users: map[string]authentication.UserDetails{},
	}

	passwordEncoder := &mockPasswordEncoder{}
	logger := &mockLogger{}
	provider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, logger)

	token := NewUsernamePasswordAuthenticationToken("nonexistent", "password")
	_, err := provider.Authenticate(context.Background(), token)

	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	if err.Error() != "bad credentials" {
		t.Errorf("expected 'bad credentials' error, got %v", err)
	}
}

func TestDaoAuthenticationProvider_Authenticate_EmptyUsername(t *testing.T) {
	t.Parallel()

	userDetailsService := &mockUserDetailsService{}
	passwordEncoder := &mockPasswordEncoder{}
	logger := &mockLogger{}
	provider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, logger)

	token := NewUsernamePasswordAuthenticationToken("", "password")
	_, err := provider.Authenticate(context.Background(), token)

	if err == nil {
		t.Fatal("expected error for empty username")
	}
	if err.Error() != "bad credentials" {
		t.Errorf("expected 'bad credentials' error, got %v", err)
	}
}

func TestAuthentication_Name(t *testing.T) {
	t.Parallel()

	token := NewUsernamePasswordAuthenticationToken("testuser", "password")
	name := token.Name()
	if name != "testuser" {
		t.Errorf("expected name 'testuser', got %s", name)
	}
}

func TestAuthentication_Name_UserDetails(t *testing.T) {
	t.Parallel()

	userDetails := &mockUserDetails{username: "uduser", password: "pass"}
	token := NewAuthenticatedUsernamePasswordAuthenticationToken(userDetails, []string{"ROLE_USER"})
	name := token.Name()
	if name != "uduser" {
		t.Errorf("expected name 'uduser', got %s", name)
	}
}

// mockAuthProvider 模拟认证提供者
type mockAuthProvider struct{}

func (m *mockAuthProvider) Authenticate(ctx context.Context, token authentication.AuthenticationToken) (authentication.Authentication, error) {
	// 创建一个已认证的 token 作为 Authentication 返回
	tok := NewAuthenticatedUsernamePasswordAuthenticationToken(token.Principal(), []string{"ROLE_USER"})
	return tok, nil
}

func (m *mockAuthProvider) Supports(token authentication.AuthenticationToken) bool {
	return true
}
