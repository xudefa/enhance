// Package enhance 提供类似 Spring Boot 的简洁启动入口
//
// 参考 Spring Boot 的 SpringApplication.run() 设计，
// 提供一行代码启动 Web 服务的能力。
//
// 示例（最简用法）：
//
//	func main() {
//	    enhance.Run()
//	}
//
// 示例（带配置选项）：
//
//	func main() {
//	    enhance.Run(
//	        boot.WithAppName("my-app"),
//	        boot.WithVersion("1.0.0"),
//	        boot.WithProfiles("dev"),
//	    )
//	}
//
// 示例（带模块）：
//
//	func main() {
//	    enhance.Run(
//	        boot.WithModulesOption(WebModule, DatabaseModule),
//	    )
//	}
package enhance

import (
	"github.com/xudefa/enhance/boot"
)

// Enhance 是 enhance 框架的主入口点
//
// 提供类似 Spring Boot 的简洁 API，一行代码启动应用。
//
// 示例：
//
//	func main() {
//	    enhance.Run()
//	}
var Enhance = &EnhanceApp{}

// EnhanceApp 应用启动器
//
// 封装 boot.Boot 的所有功能，提供更简洁的 API。
type EnhanceApp struct{}

// Run 启动应用并等待信号（一行启动）
//
// 参考 Spring Boot 的 SpringApplication.run() 方法，
// 一行代码直接启动应用，自动加载 JSON 配置文件，阻塞直到应用
// 收到 SIGINT/SIGTERM 信号后自动优雅关闭。
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
//	    enhance.Run()
//	}
//
// 示例（带选项）：
//
//	func main() {
//	    enhance.Run(
//	        boot.WithAppName("my-app"),
//	        boot.WithVersion("1.0.0"),
//	        boot.WithProfiles("dev"),
//	    )
//	}
//
// 示例（带模块）：
//
//	func main() {
//	    enhance.Run(
//	        boot.WithAppName("my-app"),
//	        boot.WithModulesOption(WebModule, DatabaseModule),
//	    )
//	}
func (e *EnhanceApp) Run(opts ...boot.BootOption) {
	boot.Run(opts...)
}

// NewApplication 创建新的应用实例（不启动）
//
// 当你需要更多控制时使用，创建后可以手动调用 Start() 和 Stop()。
//
// 示例：
//
//	func main() {
//	    app, err := enhance.NewApplication(
//	        boot.WithAppName("my-app"),
//	    )
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    app.Start()
//	    // ... 自定义逻辑
//	    app.Stop()
//	}
func NewApplication(opts ...boot.BootOption) (*boot.Boot, error) {
	return boot.NewApplication(opts...)
}

// Run 包级别的便捷函数，直接启动应用
//
// 这是最常用的入口点，一行代码启动整个应用。
//
// 示例：
//
//	func main() {
//	    enhance.Run()
//	}
func Run(opts ...boot.BootOption) {
	boot.Run(opts...)
}
