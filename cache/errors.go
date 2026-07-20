package cache

import "errors"

var (
	ErrNotFound  = errors.New("cache: key not found")
	ErrCacheMiss = errors.New("cache: key expired or not found")
)
