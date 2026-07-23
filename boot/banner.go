// Package boot 提供应用启动器功能，用于 enhance 框架。
package boot

import (
	"github.com/xudefa/enhance/boot/banner"
)

// DefaultBanner 默认横幅实例
var DefaultBanner banner.Banner = banner.NewLegacyBanner(
	banner.WithLines([]string{
		`
#####  #   #  #   #   ###   #   #  #####  #####
#      ##  #  #   #  #   #  ##  #  #      #    
#####  # # #  #####  #####  # # #  #      #####
#      #  ##  #   #  #   #  #  ##  #      #    
#####  #   #  #   #  #   #  #   #  #####  #####
		`,
	}),
)

const defaultBannerText = `
#####  #   #  #   #   ###   #   #  #####  #####
#      ##  #  #   #  #   #  ##  #  #      #    
#####  # # #  #####  #####  # # #  #      #####
#      #  ##  #   #  #   #  #  ##  #      #    
#####  #   #  #   #  #   #  #   #  #####  #####`

// NewBanner 创建旧版横幅
//
// 参数 opts: 可选配置项，如 banner.WithLines, banner.WithAppName, banner.WithProfiles
func NewBanner(opts ...banner.LegacyOption) banner.Banner {
	return banner.NewLegacyBanner(opts...)
}
