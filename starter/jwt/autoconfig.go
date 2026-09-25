package jwt

import (
	"fmt"
	"reflect"
	"time"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security"
)

func init() {
	boot.RegisterAutoConfigWith(&JwtAutoConfiguration{},
		boot.WithConditions(
			// 约定优于配置：当 security.enabled 和 security.jwt.enabled 未配置时，默认为 true（即默认启用）
			condition.OnPropertyOrDefault(SecurityEnabled, ConditionTrue, ConditionTrue),
			condition.OnPropertyOrDefault(JWTEnabled, ConditionTrue, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityAuthentication)), // 认证层，在安全核心之前执行
	)
}

// JwtAutoConfiguration JWT 自动配置类。
//
// 当配置文件中启用 JWT 时自动生效（security.enabled=true 且 security.jwt.enabled=true）。
// 负责注册 TokenProvider 和 JwtAuthenticationFilter 到 IoC 容器中。
//
// 执行顺序：Order = -1000，确保在 SecurityAutoConfiguration 之前执行，
// 这样 SecurityAutoConfiguration 构建安全过滤器链时能够找到 JWT 过滤器。
type JwtAutoConfiguration struct {
	logger log.Logger
}

// JwtConfig JWT 认证配置。
type JwtConfig struct {
	Enabled                bool   `json:"enabled" mapstructure:"enabled"`
	SecretKey              string `json:"secret-key" mapstructure:"secret-key"`
	Issuer                 string `json:"issuer" mapstructure:"issuer"`
	ExpiresDuration        int    `json:"expires-duration" mapstructure:"expires-duration"`
	RefreshExpiresDuration int    `json:"refresh-expires-duration" mapstructure:"refresh-expires-duration"`
	ExcludePaths           string `json:"exclude-paths" mapstructure:"exclude-paths"`
	SigningMethod          string `json:"signing-method" mapstructure:"signing-method"`
}

// Configure 配置 JWT 认证。
//
// 该方法在自动配置阶段调用，负责：
//  1. 从 Environment 中读取 JWT 配置（密钥、排除路径等）
//  2. 创建并注册 TokenProvider（用于生成和解析 Token）
//  3. 创建并注册 JwtAuthenticationFilter（用于请求认证）
//  4. 查找并注入 UserDetailsService（如果存在）
//
// 注意：该方法不需要检查 JWT 是否启用，因为 init() 中的 condition.OnProperty 已经确保
// 只有在 security.enabled=true 且 security.jwt.enabled=true 时才会执行此配置。
func (c *JwtAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	container := ctx.Container()
	env := ctx.Environment()

	c.resolveLogger(ctx, container)
	c.logger.Info(ctx.Context(), "configuring JWT authentication...")

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load JWT config: %w", err)
	}

	c.warnAboutConfig(ctx, cfg)

	tokenProvider, err := c.registerTokenProvider(ctx, container, cfg)
	if err != nil {
		return fmt.Errorf("failed to register token provider: %w", err)
	}

	if err := c.registerAuthFilter(ctx, container, tokenProvider, cfg); err != nil {
		return fmt.Errorf("failed to register JWT authentication filter: %w", err)
	}

	// 打印配置信息
	c.logger.Info(ctx.Context(), "JWT 配置",
		log.KeyValue{Key: LogFieldSecret, Value: maskSecret(cfg.SecretKey)},
		log.KeyValue{Key: LogFieldExclude, Value: cfg.ExcludePaths},
	)

	return nil
}

// resolveLogger 从容器获取日志记录器，缺省时使用默认实现。
func (c *JwtAutoConfiguration) resolveLogger(ctx boot.ApplicationContext, container core.Container) {
	if logger, err := core.GetByName[log.Logger](container, ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
		c.logger.Warn(ctx.Context(), "failed to get Logger from container, using default slog",
			log.KeyValue{Key: "error", Value: err.Error()})
	}
}

// warnAboutConfig 校验配置并给出默认密钥/签名方法提示。
func (c *JwtAutoConfiguration) warnAboutConfig(ctx boot.ApplicationContext, cfg *JwtConfig) {
	if cfg.SecretKey == "" {
		cfg.SecretKey = DefaultJWTSecretKey
		c.logger.Warn(ctx.Context(), "using default secret key, configure "+JWTSecretKey)
	}

	if cfg.SigningMethod != "" && cfg.SigningMethod != DefaultJWTSigningMethod {
		c.logger.Warn(ctx.Context(), "only HS256 signing method is supported, ignoring configured signing-method",
			log.KeyValue{Key: "signing-method", Value: cfg.SigningMethod})
	}
}

