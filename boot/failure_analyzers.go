package boot

import (
	"errors"
	"strings"

	"github.com/xudefa/enhance/core"
)

// 配置错误。
var (
	// ErrPropertyNotFound 配置项未找到。
	ErrPropertyNotFound = errors.New("property not found")

	// ErrTypeConversion 类型转换失败。
	ErrTypeConversion = errors.New("type conversion failed")
)

func init() {
	// 注册内置失败分析器到全局注册表
	globalAnalyzerRegistry.Register(NewBeanNotFoundAnalyzer())
	globalAnalyzerRegistry.Register(NewCircularDependencyAnalyzer())
	globalAnalyzerRegistry.Register(NewDuplicateBeanAnalyzer())
	globalAnalyzerRegistry.Register(NewPortInUseAnalyzer())
	globalAnalyzerRegistry.Register(NewConfigLoadAnalyzer())
}

// BeanNotFoundAnalyzer Bean 未找到错误分析器
type BeanNotFoundAnalyzer struct{}

// NewBeanNotFoundAnalyzer 创建 Bean 未找到错误分析器
func NewBeanNotFoundAnalyzer() *BeanNotFoundAnalyzer {
	return &BeanNotFoundAnalyzer{}
}

// CanAnalyze 检查是否能分析该错误
func (a *BeanNotFoundAnalyzer) CanAnalyze(err error) bool {
	return errors.Is(err, core.ErrBeanNotFound) ||
		strings.Contains(err.Error(), "bean not found")
}

// Analyze 分析错误并返回失败报告
func (a *BeanNotFoundAnalyzer) Analyze(err error) *FailureReport {
	return &FailureReport{
		Headline:    "Bean Not Found",
		Description: "无法找到所需的 Bean 实例",
		Action:      "检查 Bean 是否已正确注册，或检查条件装配是否满足",
		Cause:       err.Error(),
		PossibleSolutions: []string{
			"确认 Bean 已通过 Register/Constructor/Factory 注册",
			"检查是否使用了 @Component 标签且包扫描已启用",
			"检查条件装配（OnProperty/OnBean/OnClass）是否满足",
			"确认 Bean 名称拼写正确",
			"检查是否存在循环依赖导致 Bean 未创建",
		},
	}
}

// CircularDependencyAnalyzer 循环依赖错误分析器
type CircularDependencyAnalyzer struct{}

// NewCircularDependencyAnalyzer 创建循环依赖错误分析器
func NewCircularDependencyAnalyzer() *CircularDependencyAnalyzer {
	return &CircularDependencyAnalyzer{}
}

// CanAnalyze 检查是否能分析该错误
func (a *CircularDependencyAnalyzer) CanAnalyze(err error) bool {
	return errors.Is(err, core.ErrCircularDependency) ||
		strings.Contains(err.Error(), "circular dependency")
}

// Analyze 分析错误并返回失败报告
func (a *CircularDependencyAnalyzer) Analyze(err error) *FailureReport {
	return &FailureReport{
		Headline:    "Circular Dependency Detected",
		Description: "检测到循环依赖，Bean A 依赖 Bean B，Bean B 又依赖 Bean A",
		Action:      "使用懒加载注入（lazy）或重新设计依赖关系",
		Cause:       err.Error(),
		PossibleSolutions: []string{
			"在一方使用懒加载注入：`inject:\"name,lazy\"`",
			"使用 Provider 模式延迟获取依赖",
			"重新设计架构，避免循环依赖",
			"使用事件驱动替代直接依赖",
			"引入中间层解耦",
		},
	}
}

// DuplicateBeanAnalyzer 重复 Bean 错误分析器
type DuplicateBeanAnalyzer struct{}

// NewDuplicateBeanAnalyzer 创建重复 Bean 错误分析器
func NewDuplicateBeanAnalyzer() *DuplicateBeanAnalyzer {
	return &DuplicateBeanAnalyzer{}
}

// CanAnalyze 检查是否能分析该错误
func (a *DuplicateBeanAnalyzer) CanAnalyze(err error) bool {
	return errors.Is(err, core.ErrBeanAlreadyExists) ||
		strings.Contains(err.Error(), "bean already exists:")
}

// Analyze 分析错误并返回失败报告
func (a *DuplicateBeanAnalyzer) Analyze(err error) *FailureReport {
	return &FailureReport{
		Headline:    "Duplicate Bean Definition",
		Description: "Bean 定义重复，同一名称或类型已被注册",
		Action:      "检查是否重复注册了相同的 Bean，或使用 Named 区分多个实例",
		Cause:       err.Error(),
		PossibleSolutions: []string{
			"检查是否在多处注册了相同名称的 Bean",
			"使用 Named(\"name\") 为多个实例指定不同名称",
			"使用 Primary() 标记主要实例",
			"检查自动配置是否重复注册",
		},
	}
}

// PortInUseAnalyzer 端口占用错误分析器
type PortInUseAnalyzer struct{}

// NewPortInUseAnalyzer 创建端口占用错误分析器
func NewPortInUseAnalyzer() *PortInUseAnalyzer {
	return &PortInUseAnalyzer{}
}

// CanAnalyze 检查是否能分析该错误
func (a *PortInUseAnalyzer) CanAnalyze(err error) bool {
	return strings.Contains(err.Error(), "address already in use") ||
		strings.Contains(err.Error(), "port") && strings.Contains(err.Error(), "bind")
}

// Analyze 分析错误并返回失败报告
func (a *PortInUseAnalyzer) Analyze(err error) *FailureReport {
	return &FailureReport{
		Headline:    "Port Already in Use",
		Description: "服务器端口已被占用，无法启动",
		Action:      "检查端口是否被其他进程占用，或更换端口",
		Cause:       err.Error(),
		PossibleSolutions: []string{
			"使用 `lsof -i :<port>` 查看占用端口的进程",
			"使用 `kill <pid>` 终止占用进程",
			"修改配置文件中的 server.port 更换端口",
			"设置环境变量 SERVER_PORT 覆盖配置",
		},
	}
}

// ConfigLoadAnalyzer 配置加载错误分析器
type ConfigLoadAnalyzer struct{}

// NewConfigLoadAnalyzer 创建配置加载错误分析器
func NewConfigLoadAnalyzer() *ConfigLoadAnalyzer {
	return &ConfigLoadAnalyzer{}
}

// CanAnalyze 检查是否能分析该错误
func (a *ConfigLoadAnalyzer) CanAnalyze(err error) bool {
	return errors.Is(err, ErrPropertyNotFound) ||
		errors.Is(err, ErrTypeConversion) ||
		strings.Contains(err.Error(), "config") ||
		strings.Contains(err.Error(), "property")
}

// Analyze 分析错误并返回失败报告
func (a *ConfigLoadAnalyzer) Analyze(err error) *FailureReport {
	return &FailureReport{
		Headline:    "Configuration Load Failed",
		Description: "配置加载失败，无法获取所需的配置项",
		Action:      "检查配置文件是否存在，配置项名称是否正确",
		Cause:       err.Error(),
		PossibleSolutions: []string{
			"检查 application.json/yaml 配置文件是否存在",
			"确认配置项名称拼写正确（注意大小写）",
			"检查环境变量是否已正确设置",
			"使用 environment.GetRequiredProperty() 获取必填配置",
			"检查 Profile 配置是否正确激活",
		},
	}
}
