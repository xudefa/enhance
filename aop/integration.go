package aop

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// AopIntegration AOP集成器
//
// 提供代码生成和运行时AOP的统一集成
type AopIntegration struct {
	config            *AopConfig
	manager           *AopManager
	proxyFactory      *GeneratedProxyFactory
	metadataExtractor *AspectMetadataExtractor
	scannedProxies    map[string]string // 扫描到的代理类型映射（typeName -> filePath）
	scannedMu         sync.RWMutex
}

// NewAopIntegration 创建AOP集成器
func NewAopIntegration(config *AopConfig) *AopIntegration {
	if config == nil {
		config = DefaultAopConfig()
	}

	return &AopIntegration{
		config:            config,
		manager:           GlobalAopManager,
		proxyFactory:      NewGeneratedProxyFactory(),
		metadataExtractor: NewAspectMetadataExtractor(),
		scannedProxies:    make(map[string]string),
	}
}

// GetManager 获取AOP管理器
func (i *AopIntegration) GetManager() *AopManager {
	return i.manager
}

// GetProxyFactory 获取代理工厂
func (i *AopIntegration) GetProxyFactory() *GeneratedProxyFactory {
	return i.proxyFactory
}

// GetMetadataExtractor 获取元数据提取器
func (i *AopIntegration) GetMetadataExtractor() *AspectMetadataExtractor {
	return i.metadataExtractor
}

// CreateProxy 创建代理对象
func (i *AopIntegration) CreateProxy(beanID string, target any) any {
	// 使用集成器自身配置，而非共享 manager 的配置，
	// 避免多个容器互相覆盖配置并造成数据竞争。
	cfg := i.config
	if cfg == nil {
		return target
	}
	switch cfg.Mode {
	case AopModeGenerated:
		return i.proxyFactory.CreateOrFallback(beanID, target, cfg.Weaver)
	case AopModeRuntime:
		return cfg.Weaver.Weave(target)
	case AopModeMixed:
		// 优先使用代码生成的代理
		if HasGeneratedProxy(beanID) {
			proxy, err := i.proxyFactory.Create(beanID, target)
			if err == nil {
				return proxy
			}
		}
		// 回退到运行时代理
		return cfg.Weaver.Weave(target)
	default:
		return target
	}
}

// RegisterAspect 注册切面
func (i *AopIntegration) RegisterAspect(aspect *AspectMeta) {
	i.manager.RegisterAspect(aspect)
}

// RegisterAspects 批量注册切面
func (i *AopIntegration) RegisterAspects(aspects ...*AspectMeta) {
	i.manager.RegisterAspects(aspects...)
}

// GetAspects 获取所有切面
func (i *AopIntegration) GetAspects() []*AspectMeta {
	return i.manager.GetAspects()
}

// RegisterProxyType 注册代理类型（供扫描器使用）
func (i *AopIntegration) RegisterProxyType(typeName string, filePath string) {
	i.scannedMu.Lock()
	defer i.scannedMu.Unlock()
	i.scannedProxies[typeName] = filePath
}

// GetScannedProxy 获取扫描到的代理类型文件路径
func (i *AopIntegration) GetScannedProxy(typeName string) (string, bool) {
	i.scannedMu.RLock()
	defer i.scannedMu.RUnlock()
	path, ok := i.scannedProxies[typeName]
	return path, ok
}

// GlobalAopIntegration 全局AOP集成器（原子访问）
var globalAopIntegration atomic.Pointer[AopIntegration]

func init() {
	globalAopIntegration.Store(NewAopIntegration(nil))
}

// GetGlobalAopIntegration 获取全局AOP集成器
func GetGlobalAopIntegration() *AopIntegration {
	return globalAopIntegration.Load()
}

// SetGlobalAopIntegration 设置全局AOP集成器
func SetGlobalAopIntegration(i *AopIntegration) {
	globalAopIntegration.Store(i)
}

