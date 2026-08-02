package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 错误定义。
var (
	ErrInvalidToken = errors.New("invalid token")       // Token 无效（签名错误或格式错误）
	ErrExpiredToken = errors.New("token expired")       // Token 已过期
	ErrEmptySecret  = errors.New("secret key is empty") // 签名密钥为空
)

// DefaultTokenProvider 默认 Token 提供者实现。
//
// 实现 TokenProvider 接口，提供基于 HMAC-SHA256 签名算法的 JWT Token 生成和验证功能。
// 基于 github.com/golang-jwt/jwt/v5 库实现，符合 JWT RFC 7519 标准。
//
// 主要功能：
//   - 生成 JWT Token：使用 HMAC-SHA256 签名，包含用户信息和权限列表
//   - 解析 Token：验证签名并提取声明（Claims）信息
//   - 验证 Token：检查签名、过期时间和有效性
//   - 刷新 Token：基于旧 Token 生成新的 Token
//
// 安全特性：
//   - 强制验证签名算法（仅允许 HMAC 系列算法）
//   - 自动检查 Token 过期时间
//   - 防止算法混淆攻击（Algorithm Confusion Attack）
//
// 使用示例：
//
//	provider := jwt.NewTokenProvider(
//	    jwt.WithSecretKey("your-secret-key"),
//	    jwt.WithExpiration(time.Hour * 24),
//	)
//	token, err := provider.GenerateToken(ctx, "admin", []string{"ROLE_ADMIN"})
type DefaultTokenProvider struct {
	secretKey         string        // HMAC-SHA256 签名密钥
	expiration        time.Duration // Token 有效期
	refreshExpiration time.Duration // 刷新后 Token 有效期
	issuer            string        // 签发者标识
	audience          string        // 受众标识
}

// NewTokenProvider 创建 Token 提供者。
func NewTokenProvider(opts ...TokenOption) *DefaultTokenProvider {
	p := &DefaultTokenProvider{
		expiration: time.Hour,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// TokenOption Token 选项函数类型。
type TokenOption func(*DefaultTokenProvider)

// WithSecretKey 设置密钥。
func WithSecretKey(key string) TokenOption {
	return func(p *DefaultTokenProvider) {
		p.secretKey = key
	}
}

// WithExpiration 设置过期时间。
func WithExpiration(d time.Duration) TokenOption {
	return func(p *DefaultTokenProvider) {
		p.expiration = d
	}
}

// WithRefreshExpiration 设置刷新后 Token 的过期时间。
// 未设置时，刷新后的 Token 复用过期时间。
func WithRefreshExpiration(d time.Duration) TokenOption {
	return func(p *DefaultTokenProvider) {
		p.refreshExpiration = d
	}
}

// WithIssuer 设置签发者。
func WithIssuer(issuer string) TokenOption {
	return func(p *DefaultTokenProvider) {
		p.issuer = issuer
	}
}

// WithAudience 设置受众。
func WithAudience(audience string) TokenOption {
	return func(p *DefaultTokenProvider) {
		p.audience = audience
	}
}

// GenerateToken 生成 JWT Token。
func (p *DefaultTokenProvider) GenerateToken(ctx context.Context, username string, authorities []string) (string, error) {
	return p.generateToken(ctx, username, authorities, p.expiration)
}

// generateToken 使用指定的有效期生成 JWT Token。
func (p *DefaultTokenProvider) generateToken(ctx context.Context, username string, authorities []string, expiration time.Duration) (string, error) {
	if p.secretKey == "" {
		return "", ErrEmptySecret
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub":         username,
		"iss":         p.issuer,
		"aud":         p.audience,
		"exp":         now.Add(expiration).Unix(),
		"iat":         now.Unix(),
		"authorities": authorities,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(p.secretKey))
	if err != nil {
		return "", fmt.Errorf("签名 Token 失败: %w", err)
	}

	return tokenString, nil
}

// ParseToken 解析 JWT Token。
func (p *DefaultTokenProvider) ParseToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	if p.secretKey == "" {
		return nil, ErrEmptySecret
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(p.secretKey), nil
	}, jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("解析 Token 失败: %w", err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return p.extractClaims(claims)
}

// ValidateToken 验证 JWT Token 是否有效。
func (p *DefaultTokenProvider) ValidateToken(ctx context.Context, tokenString string) error {
	_, err := p.ParseToken(ctx, tokenString)
	return err
}

// RefreshToken 刷新 JWT Token。
func (p *DefaultTokenProvider) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	claims, err := p.ParseToken(ctx, tokenString)
	if err != nil {
		return "", err
	}

	refreshExpiration := p.refreshExpiration
	if refreshExpiration <= 0 {
		refreshExpiration = p.expiration
	}

	return p.generateToken(ctx, claims.Subject, claims.Authorities, refreshExpiration)
}

// extractClaims 从 JWT MapClaims 中提取 TokenClaims。
func (p *DefaultTokenProvider) extractClaims(claims jwt.MapClaims) (*TokenClaims, error) {
	subject, _ := claims.GetSubject()

	issuer, _ := claims.GetIssuer()

	audience, _ := claims.GetAudience()
	audienceStr := ""
	if len(audience) > 0 {
		audienceStr = audience[0]
	}

	expiration, err := claims.GetExpirationTime()
	if err != nil {
		return nil, fmt.Errorf("获取过期时间失败: %w", err)
	}

	issuedAt, err := claims.GetIssuedAt()
	if err != nil {
		return nil, fmt.Errorf("获取签发时间失败: %w", err)
	}

	authorities := []string{}
	if auths, ok := claims["authorities"].([]any); ok {
		for _, auth := range auths {
			if a, ok := auth.(string); ok {
				authorities = append(authorities, a)
			}
		}
	}

	return &TokenClaims{
		Subject:     subject,
		Issuer:      issuer,
		Audience:    audienceStr,
		Expiration:  expiration.Time,
		IssuedAt:    issuedAt.Time,
		Authorities: authorities,
	}, nil
}
