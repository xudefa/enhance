package security

import "context"

// ==================== 配置键常量 ====================

const (
	// Casbin 配置
	CasbinEnabled          = "security.casbin.enabled"
	CasbinModelType        = "security.casbin.model-type"
	CasbinModelPath        = "security.casbin.model-path"
	CasbinModelText        = "security.casbin.model-text"
	CasbinPolicyType       = "security.casbin.policy-type"
	CasbinPolicyPath       = "security.casbin.policy-path"
	CasbinPolicyText       = "security.casbin.policy-text"
	CasbinAutoLoad         = "security.casbin.auto-load"
	CasbinAutoLoadInterval = "security.casbin.auto-load-interval"

	// casbin 字段常量
	CasbinLogFieldModel  = "model-path"
	CasbinLogFieldPolicy = "policy-path"
)

// ==================== 默认值常量 ====================

const (
	// Casbin 默认值
	DefaultCasbinModelType        = "file"
	DefaultCasbinModelPath        = "./config/casbin_model.conf"
	DefaultCasbinPolicyType       = "file"
	DefaultCasbinPolicyPath       = "./config/casbin_policy.csv"
	DefaultCasbinAutoLoad         = false
	DefaultCasbinAutoLoadInterval = 5
)

// CasbinEnforcer Casbin 执行器接口。
//
// 封装 Casbin 的核心功能，提供权限检查。
// 实现类：DefaultCasbinEnforcer（starter/casbin）、GormCasbinEnforcer（starter/casbin-gorm）、XormCasbinEnforcer（starter/casbin-xorm）
type CasbinEnforcer interface {
	// Enforce 检查请求是否允许。
	// 参数：subject - 主体（用户/角色），object - 资源，action - 操作
	// 返回：是否允许，错误
	Enforce(ctx context.Context, subject, object, action string) (bool, error)
	// AddPolicy 添加策略。
	// 参数：sub - 主体，obj - 资源，act - 操作
	// 返回：错误
	AddPolicy(ctx context.Context, sub, obj, act string) error
	// RemovePolicy 移除策略。
	// 参数：sub - 主体，obj - 资源，act - 操作
	// 返回：错误
	RemovePolicy(ctx context.Context, sub, obj, act string) error
	// GetPolicy 获取所有策略。
	// 返回：策略列表（二维数组），错误
	GetPolicy(ctx context.Context) ([][]string, error)
	// LoadPolicy 重新加载策略。
	// 返回：错误
	LoadPolicy(ctx context.Context) error
	// SavePolicy 保存策略。
	// 返回：错误
	SavePolicy(ctx context.Context) error
}

// CasbinVoter Casbin 投票者实现。
//
// 实现 AccessDecisionVoter 接口，将 Casbin 的权限检查集成到 enhance 的访问决策机制中。
//
// 工作流程：
//  1. 检查用户是否已认证（未认证则弃权）
//  2. 从 Authentication 中提取用户名作为 subject
//  3. 从 SecurityRequest 中提取 URI 和 HTTP 方法作为 object 和 action
//  4. 调用 CasbinEnforcer.Enforce() 进行权限检查
//  5. 返回投票结果（ACCESS_GRANTED / ACCESS_DENIED / ACCESS_ABSTAIN）
//
// 投票策略：
//   - 未认证用户：ACCESS_ABSTAIN（弃权，让其他 Voter 决定）
//   - Casbin 允许：ACCESS_GRANTED（授予访问）
//   - Casbin 拒绝：ACCESS_DENIED（拒绝访问）
//   - 发生错误：ACCESS_ABSTAIN（弃权，让其他 Voter 决定）
//
// 注意：CasbinVoter 的投票结果会被 AffirmativeBased 决策管理器使用，
// 只要有一个 Voter 返回 ACCESS_GRANTED 就允许访问。
type CasbinVoter struct {
	enforcer CasbinEnforcer // Casbin 执行器，负责权限检查
}

// NewCasbinVoter 创建 Casbin 投票者。
func NewCasbinVoter(enforcer CasbinEnforcer) *CasbinVoter {
	return &CasbinVoter{
		enforcer: enforcer,
	}
}

// Vote 投票决定是否授予访问权限。
func (v *CasbinVoter) Vote(ctx context.Context, authentication Authentication, object any, attributes []string) int {
	// 未认证则弃权
	if authentication == nil || !authentication.Authenticated() {
		return ACCESS_ABSTAIN
	}

	// 检查对象类型
	request, ok := object.(SecurityRequest)
	if !ok {
		return ACCESS_ABSTAIN
	}

	// 获取请求信息
	subject := authentication.Name()
	obj := request.GetURI()
	action := request.GetMethod()

	// 使用 Casbin 进行权限检查
	allowed, err := v.enforcer.Enforce(ctx, subject, obj, action)
	if err != nil {
		// 出错时弃权，让其他 Voter 决定
		return ACCESS_ABSTAIN
	}

	if allowed {
		return ACCESS_GRANTED
	}

	// 拒绝时返回 DENIED
	return ACCESS_DENIED
}

// Supports 是否支持该属性。
func (v *CasbinVoter) Supports(ctx context.Context, authentication Authentication, object any, attributes []string) bool {
	// CasbinVoter 支持所有请求
	return true
}
