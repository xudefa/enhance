package config

import (
	"encoding/json"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/xudefa/enhance/config/environment"
)

// globalWatchers stores Watch callbacks keyed by key pattern.
var (
	globalWatchers  = make(map[string][]WatchCallback)
	globalWatcherMu sync.Mutex
)

// NewConfig 创建配置实例。
func NewConfig() Config {
	return &memoryConfig{data: make(map[string]any)}
}

// Get 获取配置值
func (c *memoryConfig) Get(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key]
}

// GetString 获取字符串值
//
// 支持字符串及可从其他类型转换的值（如 JSON 加载后的 float64 数值）。
func (c *memoryConfig) GetString(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.data[key]; ok {
		converted, err := environment.NewTypeConverter().ConvertTo(v, reflect.TypeOf(""))
		if err == nil {
			return converted.String()
		}
	}
	return ""
}

// GetInt 获取整数值
//
// 支持整数及可从其他数值类型转换的值（如 JSON 加载后的 float64）。
func (c *memoryConfig) GetInt(key string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.data[key]; ok {
		converted, err := environment.NewTypeConverter().ConvertTo(v, reflect.TypeOf(0))
		if err == nil {
			return converted.Interface().(int)
		}
	}
	return 0
}

// GetBool 获取布尔值
func (c *memoryConfig) GetBool(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.data[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GetAll 获取所有配置值
func (c *memoryConfig) GetAll() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]any, len(c.data))
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// Set 设置配置值
func (c *memoryConfig) Set(key string, value any) {
	c.mu.Lock()
	c.data[key] = value
	c.mu.Unlock()

	globalWatcherMu.Lock()
	list := globalWatchers[key]
	globalWatcherMu.Unlock()
	if len(list) == 0 {
		return
	}

	event := WatchEvent{
		Type:      EventModify,
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
		Source:    "memoryConfig",
	}
	for _, cb := range list {
		cb(event)
	}
}

// Load 从文件加载配置
func (c *memoryConfig) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	// 原子替换整个 map，读者不会看到中间状态
	c.mu.Lock()
	c.data = m
	c.mu.Unlock()

	return nil
}

// Save 保存配置到文件
func (c *memoryConfig) Save(path string) error {
	data := c.GetAll()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, jsonData, 0644)
}

// Watch 注册配置变更监听器。
//
// 当指定 key 的配置值发生变化时，触发回调函数。
//
// 示例:
//
//	config.Watch("database.url", func(event config.WatchEvent) {
//	    log.Printf("配置变更: %s -> %v", event.Key, event.Value)
//	})
func Watch(key string, callback WatchCallback) error {
	globalWatcherMu.Lock()
	globalWatchers[key] = append(globalWatchers[key], callback)
	globalWatcherMu.Unlock()
	return nil
}
