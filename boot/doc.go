// Package boot 提供应用启动器功能，用于 enhance 框架。
//
// 该模块负责应用的生命周期管理、自动配置执行、组件扫描和注册等核心功能。
// 参考 Spring Boot 的 SpringApplication 设计。
//
// # 架构设计
//
//   - AutoConfiguration: 自动配置接口，支持条件化配置
//   - Starter: 启动器接口，支持模块化启动
//   - FailureAnalyzer: 失败分析器接口，提供友好的错误提示
//   - Banner: 启动横幅接口，支持多种格式的启动横幅显示
//   - Module: 可组合的配置单元，包含 Bean 注册和 Starter
//   - BeanProvider: Bean 提供者函数类型
//   - ApplicationContext: 自动配置看到的上下文接口
//   - OrderPriority: 执行顺序优先级枚举，用于规范自动配置的执行顺序
//
// # 核心功能
//
//   - 自动配置执行（AutoConfiguration）
//   - 组件扫描和注册
//   - 生命周期管理
//   - 启动横幅（Banner）显示
//   - 启动失败分析
//   - 优雅关闭支持
//
// # 执行顺序设计
//
// 自动配置的执行顺序通过 OrderPriority 枚举定义，值越小优先级越高：
//
//   - OrderPriorityInfrastructure (-3000): 基础设施层（日志、配置中心等）
//   - OrderPriorityDataLayer (-2000): 数据层（数据库、缓存等）
//   - OrderPriorityAuthentication (-1500): 认证层（JWT、OAuth2 等）
//   - OrderPriorityAuthorizationGorm (-1300): 授权层-GORM 适配器（CasbinGorm 等）
//   - OrderPriorityAuthorization (-1200): 授权层（Casbin、RBAC 等）
//   - OrderPrioritySecurityCore (-100): 安全核心层（过滤器链、访问控制等）
//   - OrderPriorityWebLayer (0): Web 层（HTTP 服务器、路由等）
//   - OrderPriorityBusinessLayer (1000): 业务层（定时任务、消息队列等）
//   - OrderPriorityMonitoringLayer (2000): 监控层（Actuator、健康检查等）
//
// 依赖关系示例：
//
//	日志(-3000) → 数据库(-2000) → 认证(-1500) → CasbinGorm(-1300) → Casbin(-1200) → 安全(-100) → Web(0) → 业务(1000) → 监控(2000)
//
// # 使用方式
//
// 创建应用实例：
//
//	app := boot.NewApplication(
//	    boot.WithName("my-app"),
//	    boot.WithWeb(true),
//	    boot.WithPort(8080),
//	)
//	app.Run()
//
// # 自动配置
//
// 自动配置类通过实现 boot.AutoConfiguration 接口，在应用启动时自动执行：
//
//	type MyAutoConfiguration struct{}
//
//	func (m *MyAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
//	    // 配置逻辑
//	    return nil
//	}
//
//	func init() {
//	    boot.RegisterAutoConfig(&MyAutoConfiguration{})
//	}
//
// # 配置选项
//
//   - WithName: 设置应用名称
//   - WithWeb: 是否启用 Web 模式
//   - WithPort: 设置服务端口
//   - WithBanner: 设置启动横幅
package boot

import (
	"reflect"
	"sync"
	"time"

	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/event"
	"github.com/xudefa/enhance/lifecycle"
)

// ApplicationContext 自动配置看到的上下文接口。
//
// 实际由 context.DefaultApplicationContext 实现。
// 这是 context.ApplicationContext 的子集，仅包含自动配置需要的方法。
type ApplicationContext interface {
	// Container 返回 IoC 容器实例。
	Container() core.Container

	// Environment 返回环境配置实例。
	Environment() *environment.Environment

	// Register 在容器中注册 Bean。
	Register(t reflect.Type, opts ...core.BeanOption) error

	// GetByType 从容器中获取指定类型的 Bean。
	GetByType(t reflect.Type) (any, error)

	// EventBus 返回事件总线访问接口。
	EventBus() EventBusResult
}

