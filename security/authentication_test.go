package security

import (
	"context"
	"errors"
	"strings"
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
	authenticated, err := provider.Authenticate(context.Background(), token)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if authenticated == nil {
		t.Fatal("expected non-nil authentication result")
	}
	// Principal 返回的是 UserDetails 对象
	if ud, ok := authenticated.Principal().(authentication.UserDetails); ok {
		if ud.Username() != "testuser" {
			t.Errorf("expected principal username 'testuser', got %s", ud.Username())
		}
	} else {
		t.Errorf("expected principal to be UserDetails, got %T", authenticated.Principal())
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
	if !errors.Is(err, ErrBadCredentials) {
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
	if !errors.Is(err, ErrBadCredentials) {
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

// ==================== User Details Tests ====================

func TestNewRole(t *testing.T) {
	t.Parallel()

	role := NewRole("ROLE_ADMIN")
	if role == nil {
		t.Fatal("expected non-nil role")
	}
	if role.Authority() != "ROLE_ADMIN" {
		t.Errorf("expected authority 'ROLE_ADMIN', got '%s'", role.Authority())
	}
}

func TestRole_Authority(t *testing.T) {
	t.Parallel()

	role := NewRole("ROLE_USER")
	if role.Authority() != "ROLE_USER" {
		t.Errorf("expected authority 'ROLE_USER', got '%s'", role.Authority())
	}
}

func TestNewAuthority(t *testing.T) {
	t.Parallel()

	auth := NewAuthority("ROLE_EDITOR")
	if auth == nil {
		t.Fatal("expected non-nil authority")
	}
	if auth.Authority() != "ROLE_EDITOR" {
		t.Errorf("expected authority 'ROLE_EDITOR', got '%s'", auth.Authority())
	}
}

func TestInMemoryUserDetails_AllFields(t *testing.T) {
	t.Parallel()

	user := NewInMemoryUserDetails(
		"testuser",
		"password123",
		[]string{"ROLE_USER", "ROLE_ADMIN"},
	)

	if user.Username() != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", user.Username())
	}
	if user.Password() != "password123" {
		t.Errorf("expected password 'password123', got '%s'", user.Password())
	}

	authorities := user.Authorities()
	if len(authorities) != 2 {
		t.Errorf("expected 2 authorities, got %d", len(authorities))
	}

	if !user.Enabled() {
		t.Error("expected user to be enabled")
	}
	if !user.AccountNonExpired() {
		t.Error("expected account to be non-expired")
	}
	if !user.AccountNonLocked() {
		t.Error("expected account to be non-locked")
	}
	if !user.CredentialsNonExpired() {
		t.Error("expected credentials to be non-expired")
	}
}

func TestInMemoryUserDetailsService_CreateAndLoad(t *testing.T) {
	t.Parallel()

	service := NewInMemoryUserDetailsService()

	service.CreateUser("user1", "pass1", []string{"ROLE_USER"})

	loaded, err := service.LoadUserByUsername(context.Background(), "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.Username() != "user1" {
		t.Errorf("expected username 'user1', got '%s'", loaded.Username())
	}
}

func TestInMemoryUserDetailsService_DeleteUser(t *testing.T) {
	t.Parallel()

	service := NewInMemoryUserDetailsService()

	service.CreateUser("user1", "pass1", []string{"ROLE_USER"})

	if service.UserCount() != 1 {
		t.Errorf("expected 1 user, got %d", service.UserCount())
	}

	service.DeleteUser("user1")
	if service.UserCount() != 0 {
		t.Errorf("expected 0 users after delete, got %d", service.UserCount())
	}
}

func TestInMemoryUserDetailsService_LoadNonExistentUser(t *testing.T) {
	t.Parallel()

	service := NewInMemoryUserDetailsService()

	_, err := service.LoadUserByUsername(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

// ==================== Password Encoder Tests ====================

func TestSha256PasswordEncoder(t *testing.T) {
	t.Parallel()
	encoder := NewSha256PasswordEncoder()

	password := "mySecurePassword123!"
	encoded := encoder.Encode(password)

	if encoded == password {
		t.Error("encoded password should be different from raw password")
	}

	if !encoder.Matches(password, encoded) {
		t.Error("should match correct password")
	}

	if encoder.Matches("wrongPassword", encoded) {
		t.Error("should not match wrong password")
	}

	encoded2 := encoder.Encode(password)
	if encoded != encoded2 {
		t.Error("SHA256 should produce same hash for same password")
	}
}

func TestDelegatingPasswordEncoder_ExtractId(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input   string
		wantId  string
		wantPwd string
		wantErr bool
	}{
		{"{bcrypt}$2a$12$abc123", "bcrypt", "$2a$12$abc123", false},
		{"{sha256}abc123", "sha256", "abc123", false},
		{"{noop}password", "noop", "password", false},
		{"invalid", "", "", true},
		{"{noclose", "", "", true},
		{"}id{password", "", "", true},
		{"{id}", "id", "", false},
		{"{}", "", "", true},
	}

	encoder := &DelegatingPasswordEncoder{}

	for _, tt := range tests {
		id, pwd, err := encoder.extractId(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("extractId(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if id != tt.wantId || pwd != tt.wantPwd {
			t.Errorf("extractId(%q) = (%q, %q), want (%q, %q)",
				tt.input, id, pwd, tt.wantId, tt.wantPwd)
		}
	}
}

func TestDelegatingPasswordEncoder_FullCycle(t *testing.T) {
	t.Parallel()
	sha256Encoder := NewSha256PasswordEncoder()
	noopEncoder := NewNoOpPasswordEncoder()

	encoders := map[string]PasswordEncoder{
		"sha256": sha256Encoder,
		"noop":   noopEncoder,
	}

	encoder, err := NewDelegatingPasswordEncoder("sha256", encoders)
	if err != nil {
		t.Fatalf("failed to create DelegatingPasswordEncoder: %v", err)
	}

	password := "myPassword"
	encoded := encoder.Encode(password)

	if !strings.HasPrefix(encoded, "{sha256}") {
		t.Errorf("encoded password should start with {sha256}, got: %s", encoded[:10])
	}

	if !encoder.Matches(password, encoded) {
		t.Error("should match correct password")
	}

	if encoder.Matches("wrongPassword", encoded) {
		t.Error("should not match wrong password")
	}
}

func TestDelegatingPasswordEncoder_UnknownId(t *testing.T) {
	t.Parallel()
	sha256Encoder := NewSha256PasswordEncoder()

	encoders := map[string]PasswordEncoder{
		"sha256": sha256Encoder,
	}

	encoder, err := NewDelegatingPasswordEncoder("sha256", encoders)
	if err != nil {
		t.Fatalf("failed to create DelegatingPasswordEncoder: %v", err)
	}

	encoded := "{unknown}somehash"

	if encoder.Matches("password", encoded) {
		t.Error("should not match unknown encoder id")
	}
}

func TestDelegatingPasswordEncoder_InvalidFormat(t *testing.T) {
	t.Parallel()
	sha256Encoder := NewSha256PasswordEncoder()

	encoders := map[string]PasswordEncoder{
		"sha256": sha256Encoder,
	}

	encoder, err := NewDelegatingPasswordEncoder("sha256", encoders)
	if err != nil {
		t.Fatalf("failed to create DelegatingPasswordEncoder: %v", err)
	}

	invalidFormats := []string{
		"",
		"no-braces",
		"{missing-close",
		"}missing-open{",
	}

	for _, encoded := range invalidFormats {
		if encoder.Matches("password", encoded) {
			t.Errorf("should not match invalid format: %s", encoded)
		}
	}
}

func TestStandardPasswordEncoder(t *testing.T) {
	t.Parallel()

	encoder := NewStandardPasswordEncoder("my-secret-key")

	password := "testPassword123"
	encoded := encoder.Encode(password)

	if encoded == password {
		t.Error("encoded password should differ from raw password")
	}

	if !encoder.Matches(password, encoded) {
		t.Error("should match correct password")
	}

	if encoder.Matches("wrongPassword", encoded) {
		t.Error("should not match wrong password")
	}

	encoder2 := NewStandardPasswordEncoder("my-secret-key")
	if encoder2.Encode(password) != encoded {
		t.Error("same secret should produce same hash")
	}

	encoder3 := NewStandardPasswordEncoder("different-secret")
	if encoder3.Encode(password) == encoded {
		t.Error("different secret should produce different hash")
	}
}

func TestMustNewDelegatingPasswordEncoder(t *testing.T) {
	t.Parallel()

	t.Run("valid creation", func(t *testing.T) {
		encoder := MustNewDelegatingPasswordEncoder("sha256", map[string]PasswordEncoder{
			"sha256": NewSha256PasswordEncoder(),
		})
		if encoder == nil {
			t.Fatal("expected non-nil encoder")
		}
	})

	t.Run("panic on invalid id", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for invalid id")
			}
		}()
		MustNewDelegatingPasswordEncoder("unknown", map[string]PasswordEncoder{
			"sha256": NewSha256PasswordEncoder(),
		})
	})
}

func TestDelegatingPasswordEncoder_EncodeWithMissingEncoder(t *testing.T) {
	t.Parallel()

	encoder := &DelegatingPasswordEncoder{
		idForEncode:      "missing",
		passwordEncoders: map[string]PasswordEncoder{},
	}

	encodedPassword := encoder.Encode("password")
	if encodedPassword != "" {
		t.Errorf("expected empty string, got '%s'", encodedPassword)
	}
}
