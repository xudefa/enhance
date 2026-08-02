// Package environment 提供环境配置管理功能，用于 enhance 框架。
//
// 该模块支持多环境配置加载、属性源管理、配置文件绑定、类型转换等环境配置相关功能。
// 参考 Spring Environment 的设计理念。
//
// # 架构设计
//
//   - Environment: 环境配置接口，定义配置访问操作
//   - PropertySource: 属性源接口，表示配置数据来源
//   - PropertyResolver: 属性解析器，解析和转换配置值
//   - TypeConverter: 类型转换器，处理字符串到类型的转换
//   - Profile: 环境配置，表示特定的运行环境
//
// # 核心功能
//
//   - 多环境配置: 支持 dev、test、prod 等多环境配置
//   - 属性源管理: 支持多个属性源的优先级管理
//   - 配置文件绑定: 支持从 JSON、YAML 等文件加载配置
//   - 类型转换: 支持字符串到各种 Go 类型的转换
//   - 属性占位符: 支持 ${key} 格式的占位符替换
//
// # 使用方式
//
// 创建环境配置：
//
//	env := environment.NewEnvironment()
//	env.AddPropertySource(environment.NewFilePropertySource("application.json"))
//	env.AddPropertySource(environment.NewEnvPropertySource())
//
// 获取属性：
//
//	port := env.GetProperty("server.port", "8080")
//	timeout := env.GetPropertyAsInt("server.timeout", 30)
//
// 设置激活的环境：
//
//	env.SetActiveProfiles("dev", "local")
//
// # 属性源优先级
//
// 属性源按添加顺序的逆序优先级（后添加的优先级更高）：
//
//   - 命令行参数（最高优先级）
//   - 环境变量
//   - 配置文件
//   - 默认值（最低优先级）
//
// # 配置文件格式
//
// 支持以下配置文件格式：
//
//   - application.json: JSON 格式配置
//   - application.yaml: YAML 格式配置
//   - application-{profile}.json: 特定环境配置
package environment

// Priority 优先级类型，值越大优先级越高。
type Priority int

const (
	// PriorityLowest 最低优先级，应用配置使用
	PriorityLowest Priority = iota
	// PriorityLow 低优先级
	PriorityLow
	// PriorityNormal 正常优先级
	PriorityNormal
	// PriorityHigh 高优先级，环境变量使用
	PriorityHigh
	// PriorityHighest 最高优先级，命令行使用
	PriorityHighest
)

// PriorityFallback 回退优先级（最低），默认值/回退配置源使用。
//
// 低于所有常规优先级（含 PriorityLowest），确保文件配置源总是优先于默认值，
// 不受添加顺序影响。
const PriorityFallback Priority = PriorityLowest - 1

// PropertySource 配置源接口。
//
// 代表一个配置数据源，具有名称和优先级。
type PropertySource interface {
	Name() string
	GetProperty(key string) (any, bool)
	Priority() Priority
	Contains(key string) bool
}

// MapPropertySource 基于 map 的内存配置源。
type MapPropertySource struct {
	name     string
	data     map[string]any
	priority Priority
}

// ArgsPropertySource 命令行参数配置源。
type ArgsPropertySource struct {
	name     string
	args     map[string]string
	priority Priority
}

// EnvPropertySource 环境变量配置源。
type EnvPropertySource struct {
	name     string
	prefix   string
	priority Priority
}