// AutoConfiguration 自动配置接口。
//
// 参考 Spring Boot 的 @Configuration + @Bean 模式。
// 每个模块实现此接口，通过 RegisterAutoConfig 注册。
type AutoConfiguration interface {
	// Configure 执行自动配置逻辑。
	Configure(ctx ApplicationContext) error
}

// Starter 应用启动器接口。
//
// 参考 Spring Boot 的 ApplicationRunner/CommandLineRunner。
// 每个集成的 Starter 管理其自身的生命周期，包括配置、启动和停止。
//
// 生命周期：
//   - Configure: 在配置阶段调用，用于注册 Bean 和设置依赖
//   - Start: 在就绪阶段调用，启动服务（如 HTTP 服务器）
//   - Stop: 在停止阶段调用，释放资源（逆序执行）
type Starter interface {
	// Name 返回启动器名称，用于依赖排序和日志输出。
	Name() string

	// Dependencies 返回依赖的其他启动器名称。
	// 启动器会按依赖关系拓扑排序后依次启动。
	Dependencies() []string

	// Configure 配置阶段调用，用于注册 Bean 和设置依赖。
	Configure(ctx ApplicationContext) error

	// Start 启动阶段调用，启动服务。
	Start(ctx ApplicationContext) error

	// Stop 停止阶段调用，释放资源。
	Stop(ctx ApplicationContext) error

	// GetCondition 返回启动条件，nil 表示始终启动。
	GetCondition() condition.Condition
}

// FailureAnalyzer 失败分析器接口。
//
// 参考 Spring Boot 的 FailureAnalyzer。
// 在应用启动失败时提供友好的错误提示，帮助开发者快速定位问题。
type FailureAnalyzer interface {
	// CanAnalyze 检查是否能分析该错误。
	CanAnalyze(err error) bool

	// Analyze 分析错误并返回失败报告。
	Analyze(err error) *FailureReport
}

// Banner 启动横幅接口。
//
// 参考 Spring Boot 的 Banner，支持多种格式的启动横幅显示。
type Banner interface {
	// Print 输出启动横幅。
	Print(ctx ApplicationContext)
}

type EventBusResult interface {
	Publish(event event.ApplicationEvent)
}

// BeanProvider Bean 提供者函数类型。
//
// 定义如何向容器注册 Bean，是 Module 中 Bean 注册的统一抽象。
type BeanProvider func(c core.Container) error

// FailureReport 失败报告。
//
// 参考 Spring Boot 的 FailureAnalysis，在应用启动失败时提供结构化的错误信息。
// 包含错误描述、建议动作、根因和可能的解决方案。
type FailureReport struct {
	Headline          string         `json:"headline"`                    // 报告标题
	Description       string         `json:"description"`                 // 错误描述
	Action            string         `json:"action"`                      // 建议动作
	Cause             string         `json:"cause"`                       // 根因
	Details           map[string]any `json:"details,omitempty"`           // 附加详情
	StackTrace        string         `json:"stackTrace,omitempty"`        // 堆栈跟踪
	PossibleSolutions []string       `json:"possibleSolutions,omitempty"` // 可能的解决方案列表
}

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

// BootConfig 启动配置。
type BootConfig struct {
	ConfigLocation string   // 配置文件路径
	ConfigType     string   // 配置文件类型 (json)
	Profiles       []string // 激活的 Profile
	AppName        string   // 应用名称
	Version        string   // 版本号

	AutoExecute bool // 是否自动执行自动配置（默认 true）
	Starters    bool // 是否自动管理启动器生命周期（默认 true）

	// 排除的自动配置列表（按类型名匹配）
	ExcludedAutoConfigs []string // 需要排除的自动配置类型名

	CustomPropertySources []environment.PropertySource // 用户自定义配置源

	// 配置中心配置
	ConfigCenterEnabled bool          // 是否启用配置中心（默认 false）
	ConfigCenterType    string        // 配置中心类型 (nacos/etcd/consul)
	ConfigCenterAddr    []string      // 配置中心地址
	ConfigCenterDataID  string        // 配置中心数据ID
	ConfigCenterGroup   string        // 配置中心分组
	ConfigCenterPrefix  string        // 配置中心前缀
	ConfigCenterTimeout time.Duration // 配置中心超时时间

	// 显式模块（Go 风格组合，替代全局 init() 注册）
	Modules []Module // 用户显式传入的模块列表

	// 生命周期钩子（Go 风格 3 阶段：OnInit/OnStart/OnStop）
	Hooks []lifecycle.Hook // 用户注册的生命周期钩子
}

