// Package boot 提供应用启动器功能，用于 enhance 框架。
//
// 该模块负责应用的生命周期管理、自动配置执行、组件扫描和注册等核心功能。
// 参考 Spring Boot 的 SpringApplication 设计。
//
// # 架构设计
//
//   - Application: 应用接口，管理完整的应用生命周期
//   - AutoConfiguration: 自动配置接口，支持条件化配置
//   - Starter: 启动器接口，支持模块化启动
//   - StarterRegistry: 启动器注册表接口，管理 Starter 的注册和依赖排序
//   - BootError: 结构化启动错误接口，提供错误码和错误信息
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
//	app, err := boot.NewApplication(
//	    boot.WithAppName("my-app"),
//	    boot.WithVersion("1.0.0"),
//	    boot.WithProfiles("dev"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	app.Start()
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
//   - WithAppName: 设置应用名称
//   - WithVersion: 设置版本号
//   - WithProfiles: 设置激活的 Profile
//   - WithConfigLocation: 设置配置文件路径
//   - WithProperty: 添加单个配置属性
package boot

import (
	"context"
	"reflect"

	"github.com/xudefa/enhance/boot/banner"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/event"
)

// 向后兼容的类型别名，重新导出 banner 子包中的类型。
type BannerMode = banner.BannerMode

const (
	BannerModeConsole = banner.BannerModeConsole
	BannerModeLog     = banner.BannerModeLog
	BannerModeOff     = banner.BannerModeOff
)

type TextBanner = banner.TextBanner
type ASCIIArtBanner = banner.ASCIIArtBanner
type LegacyBanner = banner.LegacyBanner

// LegacyBannerOption 是 banner.LegacyOption 的类型别名。
type LegacyBannerOption = banner.LegacyOption

var (
	NewLegacyBanner    = banner.NewLegacyBanner
	BannerWithLines    = banner.WithLines
	BannerWithAppName  = banner.WithAppName
	BannerWithProfiles = banner.WithProfiles
)

// ==================== 向后兼容的类型别名 ====================

// BootErrorStruct 是 bootError 实现类型的别名。
//
// 在重构前 BootError 是一个导出字段的结构体，现在改为接口。
// 此别名保留对底层实现类型的访问，便于需要类型断言的场景。
type BootErrorStruct = bootError

// StarterRegistryStruct 是 starterRegistryImpl 实现类型的别名。
//
// 在重构前 StarterRegistry 是一个结构体，现在改为接口。
// 此别名保留对底层实现类型的访问，便于需要类型断言的场景。
type StarterRegistryStruct = starterRegistryImpl

// ==================== Application 接口 ====================

// Application 应用接口，管理应用的完整生命周期。
//
// Boot 结构体实现了此接口的所有方法（Context() 返回具体类型）。
// NewApplication 返回 *Boot，可直接调用 Start/Stop/WaitForSignal。
// 需要接口类型的场景（如测试 mock）可使用此接口。
//
// 示例：
//
//	app, _ := boot.NewApplication(boot.WithAppName("my-app"))
//	app.Start()  // 直接使用 *Boot
//
//	// 需要接口类型时：
//	var iface boot.Application = app
//	iface.Start()
type Application interface {
	// Start 启动应用，执行完整的初始化流程。
	// 包括配置加载、自动配置执行、Starter 配置和启动。
	Start() error

	// Stop 停止应用，逆序释放所有资源。
	Stop() error

	// WaitForSignal 阻塞等待 SIGINT/SIGTERM 信号，收到后自动执行优雅关闭。
	WaitForSignal()

	// Context 返回应用上下文，包含 IoC 容器和环境配置。
	Context() ApplicationContext

	// Config 返回启动配置。
	Config() *BootConfig
}

// ==================== ApplicationContext 接口 ====================

