package boot

import (
	"context"

	"github.com/xudefa/enhance/config/environment"
	contextpkg "github.com/xudefa/enhance/context"
	"github.com/xudefa/enhance/core"
)

// pluginAppCtx 实现 PluginContext 接口，桥接 ApplicationContext 和 PluginContext。
type pluginAppCtx struct {
	ctx     *contextpkg.DefaultApplicationContext
	rootCtx context.Context
	manager *PluginManager
}

// newPluginAppCtx 创建插件应用上下文
func newPluginAppCtx(ctx *contextpkg.DefaultApplicationContext, rootCtx context.Context) PluginContext {
	return &pluginAppCtx{
		ctx:     ctx,
		rootCtx: rootCtx,
		manager: GetPluginManager(),
	}
}

// Container 返回 IoC 容器
func (p *pluginAppCtx) Container() core.Container {
	return p.ctx.Container()
}

// Environment 返回环境配置
func (p *pluginAppCtx) Environment() *environment.Environment {
	return p.ctx.Environment()
}

// GetPlugin 获取指定名称的插件实例
func (p *pluginAppCtx) GetPlugin(name string) (Plugin, bool) {
	if p.manager != nil {
		return p.manager.Get(name)
	}
	return nil, false
}

// Config 获取配置值
func (p *pluginAppCtx) Config(key string) (any, bool) {
	return p.ctx.Environment().GetProperty(key)
}