// BootOption 启动选项函数。
type BootOption func(*BootConfig)

// Module 可组合的配置单元。
//
// Module 是 Go 风格的显式组合方式，替代全局 init() 注册。
// 每个 Module 可以独立测试、独立复用。
type Module struct {
	moduleName string                       // 模块名称，用于日志和调试
	beans      []BeanProvider               // 要注册的 Bean 提供者列表
	invokes    []func(core.Container) error // 安装时立即调用的函数列表
	hooks      []lifecycle.Hook             // 生命周期钩子列表
	starters   []Starter                    // 要启动的 Starter 列表
	conditions []condition.Condition        // 模块生效的条件
}

// StarterRegistry 启动器注册表。
//
// 管理所有 Starter 的注册和依赖排序。
// 使用 Kahn 算法进行拓扑排序，确保依赖的启动器先启动。
type StarterRegistry struct {
	mu       sync.RWMutex
	starters []Starter
}

// FailureAnalyzerRegistry 失败分析器注册表。
//
// 管理所有 FailureAnalyzer 的注册和查询。
// 分析时按注册顺序遍历，返回第一个匹配的失败报告。
type FailureAnalyzerRegistry struct {
	mu        sync.RWMutex
	analyzers []FailureAnalyzer
}

// SimpleFailureAnalyzer 简单的失败分析器。
//
// 通过传入的分析函数创建分析器，适用于简单的错误分析场景。
type SimpleFailureAnalyzer struct {
	analyzeFn func(err error) *FailureReport
}

// BootError 结构化启动错误。
//
// 包含错误发生阶段、原始错误、分析结果和修复建议，便于调试和错误处理。
//
// 设计模式: Adapter（适配原始错误为结构化格式）
type BootError struct {
	Phase       string   // 错误发生的阶段
	Original    error    // 原始错误
	Analyzed    string   // FailureAnalyzer 分析结果
	Suggestions []string // 修复建议
}

// TextBanner 文本横幅。
//
// 使用模板和属性渲染启动横幅。
type TextBanner struct {
	Template   string         // 横幅文本模板
	Properties map[string]any // 模板属性键值对
}

// ASCIIArtBanner ASCII 艺术横幅。
type ASCIIArtBanner struct {
	Art   string // ASCII 艺术文本
	Color string // 显示颜色（预留）
}

// CustomTemplateBanner 自定义模板横幅。
//
// 支持从文件加载横幅模板。
type CustomTemplateBanner struct {
	TemplatePath string         // 模板文件路径
	Data         map[string]any // 模板数据
}

// LegacyBanner 旧版横幅实现。
//
// 使用预定义的 ASCII 艺术行列表渲染启动横幅。
type LegacyBanner struct {
	lines []string // ASCII 艺术行列表
}

// 以下类型定义在其他文件中，此处仅作文档说明：
// - ApplicationOption: application.go（包含完整实现）
// - Boot: boot.go（包含完整实现）
// - StarterManager: starter_manager.go（包含完整实现）
// - ConfigCenterProvider: configcenter.go（包含完整实现）
// - ConfigCenterAdapter: configcenter.go（包含完整实现）
// - FailureAnalyzerAdapter: failure_analyzers.go（包含完整实现）
// - Report: report.go（包含完整实现）
