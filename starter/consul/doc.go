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
	// ConsulEnabled 是否启用 Consul。
	ConsulEnabled = "consul.enabled"
	// ConsulHost Consul 服务器地址。
	ConsulHost = "consul.host"
	// ConsulPort Consul 服务器端口。
	ConsulPort = "consul.port"
	// ConsulToken Consul 访问令牌。
	ConsulToken = "consul.token"
)

// ==================== 默认值常量 ====================

const (
	// DefaultConsulHost 默认 Consul 主机地址。
	DefaultConsulHost = "localhost"
	// DefaultConsulPort 默认 Consul 端口。
	DefaultConsulPort = 8500

	// 条件值常量
	ConditionTrue = "true"
)
