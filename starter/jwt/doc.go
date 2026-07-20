// Package jwt 提供基于 JSON Web Token (JWT) 的无状态认证机制。
//
// 本模块是 enhance 框架的第三方认证集成模块，基于 github.com/golang-jwt/jwt/v5 实现，
// 遵循 Spring Security JWT 的设计理念，与 security 模块无缝集成。
//
// # 模块独立性
//
// 本模块采用独立模块设计（拥有独立的 go.mod），依赖隔离确保：
//   - 用户只使用 enhance 核心模块时，不会引入 golang-jwt 依赖
//   - 用户显式引入本模块时，才会下载 golang-jwt 及其间接依赖
//   - 避免依赖污染，保持用户项目的依赖树清晰
//
// # 架构设计
//
// 核心组件：
//   - TokenProvider: Token 提供者接口，负责 JWT 的生成、解析和验证
//   - JwtAuthenticationFilter: JWT 认证过滤器，实现 security.SecurityFilter 接口
//   - TokenClaims: Token 声明结构体，包含用户身份和权限信息
//
// 自动配置：
//   - JwtAutoConfiguration: 自动配置类，根据配置文件自动启用 JWT 认证
//   - 当 security.enabled=true 且 security.jwt.enabled=true 时自动生效
//
// # 核心功能
//
//   - Token 生成: 根据用户名和权限列表生成 JWT Token（HMAC-SHA256 签名）
//   - Token 验证: 验证 JWT 的签名、过期时间和有效性
//   - Token 解析: 解析 Token 并提取用户声明信息
//   - 自动集成: JWT 过滤器自动插入到安全过滤器链中
//
// # 快速开始
//
// 1. 在配置文件中启用 JWT：
//
//	{
//	  "security": {
//	    "enabled": true,
//	    "jwt": {
//	      "enabled": true,
//	      "secret-key": "your-secret-key-here",
//	      "exclude-paths": "/login,/register,/public/**"
//	    }
//	  }
//	}
//
// 2. 在控制器中注入 TokenProvider 生成 Token：
//
//	type AuthController struct {
//	    TokenProvider *jwt.DefaultTokenProvider
//	}
//
//	func (c *AuthController) Login(ctx mvc.Context) {
//	    token, err := c.TokenProvider.GenerateToken(ctx.Context(), username, []string{"ROLE_USER"})
//	    // ...
//	}
//
// 3. 客户端在请求头中携带 Token：
//
//	Authorization: Bearer <token>
//
// # 配置说明
//
//   - security.jwt.enabled: 是否启用 JWT（默认 false）
//   - security.jwt.secret-key: JWT 签名密钥（生产环境必须配置）
//   - security.jwt.exclude-paths: 不需要 JWT 认证的路径（逗号分隔，支持通配符）
//
// # 与 security 模块集成
//
// JWT 模块自动与 security 模块集成：
//   - JwtAuthenticationFilter 实现 security.SecurityFilter 接口
//   - 自动注册到安全过滤器链中（在 BasicAuthenticationFilter 之前执行）
//   - 认证成功后，用户信息自动设置到 SecurityContext 中
//   - 与 Casbin 授权模块兼容，可组合使用
//
// # 执行顺序
//
// 安全过滤器链中的执行顺序：
//  1. SecurityContextHolderFilter（保存并清除认证上下文）
//  2. JwtAuthenticationFilter（JWT Token 认证）
//  3. BasicAuthenticationFilter（Basic 认证，跳过 Bearer Token）
//  4. AnonymousAuthenticationFilter（匿名认证）
//  5. FilterSecurityInterceptor（访问决策）
//
// # 依赖说明
//
// 本模块依赖：
//   - github.com/golang-jwt/jwt/v5: JWT 标准实现
//   - github.com/xudefa/enhance/security: 安全框架核心
//
// 用户项目引入本模块后，会自动引入上述依赖。
package jwt

import (
	"context"
	"time"
)

// ==================== 配置键常量 ====================

const (
	// HTTP Header 常量
	HeaderAuthorization = "Authorization"
	HeaderBearerPrefix  = "Bearer "

	// JWT 配置
	JWTEnabled                = "security.jwt.enabled"
	JWTSecretKey              = "security.jwt.secret-key"
	JWTIssuer                 = "security.jwt.issuer"
	JWTExpiresDuration        = "security.jwt.expires-duration"
	JWTRefreshExpiresDuration = "security.jwt.refresh-expires-duration"
	JWTExcludePaths           = "security.jwt.exclude-paths"
	JWTSigningMethod          = "security.jwt.signing-method"

	// Security 配置
	SecurityEnabled = "security.enabled"

	// 日志字段常量
	LogFieldError   = "error"
	LogFieldSecret  = "secret-key"
	LogFieldExclude = "exclude-paths"
)

// ==================== 默认值常量 ====================

const (
	// JWT 默认值
	DefaultJWTSecretKey              = "enhanceJwtSecret"
	DefaultJWTIssuer                 = "enhance"
	DefaultJWTExpiresDuration        = 600
	DefaultJWTRefreshExpiresDuration = 3600
	DefaultJWTExcludePaths           = "/login,/register,/health,/actuator/health"
	DefaultJWTSigningMethod          = "HS256"

	// 条件值常量
	ConditionTrue = "true"
)

// TokenProvider Token 提供者接口。
//
// 负责 JWT Token 的生成、解析和验证。
// 实现类：DefaultTokenProvider
type TokenProvider interface {
	// GenerateToken 生成 JWT Token。
	// 参数：username - 用户名，authorities - 权限列表
	// 返回：JWT Token 字符串，错误
	GenerateToken(ctx context.Context, username string, authorities []string) (string, error)
	// ParseToken 解析 JWT Token。
	// 参数：token - JWT Token 字符串
	// 返回：TokenClaims 指针，错误
	ParseToken(ctx context.Context, token string) (*TokenClaims, error)
	// ValidateToken 验证 JWT Token 是否有效。
	// 参数：token - JWT Token 字符串
	// 返回：错误（验证失败时返回具体错误）
	ValidateToken(ctx context.Context, token string) error
	// RefreshToken 刷新 JWT Token。
	// 参数：token - 旧的 JWT Token
	// 返回：新的 JWT Token 字符串，错误
	RefreshToken(ctx context.Context, token string) (string, error)
}

// TokenClaims Token 声明结构体。
//
// 包含 JWT Token 中的标准声明和自定义声明。
// 对应 JWT 标准中的 payload 部分。
type TokenClaims struct {
	Subject     string    `json:"sub"`           // 主题（用户名）
	Issuer      string    `json:"iss,omitempty"` // 签发者
	Audience    string    `json:"aud,omitempty"` // 受众
	Expiration  time.Time `json:"exp"`           // 过期时间
	IssuedAt    time.Time `json:"iat"`           // 签发时间
	Authorities []string  `json:"authorities"`   // 权限列表
}

// TokenConfig Token 配置。
//
// 用于初始化 TokenProvider 的配置参数。
type TokenConfig struct {
	SecretKey  string        `json:"secret-key"` // 签名密钥（HMAC-SHA256）
	Expiration time.Duration `json:"expiration"` // Token 有效期
	Issuer     string        `json:"issuer"`     // 签发者标识
	Audience   string        `json:"audience"`   // 受众标识
}
