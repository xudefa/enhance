package validation

import (
	"regexp"
	"sync"
)

// regexCache 预编译正则表达式缓存
//
// 性能优化：避免每次验证都重新编译正则表达式，显著减少内存分配。
// 使用 sync.Map 保证并发安全，适合读多写少场景。
var regexCache sync.Map // map[string]*regexp.Regexp

// compileRegex 获取或编译正则表达式
func compileRegex(pattern string) *regexp.Regexp {
	if cached, ok := regexCache.Load(pattern); ok {
		return cached.(*regexp.Regexp)
	}

	// 使用 LoadOrStore 避免并发重复编译
	actual, loaded := regexCache.LoadOrStore(pattern, regexp.MustCompile(pattern))
	if loaded {
		return actual.(*regexp.Regexp)
	}
	return actual.(*regexp.Regexp)
}

// validationErrorsPool 复用 ValidationError 切片
//
// 性能优化：减少验证过程中的内存分配和 GC 压力。
var validationErrorsPool = sync.Pool{
	New: func() any {
		s := make([]ValidationError, 0, 16)
		return &s
	},
}

// acquireValidationErrors 从池中获取 ValidationError 切片
func acquireValidationErrors() *[]ValidationError {
	p := validationErrorsPool.Get().(*[]ValidationError)
	*p = (*p)[:0]
	return p
}

// releaseValidationErrors 归还 ValidationError 切片到池中
func releaseValidationErrors(p *[]ValidationError) {
	for i := range *p {
		(*p)[i] = ValidationError{}
	}
	validationErrorsPool.Put(p)
}