// registerTokenProvider 创建并注册 TokenProvider。
func (c *JwtAutoConfiguration) registerTokenProvider(ctx boot.ApplicationContext, container core.Container, cfg *JwtConfig) (*DefaultTokenProvider, error) {
	tokenProvider := NewTokenProvider(
		WithSecretKey(cfg.SecretKey),
		WithExpiration(time.Duration(cfg.ExpiresDuration)*time.Second),
		WithRefreshExpiration(time.Duration(cfg.RefreshExpiresDuration)*time.Second),
		WithIssuer(cfg.Issuer),
	)

	if err := container.RegisterInstance(tokenProvider, reflect.TypeFor[*DefaultTokenProvider]()); err != nil {
		return nil, fmt.Errorf("failed to register TokenProvider: %w", err)
	}
	c.logger.Info(ctx.Context(), "TokenProvider registered")

	return tokenProvider, nil
}

// registerAuthFilter 创建并注册 JWT 认证过滤器。
func (c *JwtAutoConfiguration) registerAuthFilter(ctx boot.ApplicationContext, container core.Container, tokenProvider *DefaultTokenProvider, cfg *JwtConfig) error {
	// 获取 UserDetailsService（可选）
	var userDetailsService security.UserDetailsService
	beans, err := container.Get(reflect.TypeFor[security.UserDetailsService]())
	if err == nil && len(beans) > 0 {
		userDetailsService, _ = beans[0].(security.UserDetailsService)
		c.logger.Info(ctx.Context(), "using registered UserDetailsService")
	}

	filterOpts := []JwtFilterOption{
		WithUserDetailsService(userDetailsService),
	}

	// 解析排除路径（支持逗号分隔的多个路径，支持通配符）
	if cfg.ExcludePaths != "" {
		paths := splitPaths(cfg.ExcludePaths)
		filterOpts = append(filterOpts, WithExcludePaths(paths...))
	}

	jwtFilter := NewJwtAuthenticationFilter(tokenProvider, filterOpts...)

	// 注册 JWT 过滤器（同时使用具体类型和接口类型注册）
	// 使用具体类型注册：允许通过 *JwtAuthenticationFilter 类型查找
	if err := core.Register[*JwtAuthenticationFilter](container,
		core.WithFactory[*JwtAuthenticationFilter](func(c ...any) (any, error) {
			return jwtFilter, nil
		}),
		core.WithScope[*JwtAuthenticationFilter](registry.Singleton),
	); err != nil {
		return fmt.Errorf("failed to register JwtAuthenticationFilter: %w", err)
	}
	// 使用接口类型注册：确保 SecurityAutoConfiguration 能通过 security.SecurityFilter 接口查找到
	// 这是关键步骤，否则 JWT 过滤器不会被添加到安全过滤器链中
	if err := core.Register[security.SecurityFilter](container,
		core.WithFactory[security.SecurityFilter](func(c ...any) (any, error) {
			return jwtFilter, nil
		}),
		core.WithScope[security.SecurityFilter](registry.Singleton),
	); err != nil {
		c.logger.Warn(ctx.Context(), "failed to register SecurityFilter interface (non-fatal)", log.KeyValue{Key: LogFieldError, Value: err.Error()})
	}
	c.logger.Info(ctx.Context(), "JwtAuthenticationFilter registered")

	return nil
}

// loadConfig 从 Environment 加载 JWT 配置。
func (c *JwtAutoConfiguration) loadConfig(env *environment.Environment) (*JwtConfig, error) {
	cfg := &JwtConfig{
		SecretKey:              DefaultJWTSecretKey,
		Issuer:                 DefaultJWTIssuer,
		ExpiresDuration:        DefaultJWTExpiresDuration,
		RefreshExpiresDuration: DefaultJWTRefreshExpiresDuration,
		ExcludePaths:           DefaultJWTExcludePaths,
		SigningMethod:          DefaultJWTSigningMethod,
	}

	if err := env.BindPrefix("security.jwt", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind JWT config: %w", err)
	}

	return cfg, nil
}

// splitPaths 分割路径字符串。
func splitPaths(pathsStr string) []string {
	if pathsStr == "" {
		return nil
	}

	var paths []string
	start := 0
	for i := 0; i < len(pathsStr); i++ {
		if pathsStr[i] == ',' {
			paths = append(paths, pathsStr[start:i])
			start = i + 1
		}
	}
	if start < len(pathsStr) {
		paths = append(paths, pathsStr[start:])
	}
	return paths
}

// maskSecret 隐藏密钥。
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}
