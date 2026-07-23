// Package boot 提供应用启动器功能，用于 enhance 框架。
package boot

import (
	"reflect"
	"sort"
	"sync"

	"github.com/xudefa/enhance/condition"
)

// ==================== AutoConfigEntry 结构体 ====================

// AutoConfigEntry 自动配置条目。
type AutoConfigEntry struct {
	Config         AutoConfiguration     // 自动配置实例
	Conditions     []condition.Condition // 条件列表
	Order          int                   // 执行顺序，值越小优先级越高
	Dependencies   []string              // 依赖的配置名称
	Override       bool                  // 是否为覆盖配置（用户自定义优先）
	OverrideTarget string                // 被覆盖的自动配置类型名
	Before         []string              // 此配置应在哪些配置之前执行
	After          []string              // 此配置应在哪些配置之后执行
}

// AutoConfigurationOption 自动配置选项函数。
type AutoConfigurationOption func(entry *AutoConfigEntry)

// OrderPriority 定义自动配置执行顺序的优先级枚举。
//
// 值越小优先级越高，越先执行。
// 第三方插件集成时应参考此枚举设置 Order 值，确保正确的依赖顺序。
//
// # 执行顺序设计原则
//
//  1. 基础设施优先（日志、配置中心等）
//  2. 数据层（数据库、缓存等）
//  3. 安全层（认证、授权等）
//  4. Web 层（HTTP 服务器、路由等）
//  5. 业务层（定时任务、消息队列等）
//  6. 监控层（Actuator、健康检查等）
//
// # 完整依赖关系图
//
//	基础设施层 (-3000)
//	  └─ 日志 (Zerolog: -3000)
//	  └─ 配置中心
//
//	数据层 (-2000)
//	  └─ 数据库 (GORM: -2000)
//	  └─ 缓存 (Redis)
//	  └─ 对象存储
//
//	认证层 (-1500)
//	  └─ JWT 认证 (JWT: -1500)
//	  └─ OAuth2
//	  └─ LDAP
//
//	授权层 (-1300 ~ -1200)
//	  └─ Casbin GORM (CasbinGorm: -1300) ← 依赖 GORM，提供 GORM 版本的 Enforcer
//	  └─ Casbin 基础 (Casbin: -1200) ← 检测容器中是否有 Enforcer，有则使用，无则创建默认
//
//	安全核心层 (-100)
//	  └─ 安全框架 (Security: -100) ← 依赖认证和授权，构建完整安全体系
//
//	Web 层 (0)
//	  └─ HTTP 服务器
//	  └─ 路由
//	  └─ 中间件
//
//	业务层 (1000)
//	  └─ 定时任务 (Schedule: 1000)
//	  └─ 消息队列
//	  └─ 事件总线
//
//	监控层 (2000)
//	  └─ 指标收集 (Metrics: 2000)
//	  └─ Actuator (Actuator: 2000) ← 最后执行，监控所有组件
//
// # 第三方插件集成指南
//
//   - Redis 缓存: OrderPriorityDataLayer (-2000)
//   - 消息队列: OrderPriorityBusinessLayer (1000)
//   - 对象存储: OrderPriorityDataLayer (-2000)
//   - 搜索引擎: OrderPriorityDataLayer (-2000)
//   - 链路追踪: OrderPriorityMonitoringLayer (2000)
//   - 指标收集: OrderPriorityMonitoringLayer (2000)
type OrderPriority int

