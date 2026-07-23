package authentication

import (
	"context"
	"sync"
)

// InMemoryUserDetails 内存用户详情实现。
//
// 存储用户的完整认证和授权信息。
type InMemoryUserDetails struct {
	username              string
	password              string
	authorities           []string
	enabled               bool
	accountNonExpired     bool
	credentialsNonExpired bool
	accountNonLocked      bool
}

// NewInMemoryUserDetails 创建内存用户详情。
//
// 默认值: enabled=true, accountNonLocked=true。
func NewInMemoryUserDetails(username, password string, authorities []string) *InMemoryUserDetails {
	return &InMemoryUserDetails{
		username:              username,
		password:              password,
		authorities:           authorities,
		enabled:               true,
		accountNonExpired:     true,
		credentialsNonExpired: true,
		accountNonLocked:      true,
	}
}

// Username 返回用户名。
func (u *InMemoryUserDetails) Username() string {
	return u.username
}

// Password 返回密码。
func (u *InMemoryUserDetails) Password() string {
	return u.password
}

// Authorities 返回权限列表。
func (u *InMemoryUserDetails) Authorities() []string {
	return u.authorities
}

// Enabled 返回是否启用。
func (u *InMemoryUserDetails) Enabled() bool {
	return u.enabled
}

// AccountNonLocked 返回账户是否未锁定。
func (u *InMemoryUserDetails) AccountNonLocked() bool {
	return u.accountNonLocked
}

// AccountNonExpired 返回账户是否未过期。
func (u *InMemoryUserDetails) AccountNonExpired() bool {
	return u.accountNonExpired
}

// CredentialsNonExpired 返回凭证是否未过期。
func (u *InMemoryUserDetails) CredentialsNonExpired() bool {
	return u.credentialsNonExpired
}

// SetEnabled 设置启用状态。
func (u *InMemoryUserDetails) SetEnabled(enabled bool) {
	u.enabled = enabled
}

// SetAccountNonLocked 设置锁定状态。
func (u *InMemoryUserDetails) SetAccountNonLocked(locked bool) {
	u.accountNonLocked = locked
}

// InMemoryUserDetailsService 内存用户详情服务。
//
// 使用内存 map 存储用户信息，线程安全。
// 适用场景：开发和测试环境，生产环境应使用数据库实现。
type InMemoryUserDetailsService struct {
	users map[string]UserDetails
	mu    sync.RWMutex
}

// NewInMemoryUserDetailsService 创建内存用户详情服务。
func NewInMemoryUserDetailsService() *InMemoryUserDetailsService {
	return &InMemoryUserDetailsService{
		users: make(map[string]UserDetails),
	}
}

// CreateUser 创建新用户。
func (s *InMemoryUserDetailsService) CreateUser(username, password string, authorities []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[username] = NewInMemoryUserDetails(username, password, authorities)
}

// LoadUserByUsername 根据用户名加载用户。
//
// 如果用户不存在返回 ErrUserNotFound。
func (s *InMemoryUserDetailsService) LoadUserByUsername(_ context.Context, username string) (UserDetails, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// DeleteUser 删除用户。
func (s *InMemoryUserDetailsService) DeleteUser(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, username)
}

// UserCount 返回用户总数。
func (s *InMemoryUserDetailsService) UserCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users)
}

// NoOpPasswordEncoder 不进行编码的密码编码器。
//
// 警告：生产环境绝对不要使用此编码器。
type NoOpPasswordEncoder struct{}

// NewNoOpPasswordEncoder 创建 NoOp 密码编码器。
func NewNoOpPasswordEncoder() *NoOpPasswordEncoder {
	return &NoOpPasswordEncoder{}
}

// Encode 直接返回原始密码。
func (e *NoOpPasswordEncoder) Encode(rawPassword string) string {
	return rawPassword
}

// Matches 直接比较原始密码和编码后密码。
func (e *NoOpPasswordEncoder) Matches(rawPassword, encodedPassword string) bool {
	return rawPassword == encodedPassword
}
