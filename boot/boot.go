package boot

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/xudefa/enhance/config/environment"
	contextpkg "github.com/xudefa/enhance/context"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/event"
	"github.com/xudefa/enhance/lifecycle"
)

// Boot 应用启动器，管理应用的完整生命周期
//
// 参考 Spring Boot 的 SpringApplication，负责：
//   - 自动配置执行（AutoConfiguration）
//   - 启动器管理（Starter 的 Configure/Start/Stop）
//   - 生命周期阶段流转
//   - 事件发布
//   - 优雅关闭
type Boot struct {
	ctx          *contextpkg.DefaultApplicationContext
	config       *BootConfig
	configLoader *environment.ConfigLoader
	starters     []Starter
	hooks        *lifecycle.HookRegistry
}

// Context 返回应用上下文
func (b *Boot) Context() *contextpkg.DefaultApplicationContext {
	return b.ctx
}

// Container 返回 IoC 容器
func (b *Boot) Container() core.Container {
	return b.ctx.Container()
}

// Environment 返回环境配置
func (b *Boot) Environment() *environment.Environment {
	return b.ctx.Environment()
}

// Start 启动应用，执行完整的生命周期
//
// 启动流程简化为 3 阶段：
//  1. PhaseInit：加载配置、注册 Bean、启动启动器
//  2. PhaseRunning：应用正常运行
func (b *Boot) Start() error {
	if b.ctx.IsRunning() {
		return nil
	}

	// === 阶段 1：初始化阶段 ===
	if b.configLoader != nil {
		configSources, err := b.configLoader.Load()
		if err != nil {
			return b.reportError("初始化", fmt.Errorf("加载配置文件失败: %w", err))
		}

		for _, source := range configSources {
			b.ctx.Environment().AddPropertySource(source)
		}
	}

	if b.config.ConfigCenterEnabled {
		if err := b.loadConfigCenterConfig(); err != nil {
			return b.reportError("初始化", fmt.Errorf("加载配置中心失败: %w", err))
		}
	}

	for _, source := range b.config.CustomPropertySources {
		b.ctx.Environment().AddPropertySourceFirst(source)
	}

	b.ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventEnvironmentPrepared})

	if b.config.AutoExecute {
		entries := GlobalRegistry().GetMatchingWithExclude(newConditionCtx(b.ctx), b.config.ExcludedAutoConfigs)
		allEntries := GlobalRegistry().GetAll()

		// 收集自动配置报告
		reportEnabled := IsAutoConfigReportEnabled() || b.ctx.Environment().GetBool("enhance.debug", false)
		if reportEnabled {
			ResetAutoConfigReport()
			b.collectAutoConfigReport(allEntries, entries)
		}

		for _, entry := range entries {
			if err := entry.Config.Configure(newAppCtx(b.ctx)); err != nil {
				return b.reportError("初始化", fmt.Errorf("自动配置 %T 失败: %w", entry.Config, err))
			}
		}

		// 打印自动配置报告
		if reportEnabled {
			GetAutoConfigReport().Print()
		}
	}

	// 安装显式模块（Go 风格组合）
	for _, module := range b.config.Modules {
		// 检查模块条件
		if !b.moduleMatches(module) {
			continue
		}
		if err := module.Install(b.ctx.Container()); err != nil {
			return b.reportError("初始化", fmt.Errorf("模块 %s 安装失败: %w", module.ModuleName(), err))
		}
		// 收集模块的 Starter
		if b.config.Starters {
			b.starters = append(b.starters, module.ModuleStarters()...)
		}
		// 收集模块的钩子
		for _, h := range module.ModuleHooks() {
			b.hooks.Register(h)
		}
	}

	// 注册全局钩子
	for _, h := range lifecycle.GlobalHookRegistry().GetAll() {
		b.hooks.Register(h)
	}

	// 注册配置中的钩子（通过 WithHook / WithHookFunc）
	for _, h := range b.config.Hooks {
		b.hooks.Register(h)
	}

	// 合并全局注册的 Starter 和模块 Starter，然后拓扑排序
	allStarters := append(b.starters, GlobalStarterRegistry().GetOrdered()...)
	b.starters = deduplicateStarters(allStarters)

	if b.config.Starters {
		for _, s := range b.starters {
			if !b.starterMatches(s) {
				continue
			}
			if err := s.Configure(newAppCtx(b.ctx)); err != nil {
				return b.reportError("初始化", fmt.Errorf("启动器 %s 配置失败: %w", s.Name(), err))
			}
		}
	}

	b.ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventContextRefreshed})

	// 执行 OnInit 钩子（Bean 注册完成后）
	if b.hooks.Count() > 0 {
		if err := b.hooks.InitAll(context.Background()); err != nil {
			return b.reportError("initializing", err)
		}
	}

	if b.config.Starters {
		for _, s := range b.starters {
			if !b.starterMatches(s) {
				continue
			}
			if err := s.Start(newAppCtx(b.ctx)); err != nil {
				return b.reportError("初始化", fmt.Errorf("启动器 %s 启动失败: %w", s.Name(), err))
			}
		}
	}

	// 执行 OnStart 钩子（应用启动前）
	if b.hooks.Count() > 0 {
		if err := b.hooks.StartAll(context.Background()); err != nil {
			return b.reportError("initializing", err)
		}
	}

	banner := DefaultBanner
	banner.Print(os.Stdout, b.config.AppName, b.config.Version, b.ctx.Environment().GetActiveProfiles())

	// === 阶段 2：运行阶段 ===
	if err := b.ctx.Lifecycle().SetPhase(lifecycle.PhaseRunning); err != nil {
		return b.reportError("running", err)
	}

	b.ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})
	b.ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventApplicationReady})

	return nil
}

