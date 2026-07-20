package boot

import (
	"github.com/xudefa/enhance/config/environment"
	contextpkg "github.com/xudefa/enhance/context"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/lifecycle"
)

// NewApplication 创建新的应用实例。
//
// 这是 enhance 框架的推荐入口点，支持 BootOption 和 ApplicationOption 混合使用。
// 自动配置（AutoConfiguration）和启动器（Starter）会在 Start() 方法中按生命周期阶段自动执行。
//
// 参数:
//   - opts: 可选的配置选项，支持 BootOption 和 ApplicationOption 混合传入
//
// 返回值:
//   - *Boot: 应用启动器实例，可用于后续的 Start/Stop 操作
//   - error: 创建失败时返回错误（通常是 ApplicationOption 执行失败）
//
// 示例（基础用法）：
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
// 示例（带模块）：
//
//	app, err := boot.NewApplication(
//	    boot.WithAppName("my-app"),
//	    boot.WithModulesOption(DatabaseModule, WebModule),
//	)
//
// 启动流程说明:
//  1. 创建 IoC 容器和环境配置
//  2. 加载配置文件并注册到环境
//  3. 执行自动配置（AutoConfiguration）
//  4. 安装显式模块（Modules）
//  5. 配置并启动所有 Starter
//  6. 发布生命周期事件
//
// 注意:
//   - 该方法只创建应用实例，不会启动应用
//   - 需要显式调用 Start() 方法才会执行完整的启动流程
//   - 支持多次调用 NewApplication 创建多个独立的应用实例
func NewApplication(opts ...any) (*Boot, error) {
	cfg := defaultBootConfig()
	appOpts := make([]ApplicationOption, 0)

	for _, opt := range opts {
		switch o := opt.(type) {
		case BootOption:
			o(cfg)
		case ApplicationOption:
			appOpts = append(appOpts, o)
		}
	}

	container := core.NewContainer()
	env := environment.NewEnvironment()
	appCtx := contextpkg.NewApplicationContext(container, env)

	for _, p := range cfg.Profiles {
		appCtx.Environment().AddActiveProfile(p)
	}

	configLoader := environment.NewConfigLoader(
		"application",
		environment.ConfigType(cfg.ConfigType),
		cfg.ConfigLocation,
		cfg.Profiles,
	)

	b := &Boot{
		ctx:          appCtx,
		config:       cfg,
		configLoader: configLoader,
		hooks:        lifecycle.NewHookRegistry(),
	}

	// 应用 ApplicationOption（用于模块组合等后处理）
	for _, opt := range appOpts {
		if err := opt(b); err != nil {
			return nil, err
		}
	}

	return b, nil
}

// NewApplicationWithOptions 是 NewApplication 的别名，保持向后兼容。
//
// Deprecated: 使用 NewApplication 替代。
func NewApplicationWithOptions(opts ...any) (*Boot, error) {
	return NewApplication(opts...)
}
