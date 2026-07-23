package environment

import "time"

// ConfigChangeEvent 配置变更事件。
type ConfigChangeEvent struct {
	EventType string            // 事件类型："modify"、"delete"、"create"
	Keys      []string          // 变更的配置键列表
	OldValues map[string]any    // 变更前的值
	NewValues map[string]any    // 变更后的值
	Source    string            // 配置源类型（如 "nacos"、"etcd"）
	timestamp time.Time         // 事件发生时间
	Metadata  map[string]string // 额外元数据
}

// Type 返回事件类型标识
func (e *ConfigChangeEvent) Type() string {
	return "ConfigChange"
}

// Timestamp 返回事件发生时间
func (e *ConfigChangeEvent) Timestamp() time.Time {
	return e.timestamp
}

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
