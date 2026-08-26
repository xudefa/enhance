package tenant

// extractSubdomain 提取子域名。
func extractSubdomain(host, baseDomain string) string {
	// 移除端口号
	if idx := indexOf(host, ':'); idx != -1 {
		host = host[:idx]
	}

	// 检查是否以 baseDomain 结尾
	if !endsWith(host, "."+baseDomain) {
		return ""
	}

	// 提取子域名
	subdomain := host[:len(host)-len("."+baseDomain)]
	return subdomain
}

// splitPath 分割路径。
func splitPath(path string) []string {
	// 移除开头的 /
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	if path == "" {
		return []string{}
	}

	segments := make([]string, 0)
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if start < i {
				segments = append(segments, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		segments = append(segments, path[start:])
	}

	return segments
}

// indexOf 查找字符位置。
func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// endsWith 检查是否以指定字符串结尾。
func endsWith(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}