const (
	// OrderPriorityInfrastructure 基础设施层优先级 (-3000)
	// 适用于：日志、配置中心、环境变量等基础设施组件
	// 这些组件必须最先初始化，为其他组件提供基础能力
	OrderPriorityInfrastructure OrderPriority = -3000

	// OrderPriorityDataLayer 数据层优先级 (-2000)
	// 适用于：数据库、缓存、对象存储、搜索引擎等数据相关组件
	// 依赖基础设施层，为上层业务提供数据访问能力
	OrderPriorityDataLayer OrderPriority = -2000

	// OrderPriorityServiceDiscovery 服务发现层优先级 (-1800)
	// 适用于：Consul、Nacos、Eureka 等服务发现组件
	// 依赖基础设施层，为其他组件提供服务注册与发现能力
	OrderPriorityServiceDiscovery OrderPriority = -1800

	// OrderPriorityAuthentication 认证层优先级 (-1500)
	// 适用于：JWT、OAuth2、LDAP 等认证相关组件
	// 依赖数据层（可能需要从数据库加载用户信息）
	OrderPriorityAuthentication OrderPriority = -1500

	// OrderPriorityAuthorizationGorm 授权层-GORM 适配器优先级 (-1300)
	// 适用于：CasbinGorm 等基于数据库的授权适配器
	// 依赖数据层，在授权基础配置之前执行，提供数据库版本的 Enforcer
	OrderPriorityAuthorizationGorm OrderPriority = -1300

	// OrderPriorityAuthorization 授权层优先级 (-1200)
	// 适用于：Casbin、RBAC、ABAC 等授权相关组件
	// 依赖认证层和数据层（需要认证信息和权限数据）
	// 会检测容器中是否已有 Enforcer，有则使用，无则创建默认
	OrderPriorityAuthorization OrderPriority = -1200

	// OrderPrioritySecurityCore 安全核心层优先级 (-100)
	// 适用于：安全过滤器链、访问控制、加密等核心安全组件
	// 依赖认证和授权层，构建完整的安全体系
	OrderPrioritySecurityCore OrderPriority = -100

	// OrderPriorityWebLayer Web 层优先级 (0)
	// 适用于：HTTP 服务器、路由、中间件、Web 框架集成等
	// 依赖安全层，确保 Web 请求经过安全过滤
	OrderPriorityWebLayer OrderPriority = 0

	// OrderPriorityMiddleware 中间件层优先级 (100)
	// 适用于：限流器、CORS、压缩等 HTTP 中间件
	// 依赖 Web 层，在 Web 服务器启动前注册中间件
	OrderPriorityMiddleware OrderPriority = 100

	// OrderPriorityBusinessLayer 业务层优先级 (1000)
	// 适用于：定时任务、消息队列、事件总线、业务逻辑等
	// 依赖 Web 层和数据层，实现核心业务功能
	OrderPriorityBusinessLayer OrderPriority = 1000

	// OrderPriorityTaskLayer 任务层优先级 (-500)
	// 适用于：定时任务（Cron）、异步任务队列（Asynq）等任务调度组件
	// 依赖数据层，在业务层之前执行，为业务提供任务调度能力
	OrderPriorityTaskLayer OrderPriority = -500

	// OrderPriorityMonitoringLayer 监控层优先级 (2000)
	// 适用于：Actuator、健康检查、指标收集、链路追踪等
	// 最后初始化，监控所有其他组件的运行状态
	OrderPriorityMonitoringLayer OrderPriority = 2000
)

// WithOrder 设置执行顺序，值越小优先级越高
//
// 推荐使用 OrderPriority 枚举常量，而非直接使用魔法数字：
//
//	boot.RegisterAutoConfigWith(&MyConfig{},
//	    boot.WithOrder(int(boot.OrderPriorityDataLayer)),
//	)
func WithOrder(order int) AutoConfigurationOption {
	return func(entry *AutoConfigEntry) {
		entry.Order = order
	}
}

// WithDependsOn 设置依赖的配置名称
func WithDependsOn(deps ...string) AutoConfigurationOption {
	return func(entry *AutoConfigEntry) {
		entry.Dependencies = append(entry.Dependencies, deps...)
	}
}

// WithConditions 设置条件
func WithConditions(conds ...condition.Condition) AutoConfigurationOption {
	return func(entry *AutoConfigEntry) {
		entry.Conditions = append(entry.Conditions, conds...)
	}
}

// WithOverride 设置覆盖目标，用户自定义配置优先于自动配置
//
// 参数:
//   - target: 被覆盖的自动配置类型名（如 "*GinAutoConfiguration"）
//
// 示例:
//
//	boot.RegisterAutoConfigWith(&MyGinConfig{},
//	    boot.WithOverride("*GinAutoConfiguration"),
//	    boot.WithOrder(-100), // 确保在原始配置之前执行
//	)
func WithOverride(target string) AutoConfigurationOption {
	return func(entry *AutoConfigEntry) {
		entry.Override = true
		entry.OverrideTarget = target
		entry.Order = -100 // 默认高优先级
	}
}

