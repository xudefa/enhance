package boot

import (
	"fmt"
	"sync"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
)

// Plugin 插件接口，定义插件的生命周期和元信息。
//
// 插件是比 Starter 更高级的抽象，提供独立的生命周期管理和依赖管理。
// 每个插件可以独立启用/禁用，并声明对其他插件的依赖。
//
// 示例:
//
//	type MyPlugin struct{}
//
//	func (p *MyPlugin) Name() string { return "my-plugin" }
//	func (p *MyPlugin) Version() string { return "1.0.0" }
//	func (p *MyPlugin) Dependencies() []string { return []string{"database", "cache"} }
//	func (p *MyPlugin) Init(ctx PluginContext) error { /* 初始化 */ return nil }
//	func (p *MyPlugin) Start(ctx PluginContext) error { /* 启动 */ return nil }
//	func (p *MyPlugin) Stop(ctx PluginContext) error { /* 停止 */ return nil }
type Plugin interface {
	// Name 返回插件名称
	Name() string
	// Version 返回插件版本
	Version() string
	// Dependencies 返回依赖的其他插件名称列表
	Dependencies() []string
	// Init 初始化插件，在依赖解析后调用
	Init(ctx PluginContext) error
	// Start 启动插件，在所有插件 Init 完成后调用
	Start(ctx PluginContext) error
	// Stop 停止插件，在应用关闭时调用
	Stop(ctx PluginContext) error
}

// PluginContext 插件上下文，提供插件运行时的环境信息。
type PluginContext interface {
	// Container 返回 IoC 容器
	Container() core.Container
	// Environment 返回环境配置
	Environment() *environment.Environment
	// GetPlugin 获取指定名称的插件实例
	GetPlugin(name string) (Plugin, bool)
	// Config 获取配置值
	Config(key string) (any, bool)
}

// PluginState 插件状态
type PluginState int

const (
	// PluginStateRegistered 插件已注册
	PluginStateRegistered PluginState = iota
	// PluginStateInitialized 插件已初始化
	PluginStateInitialized
	// PluginStateStarted 插件已启动
	PluginStateStarted
	// PluginStateStopped 插件已停止
	PluginStateStopped
	// PluginStateError 插件错误
	PluginStateError
)

// String 返回插件状态的字符串表示。
func (s PluginState) String() string {
	switch s {
	case PluginStateRegistered:
		return "Registered"
	case PluginStateInitialized:
		return "Initialized"
	case PluginStateStarted:
		return "Started"
	case PluginStateStopped:
		return "Stopped"
	case PluginStateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// PluginInfo 插件信息
type PluginInfo struct {
	Name    string
	Version string
	State   PluginState
}

// PluginManager 插件管理器，负责插件的注册、依赖解析和生命周期管理。
type PluginManager struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
	infos   map[string]*PluginInfo
	ctx     PluginContext
}

// NewPluginManager 创建插件管理器
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
		infos:   make(map[string]*PluginInfo),
	}
}

// Register 注册插件
func (pm *PluginManager) Register(plugin Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	name := plugin.Name()
	if _, exists := pm.plugins[name]; exists {
		return fmt.Errorf("plugin %q already registered", name)
	}

	pm.plugins[name] = plugin
	pm.infos[name] = &PluginInfo{
		Name:    name,
		Version: plugin.Version(),
		State:   PluginStateRegistered,
	}
	return nil
}

// Get 获取插件
func (pm *PluginManager) Get(name string) (Plugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	plugin, ok := pm.plugins[name]
	return plugin, ok
}

// List 列出所有插件信息
func (pm *PluginManager) List() []PluginInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	infos := make([]PluginInfo, 0, len(pm.infos))
	for _, info := range pm.infos {
		infos = append(infos, *info)
	}
	return infos
}

// SetContext 设置插件上下文
func (pm *PluginManager) SetContext(ctx PluginContext) {
	pm.ctx = ctx
}

