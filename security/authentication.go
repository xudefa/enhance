package security

import (
	"context"
	"errors"

	"github.com/xudefa/enhance/log"
)

// UsernamePasswordAuthenticationToken 用户名密码认证令牌
// 用于在认证过程中传递和存储用户凭证
// 认证前：包含用户名和密码（authenticated=false）
// 认证后：包含用户详情和权限列表（authenticated=true）
type UsernamePasswordAuthenticationToken struct {
	principal     any
	credentials   any
	authorities   []string
	authenticated bool
}

// NewUsernamePasswordAuthenticationToken 创建未认证的用户名密码令牌
// 用于认证请求的初始阶段
func NewUsernamePasswordAuthenticationToken(principal, credentials any) *UsernamePasswordAuthenticationToken {
	return &UsernamePasswordAuthenticationToken{
		principal:     principal,
		credentials:   credentials,
		authorities:   []string{},
		authenticated: false,
	}
}

// NewAuthenticatedUsernamePasswordAuthenticationToken 创建已认证的用户名密码令牌
// 用于认证成功后的令牌
func NewAuthenticatedUsernamePasswordAuthenticationToken(principal any, authorities []string) *UsernamePasswordAuthenticationToken {
	return &UsernamePasswordAuthenticationToken{
		principal:     principal,
		credentials:   nil,
		authorities:   authorities,
		authenticated: true,
	}
}

func (t *UsernamePasswordAuthenticationToken) Principal() any {
	return t.principal
}

func (t *UsernamePasswordAuthenticationToken) Credentials() any {
	return t.credentials
}

func (t *UsernamePasswordAuthenticationToken) Authorities() []string {
	return t.authorities
}

func (t *UsernamePasswordAuthenticationToken) Authenticated() bool {
	return t.authenticated
}

// Name 返回认证主体名称
// 支持字符串和UserDetails两种principal类型
func (t *UsernamePasswordAuthenticationToken) Name() string {
	if name, ok := t.principal.(string); ok {
		return name
	}
	if userDetails, ok := t.principal.(UserDetails); ok {
		return userDetails.Username()
	}
	return ""
}

// SetAuthenticated 设置认证状态
func (t *UsernamePasswordAuthenticationToken) SetAuthenticated(authenticated bool) {
	t.authenticated = authenticated
}

// SetAuthorities 设置授权列表
func (t *UsernamePasswordAuthenticationToken) SetAuthorities(authorities []string) {
	t.authorities = authorities
}

// ProviderManager 认证提供者管理器
// 管理多个认证提供者，按顺序尝试认证
// 执行流程：遍历所有支持的提供者，返回第一个成功认证的结果
type ProviderManager struct {
	providers []AuthenticationProvider
}

// NewProviderManager 创建认证提供者管理器
func NewProviderManager(providers ...AuthenticationProvider) *ProviderManager {
	return &ProviderManager{
		providers: providers,
	}
}

