package banner

import (
	"fmt"
	"os"
	"strings"
)

// ASCIIArtBanner ASCII 艺术横幅实现。
//
// 使用自定义的 ASCII 艺术文本作为启动横幅。
// 支持可选的颜色设置（预留功能）。
//
// 示例：
//
//	banner := NewASCIIArtBanner("  ___  \n /   \\ \n|     |", "")
//	banner.Print("1.0.0")
type ASCIIArtBanner struct {
	art   string     // ASCII 艺术文本
	color string     // 显示颜色（预留）
	mode  BannerMode // 输出模式
}

// NewASCIIArtBanner 创建 ASCII 艺术横幅实例。
//
// 参数:
//   - art: ASCII 艺术文本，支持多行文本
//   - color: 显示颜色（预留，当前未使用）
//
// 返回值:
//   - Banner: ASCII 艺术横幅实例，已设置默认控制台输出模式
func NewASCIIArtBanner(art string, color string) Banner {
	return &ASCIIArtBanner{
		art:   art,
		color: color,
		mode:  BannerModeConsole,
	}
}

// Print 打印 ASCII 艺术横幅到标准输出。
//
// 输出 ASCII 艺术文本和版本信息到 os.Stdout。
// 如果 art 为空，则输出简单的版本信息。
//
// 参数 version 为应用版本号，会显示在横幅底部。
func (b *ASCIIArtBanner) Print(version string) error {
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
		return fmt.Errorf("failed to print ascii art banner: %w", err)
	}
	return nil
}

// Mode 返回当前横幅的输出模式。
func (b *ASCIIArtBanner) Mode() BannerMode {
	return b.mode
}

// SetMode 设置横幅的输出模式。
func (b *ASCIIArtBanner) SetMode(mode BannerMode) {
	b.mode = mode
}

// render 渲染 ASCII 艺术横幅。
func (b *ASCIIArtBanner) render(version string) string {
	var sb strings.Builder

	if b.art != "" {
		sb.WriteString(b.art)
		sb.WriteString("\n")
	}

	if version != "" {
		sb.WriteString(fmt.Sprintf(":: Application :: v%s", version))
	}

	return sb.String()
}
