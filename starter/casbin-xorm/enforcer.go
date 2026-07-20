package casbinxorm

import (
	"context"

	"github.com/casbin/casbin/v2"
	xormadapter "github.com/casbin/xorm-adapter/v2"
	"github.com/xudefa/enhance/security"
)

// XormCasbinEnforcer 基于 XORM 的 Casbin 执行器实现。
//
// 封装官方 casbin.Enforcer 和 xorm-adapter，提供基于数据库的权限管理能力。
// 实现了 security.CasbinEnforcer 接口，策略持久化到数据库中。
type XormCasbinEnforcer struct {
	*casbin.Enforcer
	adapter *xormadapter.Adapter
}

// Enforce 检查请求是否允许。
func (e *XormCasbinEnforcer) Enforce(ctx context.Context, subject, object, action string) (bool, error) {
	return e.Enforcer.Enforce(subject, object, action)
}

// AddPolicy 添加策略。
func (e *XormCasbinEnforcer) AddPolicy(ctx context.Context, sub, obj, act string) error {
	_, err := e.Enforcer.AddPolicy(sub, obj, act)
	return err
}

// RemovePolicy 移除策略。
func (e *XormCasbinEnforcer) RemovePolicy(ctx context.Context, sub, obj, act string) error {
	_, err := e.Enforcer.RemovePolicy(sub, obj, act)
	return err
}

// GetPolicy 获取所有策略。
func (e *XormCasbinEnforcer) GetPolicy(ctx context.Context) ([][]string, error) {
	return e.Enforcer.GetPolicy()
}

// LoadPolicy 重新加载策略。
func (e *XormCasbinEnforcer) LoadPolicy(ctx context.Context) error {
	return e.Enforcer.LoadPolicy()
}

// SavePolicy 保存策略。
func (e *XormCasbinEnforcer) SavePolicy(ctx context.Context) error {
	return e.Enforcer.SavePolicy()
}

// 确保 XormCasbinEnforcer 实现了 CasbinEnforcer 接口。
var _ security.CasbinEnforcer = (*XormCasbinEnforcer)(nil)
