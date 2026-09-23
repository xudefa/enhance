package boot

import "errors"

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

// 失败分析器实现已按"一实现一文件"原则拆分到独立文件中：
//   - bean_not_found_analyzer.go: BeanNotFoundAnalyzer
//   - circular_dependency_analyzer.go: CircularDependencyAnalyzer
//   - duplicate_bean_analyzer.go: DuplicateBeanAnalyzer
//   - port_in_use_analyzer.go: PortInUseAnalyzer
//   - config_load_analyzer.go: ConfigLoadAnalyzer
