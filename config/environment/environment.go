package environment

import (
	"os"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/xudefa/enhance/config/refresh"
)

// Environment 环境配置管理器
//
// 参考 Spring 的 Environment 抽象，提供分层配置源管理和 Profile 机制。
// 配置源按优先级排序，高优先级覆盖低优先级。
// 支持命令行参数（最高优先级）> 环境变量 > 配置文件（最低优先级）。
type Environment struct {
	mu              sync.RWMutex                      // 保护所有字段的读写锁
	sources         []PropertySource                  // 配置源列表，按优先级升序排列
	activeProfiles  []string                          // 当前激活的 Profile 列表
	configListeners []func(refresh.ConfigChangeEvent) // 配置变更监听器
}

// NewEnvironment 创建环境配置管理器
//
// 默认配置源优先级（从高到低）：
//  1. 命令行参数（--key=value）
//  2. 环境变量（GO_BOOT_ 前缀）
//  3. 应用 JSON 配置文件（application.json）
func NewEnvironment() *Environment {
	args := NewArgsPropertySource("args", os.Args)
	envSource := NewEnvPropertySource("env", "GO_BOOT")

	sources := make([]PropertySource, 2, 5)
	sources[0] = args
	sources[1] = envSource

	// 尝试加载应用 JSON 配置文件
	if applicationConfigFile := FindApplicationConfigFile(); applicationConfigFile != "" {
		if jsonSource := NewJSONPropertySourceOrDefault("application-config", applicationConfigFile); jsonSource != nil {
			sources = append(sources, jsonSource)
		}
	}

	environment := &Environment{
		sources: sources,
	}

	// 排序配置源以确保正确的优先级顺序
	environment.sortSources()

	return environment
}

// AddPropertySource 添加配置源到环境.
//
// 新添加的配置源会按优先级排序,高优先级覆盖低优先级.
//
// 参数:
//   - source: 配置源实例
func (e *Environment) AddPropertySource(source PropertySource) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sources = append(e.sources, source)
	e.sortSourcesIfNeeded()
}

// AddPropertySourceFirst 添加最高优先级的配置源.
//
// 新添加的配置源会插入到配置源列表头部,成为最高优先级的来源.
//
// 参数:
//   - source: 配置源实例
func (e *Environment) AddPropertySourceFirst(source PropertySource) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sources = append([]PropertySource{source}, e.sources...)
	e.sortSourcesIfNeeded()
}

// GetPropertySources 获取所有配置源列表.
//
// 返回按优先级排序的配置源副本,高优先级在后.
//
// 返回:
//   - []PropertySource: 配置源列表
func (e *Environment) GetPropertySources() []PropertySource {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := make([]PropertySource, len(e.sources))
	copy(result, e.sources)
	return result
}

// GetActiveProfiles 获取当前激活的 Profile 列表.
//
// 返回:
//   - []string: Profile 名称列表
func (e *Environment) GetActiveProfiles() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := make([]string, len(e.activeProfiles))
	copy(result, e.activeProfiles)
	return result
}

// AddActiveProfile 激活指定 Profile.
//
// 如果该 Profile 已经激活,则忽略.
//
// 参数:
//   - profile: Profile 名称
func (e *Environment) AddActiveProfile(profile string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if slices.Contains(e.activeProfiles, profile) {
		return
	}
	e.activeProfiles = append(e.activeProfiles, profile)
}

// AcceptsProfile 检查指定 Profile 是否被当前环境接受
//
// 支持否定前缀 "!"，如 "!dev" 表示非 dev 环境时匹配。
// 检查流程：
//  1. 检查是否有否定前缀
//  2. 在激活的 Profile 列表中查找
//  3. 含否定前缀时：Profile 不在激活列表中返回 true
//  4. 无否定前缀时：Profile 在激活列表中返回 true
func (e *Environment) AcceptsProfile(profile string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	negate := false
	if strings.HasPrefix(profile, "!") {
		negate = true
		profile = profile[1:]
	}
	if slices.Contains(e.activeProfiles, profile) {
		return !negate
	}
	return negate
}

// RemovePropertySource 按名称移除配置源
func (e *Environment) RemovePropertySource(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, src := range e.sources {
		if src.Name() == name {
			e.sources = append(e.sources[:i], e.sources[i+1:]...)
			return
		}
	}
}

// RemoveProfile 移除激活的 Profile
func (e *Environment) RemoveProfile(profile string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, p := range e.activeProfiles {
		if p == profile {
			e.activeProfiles = append(e.activeProfiles[:i], e.activeProfiles[i+1:]...)
			return
		}
	}
}

func (e *Environment) sortSources() {
	sort.Slice(e.sources, func(i, j int) bool {
		return e.sources[i].Priority() < e.sources[j].Priority()
	})
}

// sortSourcesIfNeeded 仅在配置源未排序时才执行排序
//
// 优化：如果新添加的配置源优先级高于所有已有配置源，
// 则无需全量排序，因为配置源列表已经基本有序。
func (e *Environment) sortSourcesIfNeeded() {
	if len(e.sources) <= 1 {
		return
	}

	// 检查是否已经按优先级升序排列
	alreadySorted := true
	for i := 1; i < len(e.sources); i++ {
		if e.sources[i-1].Priority() > e.sources[i].Priority() {
			alreadySorted = false
			break
		}
	}

	if !alreadySorted {
		sort.Slice(e.sources, func(i, j int) bool {
			return e.sources[i].Priority() < e.sources[j].Priority()
		})
	}
}

// AddConfigChangeListener 添加配置变更监听器
func (e *Environment) AddConfigChangeListener(listener func(refresh.ConfigChangeEvent)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.configListeners = append(e.configListeners, listener)
}

// notifyConfigChange 通知所有配置变更监听器
func (e *Environment) notifyConfigChange(event refresh.ConfigChangeEvent) {
	e.mu.RLock()
	listeners := make([]func(refresh.ConfigChangeEvent), len(e.configListeners))
	copy(listeners, e.configListeners)
	e.mu.RUnlock()

	for _, listener := range listeners {
		go listener(event) // 异步通知
	}
}
