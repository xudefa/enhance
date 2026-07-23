package authorization

import (
	"context"
	"strings"
)

// affirmativeBased 肯定优先访问决策管理器。
//
// 只要有一个投票者授予访问权限就允许访问（最宽松策略）。
// 适用场景：多个权限满足其一即可访问。
type affirmativeBased struct {
	voters                     []AccessDecisionVoter
	allowIfAllAbstainDecisions bool
}

// NewAffirmativeBased 创建肯定优先决策管理器。
func NewAffirmativeBased(voters ...AccessDecisionVoter) AccessDecisionManager {
	return &affirmativeBased{
		voters:                     voters,
		allowIfAllAbstainDecisions: false,
	}
}

// Decide 决定是否授予访问权限。
//
// 策略：只要有一个投票者授予权限就通过。
// 如果所有投票者都弃权，根据配置决定是否允许。
func (m *affirmativeBased) Decide(ctx context.Context, authentication Authentication, resource string, attributes []string) error {
	grant := 0
	deny := 0
	abstain := 0

	for _, voter := range m.voters {
		if !m.supportsAny(voter, attributes) {
			continue
		}
		result := voter.Vote(ctx, authentication, resource, attributes)
		switch result {
		case AccessGranted:
			grant++
		case AccessDenied:
			deny++
		case AccessAbstain:
			abstain++
		}
	}

	if grant > 0 {
		return nil
	}

	if deny > 0 {
		return ErrAccessDenied
	}

	if m.allowIfAllAbstainDecisions || abstain == 0 {
		return nil
	}

	return ErrAccessDenied
}

// Supports 检查是否有任何投票者支持该属性。
func (m *affirmativeBased) Supports(attribute string) bool {
	for _, voter := range m.voters {
		if voter.Supports(attribute) {
			return true
		}
	}
	return false
}

// supportsAny 检查投票者是否支持任何属性。
func (m *affirmativeBased) supportsAny(voter AccessDecisionVoter, attributes []string) bool {
	for _, attr := range attributes {
		if voter.Supports(attr) {
			return true
		}
	}
	return false
}

// AddVoter 添加投票者。
func (m *affirmativeBased) AddVoter(voter AccessDecisionVoter) {
	m.voters = append(m.voters, voter)
}

// SetAllowIfAllAbstainDecisions 设置是否在所有投票者都弃权时允许访问。
func (m *affirmativeBased) SetAllowIfAllAbstainDecisions(allow bool) {
	m.allowIfAllAbstainDecisions = allow
}

// unanimousBased 一致通过访问决策管理器。
//
// 只有所有投票者都不拒绝才允许访问（最严格策略）。
// 适用场景：需要满足所有权限要求才能访问。
type unanimousBased struct {
	voters                     []AccessDecisionVoter
	allowIfAllAbstainDecisions bool
}

// NewUnanimousBased 创建一致通过决策管理器。
func NewUnanimousBased(voters ...AccessDecisionVoter) AccessDecisionManager {
	return &unanimousBased{
		voters:                     voters,
		allowIfAllAbstainDecisions: false,
	}
}

// Decide 决定是否授予访问权限。
//
// 策略：所有投票者都不拒绝，且至少有一个授予权限时通过。
func (m *unanimousBased) Decide(ctx context.Context, authentication Authentication, resource string, attributes []string) error {
	deny := 0
	grant := 0
	abstain := 0

	for _, voter := range m.voters {
		if !m.supportsAny(voter, attributes) {
			continue
		}
		result := voter.Vote(ctx, authentication, resource, attributes)
		switch result {
		case AccessGranted:
			grant++
		case AccessDenied:
			deny++
		case AccessAbstain:
			abstain++
		}
	}

	if deny > 0 {
		return ErrAccessDenied
	}

	if grant > 0 {
		return nil
	}

	if m.allowIfAllAbstainDecisions || abstain == 0 {
		return nil
	}

	return ErrAccessDenied
}

// Supports 检查是否有任何投票者支持该属性。
func (m *unanimousBased) Supports(attribute string) bool {
	for _, voter := range m.voters {
		if voter.Supports(attribute) {
			return true
		}
	}
	return false
}

