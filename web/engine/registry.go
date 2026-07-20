package engine

import (
	"fmt"
	"sync"

	"github.com/xudefa/enhance/web/core"
)

// Registry 网络引擎注册表。
//
// 管理所有已注册的网络引擎工厂,支持运行时切换。
type Registry struct {
	mu            sync.RWMutex
	factories     map[Type]Factory
	defaultEngine Type
}

// GlobalRegistry 全局网络引擎注册表。
var GlobalRegistry = NewRegistry()

// NewRegistry 创建网络引擎注册表。
func NewRegistry() *Registry {
	return &Registry{
		factories:     make(map[Type]Factory),
		defaultEngine: StdLib,
	}
}

// Register 注册网络引擎工厂。
//
// 参数:
//   - factory: 引擎工厂实例
//
// 注意:
//   - 同一引擎类型只能注册一次,重复注册会 panic
//   - 注册是线程安全的
func (r *Registry) Register(factory Factory) {
	r.mu.Lock()
	defer r.mu.Unlock()

	engineType := factory.Type()
	if _, exists := r.factories[engineType]; exists {
		panic(fmt.Sprintf("engine %s already registered", engineType))
	}

	r.factories[engineType] = factory
}

// Get 获取指定类型的引擎工厂。
//
// 参数:
//   - engineType: 引擎类型
//
// 返回值:
//   - Factory: 引擎工厂实例
//   - error: 引擎未注册时返回错误
func (r *Registry) Get(engineType Type) (Factory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.factories[engineType]
	if !exists {
		return nil, fmt.Errorf("engine %s not registered", engineType)
	}

	return factory, nil
}

// SetDefault 设置默认引擎类型。
func (r *Registry) SetDefault(engineType Type) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultEngine = engineType
}

// GetDefault 获取默认引擎类型。
func (r *Registry) GetDefault() Type {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.defaultEngine
}

// CreateRouter 使用默认引擎创建路由器。
func (r *Registry) CreateRouter() (core.Router, error) {
	factory, err := r.Get(r.GetDefault())
	if err != nil {
		return nil, err
	}
	return factory.CreateRouter()
}

// CreateServer 使用默认引擎创建服务器。
func (r *Registry) CreateServer(opts ...ServerOption) (core.Server, error) {
	factory, err := r.Get(r.GetDefault())
	if err != nil {
		return nil, err
	}
	return factory.CreateServer(opts...)
}

// ListEngines 获取所有已注册的引擎类型。
func (r *Registry) ListEngines() []Type {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]Type, 0, len(r.factories))
	for t := range r.factories {
		types = append(types, t)
	}
	return types
}

// HasEngine 检查引擎是否已注册。
func (r *Registry) HasEngine(engineType Type) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.factories[engineType]
	return exists
}

// WithHost 设置服务器监听地址。
func WithHost(host string) ServerOption {
	return func(c *ServerConfig) {
		c.Host = host
	}
}

// WithPort 设置服务器监听端口。
func WithPort(port int) ServerOption {
	return func(c *ServerConfig) {
		c.Port = port
	}
}

// WithReadTimeout 设置读取超时时间(秒)。
func WithReadTimeout(timeout int) ServerOption {
	return func(c *ServerConfig) {
		c.ReadTimeout = timeout
	}
}

// WithWriteTimeout 设置写入超时时间(秒)。
func WithWriteTimeout(timeout int) ServerOption {
	return func(c *ServerConfig) {
		c.WriteTimeout = timeout
	}
}

// WithIdleTimeout 设置空闲超时时间(秒)。
func WithIdleTimeout(timeout int) ServerOption {
	return func(c *ServerConfig) {
		c.IdleTimeout = timeout
	}
}

// WithTLS 设置 TLS 证书和密钥文件。
func WithTLS(certFile, keyFile string) ServerOption {
	return func(c *ServerConfig) {
		c.TLSCertFile = certFile
		c.TLSKeyFile = keyFile
	}
}

// DefaultServerConfig 返回默认服务器配置。
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Host:         "0.0.0.0",
		Port:         8080,
		ReadTimeout:  30,
		WriteTimeout: 30,
		IdleTimeout:  120,
	}
}
