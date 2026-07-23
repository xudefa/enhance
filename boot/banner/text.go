package banner

import (
	"fmt"
	"os"
	"strings"
)

// TextBanner 文本横幅实现。
//
// 使用模板和属性渲染启动横幅，支持简单的属性替换。
// 模板中的 {{key}} 会被替换为对应的属性值。
//
// 示例：
//
//	banner := NewTextBanner("Hello {{name}} v{{version}}", map[string]any{
//	    "name": "my-app",
//	    "version": "1.0.0",
//	})
//	banner.Print("1.0.0")
type TextBanner struct {
	template   string         // 横幅文本模板
	properties map[string]any // 模板属性键值对
	mode       BannerMode     // 输出模式
}

// NewTextBanner 创建文本横幅实例。
//
// 参数:
//   - template: 横幅文本模板，支持 {{key}} 格式的属性替换
//   - properties: 模板属性键值对，用于替换模板中的占位符
//
// 返回值:
//   - Banner: 文本横幅实例，已设置默认控制台输出模式
func NewTextBanner(template string, properties map[string]any) Banner {
	if properties == nil {
		properties = make(map[string]any)
	}
	return &TextBanner{
		template:   template,
		properties: properties,
		mode:       BannerModeConsole,
	}
}

// Print 打印文本横幅到标准输出。
//
// 渲染模板并输出到 os.Stdout，模板中的 {{key}} 会被替换为属性值。
// 如果模板为空，则输出简单的版本信息。
//
// 参数 version 为应用版本号，会自动添加到 properties 中。
func (b *TextBanner) Print(version string) error {
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
		return fmt.Errorf("failed to print text banner: %w", err)
	}
	return nil
}

// Mode 返回当前横幅的输出模式。
func (b *TextBanner) Mode() BannerMode {
	return b.mode
}

// SetMode 设置横幅的输出模式。
func (b *TextBanner) SetMode(mode BannerMode) {
	b.mode = mode
}

// render 渲染模板，替换占位符。
func (b *TextBanner) render(version string) string {
	text := b.template
	if text == "" {
		return fmt.Sprintf(":: Application :: v%s", version)
	}

	// 自动添加 version 属性
	props := make(map[string]any, len(b.properties)+1)
	for k, v := range b.properties {
		props[k] = v
	}
	props["version"] = version

	// 替换 {{key}} 占位符
	for key, value := range props {
		placeholder := fmt.Sprintf("{{%s}}", key)
		text = strings.ReplaceAll(text, placeholder, fmt.Sprintf("%v", value))
	}

	return text
}