// supportsAny 检查投票者是否支持任何属性。
func (m *unanimousBased) supportsAny(voter AccessDecisionVoter, attributes []string) bool {
	for _, attr := range attributes {
		if voter.Supports(attr) {
			return true
		}
	}
	return false
}

// AddVoter 添加投票者。
func (m *unanimousBased) AddVoter(voter AccessDecisionVoter) {
	m.voters = append(m.voters, voter)
}

// SetAllowIfAllAbstainDecisions 设置是否在所有投票者都弃权时允许访问。
func (m *unanimousBased) SetAllowIfAllAbstainDecisions(allow bool) {
	m.allowIfAllAbstainDecisions = allow
}

// consensusBased 共识优先访问决策管理器。
//
// 根据多数投票结果决定访问权限（民主策略）。
// 适用场景：需要综合考虑多个权限维度。
type consensusBased struct {
	voters                     []AccessDecisionVoter
	allowIfEqualGrantedDenied  bool
	allowIfAllAbstainDecisions bool
}

// NewConsensusBased 创建共识优先决策管理器。
func NewConsensusBased(voters ...AccessDecisionVoter) AccessDecisionManager {
	return &consensusBased{
		voters:                     voters,
		allowIfEqualGrantedDenied:  false,
		allowIfAllAbstainDecisions: false,
	}
}

// Decide 决定是否授予访问权限。
//
// 策略：根据多数票决定，相同数量时按配置处理。
func (m *consensusBased) Decide(ctx context.Context, authentication Authentication, resource string, attributes []string) error {
	grant := 0
	deny := 0
	abstain := 0

	for _, voter := range m.voters {
		if !m.supportsAny(voter, attributes) {
			continue
		}
		result := voter.Vote(ctx, authentication, resource, attributes)
		switch result {
		case AccessGranted:
			grant++
		case AccessDenied:
			deny++
		case AccessAbstain:
			abstain++
		}
	}

	if grant == 0 && deny == 0 {
		if m.allowIfAllAbstainDecisions || abstain == 0 {
			return nil
		}
		return ErrAccessDenied
	}

	if grant > deny {
		return nil
	}

	if deny > grant {
		return ErrAccessDenied
	}

	if m.allowIfEqualGrantedDenied {
		return nil
	}

	return ErrAccessDenied
}

// Supports 检查是否有任何投票者支持该属性。
func (m *consensusBased) Supports(attribute string) bool {
	for _, voter := range m.voters {
		if voter.Supports(attribute) {
			return true
		}
	}
	return false
}

// supportsAny 检查投票者是否支持任何属性。
func (m *consensusBased) supportsAny(voter AccessDecisionVoter, attributes []string) bool {
	for _, attr := range attributes {
		if voter.Supports(attr) {
			return true
		}
	}
	return false
}

// AddVoter 添加投票者。
func (m *consensusBased) AddVoter(voter AccessDecisionVoter) {
	m.voters = append(m.voters, voter)
}

// SetAllowIfEqualGrantedDenied 设置当授权和拒绝票数相同时的处理方式。
func (m *consensusBased) SetAllowIfEqualGrantedDenied(allow bool) {
	m.allowIfEqualGrantedDenied = allow
}

// SetAllowIfAllAbstainDecisions 设置是否在所有投票者都弃权时允许访问。
func (m *consensusBased) SetAllowIfAllAbstainDecisions(allow bool) {
	m.allowIfAllAbstainDecisions = allow
}

// webExpressionVoter Web 表达式投票者。
//
// 支持 Spring Security 风格的 Web 表达式语言。
// 用于解析和执行权限表达式（如 hasRole、hasAuthority 等）。
type webExpressionVoter struct{}

// NewWebExpressionVoter 创建 Web 表达式投票者。
func NewWebExpressionVoter() AccessDecisionVoter {
	return &webExpressionVoter{}
}

