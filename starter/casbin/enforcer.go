package casbin

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security"
)

// DefaultCasbinEnforcer 默认 Casbin 执行器实现。
//
// 基于 github.com/casbin/casbin/v2 官方库实现，提供完整的 RBAC 权限控制能力。
// 实现了 security.CasbinEnforcer 接口，用于：
//   - 策略管理：添加、移除、查询权限策略
//   - 角色管理：用户与角色的映射关系
//   - 权限检查：根据策略和角色判断是否允许访问
//   - 模型管理：支持从文件或字符串加载 Casbin 模型
//   - 策略持久化：支持从文件或字符串加载/保存策略
type DefaultCasbinEnforcer struct {
	logger   log.Logger
	mu       sync.RWMutex
	enforcer *casbin.Enforcer
}

// NewCasbinEnforcer 创建 Casbin 执行器。
func NewCasbinEnforcer(logger log.Logger, modelPath, policyPath string) *DefaultCasbinEnforcer {
	// 初始化 Casbin 执行器
	enforcer, err := casbin.NewEnforcer(modelPath, policyPath)
	if err != nil {
		logger.Error(context.Background(), "创建 Casbin 执行器失败", log.KeyValue{Key: "error", Value: err.Error()})
		panic(fmt.Sprintf("创建 Casbin 执行器失败: %v", err))
	}
	logger.Info(context.Background(), "Casbin 执行器初始化成功", log.KeyValue{Key: "model", Value: modelPath})
	logger.Info(context.Background(), "Casbin 策略路径", log.KeyValue{Key: "policy", Value: policyPath})

	return &DefaultCasbinEnforcer{
		enforcer: enforcer,
		logger:   logger,
	}
}

// Enforce 检查请求是否允许。
func (e *DefaultCasbinEnforcer) Enforce(ctx context.Context, subject, object, action string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.Enforce(subject, object, action)
}

// AddPolicy 添加策略。
func (e *DefaultCasbinEnforcer) AddPolicy(ctx context.Context, sub, obj, act string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.enforcer.AddPolicy(sub, obj, act)
	return err
}

// RemovePolicy 移除策略。
func (e *DefaultCasbinEnforcer) RemovePolicy(ctx context.Context, sub, obj, act string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.enforcer.RemovePolicy(sub, obj, act)
	return err
}

// GetPolicy 获取所有策略。
func (e *DefaultCasbinEnforcer) GetPolicy(ctx context.Context) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	policies, err := e.enforcer.GetPolicy()
	return policies, err
}

// LoadPolicy 重新加载策略。
func (e *DefaultCasbinEnforcer) LoadPolicy(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.LoadPolicy()
}

// SavePolicy 保存策略。
func (e *DefaultCasbinEnforcer) SavePolicy(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.SavePolicy()
}

// AddRole 添加角色映射（便捷方法）。
func (e *DefaultCasbinEnforcer) AddRole(ctx context.Context, user, role string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.enforcer.AddRoleForUser(user, role)
	return err
}

// GetRolesForUser 获取用户的所有角色（便捷方法）。
func (e *DefaultCasbinEnforcer) GetRolesForUser(ctx context.Context, user string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	roles, err := e.enforcer.GetRolesForUser(user)
	return roles, err
}

// AddRoleForUser 为用户添加角色（便捷方法）。
func (e *DefaultCasbinEnforcer) AddRoleForUser(ctx context.Context, user, role string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.AddRoleForUser(user, role)
}

// DeleteRoleForUser 删除用户角色（便捷方法）。
func (e *DefaultCasbinEnforcer) DeleteRoleForUser(ctx context.Context, user, role string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.DeleteRoleForUser(user, role)
}

// GetUsersForRole 获取角色的所有用户（便捷方法）。
func (e *DefaultCasbinEnforcer) GetUsersForRole(ctx context.Context, role string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	users, err := e.enforcer.GetUsersForRole(role)
	return users, err
}

// HasRoleForUser 检查用户是否有指定角色（便捷方法）。
func (e *DefaultCasbinEnforcer) HasRoleForUser(ctx context.Context, user, role string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enforcer.HasRoleForUser(user, role)
}

// DeleteRole 删除角色（便捷方法）。
func (e *DefaultCasbinEnforcer) DeleteRole(ctx context.Context, role string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.DeleteRole(role)
}

// DeleteUser 删除用户（便捷方法）。
func (e *DefaultCasbinEnforcer) DeleteUser(ctx context.Context, user string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.DeleteUser(user)
}

// AddPolicies 批量添加策略（便捷方法）。
func (e *DefaultCasbinEnforcer) AddPolicies(ctx context.Context, policies [][]string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.AddPolicies(policies)
}

// RemovePolicies 批量删除策略（便捷方法）。
func (e *DefaultCasbinEnforcer) RemovePolicies(ctx context.Context, policies [][]string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enforcer.RemovePolicies(policies)
}

// GetPermissionsForUser 获取用户的权限（便捷方法）。
func (e *DefaultCasbinEnforcer) GetPermissionsForUser(ctx context.Context, user string, domain ...string) [][]string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	permissions, _ := e.enforcer.GetPermissionsForUser(user, domain...)
	return permissions
}

// HasPermissionForUser 检查用户是否有指定权限（便捷方法）。
func (e *DefaultCasbinEnforcer) HasPermissionForUser(ctx context.Context, user string, permission ...string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	hasPermission, _ := e.enforcer.HasPermissionForUser(user, permission...)
	return hasPermission
}

// ClearPolicy 清空所有策略（便捷方法）。
func (e *DefaultCasbinEnforcer) ClearPolicy(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.enforcer.ClearPolicy()
	return nil
}

// EnableAutoSave 启用或禁用自动保存策略（便捷方法）。
func (e *DefaultCasbinEnforcer) EnableAutoSave(autoSave bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.enforcer != nil {
		e.enforcer.EnableAutoSave(autoSave)
	}
}

// EnableEnforce 启用或禁用强制检查（便捷方法）。
func (e *DefaultCasbinEnforcer) EnableEnforce(enable bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.enforcer != nil {
		e.enforcer.EnableEnforce(enable)
	}
}

// GetSubject 从 Casbin 规则中提取 subject 字段。
func GetSubject(rule []string) string {
	if len(rule) > 0 {
		return rule[0]
	}
	return ""
}

// GetObject 从 Casbin 规则中提取 object 字段。
func GetObject(rule []string) string {
	if len(rule) > 1 {
		return rule[1]
	}
	return ""
}

// GetAction 从 Casbin 规则中提取 action 字段。
func GetAction(rule []string) string {
	if len(rule) > 2 {
		return rule[2]
	}
	return ""
}

// IsRoleRule 检查是否为角色规则。
func IsRoleRule(rule []string) bool {
	return len(rule) > 0 && strings.HasPrefix(rule[0], "g")
}

// IsPolicyRule 检查是否为策略规则。
func IsPolicyRule(rule []string) bool {
	return len(rule) > 0 && strings.HasPrefix(rule[0], "p")
}

// 确保 DefaultCasbinEnforcer 实现了 security.CasbinEnforcer 接口。
var _ security.CasbinEnforcer = (*DefaultCasbinEnforcer)(nil)
