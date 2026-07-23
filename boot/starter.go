// Package boot 提供应用启动器功能，用于 enhance 框架。
package boot

import (
	"sync"
	"sync/atomic"
)

var globalStarterRegistry atomic.Pointer[starterRegistryImpl]

func init() {
	globalStarterRegistry.Store(newStarterRegistryImpl())
}

// newStarterRegistryImpl 创建 StarterRegistry 接口的内部实现。
func newStarterRegistryImpl() *starterRegistryImpl {
	return &starterRegistryImpl{}
}

// NewStarterRegistry 创建启动器注册表。
//
// 返回 StarterRegistry 接口，隐藏实现细节。
func NewStarterRegistry() StarterRegistry {
	return newStarterRegistryImpl()
}

// starterRegistryImpl StarterRegistry 接口的具体实现。
//
// 使用互斥锁保证并发安全，使用 Kahn 算法进行拓扑排序。
type starterRegistryImpl struct {
	mu       sync.RWMutex
	starters []Starter
}

// Register 注册一个 Starter 到注册表。
func (r *starterRegistryImpl) Register(starter Starter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.starters = append(r.starters, starter)
}

// Add 添加启动器到注册表（向后兼容别名）。
func (r *starterRegistryImpl) Add(starter Starter) {
	r.Register(starter)
}

// Get 根据名称获取已注册的 Starter，未找到返回 nil。
func (r *starterRegistryImpl) Get(name string) Starter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.starters {
		if s.Name() == name {
			return s
		}
	}
	return nil
}

// GetAll 获取所有注册的启动器（返回副本）。
func (r *starterRegistryImpl) GetAll() []Starter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Starter, len(r.starters))
	copy(result, r.starters)
	return result
}

// GetOrdered 按依赖关系拓扑排序获取启动器。
//
// 使用 Kahn 算法进行拓扑排序：
//  1. 构建依赖关系图和入度表
//  2. 将入度为 0 的启动器加入队列
//  3. 依次出队并减少依赖其的启动器的入度
//  4. 如果存在循环依赖，回退到原始注册顺序
func (r *starterRegistryImpl) GetOrdered() []Starter {
	r.mu.RLock()
	starters := make([]Starter, len(r.starters))
	copy(starters, r.starters)
	r.mu.RUnlock()

	nameMap := make(map[string]Starter, len(starters))
	for _, s := range starters {
		nameMap[s.Name()] = s
	}

	inDegree := make(map[string]int)
	depMap := make(map[string][]string)
	for _, s := range starters {
		name := s.Name()
		if _, ok := inDegree[name]; !ok {
			inDegree[name] = 0
		}
		for _, dep := range s.Dependencies() {
			if _, exists := nameMap[dep]; exists {
				depMap[dep] = append(depMap[dep], name)
				inDegree[name]++
			}
		}
	}

	queue := make([]string, 0)
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	result := make([]Starter, 0, len(starters))
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		result = append(result, nameMap[name])
		for _, dep := range depMap[name] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if len(result) != len(starters) {
		return starters
	}

	return result
}

// RegisterStarter 注册启动器到全局注册表。
//
// 在模块的 init() 中调用：
//
//	func init() {
//	    boot.RegisterStarter(&MyStarter{})
//	}
func RegisterStarter(starter Starter) {
	globalStarterRegistry.Load().Register(starter)
}

// GlobalStarterRegistry 返回全局启动器注册表。
func GlobalStarterRegistry() StarterRegistry {
	return globalStarterRegistry.Load()
}
