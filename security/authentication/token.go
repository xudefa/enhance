package authentication

// UsernamePasswordToken 用户名密码认证令牌。
//
// 用于基于用户名和密码的认证流程。
// 认证前包含用户名和密码，认证后包含用户主体和权限列表。
type UsernamePasswordToken struct {
	principal     any
	credentials   any
	authorities   []string
	authenticated bool
}

// NewUsernamePasswordToken 创建未认证的用户名密码令牌。
//
// 用于认证请求的初始阶段，包含用户名和密码。
func NewUsernamePasswordToken(principal, credentials any) *UsernamePasswordToken {
	return &UsernamePasswordToken{
		principal:     principal,
		credentials:   credentials,
		authorities:   nil,
		authenticated: false,
	}
}

// NewAuthenticatedUsernamePasswordToken 创建已认证的用户名密码令牌。
//
// 用于认证成功后的令牌，包含用户主体和权限列表。
func NewAuthenticatedUsernamePasswordToken(principal any, credentials any, authorities []string) *UsernamePasswordToken {
	return &UsernamePasswordToken{
		principal:     principal,
		credentials:   credentials,
		authorities:   authorities,
		authenticated: true,
	}
}

// Principal 返回主体标识。
func (t *UsernamePasswordToken) Principal() any {
	return t.principal
}

// Credentials 返回凭证信息。
func (t *UsernamePasswordToken) Credentials() any {
	return t.credentials
}

// Authorities 返回权限列表。
func (t *UsernamePasswordToken) Authorities() []string {
	return t.authorities
}

// Authenticated 返回是否已认证。
func (t *UsernamePasswordToken) Authenticated() bool {
	return t.authenticated
}

// SetAuthorities 设置权限列表。
func (t *UsernamePasswordToken) SetAuthorities(authorities []string) {
	t.authorities = authorities
}

// SetAuthenticated 设置认证状态。
func (t *UsernamePasswordToken) SetAuthenticated(authenticated bool) {
	t.authenticated = authenticated
}

// Name 返回主体名称。
//
// 支持 string 和 UserDetails 两种 principal 类型。
func (t *UsernamePasswordToken) Name() string {
	if name, ok := t.principal.(string); ok {
		return name
	}
	if userDetails, ok := t.principal.(UserDetails); ok {
		return userDetails.Username()
	}
	return ""
}

// anonymousToken 匿名认证令牌。
type anonymousToken struct{}

func (t *anonymousToken) Principal() any {
	return "anonymousUser"
}

func (t *anonymousToken) Credentials() any {
	return nil
}

func (t *anonymousToken) Authenticated() bool {
	return false
}

// NewAnonymousToken 创建匿名认证令牌。
func NewAnonymousToken() AuthenticationToken {
	return &anonymousToken{}
}
