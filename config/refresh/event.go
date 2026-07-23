package refresh

import "github.com/xudefa/enhance/config/environment"

// NewConfigChangeEvent 创建配置变更事件
//
// 参数：
//   - eventType: 事件类型（"modify"、"delete"、"create"）
//   - keys: 变更的配置键列表
//   - oldValues: 变更前的值
//   - newValues: 变更后的值
//   - source: 配置源类型
func NewConfigChangeEvent(eventType string, keys []string, oldValues, newValues map[string]any, source string) environment.ConfigChangeEvent {
	return environment.NewConfigChangeEvent(eventType, keys, oldValues, newValues, source)
}
