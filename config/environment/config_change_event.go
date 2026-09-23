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

// configChangeEventOptions 保存配置变更事件的可选参数。
type configChangeEventOptions struct {
	keys      []string
	oldValues map[string]any
	newValues map[string]any
	source    string
}

// ConfigChangeEventOption 配置变更事件的可选参数。
type ConfigChangeEventOption func(*configChangeEventOptions)

// WithEventKeys 设置变更的配置键列表。
func WithEventKeys(keys []string) ConfigChangeEventOption {
	return func(o *configChangeEventOptions) {
		o.keys = keys
	}
}

// WithEventValues 设置变更前后的值。
func WithEventValues(oldValues, newValues map[string]any) ConfigChangeEventOption {
	return func(o *configChangeEventOptions) {
		o.oldValues = oldValues
		o.newValues = newValues
	}
}

// WithEventSource 设置配置源类型（如 "nacos"、"etcd"）。
func WithEventSource(source string) ConfigChangeEventOption {
	return func(o *configChangeEventOptions) {
		o.source = source
	}
}

// NewConfigChangeEvent 创建配置变更事件
//
// 参数：
//   - eventType: 事件类型（"modify"、"delete"、"create"）
//   - opts: 可选参数（WithEventKeys / WithEventValues / WithEventSource）
func NewConfigChangeEvent(eventType string, opts ...ConfigChangeEventOption) ConfigChangeEvent {
	options := &configChangeEventOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return ConfigChangeEvent{
		EventType: eventType,
		Keys:      options.keys,
		OldValues: options.oldValues,
		NewValues: options.newValues,
		Source:    options.source,
		timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}
}
