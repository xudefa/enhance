package security

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
)

// NoOpPasswordEncoder 不进行编码的密码编码器
// 适用场景：仅用于开发和测试环境，明文存储密码
// 警告：生产环境绝对不要使用此编码器
type NoOpPasswordEncoder struct{}

// NewNoOpPasswordEncoder 创建NoOp密码编码器
func NewNoOpPasswordEncoder() *NoOpPasswordEncoder {
	return &NoOpPasswordEncoder{}
}

// Encode 直接返回原始密码
func (e *NoOpPasswordEncoder) Encode(rawPassword string) string {
	return rawPassword
}

// Matches 直接比较原始密码和编码后密码
func (e *NoOpPasswordEncoder) Matches(rawPassword, encodedPassword string) bool {
	return subtle.ConstantTimeCompare([]byte(rawPassword), []byte(encodedPassword)) == 1
}

// Sha256PasswordEncoder 基于 SHA256 的密码编码器
//
// Deprecated: 此实现不使用 salt，易受彩虹表攻击。
// 仅适用于兼容性场景，不推荐用于生产环境。
// 生产环境应使用 BCryptPasswordEncoder 或 Argon2PasswordEncoder。
type Sha256PasswordEncoder struct{}

// NewSha256PasswordEncoder 创建 SHA256 密码编码器
func NewSha256PasswordEncoder() *Sha256PasswordEncoder {
	return &Sha256PasswordEncoder{}
}

// Encode 使用 SHA256 算法编码密码
func (e *Sha256PasswordEncoder) Encode(rawPassword string) string {
	hash := sha256.Sum256([]byte(rawPassword))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// Matches 比较编码后的密码
func (e *Sha256PasswordEncoder) Matches(rawPassword, encodedPassword string) bool {
	return subtle.ConstantTimeCompare([]byte(e.Encode(rawPassword)), []byte(encodedPassword)) == 1
}

// StandardPasswordEncoder 标准密码编码器
// 使用密钥对密码进行SHA256哈希
type StandardPasswordEncoder struct {
	secret string
}

// NewStandardPasswordEncoder 创建标准密码编码器
// secret: 用于加盐的密钥
func NewStandardPasswordEncoder(secret string) *StandardPasswordEncoder {
	return &StandardPasswordEncoder{secret: secret}
}

// Encode 使用密钥对密码进行编码
func (e *StandardPasswordEncoder) Encode(rawPassword string) string {
	combined := e.secret + rawPassword
	hash := sha256.Sum256([]byte(combined))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// Matches 比较编码后的密码
func (e *StandardPasswordEncoder) Matches(rawPassword, encodedPassword string) bool {
	return subtle.ConstantTimeCompare([]byte(e.Encode(rawPassword)), []byte(encodedPassword)) == 1
}

// DelegatingPasswordEncoder 委托密码编码器
// 职责：支持多种编码器，可根据编码ID选择合适的编码器
// 编码格式：{编码器ID}encodedPassword，例如：{bcrypt}$2a$10$...
// 优势：支持密码编码算法迁移，历史密码无需重新编码
type DelegatingPasswordEncoder struct {
	idForEncode      string
	passwordEncoders map[string]PasswordEncoder
}

// NewDelegatingPasswordEncoder 创建委托密码编码器
// idForEncode: 默认使用的编码器ID
// passwordEncoders: 可用的编码器映射
func NewDelegatingPasswordEncoder(idForEncode string, passwordEncoders map[string]PasswordEncoder) *DelegatingPasswordEncoder {
	// 在构造时检查编码器是否存在，避免运行时panic
	if _, exists := passwordEncoders[idForEncode]; !exists {
		panic(fmt.Sprintf("encoder not found for id: %s", idForEncode))
	}
	encoders := make(map[string]PasswordEncoder, len(passwordEncoders))
	for k, v := range passwordEncoders {
		encoders[k] = v
	}
	return &DelegatingPasswordEncoder{
		idForEncode:      idForEncode,
		passwordEncoders: encoders,
	}
}

// Encode 使用默认编码器编码密码
func (e *DelegatingPasswordEncoder) Encode(rawPassword string) string {
	encoder, exists := e.passwordEncoders[e.idForEncode]
	if !exists {
		// 理论上不应该发生（构造时已检查），防御性处理
		return ""
	}
	encoded := encoder.Encode(rawPassword)
	return fmt.Sprintf("{%s}%s", e.idForEncode, encoded)
}

// Matches 根据编码ID选择合适的编码器进行匹配
func (e *DelegatingPasswordEncoder) Matches(rawPassword, encodedPassword string) bool {
	id, password, err := e.extractId(encodedPassword)
	if err != nil {
		return false
	}

	encoder, exists := e.passwordEncoders[id]
	if !exists {
		return false
	}

	return encoder.Matches(rawPassword, password)
}

// extractId 从编码后的密码中提取编码器ID和实际密码
// 格式: {id}encodedPassword
func (e *DelegatingPasswordEncoder) extractId(encodedPassword string) (string, string, error) {
	if len(encodedPassword) < 3 || encodedPassword[0] != '{' {
		return "", "", fmt.Errorf("invalid encoded password format: missing prefix '{'")
	}

	end := strings.Index(encodedPassword, "}")
	if end == -1 {
		return "", "", fmt.Errorf("invalid encoded password format: missing suffix '}'")
	}

	id := encodedPassword[1:end]
	if id == "" {
		return "", "", fmt.Errorf("invalid encoded password format: empty encoder id")
	}

	password := encodedPassword[end+1:]
	return id, password, nil
}
