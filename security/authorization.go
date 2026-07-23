package security

import (
	"context"
	"strings"

	"github.com/xudefa/enhance/security/authorization"
)

// 访问决策投票结果常量
const (
	ACCESS_GRANTED = 1  // 允许访问
	ACCESS_DENIED  = -1 // 拒绝访问
	ACCESS_ABSTAIN = 0  // 投票结果为 abstain 弃权
)

// DefaultRolePrefix is the standard prefix prepended to role names.
const DefaultRolePrefix = "ROLE_"

// WebExpressionVoter Web表达式投票者
type WebExpressionVoter struct{}

func NewWebExpressionVoter() *WebExpressionVoter {
	return &WebExpressionVoter{}
}

// Vote 投票决定访问权限
func (v *WebExpressionVoter) Vote(ctx context.Context, authentication authorization.Authentication, resource string, attributes []string) int {
	if len(attributes) == 0 {
		return ACCESS_ABSTAIN
	}

	for _, attribute := range attributes {
		if attribute == "permitAll" {
			return ACCESS_GRANTED
		}

		if attribute == "denyAll" {
			return ACCESS_DENIED
		}

		if attribute == "authenticated" {
			if authentication != nil && authentication.Authenticated() {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}

		if strings.HasPrefix(attribute, "hasRole('") && strings.HasSuffix(attribute, "')") {
			role := strings.TrimPrefix(attribute, "hasRole('")
			role = strings.TrimSuffix(role, "')")
			if v.hasRole(authentication, role) {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}

		if strings.HasPrefix(attribute, "hasAnyRole('") && strings.HasSuffix(attribute, "')") {
			rolesStr := strings.TrimPrefix(attribute, "hasAnyRole('")
			rolesStr = strings.TrimSuffix(rolesStr, "')")
			roles := strings.Split(rolesStr, "','")
			if v.hasAnyRole(authentication, roles) {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}

		if strings.HasPrefix(attribute, "hasAuthority('") && strings.HasSuffix(attribute, "')") {
			authority := strings.TrimPrefix(attribute, "hasAuthority('")
			authority = strings.TrimSuffix(authority, "')")
			if v.hasAuthority(authentication, authority) {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}

		if strings.HasPrefix(attribute, "hasAnyAuthority('") && strings.HasSuffix(attribute, "')") {
			authoritiesStr := strings.TrimPrefix(attribute, "hasAnyAuthority('")
			authoritiesStr = strings.TrimSuffix(authoritiesStr, "')")
			authorities := strings.Split(authoritiesStr, "','")
			if v.hasAnyAuthority(authentication, authorities) {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}
	}

	return ACCESS_ABSTAIN
}

// Supports 是否支持该属性
func (v *WebExpressionVoter) Supports(attribute string) bool {
	return true
}

func (v *WebExpressionVoter) hasRole(authentication authorization.Authentication, role string) bool {
	if authentication == nil {
		return false
	}
	authorities := authentication.Authorities()
	for _, auth := range authorities {
		if auth == DefaultRolePrefix+role || auth == role {
			return true
		}
	}
	return false
}

func (v *WebExpressionVoter) hasAnyRole(authentication authorization.Authentication, roles []string) bool {
	if authentication == nil {
		return false
	}
	for _, role := range roles {
		if v.hasRole(authentication, role) {
			return true
		}
	}
	return false
}

func (v *WebExpressionVoter) hasAuthority(authentication authorization.Authentication, authority string) bool {
	if authentication == nil {
		return false
	}
	authorities := authentication.Authorities()
	for _, auth := range authorities {
		if auth == authority {
			return true
		}
	}
	return false
}

func (v *WebExpressionVoter) hasAnyAuthority(authentication authorization.Authentication, authorities []string) bool {
	if authentication == nil {
		return false
	}
	for _, authority := range authorities {
		if v.hasAuthority(authentication, authority) {
			return true
		}
	}
	return false
}

// RoleVoter 角色投票者
type RoleVoter struct {
	rolePrefix string
}

func NewRoleVoter() *RoleVoter {
	return &RoleVoter{
		rolePrefix: DefaultRolePrefix,
	}
}

// Vote 投票决定访问权限
func (v *RoleVoter) Vote(ctx context.Context, authentication authorization.Authentication, resource string, attributes []string) int {
	if len(attributes) == 0 {
		return ACCESS_ABSTAIN
	}
	for _, attribute := range attributes {
		if v.Supports(attribute) {
			role := strings.TrimPrefix(attribute, v.rolePrefix)
			if v.hasRole(authentication, role) {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}
	}
	return ACCESS_ABSTAIN
}

// Supports 是否支持该属性
func (v *RoleVoter) Supports(attribute string) bool {
	return strings.HasPrefix(attribute, v.rolePrefix)
}

func (v *RoleVoter) hasRole(authentication authorization.Authentication, role string) bool {
	if authentication == nil {
		return false
	}
	authorities := authentication.Authorities()
	for _, auth := range authorities {
		if auth == v.rolePrefix+role || auth == role {
			return true
		}
	}
	return false
}

func (v *RoleVoter) SetRolePrefix(prefix string) {
	v.rolePrefix = prefix
}

// AuthenticatedVoter 认证投票者
type AuthenticatedVoter struct{}

func NewAuthenticatedVoter() *AuthenticatedVoter {
	return &AuthenticatedVoter{}
}

// Vote 投票决定访问权限
func (v *AuthenticatedVoter) Vote(ctx context.Context, authentication authorization.Authentication, resource string, attributes []string) int {
	if len(attributes) == 0 {
		return ACCESS_ABSTAIN
	}
	for _, attribute := range attributes {
		if attribute == "IS_AUTHENTICATED_FULLY" {
			if authentication != nil && authentication.Authenticated() {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}
		if attribute == "IS_AUTHENTICATED_REMEMBERED" {
			if authentication != nil && authentication.Authenticated() {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}
		if attribute == "IS_AUTHENTICATED_ANONYMOUSLY" {
			return ACCESS_GRANTED
		}
	}
	return ACCESS_ABSTAIN
}

// Supports 是否支持该属性
func (v *AuthenticatedVoter) Supports(attribute string) bool {
	return true
}

// AffirmativeBased 肯定优先访问决策管理器
type AffirmativeBased struct {
	decisionVoters             []AccessDecisionVoter
	allowIfAllAbstainDecisions bool
}

func NewAffirmativeBased(voters ...AccessDecisionVoter) *AffirmativeBased {
	return &AffirmativeBased{
		decisionVoters:             voters,
		allowIfAllAbstainDecisions: false,
	}
}

// Decide 决定是否授予访问权限
func (m *AffirmativeBased) Decide(ctx context.Context, authentication authorization.Authentication, resource string, attributes []string) error {
	grant := 0
	deny := 0
	abstain := 0

	for _, voter := range m.decisionVoters {
		result := voter.Vote(ctx, authentication, resource, attributes)
		switch result {
		case ACCESS_GRANTED:
			grant++
		case ACCESS_DENIED:
			deny++
		case ACCESS_ABSTAIN:
			abstain++
		}
	}

	if grant > 0 {
		return nil
	}
	if deny > 0 {
		return ErrAccessDenied
	}
	if m.allowIfAllAbstainDecisions {
		return nil
	}
	return ErrAccessDenied
}

// Supports 是否支持该决策属性
func (m *AffirmativeBased) Supports(attribute string) bool {
	return true
}

func (m *AffirmativeBased) AddVoter(voter AccessDecisionVoter) {
	m.decisionVoters = append(m.decisionVoters, voter)
}

func (m *AffirmativeBased) SetAllowIfAllAbstainDecisions(allow bool) {
	m.allowIfAllAbstainDecisions = allow
}

// UnanimousBased 一致通过访问决策管理器
type UnanimousBased struct {
	decisionVoters             []AccessDecisionVoter
	allowIfAllAbstainDecisions bool
}

func NewUnanimousBased(voters ...AccessDecisionVoter) *UnanimousBased {
	return &UnanimousBased{
		decisionVoters:             voters,
		allowIfAllAbstainDecisions: false,
	}
}

// Decide 决定是否授予访问权限
func (m *UnanimousBased) Decide(ctx context.Context, authentication authorization.Authentication, resource string, attributes []string) error {
	deny := 0
	grant := 0
	abstain := 0

	for _, voter := range m.decisionVoters {
		result := voter.Vote(ctx, authentication, resource, attributes)
		switch result {
		case ACCESS_GRANTED:
			grant++
		case ACCESS_DENIED:
			deny++
		case ACCESS_ABSTAIN:
			abstain++
		}
	}

	if deny > 0 {
		return ErrAccessDenied
	}
	if grant > 0 {
		return nil
	}
	if m.allowIfAllAbstainDecisions {
		return nil
	}
	return ErrAccessDenied
}

// Supports 是否支持该决策属性
func (m *UnanimousBased) Supports(attribute string) bool {
	return true
}

func (m *UnanimousBased) AddVoter(voter AccessDecisionVoter) {
	m.decisionVoters = append(m.decisionVoters, voter)
}

func (m *UnanimousBased) SetAllowIfAllAbstainDecisions(allow bool) {
	m.allowIfAllAbstainDecisions = allow
}

// ConsensusBased 共识优先访问决策管理器
type ConsensusBased struct {
	decisionVoters             []AccessDecisionVoter
	allowIfEqualGrantedDenied  bool
	allowIfAllAbstainDecisions bool
}

func NewConsensusBased(voters ...AccessDecisionVoter) *ConsensusBased {
	return &ConsensusBased{
		decisionVoters:             voters,
		allowIfEqualGrantedDenied:  false,
		allowIfAllAbstainDecisions: false,
	}
}

// Decide 决定是否授予访问权限
func (m *ConsensusBased) Decide(ctx context.Context, authentication authorization.Authentication, resource string, attributes []string) error {
	grant := 0
	deny := 0
	abstain := 0

	for _, voter := range m.decisionVoters {
		result := voter.Vote(ctx, authentication, resource, attributes)
		switch result {
		case ACCESS_GRANTED:
			grant++
		case ACCESS_DENIED:
			deny++
		case ACCESS_ABSTAIN:
			abstain++
		}
	}

	if grant > deny {
		return nil
	}
	if deny > grant {
		return ErrAccessDenied
	}
	if grant == deny {
		if m.allowIfEqualGrantedDenied {
			return nil
		}
		return ErrAccessDenied
	}
	if m.allowIfAllAbstainDecisions {
		return nil
	}
	return ErrAccessDenied
}

// Supports 是否支持该决策属性
func (m *ConsensusBased) Supports(attribute string) bool {
	return true
}

func (m *ConsensusBased) AddVoter(voter AccessDecisionVoter) {
	m.decisionVoters = append(m.decisionVoters, voter)
}

func (m *ConsensusBased) SetAllowIfEqualGrantedDenied(allow bool) {
	m.allowIfEqualGrantedDenied = allow
}

func (m *ConsensusBased) SetAllowIfAllAbstainDecisions(allow bool) {
	m.allowIfAllAbstainDecisions = allow
}
