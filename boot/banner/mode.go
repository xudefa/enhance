package banner

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