// ApplicationContext 自动配置看到的上下文接口。
//
// 实际由 context.DefaultApplicationContext 实现。
// 这是 context.ApplicationContext 的子集，仅包含自动配置需要的方法。
type ApplicationContext interface {
	// Context 返回应用级别的 Go 标准 context.Context。
	//
	// 该 context 在应用启动时创建，在应用停止时取消。
	// Starters 和 AutoConfigurations 应使用此 context 进行日志记录、
	// 超时控制和取消传播，避免使用 context.Background()。
	Context() context.Context

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

// ==================== AutoConfiguration 接口 ====================

// AutoConfiguration 自动配置接口。
//
// 参考 Spring Boot 的 @Configuration + @Bean 模式。
// 每个模块实现此接口，通过 RegisterAutoConfig 注册。
type AutoConfiguration interface {
	// Configure 执行自动配置逻辑。
	Configure(ctx ApplicationContext) error
}

// ==================== Starter 接口 ====================

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

// ==================== StarterRegistry 接口 ====================

// StarterRegistry 启动器注册表接口。
//
// 管理所有 Starter 的注册和依赖排序。
// 使用 Kahn 算法进行拓扑排序，确保依赖的启动器先启动。
type StarterRegistry interface {
	// Register 注册一个 Starter 到注册表。
	Register(starter Starter)

	// Get 根据名称获取已注册的 Starter，未找到返回 nil。
	Get(name string) Starter

	// GetAll 获取所有已注册的 Starter（返回副本）。
	GetAll() []Starter

	// GetOrdered 按依赖关系拓扑排序获取 Starter。
	// 如果存在循环依赖，回退到原始注册顺序。
	GetOrdered() []Starter
}

// ==================== BootError 接口 ====================

// BootError 结构化启动错误接口。
//
// 提供标准化的错误信息访问方式，包含错误码、错误消息和原始错误。
// 通过 NewBootErr 创建实例，通过 errors.As 提取。
//
// 示例：
//
//	var bootErr boot.BootError
//	if errors.As(err, &bootErr) {
//	    fmt.Println(bootErr.Code(), bootErr.Message())
//	}
type BootError interface {
	// Code 返回错误码，用于程序化错误处理。
	Code() string

	// Message 返回人类可读的错误消息。
	Message() string

	// Cause 返回原始错误，用于错误链追踪。
	Cause() error

	// Error 实现 error 接口。
	Error() string

	// Unwrap 实现 errors.Unwrap 接口，支持 errors.Is/As。
	Unwrap() error
}

// ==================== FailureAnalyzer 接口 ====================

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

// ==================== Banner 类型别名 ====================

// Banner 是 banner.Banner 的类型别名，用于向后兼容。
//
// 新代码应直接使用 banner.Banner。
type Banner = banner.Banner

// ==================== EventBusResult 接口 ====================

// EventBusResult 事件总线访问接口。
type EventBusResult interface {
	Publish(event event.ApplicationEvent)
}

// ==================== BeanProvider 类型 ====================

// BeanProvider Bean 提供者函数类型。
//
// 定义如何向容器注册 Bean，是 Module 中 Bean 注册的统一抽象。
type BeanProvider func(c core.Container) error

// 以下类型定义在其他文件中，此处仅作文档说明：
// - ApplicationOption: application.go（包含完整实现）
// - Boot: boot.go（包含完整实现，实现 Application 接口方法）
// - BootConfig: config.go（启动配置结构体）
// - starterRegistryImpl: starter.go（StarterRegistry 接口的具体实现）
// - bootError: boot_error.go（BootError 接口的具体实现）
// - ConfigCenterProvider: configcenter.go（包含完整实现）
// - ConfigCenterAdapter: configcenter.go（包含完整实现）
// - FailureAnalyzerAdapter: failure_analyzers.go（包含完整实现）
// - Report: report.go（包含完整实现）
// - FailureReport: failure_analyzer.go（失败报告结构体）
// - FailureAnalyzerRegistry: failure_analyzer.go（失败分析器注册表）
// - SimpleFailureAnalyzer: failure_analyzer.go（简单失败分析器）
// - AutoConfigEntry: autoconfig.go（自动配置条目）
// - AutoConfigurationOption: autoconfig.go（自动配置选项函数）
// - Module: module.go（可组合的配置单元）