// WithBefore 设置此配置应在哪些配置之前执行
//
// 参数:
//   - configs: 配置类型名列表
//
// 示例:
//
//	boot.RegisterAutoConfigWith(&DatabaseConfig{},
//	    boot.WithBefore("webConfig", "cacheConfig"),
//	)
func WithBefore(configs ...string) AutoConfigurationOption {
	return func(entry *AutoConfigEntry) {
		entry.Before = append(entry.Before, configs...)
	}
}

// WithAfter 设置此配置应在哪些配置之后执行
//
// 参数:
//   - configs: 配置类型名列表
//
// 示例:
//
//	boot.RegisterAutoConfigWith(&WebConfig{},
//	    boot.WithAfter("databaseConfig"),
//	)
func WithAfter(configs ...string) AutoConfigurationOption {
	return func(entry *AutoConfigEntry) {
		entry.After = append(entry.After, configs...)
	}
}

// AutoConfigRegistry 自动配置注册表
type AutoConfigRegistry struct {
	mu      sync.RWMutex
	entries []AutoConfigEntry
}

var globalRegistry = NewAutoConfigRegistry()

// NewAutoConfigRegistry 创建注册表
func NewAutoConfigRegistry() *AutoConfigRegistry {
	return &AutoConfigRegistry{}
}

// RegisterAutoConfig 注册自动配置到全局注册表
//
// 在模块的 init() 中调用：
//
//	func init() {
//	    boot.RegisterAutoConfig(&CircuitAutoConfiguration{},
//	        condition.OnProperty("circuit.enabled", "true"),
//	    )
//	}
func RegisterAutoConfig(config AutoConfiguration, conditions ...condition.Condition) {
	globalRegistry.Add(AutoConfigEntry{
		Config:     config,
		Conditions: conditions,
		Order:      len(globalRegistry.GetAll()),
	})
}

// RegisterAutoConfigWith 注册自动配置到全局注册表（支持选项）
func RegisterAutoConfigWith(config AutoConfiguration, opts ...AutoConfigurationOption) {
	entry := AutoConfigEntry{
		Config: config,
		Order:  len(globalRegistry.GetAll()),
	}
	for _, opt := range opts {
		opt(&entry)
	}
	globalRegistry.Add(entry)
}

// Add 添加自动配置条目
func (r *AutoConfigRegistry) Add(entry AutoConfigEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entry)
}

// GetAll 获取所有注册的自动配置
func (r *AutoConfigRegistry) GetAll() []AutoConfigEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]AutoConfigEntry, len(r.entries))
	copy(result, r.entries)
	return result
}

// GetMatching 获取匹配条件的自动配置（按 Order 排序，支持覆盖机制和 Before/After 排序）
//
// 处理逻辑：
//  1. 收集所有 Override 配置的目标类型
//  2. 过滤被覆盖的配置（跳过被覆盖的自动配置）
//  3. 按 Order、Before、After 排序返回
func (r *AutoConfigRegistry) GetMatching(ctx condition.ConditionContext) []AutoConfigEntry {
	return r.GetMatchingWithExclude(ctx, nil)
}