// InitAll 初始化所有已注册的插件（按依赖顺序）
func (pm *PluginManager) InitAll() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	ordered, err := pm.resolveDependencies()
	if err != nil {
		return fmt.Errorf("解析插件依赖关系失败: %w", err)
	}

	for _, name := range ordered {
		plugin := pm.plugins[name]
		pluginInfo := pm.infos[name]

		if pluginInfo.State != PluginStateRegistered {
			continue
		}

		if err := plugin.Init(pm.ctx); err != nil {
			pluginInfo.State = PluginStateError
			return fmt.Errorf("plugin %q init failed: %w", name, err)
		}
		pluginInfo.State = PluginStateInitialized
	}
	return nil
}

// StartAll 启动所有已初始化的插件（按依赖顺序）
func (pm *PluginManager) StartAll() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	ordered, err := pm.resolveDependencies()
	if err != nil {
		return fmt.Errorf("解析插件依赖关系失败: %w", err)
	}

	for _, name := range ordered {
		plugin := pm.plugins[name]
		pluginInfo := pm.infos[name]

		if pluginInfo.State != PluginStateInitialized {
			continue
		}

		if err := plugin.Start(pm.ctx); err != nil {
			pluginInfo.State = PluginStateError
			return fmt.Errorf("plugin %q start failed: %w", name, err)
		}
		pluginInfo.State = PluginStateStarted
	}
	return nil
}

// StopAll 停止所有已启动的插件（逆序）
func (pm *PluginManager) StopAll() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	ordered, err := pm.resolveDependencies()
	if err != nil {
		return fmt.Errorf("解析插件依赖关系失败: %w", err)
	}

	// 逆序停止
	for i := len(ordered) - 1; i >= 0; i-- {
		name := ordered[i]
		plugin := pm.plugins[name]
		pluginInfo := pm.infos[name]

		if pluginInfo.State != PluginStateStarted {
			continue
		}

		if err := plugin.Stop(pm.ctx); err != nil {
			pluginInfo.State = PluginStateError
			return fmt.Errorf("plugin %q stop failed: %w", name, err)
		}
		pluginInfo.State = PluginStateStopped
	}
	return nil
}

// resolveDependencies 解析插件依赖顺序（拓扑排序）
func (pm *PluginManager) resolveDependencies() ([]string, error) {
	// 构建依赖图
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for name, plugin := range pm.plugins {
		if _, exists := inDegree[name]; !exists {
			inDegree[name] = 0
		}
		for _, dep := range plugin.Dependencies() {
			graph[dep] = append(graph[dep], name)
			inDegree[name]++
			if _, exists := inDegree[dep]; !exists {
				inDegree[dep] = 0
			}
		}
	}

	// Kahn 算法拓扑排序
	queue := make([]string, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	ordered := make([]string, 0)
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		ordered = append(ordered, node)

		for _, neighbor := range graph[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(ordered) != len(pm.plugins) {
		return nil, fmt.Errorf("circular dependency detected among plugins")
	}

	return ordered, nil
}

// globalPluginManager 全局插件管理器
var globalPluginManager = NewPluginManager()

// GetPluginManager 获取全局插件管理器
func GetPluginManager() *PluginManager {
	return globalPluginManager
}

// RegisterPlugin 便捷函数，注册插件到全局管理器
func RegisterPlugin(plugin Plugin) error {
	return globalPluginManager.Register(plugin)
}

// WithPlugin 添加插件到 Boot 配置
func WithPlugin(plugin Plugin) BootOption {
	return func(cfg *BootConfig) {
		cfg.Plugins = append(cfg.Plugins, plugin)
	}
}

// WithPlugins 添加多个插件到 Boot 配置
func WithPlugins(plugins ...Plugin) BootOption {
	return func(cfg *BootConfig) {
		cfg.Plugins = append(cfg.Plugins, plugins...)
	}
}
