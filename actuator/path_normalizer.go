package actuator

import "strings"

// PathNormalizer 路径标准化工具
type PathNormalizer struct{}

// NormalizePath 标准化路径,确保路径格式正确
func (PathNormalizer) NormalizePath(path string) string {
	if path == "" {
		return "/"
	}

	// 确保路径以 / 开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 移除末尾的 / (除非路径就是 /)
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	// 将连续的 // 替换为单个 /
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	return path
}

// EnsureLeadingSlash 确保路径以 / 开头
func EnsureLeadingSlash(path string) string {
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

// JoinPath 拼接路径
func JoinPath(base, path string) string {
	base = strings.TrimSuffix(base, "/")
	path = EnsureLeadingSlash(path)
	return base + path
}