// Stop 停止应用
//
// 流程：
//  1. 逆序停止启动器
//  2. 执行 OnStop 钩子
//  3. 发布停止事件
//
// 注意：允许从任何阶段调用 Stop，以支持 Start() 部分失败时的资源清理。
func (b *Boot) Stop() error {
	phase := b.ctx.Lifecycle().GetPhase()

	// 已停止，无需清理
	if phase == lifecycle.PhaseStopped {
		return nil
	}

	// 逆序停止启动器（仅当启动器已启动后才执行停止）
	if b.config.Starters && phase == lifecycle.PhaseRunning {
		for i := len(b.starters) - 1; i >= 0; i-- {
			s := b.starters[i]
			if !b.starterMatches(s) {
				continue
			}
			if err := s.Stop(newAppCtx(b.ctx)); err != nil {
				fmt.Fprintf(os.Stderr, "starter %s stop error: %v\n", s.Name(), err)
			}
		}
	}

	// 执行 OnStop 钩子（逆序释放资源）
	if b.hooks.Count() > 0 {
		if err := b.hooks.StopAll(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "hook stop error: %v\n", err)
		}
	}

	if err := b.ctx.Lifecycle().SetPhase(lifecycle.PhaseStopped); err != nil {
		_ = b.reportError("stopping", err)
		fmt.Fprintf(os.Stderr, "[enhance] failed to set phase to stopped: %v\n", err)
	}

	b.ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventApplicationStopped})
	return nil
}

// IsRunning 检查应用是否运行中
func (b *Boot) IsRunning() bool {
	return b.ctx.Lifecycle().GetPhase() == lifecycle.PhaseRunning
}

// BindConfig 将配置绑定到指定类型的结构体（泛型版本）
//
// 示例:
//
//	cfg, err := bootApp.BindConfig[ServerConfig]()
//	// 或带前缀:
//	cfg, err := bootApp.BindConfig[ServerConfig](boot.WithConfigPrefix("server"))
func BindConfig[T any](b *Boot, opts ...BindConfigOption) (T, error) {
	var zero T
	target := new(T)

	cfgOpts := &bindConfigOpts{}
	for _, opt := range opts {
		opt(cfgOpts)
	}

	var err error
	if cfgOpts.prefix != "" {
		err = b.ctx.Environment().BindPrefix(cfgOpts.prefix, target)
	} else {
		err = b.ctx.Environment().Bind(target)
	}
	if err != nil {
		return zero, err
	}
	return *target, nil
}

// BindConfigOption 配置绑定选项
type BindConfigOption func(*bindConfigOpts)

type bindConfigOpts struct {
	prefix string
}

// WithConfigPrefix 设置配置绑定前缀
func WithConfigPrefix(prefix string) BindConfigOption {
	return func(o *bindConfigOpts) {
		o.prefix = prefix
	}
}

// WaitForSignal 等待终止信号，收到信号后自动执行优雅关闭
func (b *Boot) WaitForSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\n接收到终止信号，正在优雅关闭...")
	signal.Stop(sigCh)
	if stopErr := b.Stop(); stopErr != nil {
		fmt.Fprintf(os.Stderr, "[enhance] failed to stop application: %v\n", stopErr)
	}
}

// starterMatches 检查启动器条件是否匹配
func (b *Boot) starterMatches(s Starter) bool {
	cond := s.GetCondition()
	if cond == nil {
		return true
	}
	return cond.Matches(newConditionCtx(b.ctx))
}

