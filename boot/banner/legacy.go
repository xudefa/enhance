package banner

import (
	"fmt"
	"os"
	"strings"
)

// LegacyOption LegacyBanner 的选项函数。
type LegacyOption func(*LegacyBanner)

// WithLines 设置 ASCII 艺术行列表。
func WithLines(lines []string) LegacyOption {
	return func(b *LegacyBanner) {
		b.lines = lines
	}
}

// WithAppName 设置应用名称。
func WithAppName(name string) LegacyOption {
	return func(b *LegacyBanner) {
		b.appName = name
	}
}

// WithProfiles 设置激活的 Profile 列表。
func WithProfiles(profiles []string) LegacyOption {
	return func(b *LegacyBanner) {
		b.profiles = profiles
	}
}

// LegacyBanner 旧版横幅实现。
//
// 兼容 boot 包中的 LegacyBanner 格式，使用预定义的 ASCII 艺术行列表
// 和应用名称、版本号、Profile 信息渲染启动横幅。
//
// 示例：
//
//	banner := NewLegacyBanner(
//	    WithLines([]string{"Line 1", "Line 2"}),
//	    WithAppName("my-app"),
//	    WithProfiles([]string{"dev"}),
//	)
//	banner.Print("1.0.0")
type LegacyBanner struct {
	lines    []string   // ASCII 艺术行列表
	appName  string     // 应用名称
	profiles []string   // 激活的 Profile 列表
	mode     BannerMode // 输出模式
}

// NewLegacyBanner 创建旧版横幅实例。
//
// 参数:
//   - opts: 可选配置项，如 WithLines, WithAppName, WithProfiles
//
// 返回值:
//   - Banner: 旧版横幅实例，已设置默认控制台输出模式
func NewLegacyBanner(opts ...LegacyOption) Banner {
	b := &LegacyBanner{
		mode: BannerModeConsole,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Print 打印旧版横幅到标准输出。
//
// 输出 ASCII 艺术行、应用名称、版本号和 Profile 信息到 os.Stdout。
// 格式为：:: appName :: vversion :: profiles(p1, p2)
//
// 参数 version 为应用版本号，会显示在横幅信息行中。
func (b *LegacyBanner) Print(version string) error {
	if b.mode == BannerModeOff {
		return nil
	}

	text := b.render(version)
	if b.mode == BannerModeLog {
		fmt.Printf("[enhance] %s\n", text)
		return nil
	}

	_, err := fmt.Fprintln(os.Stdout, text)
	if err != nil {
		return fmt.Errorf("failed to print legacy banner: %w", err)
	}
	return nil
}

// Mode 返回当前横幅的输出模式。
func (b *LegacyBanner) Mode() BannerMode {
	return b.mode
}

// SetMode 设置横幅的输出模式。
func (b *LegacyBanner) SetMode(mode BannerMode) {
	b.mode = mode
}

// render 渲染旧版横幅。
func (b *LegacyBanner) render(version string) string {
	var sb strings.Builder

	for _, line := range b.lines {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	appName := b.appName
	if appName == "" {
		appName = "Application"
	}

	profileStr := formatProfiles(b.profiles)
	sb.WriteString(fmt.Sprintf(":: %s :: v%s :: profiles(%s)", appName, version, profileStr))

	return sb.String()
}

// formatProfiles 格式化 Profile 列表为逗号分隔的字符串。
func formatProfiles(profiles []string) string {
	if len(profiles) == 0 {
		return "default"
	}
	return strings.Join(profiles, ", ")
}
