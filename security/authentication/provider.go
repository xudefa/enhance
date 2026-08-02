package authentication

import (
	"context"
	"errors"
	"fmt"

	"github.com/xudefa/enhance/log"
)

// daoAuthenticationProvider 基于数据源的认证提供者。
//
// 从 UserDetailsService 加载用户并验证密码。
// 认证流程：
//  1. 从认证令牌中提取用户名
//  2. 通过 UserDetailsService 加载用户详情
//  3. 验证密码是否匹配
//  4. 检查用户状态（是否启用、是否锁定等）
//  5. 认证成功后返回包含用户详情的认证信息
type daoAuthenticationProvider struct {
	userDetailsService UserDetailsService
	passwordEncoder    PasswordEncoder
	logger             log.Logger
}

// NewDaoAuthenticationProvider 创建基于数据源的认证提供者。
//
// 参数:
//   - userDetailsService: 用户详情服务，用于加载用户信息
//   - passwordEncoder: 密码编码器，用于验证密码
//   - logger: 日志记录器，可为 nil
//
// 返回:
//   - AuthenticationProvider: 认证提供者接口
func NewDaoAuthenticationProvider(
	userDetailsService UserDetailsService,
	passwordEncoder PasswordEncoder,
	logger log.Logger,
) AuthenticationProvider {
	return &daoAuthenticationProvider{
		userDetailsService: userDetailsService,
		passwordEncoder:    passwordEncoder,
		logger:             logger,
	}
}

// Authenticate 执行认证逻辑。
func (p *daoAuthenticationProvider) Authenticate(ctx context.Context, token AuthenticationToken) (Authentication, error) {
	var username string
	switch principal := token.Principal().(type) {
	case string:
		username = principal
	case UserDetails:
		username = principal.Username()
	default:
		return nil, fmt.Errorf("unsupported principal type: %T", token.Principal())
	}
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
	if creds, ok := token.Credentials().(string); ok {
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

	if !user.AccountNonExpired() {
		p.logger.Warn(ctx, "用户账户已过期", log.KeyValue{Key: "username", Value: username})
		return nil, errors.New("user account is expired")
	}

	if !user.CredentialsNonExpired() {
		p.logger.Warn(ctx, "用户凭证已过期", log.KeyValue{Key: "username", Value: username})
		return nil, errors.New("user credentials are expired")
	}

	p.logger.Info(ctx, "用户认证成功", log.KeyValue{Key: "username", Value: username})
	return NewAuthenticatedUsernamePasswordToken(user, nil, user.Authorities()), nil
}

// Supports 判断是否支持该令牌类型。
func (p *daoAuthenticationProvider) Supports(token AuthenticationToken) bool {
	_, ok := token.(*UsernamePasswordToken)
	return ok
}

// anonymousAuthenticationProvider 匿名认证提供者。
//
// 为未认证请求创建匿名令牌，通常作为最后一个提供者使用。
type anonymousAuthenticationProvider struct{}

// NewAnonymousAuthenticationProvider 创建匿名认证提供者。
//
// 匿名提供者为未认证请求自动创建匿名令牌，
// 通常作为认证链中的最后一个提供者使用。
func NewAnonymousAuthenticationProvider() AuthenticationProvider {
	return &anonymousAuthenticationProvider{}
}

// Authenticate 创建匿名认证令牌。
//
// 只有当未发起实际认证请求时，才创建匿名令牌。
func (p *anonymousAuthenticationProvider) Authenticate(_ context.Context, token AuthenticationToken) (Authentication, error) {
	if token == nil {
		return NewAuthenticatedUsernamePasswordToken("anonymousUser", nil, []string{"ROLE_ANONYMOUS"}), nil
	}
	// 如果令牌已认证或包含凭证，不创建匿名令牌
	if token.Authenticated() || token.Credentials() != nil {
		return nil, nil
	}
	return NewAuthenticatedUsernamePasswordToken("anonymousUser", nil, []string{"ROLE_ANONYMOUS"}), nil
}

// Supports 只支持 UsernamePasswordToken 类型。
func (p *anonymousAuthenticationProvider) Supports(token AuthenticationToken) bool {
	_, ok := token.(*UsernamePasswordToken)
	return ok
}