// CreateProxy 创建代理对象（使用全局集成器）
func CreateProxy(beanID string, target any) any {
	return GetGlobalAopIntegration().CreateProxy(beanID, target)
}

// RegisterAspectToGlobal 注册切面到全局集成器
func RegisterAspectToGlobal(aspect *AspectMeta) {
	GetGlobalAopIntegration().RegisterAspect(aspect)
}

// GetGlobalAspects 获取全局切面
func GetGlobalAspects() []*AspectMeta {
	return GetGlobalAopIntegration().GetAspects()
}

// AutoRegister 自动注册切面
//
// 从代码生成的代理中自动提取并注册切面
func AutoRegister(beanID string) error {
	aspects := GetGlobalAopIntegration().GetMetadataExtractor().ExtractFromBeanID(beanID)
	if len(aspects) == 0 {
		return fmt.Errorf("no aspects found for bean: %s", beanID)
	}

	GetGlobalAopIntegration().RegisterAspects(aspects...)
	return nil
}

// AutoRegisterAll 自动注册所有切面
//
// 从所有代码生成的代理中自动提取并注册切面
func AutoRegisterAll() error {
	beanIDs := GlobalGeneratedProxyRegistry.List()
	for _, beanID := range beanIDs {
		if err := AutoRegister(beanID); err != nil {
			// 继续处理其他bean，不中断
			continue
		}
	}
	return nil
}

// BuildTagChecker 构建标签检查器。
//
// 通过编译期注入的布尔常量判断构建标签，避免运行时解析开销。
// 使用方式：在包含 goaop 标签的文件中定义 //go:build goaop，并设置
//
//	const hasGoAopBuildTag = true
//
// 默认文件（无标签）中设置
//
//	const hasGoAopBuildTag = false
type BuildTagChecker struct{}

// NewBuildTagChecker 创建构建标签检查器
func NewBuildTagChecker() *BuildTagChecker {
	return &BuildTagChecker{}
}

// HasTag 检查是否有指定标签。
//
// 目前仅支持 "goaop" 标签，通过 hasGoAopBuildTag 编译期常量判断。
func (c *BuildTagChecker) HasTag(tag string) bool {
	if tag == "goaop" {
		return hasGoAopBuildTag
	}
	return false
}

// IsGeneratedMode 检查是否为代码生成模式
func (c *BuildTagChecker) IsGeneratedMode() bool {
	return c.HasTag("goaop")
}

// IsRuntimeMode 检查是否为运行时模式
func (c *BuildTagChecker) IsRuntimeMode() bool {
	return !c.IsGeneratedMode()
}

// GetOptimalMode 获取最优模式
func (c *BuildTagChecker) GetOptimalMode() AopMode {
	if c.IsGeneratedMode() {
		return AopModeGenerated
	}
	return AopModeRuntime
}

// GlobalBuildTagChecker 全局构建标签检查器
var GlobalBuildTagChecker = NewBuildTagChecker()

// DetectOptimalMode 检测最优AOP模式
func DetectOptimalMode() AopMode {
	return GlobalBuildTagChecker.GetOptimalMode()
}

// ConfigureAopManager 配置AOP管理器
//
// 根据构建标签自动配置最优的AOP模式
func ConfigureAopManager() *AopConfig {
	config := DefaultAopConfig()
	config.Mode = DetectOptimalMode()
	return config
}

// InitializeAop 初始化AOP
//
// 自动配置并初始化AOP系统
func InitializeAop() {
	config := ConfigureAopManager()
	SetGlobalAopIntegration(NewAopIntegration(config))

	// 如果是代码生成模式，自动注册切面
	if config.Mode == AopModeGenerated || config.Mode == AopModeMixed {
		if err := AutoRegisterAll(); err != nil {
			fmt.Printf("[enhance] failed to auto register aspects: %v\n", err)
		}
	}
}

// GetProxyWithAutoMode 使用自动模式获取代理
func GetProxyWithAutoMode(beanID string, target any) any {
	return GetGlobalAopIntegration().CreateProxy(beanID, target)
}
