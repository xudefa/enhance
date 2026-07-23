package cache

import "errors"

var (
	// ErrNotFound 缓存键不存在。
	ErrNotFound = errors.New("cache: key not found")
	// ErrCacheMiss 缓存键已过期或不存在。
	ErrCacheMiss = errors.New("cache: key expired or not found")
)
