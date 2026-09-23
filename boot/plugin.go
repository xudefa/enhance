package boot

import (
	"fmt"
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
