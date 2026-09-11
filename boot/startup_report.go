package boot

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// StartupReport 启动报告，提供友好的启动信息输出。
//
// 参考 Spring Boot 的 StartupInfoLogger，在应用启动完成后输出：
//   - 应用名称和版本
//   - 启动耗时
//   - 已注册的 Bean 数量
//   - 已启用的自动配置
//   - 已启动的 Starter
//   - 已安装的模块
//   - 警告信息（如未使用的 Bean）
//   - 服务地址和 Actuator 端点
type StartupReport struct {
	mu sync.Mutex

	// 基本信息
	appName   string
	version   string
	startTime time.Time
	elapsed   time.Duration

	// 统计信息
	beanCount       int
	autoConfigCount int
	starterCount    int
	moduleCount     int

	// 详细信息
	starters    []string
	modules     []string
	autoConfigs []string
	warnings    []string

	// 网络信息
	serverAddr   string
	actuatorAddr string

	// 开关
	enabled bool
}

// NewStartupReport 创建启动报告
func NewStartupReport() *StartupReport {
	return &StartupReport{
		starters:    make([]string, 0),
		modules:     make([]string, 0),
		autoConfigs: make([]string, 0),
		warnings:    make([]string, 0),
		enabled:     true,
	}
}

// SetAppInfo 设置应用基本信息
func (r *StartupReport) SetAppInfo(name, version string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.appName = name
	r.version = version
}

// StartTiming 开始计时
func (r *StartupReport) StartTiming() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.startTime = time.Now()
}

// StopTiming 停止计时
func (r *StartupReport) StopTiming() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.elapsed = time.Since(r.startTime)
}

// SetBeanCount 设置 Bean 数量
func (r *StartupReport) SetBeanCount(count int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.beanCount = count
}

// SetAutoConfigCount 设置自动配置数量
func (r *StartupReport) SetAutoConfigCount(count int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.autoConfigCount = count
}

// SetStarterCount 设置 Starter 数量
func (r *StartupReport) SetStarterCount(count int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.starterCount = count
}

// SetModuleCount 设置模块数量
func (r *StartupReport) SetModuleCount(count int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.moduleCount = count
}

// AddStarter 添加已启动的 Starter
func (r *StartupReport) AddStarter(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.starters = append(r.starters, name)
}

// AddModule 添加已安装的模块
func (r *StartupReport) AddModule(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.modules = append(r.modules, name)
}

// AddAutoConfig 添加已启用的自动配置
func (r *StartupReport) AddAutoConfig(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.autoConfigs = append(r.autoConfigs, name)
}

// AddWarning 添加警告信息
func (r *StartupReport) AddWarning(warning string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.warnings = append(r.warnings, warning)
}

// SetServerAddr 设置服务器地址
func (r *StartupReport) SetServerAddr(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.serverAddr = addr
}

// SetActuatorAddr 设置 Actuator 地址
func (r *StartupReport) SetActuatorAddr(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.actuatorAddr = addr
}

// Disable 禁用启动报告
func (r *StartupReport) Disable() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled = false
}

// Print 打印启动报告
func (r *StartupReport) Print() {
	r.mu.Lock()
	if !r.enabled {
		r.mu.Unlock()
		return
	}

	// 复制数据，避免长时间持有锁
	report := &StartupReport{
		appName:         r.appName,
		version:         r.version,
		elapsed:         r.elapsed,
		beanCount:       r.beanCount,
		autoConfigCount: r.autoConfigCount,
		starterCount:    r.starterCount,
		moduleCount:     r.moduleCount,
		starters:        make([]string, len(r.starters)),
		modules:         make([]string, len(r.modules)),
		autoConfigs:     make([]string, len(r.autoConfigs)),
		warnings:        make([]string, len(r.warnings)),
		serverAddr:      r.serverAddr,
		actuatorAddr:    r.actuatorAddr,
		enabled:         r.enabled,
	}
	copy(report.starters, r.starters)
	copy(report.modules, r.modules)
	copy(report.autoConfigs, r.autoConfigs)
	copy(report.warnings, r.warnings)
	r.mu.Unlock()

	fmt.Println(report.format())
}

// format 格式化启动报告
func (r *StartupReport) format() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("┌─────────────────────────────────────────────────────────────┐\n")
	sb.WriteString("│                     🚀 enhance 启动报告                      │\n")
	sb.WriteString("└─────────────────────────────────────────────────────────────┘\n\n")

	// 应用信息
	sb.WriteString(fmt.Sprintf("  📦 应用: %s v%s\n", r.appName, r.version))
	sb.WriteString(fmt.Sprintf("  ⏱️  启动耗时: %s\n\n", r.elapsed.Round(time.Millisecond)))

	// 统计信息
	sb.WriteString("  📊 统计信息:\n")
	sb.WriteString(fmt.Sprintf("    ✅ IoC 容器: %d 个 Bean 已注册\n", r.beanCount))
	sb.WriteString(fmt.Sprintf("    ✅ 自动配置: %d 个配置已应用\n", r.autoConfigCount))
	sb.WriteString(fmt.Sprintf("    ✅ Starters: %d 个已启动\n", r.starterCount))
	sb.WriteString(fmt.Sprintf("    ✅ 模块: %d 个已安装\n\n", r.moduleCount))

	// 已启动的 Starter
	if len(r.starters) > 0 {
		sb.WriteString("  🔌 已启动的 Starters:\n")
		for _, s := range r.sortStrings(r.starters) {
			sb.WriteString(fmt.Sprintf("    ✅ %s\n", s))
		}
		sb.WriteString("\n")
	}

	// 已安装的模块
	if len(r.modules) > 0 {
		sb.WriteString("  📦 已安装的模块:\n")
		for _, m := range r.sortStrings(r.modules) {
			sb.WriteString(fmt.Sprintf("    ✅ %s\n", m))
		}
		sb.WriteString("\n")
	}

	// 警告信息
	if len(r.warnings) > 0 {
		sb.WriteString("  ⚠️  警告:\n")
		for _, w := range r.warnings {
			sb.WriteString(fmt.Sprintf("    ⚠️  %s\n", w))
		}
		sb.WriteString("\n")
	}

	// 网络信息
	if r.serverAddr != "" {
		sb.WriteString(fmt.Sprintf("  🌐 服务地址: http://%s\n", r.serverAddr))
	}
	if r.actuatorAddr != "" {
		sb.WriteString(fmt.Sprintf("  📝 Actuator: http://%s/actuator\n", r.actuatorAddr))
	}

	sb.WriteString("\n")
	return sb.String()
}

// sortStrings 对字符串切片排序
func (r *StartupReport) sortStrings(ss []string) []string {
	sorted := make([]string, len(ss))
	copy(sorted, ss)
	sort.Strings(sorted)
	return sorted
}

// globalStartupReport 全局启动报告实例
var globalStartupReport = NewStartupReport()

// GetStartupReport 获取全局启动报告实例
func GetStartupReport() *StartupReport {
	return globalStartupReport
}

// ResetStartupReport 重置全局启动报告
func ResetStartupReport() {
	globalStartupReport = NewStartupReport()
}