// Vote 投票决定访问权限。
//
// 支持的表达式：
//   - permitAll: 允许所有人访问
//   - denyAll: 拒绝所有人访问
//   - authenticated: 仅允许已认证用户
//   - hasRole('ROLE'): 检查是否具有指定角色
//   - hasAnyRole('ROLE1','ROLE2'): 检查是否具有任一指定角色
//   - hasAuthority('AUTHORITY'): 检查是否具有指定权限
//   - hasAnyAuthority('AUTH1','AUTH2'): 检查是否具有任一指定权限
func (v *webExpressionVoter) Vote(_ context.Context, authentication Authentication, _ string, attributes []string) int {
	if len(attributes) == 0 {
		return AccessAbstain
	}

	for _, attribute := range attributes {
		result := v.evaluateAttribute(authentication, attribute)
		if result != AccessAbstain {
			return result
		}
	}

	return AccessAbstain
}

// Supports 支持所有属性。
func (v *webExpressionVoter) Supports(_ string) bool {
	return true
}

// evaluateAttribute 评估单个属性表达式。
func (v *webExpressionVoter) evaluateAttribute(authentication Authentication, attribute string) int {
	switch {
	case attribute == "permitAll":
		return AccessGranted
	case attribute == "denyAll":
		return AccessDenied
	case attribute == "authenticated":
		return v.checkAuthenticated(authentication)
	case strings.HasPrefix(attribute, "hasRole('") && strings.HasSuffix(attribute, "')"):
		role := extractExpressionArg(attribute, "hasRole('", "')")
		return v.checkRole(authentication, role)
	case strings.HasPrefix(attribute, "hasAnyRole('") && strings.HasSuffix(attribute, "')"):
		roles := splitExpressionArgs(attribute, "hasAnyRole('", "')")
		return v.checkAnyRole(authentication, roles)
	case strings.HasPrefix(attribute, "hasAuthority('") && strings.HasSuffix(attribute, "')"):
		authority := extractExpressionArg(attribute, "hasAuthority('", "')")
		return v.checkAuthority(authentication, authority)
	case strings.HasPrefix(attribute, "hasAnyAuthority('") && strings.HasSuffix(attribute, "')"):
		authorities := splitExpressionArgs(attribute, "hasAnyAuthority('", "')")
		return v.checkAnyAuthority(authentication, authorities)
	}
	return AccessAbstain
}

// checkAuthenticated 检查是否已认证。
func (v *webExpressionVoter) checkAuthenticated(authentication Authentication) int {
	if authentication != nil && authentication.Authenticated() {
		return AccessGranted
	}
	return AccessDenied
}

// checkRole 检查是否具有指定角色。
func (v *webExpressionVoter) checkRole(authentication Authentication, role string) int {
	if authentication == nil {
		return AccessDenied
	}
	for _, auth := range authentication.Authorities() {
		if auth == "ROLE_"+role || auth == role {
			return AccessGranted
		}
	}
	return AccessDenied
}

// checkAnyRole 检查是否具有任一指定角色。
func (v *webExpressionVoter) checkAnyRole(authentication Authentication, roles []string) int {
	for _, role := range roles {
		if v.checkRole(authentication, role) == AccessGranted {
			return AccessGranted
		}
	}
	return AccessDenied
}

// checkAuthority 检查是否具有指定权限。
func (v *webExpressionVoter) checkAuthority(authentication Authentication, authority string) int {
	if authentication == nil {
		return AccessDenied
	}
	for _, auth := range authentication.Authorities() {
		if auth == authority {
			return AccessGranted
		}
	}
	return AccessDenied
}

// checkAnyAuthority 检查是否具有任一指定权限。
func (v *webExpressionVoter) checkAnyAuthority(authentication Authentication, authorities []string) int {
	for _, authority := range authorities {
		if v.checkAuthority(authentication, authority) == AccessGranted {
			return AccessGranted
		}
	}
	return AccessDenied
}

// extractExpressionArg 从表达式中提取参数。
func extractExpressionArg(attribute, prefix, suffix string) string {
	arg := strings.TrimPrefix(attribute, prefix)
	return strings.TrimSuffix(arg, suffix)
}

// splitExpressionArgs 从表达式中提取并分割参数。
func splitExpressionArgs(attribute, prefix, suffix string) []string {
	argsStr := extractExpressionArg(attribute, prefix, suffix)
	return strings.Split(argsStr, "','")
}
