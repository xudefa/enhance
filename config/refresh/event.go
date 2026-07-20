package refresh

import "time"

// NewConfigChangeEvent 创建配置变更事件
//
// 参数：
//   - eventType: 事件类型（"modify"、"delete"、"create"）
//   - keys: 变更的配置键列表
//   - oldValues: 变更前的值
//   - newValues: 变更后的值
//   - source: 配置源类型
func NewConfigChangeEvent(eventType string, keys []string, oldValues, newValues map[string]any, source string) ConfigChangeEvent {
	return ConfigChangeEvent{
		EventType: eventType,
		Keys:      keys,
		OldValues: oldValues,
		NewValues: newValues,
		Source:    source,
		timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}
}
