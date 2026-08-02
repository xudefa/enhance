package generator

import (
	"encoding/json"
	"os"
	"sync"
)

// Registry 代理注册表
//
// 记录 Bean 标识与生成文件路径的映射关系，支持持久化到 JSON 文件。
// 使用 sync.Map 优化并发读取性能（读多写少场景）。
type Registry struct {
	proxies sync.Map // string -> string，Bean 标识 -> 代理文件路径映射
}

// NewRegistry 创建代理注册表
func NewRegistry() *Registry {
	return &Registry{}
}

// Register 注册代理映射
func (r *Registry) Register(beanID, filePath string) {
	r.proxies.Store(beanID, filePath)
}

// Get 根据 Bean 标识获取代理文件路径
func (r *Registry) Get(beanID string) (string, bool) {
	value, ok := r.proxies.Load(beanID)
	if !ok {
		return "", false
	}
	s, _ := value.(string)
	return s, true
}

// List 获取所有代理映射的副本
func (r *Registry) List() map[string]string {
	result := make(map[string]string)
	r.proxies.Range(func(key, value any) bool {
		k, _ := key.(string)
		v, _ := value.(string)
		result[k] = v
		return true
	})
	return result
}

// Save 将注册表持久化到 JSON 文件
func (r *Registry) Save(filePath string) error {
	data, err := json.MarshalIndent(r.List(), "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// Load 从 JSON 文件加载注册表
func (r *Registry) Load(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var proxies map[string]string
	if err := json.Unmarshal(data, &proxies); err != nil {
		return err
	}

	// 清空现有数据并加载新数据
	r.proxies.Range(func(key, value any) bool {
		r.proxies.Delete(key)
		return true
	})
	for k, v := range proxies {
		r.proxies.Store(k, v)
	}
	return nil
}

// Clear 清空注册表
func (r *Registry) Clear() {
	r.proxies.Range(func(key, value any) bool {
		r.proxies.Delete(key)
		return true
	})
}
