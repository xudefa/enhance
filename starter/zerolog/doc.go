// Package zerolog 提供基于 Zerolog 的日志集成能力。
//
// 本模块是 enhance 框架的日志集成模块，基于 github.com/rs/zerolog 实现，
// 提供高性能、低分配的结构化日志记录能力。
//
// 官方文档：https://github.com/rs/zerolog
//
// # 模块独立性
//
// 本模块采用独立模块设计（拥有独立的 go.mod），依赖隔离确保：
//   - 用户只使用 enhance 核心模块时，不会引入 Zerolog 依赖
//   - 用户显式引入本模块时，才会下载 Zerolog 及其间接依赖
//   - 避免依赖污染，保持用户项目的依赖树清晰
//
// # 架构设计
//
// 核心组件：
//   - ZerologLogger: 实现 enhance/log.Logger 接口的 Zerolog 适配器
//   - ZerologAutoConfiguration: 自动配置类，根据配置文件创建和注册日志器
//
// 自动配置：
//   - 当 zerolog.enabled=true 时自动生效
//   - 自动创建 Zerolog 实例并注册到 IoC 容器
//   - 支持日志级别、输出格式、采样等配置
//
// # 快速开始
//
// 1. 在项目中引入模块：
//
//	import _ "github.com/xudefa/enhance/starter/zerolog"
//
// 2. 在配置文件中启用 Zerolog：
//
//	{
//	  "zerolog": {
//	    "enabled": true,
//	    "level": "info",
//	    "format": "json",
//	    "time-format": "2006-01-02 15:04:05",
//	    "add-source": false,
//	    "output-path": ""
//	  }
//	}
//
// 3. 在代码中使用：
//
//	logger := log.Build() // 从容器中获取日志记录器
//	logger.Info(context.Background(), "Application started",
//	    log.KeyValue{Key: "version", Value: "1.0.0"})
//
// # 配置说明
//
//   - zerolog.enabled: 是否启用 Zerolog（默认 false）
//   - zerolog.level: 日志级别（debug/info/warn/error，默认 info）
//   - zerolog.format: 输出格式（json/console，默认 json）
//   - zerolog.time-format: 时间格式（默认 RFC3339）
//   - zerolog.add-source: 是否添加源码位置（默认 false）
//   - zerolog.output-path: 日志文件输出路径（默认 stdout）
//
// # 依赖说明
//
// 本模块依赖：
//   - github.com/rs/zerolog: Zerolog 核心库
//
// 用户项目引入本模块后，会自动引入上述依赖。
package zerolog

// ==================== 配置键常量 ====================

const (
	// ZeroLog 配置
	ZeroLogEnabled    = "log.zerolog.enabled"
	ZeroLogLevel      = "log.zerolog.level"
	ZeroLogFormat     = "log.zerolog.format"
	ZeroLogTimeFormat = "log.zerolog.time-format"
	ZeroLogAddSource  = "log.zerolog.add-source"
	ZeroLogOutputPath = "log.zerolog.output-path"

	// 日志级别常量
	LogLevelDebug   = "debug"
	LogLevelInfo    = "info"
	LogLevelWarn    = "warn"
	LogLevelWarning = "warning"
	LogLevelError   = "error"
	LogLevelFatal   = "fatal"
	LogLevelPanic   = "panic"

	// 日志格式常量
	LogFormatConsole = "console"
	LogFormatJSON    = "json"

	// 日志字段常量
	LogFieldLevel  = "level"
	LogFieldFormat = "format"
	LogFieldOutput = "output-path"
)

// ==================== 默认值常量 ====================

const (
	// ZeroLog 默认值
	DefaultZeroLogLevel      = "info"
	DefaultZeroLogFormat     = "json"
	DefaultZeroLogTimeFormat = "2006-01-02 15:04:05"
	DefaultZeroLogAddSource  = false
	DefaultZeroLogOutputPath = ""

	// 条件值常量
	ConditionTrue = "true"
)
