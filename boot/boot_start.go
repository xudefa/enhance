package boot

import (
	"fmt"

	"github.com/xudefa/enhance/event"
	"github.com/xudefa/enhance/lifecycle"
)

// startPhaseInit 执行启动阶段 1：加载配置、注册 Bean、配置并启动启动器。
func (b *Boot) startPhaseInit(report *StartupReport) error {
	if err := b.loadConfiguration(); err != nil {
		return err
	}

	b.ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventEnvironmentPrepared})

	if b.config.AutoExecute {
		if err := b.runAutoConfigurations(report); err != nil {
			return err
		}
	}

	if err := b.installModules(report); err != nil {
		return err
	}

	b.registerHooks()

	// 合并全局注册的 Starter 和模块 Starter，然后拓扑排序
	b.starters = deduplicateStarters(append(b.starters, GlobalStarterRegistry().GetOrdered()...))

	if b.config.Starters {
		if err := b.configureStarters(); err != nil {
			return err
		}
	}

	b.ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventContextRefreshed})

	// 执行 OnInit 钩子（Bean 注册完成后）
	if b.hooks.Count() > 0 {
		if err := b.hooks.InitAll(b.rootCtx); err != nil {
			return b.reportError("initializing", err)
		}
	}

	if b.config.Starters {
		if err := b.startStarters(report); err != nil {
			return err
		}
	}

	return b.runStartHooks()
}

// loadConfiguration 加载配置文件、配置中心配置和自定义属性源，并注入环境。
func (b *Boot) loadConfiguration() error {
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

	return nil
}

// runAutoConfigurations 执行满足条件的自动配置并收集报告信息。
func (b *Boot) runAutoConfigurations(report *StartupReport) error {
	entries := GlobalRegistry().GetMatchingWithExclude(newConditionCtx(b.ctx), b.config.ExcludedAutoConfigs)
	allEntries := GlobalRegistry().GetAll()

	// 收集自动配置报告
	reportEnabled := IsAutoConfigReportEnabled() || b.ctx.Environment().GetBool("enhance.debug", false)
	if reportEnabled {
		ResetAutoConfigReport()
		b.collectAutoConfigReport(allEntries, entries)
	}

	for _, entry := range entries {
		if err := entry.Config.Configure(newAppCtx(b.ctx, b.rootCtx)); err != nil {
			return b.reportError("初始化", fmt.Errorf("自动配置 %T 失败: %w", entry.Config, err))
		}
	}

	// 打印自动配置报告
	if reportEnabled {
		GetAutoConfigReport().Print()
	}

	// 收集启动报告信息
	report.SetAutoConfigCount(len(entries))
	for _, entry := range entries {
		report.AddAutoConfig(fmt.Sprintf("%T", entry.Config))
	}

	return nil
}

// installModules 安装显式模块，收集模块的 Starter 和钩子。
func (b *Boot) installModules(report *StartupReport) error {
	// 安装显式模块（Go 风格组合）
	moduleCount := 0
	for _, module := range b.config.Modules {
		// 检查模块条件
		if !b.moduleMatches(module) {
			continue
		}
		if err := module.Install(b.ctx.Container()); err != nil {
			return b.reportError("初始化", fmt.Errorf("模块 %s 安装失败: %w", module.ModuleName(), err))
		}
		moduleCount++
		if module.ModuleName() != "" {
			report.AddModule(module.ModuleName())
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
	report.SetModuleCount(moduleCount)
	return nil
}

// registerHooks 注册全局钩子与配置中的钩子。
func (b *Boot) registerHooks() {
	// 注册全局钩子
	for _, h := range lifecycle.GlobalHookRegistry().GetAll() {
		b.hooks.Register(h)
	}

	// 注册配置中的钩子（通过 WithHook / WithHookFunc）
	for _, h := range b.config.Hooks {
		b.hooks.Register(h)
	}
}
