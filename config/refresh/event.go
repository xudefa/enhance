package refresh

import (
	"github.com/xudefa/enhance/config/environment"
)

// ConfigChangeEventOption refresh 配置变更事件的可选参数。
type ConfigChangeEventOption = environment.ConfigChangeEventOption

// NewConfigChangeEvent 创建配置变更事件
//
// 参数：
//   - eventType: 事件类型（"modify"、"delete"、"create"）
//   - opts: 可选参数（environment.WithEventKeys / WithEventValues / WithEventSource）
func NewConfigChangeEvent(eventType string, opts ...ConfigChangeEventOption) environment.ConfigChangeEvent {
	return environment.NewConfigChangeEvent(eventType, opts...)
}
