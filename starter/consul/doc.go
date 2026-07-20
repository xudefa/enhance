// Package consul 提供 Consul 服务发现自动配置。
//
// Consul 是 HashiCorp 推出的服务网格解决方案。
//
// 功能特性：
//   - 自动配置 Consul 客户端
//   - 服务注册与发现
//   - 健康检查
//   - KV 存储
//
// 配置示例：
//
//	{
//	  "consul": {
//	    "enabled": true,
//	    "host": "localhost",
//	    "port": 8500
//	  }
//	}
//
// 使用示例：
//
//	client := core.MustGetBean[*consulapi.Client](app.Container())
//	entries, _, _ := client.Health().Service("web", "", true, nil)
package consul

// ==================== 配置键常量 ====================

const (
	// Consul 配置
	ConsulEnabled = "consul.enabled"
	ConsulHost    = "consul.host"
	ConsulPort    = "consul.port"
	ConsulToken   = "consul.token"
)

// ==================== 默认值常量 ====================

const (
	// Consul 默认值
	DefaultConsulHost = "localhost"
	DefaultConsulPort = 8500

	// 条件值常量
	ConditionTrue = "true"
)
