package banner

import (
	"fmt"
	"os"
)

// String 返回横幅模式的字符串表示。
func (m BannerMode) String() string {
	switch m {
	case BannerModeConsole:
		return "console"
	case BannerModeLog:
		return "log"
	case BannerModeOff:
		return "off"
	default:
		return "unknown"
	}
}

// printBanner 横幅打印公共逻辑，消除 TextBanner/LegacyBanner/ASCIIArtBanner 重复
func printBanner(mode BannerMode, text string, name string) error {
	if mode == BannerModeOff {
		return nil
	}
	if mode == BannerModeLog {
		fmt.Printf("[enhance] %s\n", text)
		return nil
	}
	if _, err := fmt.Fprintln(os.Stdout, text); err != nil {
		return fmt.Errorf("failed to print %s: %w", name, err)
	}
	return nil
}