// GetMatchingWithExclude 获取匹配条件的自动配置，支持排除列表
//
// 参数:
//   - ctx: 条件判断上下文
//   - excluded: 需要排除的自动配置类型名列表
//
// 处理逻辑：
//  1. 收集所有 Override 配置的目标类型
//  2. 过滤被覆盖的配置（跳过被覆盖的自动配置）
//  3. 过滤被排除的配置
//  4. 按 Order、Before、After 排序返回
func (r *AutoConfigRegistry) GetMatchingWithExclude(ctx condition.ConditionContext, excluded []string) []AutoConfigEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 构建排除集合
	excludedSet := make(map[string]bool)
	for _, name := range excluded {
		excludedSet[name] = true
	}

	// 第一遍：收集覆盖目标
	overrideTargets := make(map[string]bool)
	for _, entry := range r.entries {
		if entry.Override && entry.OverrideTarget != "" {
			overrideTargets[entry.OverrideTarget] = true
		}
	}

	// 第二遍：过滤并匹配配置
	matched := make([]AutoConfigEntry, 0, len(r.entries))
	for _, entry := range r.entries {
		// 检查是否被覆盖
		configType := reflect.TypeOf(entry.Config).String()
		if overrideTargets[configType] {
			continue // 被覆盖的配置跳过
		}

		// 检查是否被排除
		if excludedSet[configType] {
			continue // 被排除的配置跳过
		}

		// 检查条件是否匹配
		if r.matchesAll(ctx, entry.Conditions) {
			matched = append(matched, entry)
		}
	}

	// 按 Order、Before、After 排序
	r.sortWithDependencies(matched)
	return matched
}

// sortWithDependencies 根据 Order、Before、After 进行排序
// 使用拓扑排序处理 Before/After 依赖关系
func (r *AutoConfigRegistry) sortWithDependencies(entries []AutoConfigEntry) {
	if len(entries) == 0 {
		return
	}

	// 构建配置类型名到索引的映射
	typeNameToIndex := make(map[string]int)
	for i, entry := range entries {
		typeName := reflect.TypeOf(entry.Config).String()
		typeNameToIndex[typeName] = i
	}

	// 构建依赖图：如果 A 应该在 B 之前，则 A -> B
	// inDegree[B]++ 表示 B 依赖 A
	inDegree := make(map[int]int)
	adjList := make(map[int][]int)

	for i, entry := range entries {
		if _, ok := inDegree[i]; !ok {
			inDegree[i] = 0
		}

		// 处理 Before：当前配置应该在指定配置之前执行
		for _, beforeTarget := range entry.Before {
			if targetIdx, exists := typeNameToIndex[beforeTarget]; exists {
				adjList[i] = append(adjList[i], targetIdx)
				inDegree[targetIdx]++
			}
		}

		// 处理 After：当前配置应该在指定配置之后执行
		for _, afterTarget := range entry.After {
			if targetIdx, exists := typeNameToIndex[afterTarget]; exists {
				adjList[targetIdx] = append(adjList[targetIdx], i)
				inDegree[i]++
			}
		}
	}

	// Kahn 算法拓扑排序
	// 使用优先队列，当有多个节点入度为0时，按 Order 排序（Order 小的优先）
	queue := make([]int, 0)
	for i := range entries {
		if inDegree[i] == 0 {
			queue = append(queue, i)
		}
	}

	// 对初始队列按 Order 排序
	sort.Slice(queue, func(i, j int) bool {
		return entries[queue[i]].Order < entries[queue[j]].Order
	})

	result := make([]int, 0, len(entries))
	for len(queue) > 0 {
		// 从队列中取出节点
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		// 减少相邻节点的入度
		newReady := make([]int, 0)
		for _, neighbor := range adjList[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				newReady = append(newReady, neighbor)
			}
		}

		// 对新就绪的节点按 Order 排序并加入队列
		if len(newReady) > 0 {
			sort.Slice(newReady, func(i, j int) bool {
				return entries[newReady[i]].Order < entries[newReady[j]].Order
			})
			queue = append(queue, newReady...)
		}
	}

	// 如果存在循环依赖，回退到按 Order 排序
	if len(result) != len(entries) {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Order < entries[j].Order
		})
		return
	}

	// 根据拓扑排序结果重新排列 entries
	sorted := make([]AutoConfigEntry, len(entries))
	for i, idx := range result {
		sorted[i] = entries[idx]
	}
	copy(entries, sorted)
}

func (r *AutoConfigRegistry) matchesAll(ctx condition.ConditionContext, conditions []condition.Condition) bool {
	if len(conditions) == 0 {
		return true
	}
	all := condition.All(conditions...)
	return all.Matches(ctx)
}

// GlobalRegistry 返回全局注册表
func GlobalRegistry() *AutoConfigRegistry {
	return globalRegistry
}