// moduleMatches 检查模块条件是否全部匹配
func (b *Boot) moduleMatches(m Module) bool {
	conds := m.conditions
	if len(conds) == 0 {
		return true
	}
	ctx := newConditionCtx(b.ctx)
	for _, cond := range conds {
		if !cond.Matches(ctx) {
			return false
		}
	}
	return true
}

// deduplicateStarters 去重启动器列表（按名称去重，保留第一个）
func deduplicateStarters(starters []Starter) []Starter {
	seen := make(map[string]bool)
	result := make([]Starter, 0, len(starters))
	for _, s := range starters {
		name := s.Name()
		if !seen[name] {
			seen[name] = true
			result = append(result, s)
		}
	}
	return result
}

// Run 启动应用并等待信号（一行启动）
//
// 参考 Spring Boot 的 SpringApplication.run() 方法，
// 一行代码直接启动应用，自动加载 JSON 配置文件，阻塞直到应用关闭。
//
// 默认行为：
//   - 自动加载 application.json 配置文件
//   - 自动执行 AutoConfiguration
//   - 自动启动所有 Starter
//   - 阻塞等待 SIGINT/SIGTERM 信号
//   - 收到信号后自动优雅关闭
//
// 示例（最简用法）：
//
//	func main() {
//	    boot.Run()
//	}
//
// 示例（带选项）：
//
//	func main() {
//	    boot.Run(
//	        boot.WithAppName("my-app"),
//	        boot.WithVersion("1.0.0"),
//	        boot.WithProfiles("dev"),
//	    )
//	}
//
// 示例（带模块）：
//
//	func main() {
//	    boot.Run(
//	        boot.WithAppName("my-app"),
//	        boot.WithModulesOption(WebModule, DatabaseModule),
//	    )
//	}
func Run(opts ...any) {
	app, err := NewApplicationFromRunOptions(opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[enhance] failed to create application: %v\n", err)
		os.Exit(1)
	}

	if err := app.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "[enhance] failed to start application: %v\n", err)
		os.Exit(1)
	}

	app.WaitForSignal()

	if err := app.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "[enhance] failed to stop application: %v\n", err)
		os.Exit(1)
	}
}

// NewApplicationFromRunOptions 从 Run 选项创建应用
//
// 支持 BootOption 和 ApplicationOption 混合使用。
func NewApplicationFromRunOptions(opts ...any) (*Boot, error) {
	return NewApplicationWithOptions(opts...)
}

// reportError 通过 FailureAnalyzer 输出友好错误提示并返回结构化错误
func (b *Boot) reportError(phase string, err error) *BootError {
	bootErr := NewBootError(phase, err)

	// 使用 FailureAnalyzer 分析
	report := globalAnalyzerRegistry.Analyze(err)
	if report != nil {
		_ = bootErr.WithAnalysis(report.Description)
		_ = bootErr.WithSuggestions(report.PossibleSolutions...)
		fmt.Fprintf(os.Stderr, "\n%s\n", formatFailure(report))
	}

	return bootErr
}

// collectAutoConfigReport 收集自动配置匹配情况
func (b *Boot) collectAutoConfigReport(allEntries []AutoConfigEntry, matchedEntries []AutoConfigEntry) {
	report := GetAutoConfigReport()

	// 构建已匹配的集合
	matchedSet := make(map[string]bool)
	for _, entry := range matchedEntries {
		name := typeName(entry.Config)
		matchedSet[name] = true
	}

	// 遍历所有条目，记录匹配结果
	for _, entry := range allEntries {
		name := typeName(entry.Config)

		// 构建条件结果
		conditions := make([]ConditionResult, 0, len(entry.Conditions))
		for _, cond := range entry.Conditions {
			ctx := newConditionCtx(b.ctx)
			matched := cond.Matches(ctx)
			conditions = append(conditions, ConditionResult{
				Condition: fmt.Sprintf("%T", cond),
				Matched:   matched,
				Message:   cond.String(),
			})
		}

		// 无条件时记录为无条件类
		if len(entry.Conditions) == 0 {
			report.RecordUnconditional(name)
			continue
		}

		// 根据是否匹配记录到正面或负面
		if matchedSet[name] {
			report.RecordPositiveMatch(name, conditions)
		} else {
			report.RecordNegativeMatch(name, conditions)
		}
	}
}

// typeName 获取类型的简短名称
func typeName(v any) string {
	t := fmt.Sprintf("%T", v)
	// 移除包路径，只保留类型名
	if idx := strings.LastIndex(t, "."); idx >= 0 {
		return t[idx+1:]
	}
	return t
}