// Authenticate 尝试通过配置的提供者进行认证
// 遍历所有提供者，返回第一个成功认证的结果
func (m *ProviderManager) Authenticate(ctx context.Context, authentication Authentication) (Authentication, error) {
	var lastErr error

	for _, provider := range m.providers {
		if provider.Supports(authentication) {
			result, err := provider.Authenticate(ctx, authentication)
			if err != nil {
				lastErr = err
				continue
			}
			if result != nil {
				return result, nil
			}
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, ErrAuthenticationFailed
}

// AddProvider 添加认证提供者
func (m *ProviderManager) AddProvider(provider AuthenticationProvider) {
	m.providers = append(m.providers, provider)
}

// DaoAuthenticationProvider 基于DAO的认证提供者
// 从UserDetailsService加载用户并验证密码
// 认证流程：
// 1. 从认证令牌中提取用户名
// 2. 通过UserDetailsService加载用户详情
// 3. 验证密码是否匹配
// 4. 检查用户状态（是否启用、是否锁定等）
// 5. 认证成功后返回包含用户详情的认证令牌
type DaoAuthenticationProvider struct {
	userDetailsService UserDetailsService
	passwordEncoder    PasswordEncoder
	logger             log.Logger
}

// NewDaoAuthenticationProvider 创建DAO认证提供者
func NewDaoAuthenticationProvider(userDetailsService UserDetailsService, passwordEncoder PasswordEncoder, logger log.Logger) *DaoAuthenticationProvider {
	return &DaoAuthenticationProvider{
		userDetailsService: userDetailsService,
		passwordEncoder:    passwordEncoder,
		logger:             logger,
	}
}

// Authenticate 执行认证逻辑
// 1. 根据用户名加载用户信息
// 2. 验证密码
// 3. 检查用户状态
func (p *DaoAuthenticationProvider) Authenticate(ctx context.Context, authentication Authentication) (Authentication, error) {
	username := authentication.Name()
	if username == "" {
		p.logger.Debug(ctx, "认证失败：用户名为空")
		return nil, ErrBadCredentials
	}

	p.logger.Debug(ctx, "尝试加载用户信息", log.KeyValue{Key: "username", Value: username})
	user, err := p.userDetailsService.LoadUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			p.logger.Debug(ctx, "用户不存在", log.KeyValue{Key: "username", Value: username})
			return nil, ErrBadCredentials
		}
		p.logger.Error(ctx, "加载用户信息失败",
			log.KeyValue{Key: "username", Value: username},
			log.KeyValue{Key: "error", Value: err.Error()},
		)
		return nil, err
	}

	p.logger.Debug(ctx, "验证用户密码", log.KeyValue{Key: "username", Value: username})
	presentedPassword := ""
	if creds, ok := authentication.Credentials().(string); ok {
		presentedPassword = creds
	}

	if !p.passwordEncoder.Matches(presentedPassword, user.Password()) {
		p.logger.Warn(ctx, "密码验证失败", log.KeyValue{Key: "username", Value: username})
		return nil, ErrBadCredentials
	}

	if !user.Enabled() {
		p.logger.Warn(ctx, "用户已禁用", log.KeyValue{Key: "username", Value: username})
		return nil, errors.New("user is disabled")
	}

	if !user.AccountNonLocked() {
		p.logger.Warn(ctx, "用户账户已锁定", log.KeyValue{Key: "username", Value: username})
		return nil, errors.New("user account is locked")
	}

	p.logger.Info(ctx, "用户认证成功", log.KeyValue{Key: "username", Value: username})
	return NewAuthenticatedUsernamePasswordAuthenticationToken(user, user.Authorities()), nil
}

// Supports 判断是否支持该认证方式
func (p *DaoAuthenticationProvider) Supports(authentication Authentication) bool {
	_, ok := authentication.(*UsernamePasswordAuthenticationToken)
	return ok
}

// AnonymousAuthenticationProvider 匿名认证提供者
// 为匿名用户创建认证令牌，当其他认证提供者都无法处理时使用
// 使用场景：用户未登录但需要访问某些公开资源时
type AnonymousAuthenticationProvider struct{}

// NewAnonymousAuthenticationProvider 创建匿名认证提供者
func NewAnonymousAuthenticationProvider() *AnonymousAuthenticationProvider {
	return &AnonymousAuthenticationProvider{}
}

// Authenticate 为匿名用户创建认证令牌
// 只有当传入的authentication为nil或未认证时，才创建匿名令牌
func (p *AnonymousAuthenticationProvider) Authenticate(ctx context.Context, authentication Authentication) (Authentication, error) {
	if authentication == nil || !authentication.Authenticated() {
		return NewAuthenticatedUsernamePasswordAuthenticationToken("anonymousUser", []string{"ROLE_ANONYMOUS"}), nil
	}
	return nil, nil
}

// Supports 只支持UsernamePasswordAuthenticationToken类型
func (p *AnonymousAuthenticationProvider) Supports(authentication Authentication) bool {
	_, ok := authentication.(*UsernamePasswordAuthenticationToken)
	return ok
}
