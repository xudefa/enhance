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

	writeStartupReportHeader(&sb)
	writeStartupAppInfo(&sb, r.appName, r.version, r.elapsed)
	writeStartupStats(&sb, startupStats{
		beanCount:       r.beanCount,
		autoConfigCount: r.autoConfigCount,
		starterCount:    r.starterCount,
		moduleCount:     r.moduleCount,
	})

	// 已启动的 Starter
	writeStartupList(&sb, "  🔌 已启动的 Starters:\n", r.sortStrings(r.starters))

	// 已安装的模块
	writeStartupList(&sb, "  📦 已安装的模块:\n", r.sortStrings(r.modules))

	// 警告信息
	writeStartupWarnings(&sb, r.warnings)

	// 网络信息
	writeStartupNetInfo(&sb, r.serverAddr, r.actuatorAddr)

	sb.WriteString("\n")
	return sb.String()
}

// writeStartupReportHeader 写入启动报告边框和标题。
func writeStartupReportHeader(sb *strings.Builder) {
	sb.WriteString("\n")
	sb.WriteString("┌─────────────────────────────────────────────────────────────┐\n")
	sb.WriteString("│                     🚀 enhance 启动报告                      │\n")
	sb.WriteString("└─────────────────────────────────────────────────────────────┘\n\n")
}

// writeStartupAppInfo 写入应用名称、版本和启动耗时。
func writeStartupAppInfo(sb *strings.Builder, appName string, version string, elapsed time.Duration) {
	sb.WriteString(fmt.Sprintf("  📦 应用: %s v%s\n", appName, version))
	sb.WriteString(fmt.Sprintf("  ⏱️  启动耗时: %s\n\n", elapsed.Round(time.Millisecond)))
}

// startupStats 启动统计信息，用于 writeStartupStats。
type startupStats struct {
	beanCount       int
	autoConfigCount int
	starterCount    int
	moduleCount     int
}

// writeStartupStats 写入 Bean、自动配置、Starters、模块的统计信息。
func writeStartupStats(sb *strings.Builder, stats startupStats) {
	sb.WriteString("  📊 统计信息:\n")
	sb.WriteString(fmt.Sprintf("    ✅ IoC 容器: %d 个 Bean 已注册\n", stats.beanCount))
	sb.WriteString(fmt.Sprintf("    ✅ 自动配置: %d 个配置已应用\n", stats.autoConfigCount))
	sb.WriteString(fmt.Sprintf("    ✅ Starters: %d 个已启动\n", stats.starterCount))
	sb.WriteString(fmt.Sprintf("    ✅ 模块: %d 个已安装\n\n", stats.moduleCount))
}

// writeStartupList 写入一个已排序的列表节（列表为空时无输出）。
func writeStartupList(sb *strings.Builder, header string, items []string) {
	if len(items) == 0 {
		return
	}
	sb.WriteString(header)
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("    ✅ %s\n", item))
	}
	sb.WriteString("\n")
}

// writeStartupWarnings 写入警告信息列表。
func writeStartupWarnings(sb *strings.Builder, warnings []string) {
	if len(warnings) == 0 {
		return
	}
	sb.WriteString("  ⚠️  警告:\n")
	for _, w := range warnings {
		sb.WriteString(fmt.Sprintf("    ⚠️  %s\n", w))
	}
	sb.WriteString("\n")
}

// writeStartupNetInfo 写入服务地址和 Actuator 地址（有值时输出）。
func writeStartupNetInfo(sb *strings.Builder, serverAddr string, actuatorAddr string) {
	if serverAddr != "" {
		sb.WriteString(fmt.Sprintf("  🌐 服务地址: http://%s\n", serverAddr))
	}
	if actuatorAddr != "" {
		sb.WriteString(fmt.Sprintf("  📝 Actuator: http://%s/actuator\n", actuatorAddr))
	}
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

// globalStartupReportMu 保护 globalStartupReport 指针替换的读写锁。
// GetStartupReport 可能被并发调用（如并行测试各自启动应用），
// ResetStartupReport 会在其中裸替换指针——若无锁即在同一地址产生
// 数据竞争（go test -race 可复现）。各报告实例的字段写入由自身 mu
// 保护，这里只需保证指针本身的读写原子可见。
var globalStartupReportMu sync.RWMutex

// GetStartupReport 获取全局启动报告实例
func GetStartupReport() *StartupReport {
	globalStartupReportMu.RLock()
	defer globalStartupReportMu.RUnlock()
	return globalStartupReport
}

// ResetStartupReport 重置全局启动报告
func ResetStartupReport() {
	globalStartupReportMu.Lock()
	defer globalStartupReportMu.Unlock()
	globalStartupReport = NewStartupReport()
}
