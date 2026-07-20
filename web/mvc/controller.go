// Package mvc 提供 MVC 控制器支持。
package mvc

import (
	"sync"

	"github.com/xudefa/enhance/web/core"
)

var (
	controllers []core.Controller
	mu          sync.RWMutex
)

// RegisterController 注册控制器到全局注册表。
//
// 控制器会在应用启动时自动被扫描并注册到路由器。
// 通常在 init() 函数中调用。
func RegisterController(ctrl core.Controller) {
	mu.Lock()
	defer mu.Unlock()
	controllers = append(controllers, ctrl)
}

// GetControllers 获取所有已注册的控制器。
func GetControllers() []core.Controller {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]core.Controller, len(controllers))
	copy(result, controllers)
	return result
}

// ClearControllers 清除所有已注册的控制器(仅用于测试)。
func ClearControllers() {
	mu.Lock()
	defer mu.Unlock()
	controllers = controllers[:0]
}
