// Package boot 提供应用启动器功能，用于 enhance 框架。
package boot

import (
	"fmt"
	"io"
	"os"
)

// DefaultBanner 默认横幅实例
var DefaultBanner = &LegacyBanner{
	lines: []string{
		`
#####  #   #  #   #   ###   #   #  #####  #####
#      ##  #  #   #  #   #  ##  #  #      #    
#####  # # #  #####  #####  # # #  #      #####
#      #  ##  #   #  #   #  #  ##  #      #    
#####  #   #  #   #  #   #  #   #  #####  #####
		`,
	},
}

const defaultBannerText = `
#####  #   #  #   #   ###   #   #  #####  #####
#      ##  #  #   #  #   #  ##  #  #      #    
#####  # # #  #####  #####  # # #  #      #####
#      #  ##  #   #  #   #  #  ##  #      #    
#####  #   #  #   #  #   #  #   #  #####  #####`

// NewBanner 创建旧版横幅
//
// 参数：
//   - lines: ASCII 艺术行列表
func NewBanner(lines []string) *LegacyBanner {
	return &LegacyBanner{lines: lines}
}

// Print 输出横幅到指定 writer
//
// 显示 ASCII 艺术横幅，附带应用名称、版本号和激活的 Profile 信息。
func (b *LegacyBanner) Print(w io.Writer, appName, version string, profiles []string) {
	for _, line := range b.lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			fmt.Printf("[enhance] failed to print banner line: %v\n", err)
			return
		}
	}
	profileStr := ""
	for i, p := range profiles {
		if i > 0 {
			profileStr += ", "
		}
		profileStr += p
	}
	if profileStr == "" {
		profileStr = "default"
	}
	if _, err := fmt.Fprintf(w, ":: %s :: v%s :: profiles(%s)\n\n", appName, version, profileStr); err != nil {
		fmt.Printf("[enhance] failed to print banner info: %v\n", err)
	}
}

// PrintBanner 使用默认横幅输出到 stdout
func PrintBanner(appName, version string, profiles []string) {
	DefaultBanner.Print(os.Stdout, appName, version, profiles)
}
