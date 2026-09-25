package actuator

import (
	"strings"
)

// PathNormalizer 路径标准化工具
type PathNormalizer struct{}

// NormalizePath 标准化路径,确保路径格式正确
func (PathNormalizer) NormalizePath(path string) string {
	if path == "" {
		return "/"
	}

	path = EnsureLeadingSlash(path)

	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	path = collapseDoubleSlashes(path)

	return path
}

// collapseDoubleSlashes 将连续的 // 替换为单个 /，一次遍历完成
func collapseDoubleSlashes(path string) string {
	if !strings.Contains(path, "//") {
		return path
	}
	var b strings.Builder
	b.Grow(len(path))
	prevSlash := false
	for _, r := range path {
		if r == '/' {
			if prevSlash {
				continue
			}
			prevSlash = true
		} else {
			prevSlash = false
		}
		b.WriteRune(r)
	}
	return b.String()
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
