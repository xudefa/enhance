package boot

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/xudefa/enhance/config"
	"github.com/xudefa/enhance/config/environment"
)

// ConfigCenterFactory 配置中心工厂函数类型
type ConfigCenterFactory func(ctx context.Context, cfg *config.ConfigCenterConfig) (config.ConfigCenter, error)

var (
	configCenterFactories = make(map[string]ConfigCenterFactory)
	factoryMutex          sync.RWMutex
)

// RegisterConfigCenterFactory 注册配置中心工厂函数
func RegisterConfigCenterFactory(centerType string, factory ConfigCenterFactory) {
	factoryMutex.Lock()
	defer factoryMutex.Unlock()
	configCenterFactories[centerType] = factory
}

// loadConfigCenterConfig 从配置中心加载配置
func (b *Boot) loadConfigCenterConfig() error {
	if len(b.config.ConfigCenterAddr) == 0 {
		return fmt.Errorf("config center address is required")
	}

	factoryMutex.RLock()
	factory, ok := configCenterFactories[b.config.ConfigCenterType]
	factoryMutex.RUnlock()

	if !ok {
		return fmt.Errorf("unsupported config center type: %s (no factory registered)", b.config.ConfigCenterType)
	}

	cfg := b.buildConfigCenterConfig()
	ctx, cancel := context.WithTimeout(b.rootCtx, b.config.ConfigCenterTimeout)
	defer cancel()

	center, err := factory(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create config center client: %w", err)
	}
	defer func() {
		if center == nil {
			return
		}
		if closeErr := center.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "[enhance] failed to close config center: %v\n", closeErr)
		}
	}()

	data, err := center.Load()
	if err != nil {
		return fmt.Errorf("failed to load config from center: %w", err)
	}

	if len(data) > 0 {
		source := environment.NewMapPropertySource("config-center", environment.PriorityNormal, data)
		b.ctx.Environment().AddPropertySource(source)
	}

	return nil
}

// buildConfigCenterConfig 构建配置中心配置
func (b *Boot) buildConfigCenterConfig() *config.ConfigCenterConfig {
	cfg := &config.ConfigCenterConfig{
		Endpoints: b.config.ConfigCenterAddr,
		Timeout:   b.config.ConfigCenterTimeout,
	}

	switch b.config.ConfigCenterType {
	case "nacos":
		dataID := b.config.ConfigCenterDataID
		if dataID == "" {
			dataID = "app-config"
		}
		group := b.config.ConfigCenterGroup
		if group == "" {
			group = "DEFAULT_GROUP"
		}
		cfg.DataID = dataID
		cfg.Group = group

	case "etcd", "consul":
		prefix := b.config.ConfigCenterPrefix
		if prefix == "" {
			prefix = "config"
			if b.config.ConfigCenterType == "etcd" {
				prefix = "/config"
			}
		}
		cfg.Prefix = prefix
	default:
		// 未知配置中心类型，仅使用基础配置
	}

	return cfg
}
