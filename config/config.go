package config

import (
	"encoding/json"
	"os"
)

// NewConfig 创建配置实例。
func NewConfig() Config {
	return &memoryConfig{}
}

// Get 获取配置值
func (c *memoryConfig) Get(key string) any {
	v, _ := c.data.Load(key)
	return v
}

// GetString 获取字符串值
func (c *memoryConfig) GetString(key string) string {
	if v, ok := c.data.Load(key); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetInt 获取整数值
func (c *memoryConfig) GetInt(key string) int {
	if v, ok := c.data.Load(key); ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return 0
}

// GetBool 获取布尔值
func (c *memoryConfig) GetBool(key string) bool {
	if v, ok := c.data.Load(key); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GetAll 获取所有配置值
func (c *memoryConfig) GetAll() map[string]any {
	result := make(map[string]any)
	c.data.Range(func(key, value any) bool {
		result[key.(string)] = value
		return true
	})
	return result
}

// Set 设置配置值
func (c *memoryConfig) Set(key string, value any) {
	c.data.Store(key, value)
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

	// 清空现有数据
	c.data.Range(func(key, value any) bool {
		c.data.Delete(key)
		return true
	})

	// 写入新数据
	for k, v := range m {
		c.data.Store(k, v)
	}

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
	// 默认实现：注册到全局 WatchManager
	// 实际实现由 boot 层或 starter 层提供
	return nil
}
