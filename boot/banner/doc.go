// Package banner 提供启动横幅功能，用于 enhance 框架。
//
// 该包定义了启动横幅的接口和多种实现，支持控制台输出、日志输出和关闭模式。
// 参考 Spring Boot 的 Banner 设计，支持 ASCII 艺术、文本模板等多种横幅格式。
//
// # 架构设计
//
//   - Banner: 启动横幅接口，定义横幅打印和模式查询方法
//   - BannerMode: 横幅模式枚举，控制横幅的输出方式
//   - TextBanner: 文本横幅实现，支持模板和属性替换
//   - ASCIIArtBanner: ASCII 艺术横幅实现，支持自定义艺术文本
//   - LegacyBanner: 旧版横幅实现，兼容 boot 包中的横幅格式
//
// # 核心功能
//
//   - 启动横幅显示: 在应用启动时打印横幅信息
//   - 多种输出模式: 支持控制台输出、日志输出和关闭模式
//   - 模板渲染: 支持文本模板和属性替换
//   - ASCII 艺术: 支持自定义 ASCII 艺术文本
//
// # 使用方式
//
// 使用文本横幅：
//
//	banner := banner.NewTextBanner("Hello {{name}} v{{version}}", map[string]any{
//	    "name": "my-app",
//	    "version": "1.0.0",
//	})
//	banner.Print("1.0.0")
//
// 使用 ASCII 艺术横幅：
//
//	banner := banner.NewASCIIArtBanner("  ___  \n /   \\ \n|     |", "")
//	banner.Print("1.0.0")
//
// 使用旧版横幅：
//
//	b := banner.NewLegacyBanner(
//	    banner.WithLines([]string{"Line 1", "Line 2"}),
//	    banner.WithAppName("my-app"),
//	    banner.WithProfiles([]string{"dev"}),
//	)
//	b.Print("1.0.0")
package banner

// Banner 启动横幅接口。
//
// 参考 Spring Boot 的 Banner，支持多种格式的启动横幅显示。
// 每个横幅实现都必须实现此接口，以便与 enhance 的启动流程集成。
type Banner interface {
	// Print 打印横幅到标准输出。
	//
	// 参数 version 为应用版本号，用于在横幅中显示版本信息。
	// 返回值 error 表示打印过程中是否发生错误。
	Print(version string) error

	// Mode 返回当前横幅的输出模式。
	//
	// 返回值 BannerMode 枚举，指示横幅将输出到控制台、日志还是关闭。
	Mode() BannerMode
}

// BannerMode 横幅模式枚举。
//
// 控制横幅的输出方式，支持控制台输出、日志输出和关闭模式。
type BannerMode int

const (
	// BannerModeConsole 控制台模式。
	//
	// 横幅直接输出到标准输出（os.Stdout）。
	BannerModeConsole BannerMode = iota

	// BannerModeLog 日志模式。
	//
	// 横幅通过日志框架输出，不直接打印到控制台。
	BannerModeLog

	// BannerModeOff 关闭模式。
	//
	// 横幅不输出任何内容，完全禁用横幅显示。
	BannerModeOff
)
