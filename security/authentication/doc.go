// Package authentication 提供认证功能支持，用于 enhance 框架。
//
// 该模块提供独立的认证抽象层，将认证逻辑与安全授权分离。
// 参考 Spring Security Authentication 模块的设计理念。
//
// # 架构设计
//
//   - AuthenticationManager: 认证管理器接口，负责处理认证请求
//   - AuthenticationToken: 认证令牌接口，传递认证凭证
//   - Authentication: 认证信息接口，包含已认证主体信息
//   - AuthenticationProvider: 认证提供者接口，定义认证逻辑
//
// # 核心功能
//
//   - 令牌认证: 支持用户名/密码等令牌类型
//   - 提供者管理: 支持多个认证提供者的链式调用
//   - 认证结果: 返回包含权限列表的认证信息
//
// # 使用方式
//
// 创建认证管理器并认证用户：
//
//	provider := authentication.NewDaoAuthenticationProvider(userDetailsService, encoder, logger)
//	manager := authentication.NewProviderManager(provider)
//	token := authentication.NewUsernamePasswordToken("admin", "password")
//	result, err := manager.Authenticate(ctx, token)
//	if err != nil {
//	    // 认证失败
//	}
//	if result.Authenticated() {
//	    // 认证成功，可以获取权限列表
//	    authorities := result.Authorities()
//	}
package authentication

import (
	"context"
	"errors"
)

// 错误定义。
var (
	// ErrAuthenticationFailed 认证失败。
	ErrAuthenticationFailed = errors.New("authentication failed")
	// ErrBadCredentials 凭证无效。
	ErrBadCredentials = errors.New("bad credentials")
	// ErrUserNotFound 用户未找到。
	ErrUserNotFound = errors.New("user not found")
)

// AuthenticationManager 认证管理器接口。
//
// 负责处理认证请求，验证用户凭证并返回认证信息。
// 通常与多个 AuthenticationProvider 配合使用。
type AuthenticationManager interface {
	// Authenticate 执行认证。
	Authenticate(ctx context.Context, token AuthenticationToken) (Authentication, error)
}

// AuthenticationToken 认证令牌接口。
//
// 代表一个认证请求，包含主体标识和凭证信息。
// 实现应包含用户身份和待验证的凭证。
type AuthenticationToken interface {
	// Principal 获取主体标识，通常是用户名或用户对象。
	Principal() any
	// Credentials 获取凭证信息，如密码或令牌。
	Credentials() any
	// Authenticated 是否已认证。
	Authenticated() bool
}

// Authentication 认证信息接口。
//
// 代表一个已认证主体的完整信息，包括身份标识、凭证和权限。
// 实现应包含用户身份、角色列表和认证状态。
type Authentication interface {
	// Principal 获取认证主体的标识，通常是用户名或用户对象。
	Principal() any
	// Credentials 获取凭证信息，如密码或令牌。
	Credentials() any
	// Authorities 返回授权列表，如角色和权限。
	Authorities() []string
	// Authenticated 是否已认证。
	Authenticated() bool
}

// AuthenticationProvider 认证提供者接口。
//
// 定义认证逻辑，支持特定类型的认证令牌。
// 每个提供者负责一种认证方式的验证。
type AuthenticationProvider interface {
	// Supports 是否支持该令牌类型。
	Supports(token AuthenticationToken) bool
	// Authenticate 执行认证。
	Authenticate(ctx context.Context, token AuthenticationToken) (Authentication, error)
}

// UserDetails 用户详情接口。
//
// 包含用户的完整认证和授权信息。
// 实现应提供用户名、密码、权限列表和账户状态。
type UserDetails interface {
	// Username 返回用户名。
	Username() string
	// Password 返回密码。
	Password() string
	// Authorities 返回权限列表。
	Authorities() []string
	// Enabled 账户是否启用。
	Enabled() bool
	// AccountNonExpired 账户是否未过期。
	AccountNonExpired() bool
	// CredentialsNonExpired 凭证是否未过期。
	CredentialsNonExpired() bool
	// AccountNonLocked 账户是否未锁定。
	AccountNonLocked() bool
}

// UserDetailsService 用户详情服务接口。
//
// 根据用户名加载用户信息，用于认证。
// 实现可以从数据库、LDAP、内存等来源加载用户。
type UserDetailsService interface {
	// LoadUserByUsername 根据用户名加载用户详情。
	LoadUserByUsername(ctx context.Context, username string) (UserDetails, error)
}

// PasswordEncoder 密码编码器接口。
//
// 用于密码的编码和验证。
type PasswordEncoder interface {
	// Encode 编码密码。
	Encode(rawPassword string) string
	// Matches 验证密码是否匹配。
	Matches(rawPassword, encodedPassword string) bool
}
