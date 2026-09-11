package boot

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/config"
)

func TestRegisterConfigCenterFactory(t *testing.T) {
	t.Parallel()

	// 保存原始工厂映射
	origFactories := make(map[string]ConfigCenterFactory)
	factoryMutex.RLock()
	for k, v := range configCenterFactories {
		origFactories[k] = v
	}
	factoryMutex.RUnlock()
	t.Cleanup(func() {
		factoryMutex.Lock()
		defer factoryMutex.Unlock()
		configCenterFactories = origFactories
	})

	RegisterConfigCenterFactory("mock", func(ctx context.Context, cfg *config.ConfigCenterConfig) (config.ConfigCenter, error) {
		return &mockConfigCenter{}, nil
	})

	// 重复注册应覆盖，不会 panic
	RegisterConfigCenterFactory("mock", func(ctx context.Context, cfg *config.ConfigCenterConfig) (config.ConfigCenter, error) {
		return &mockConfigCenter{}, nil
	})
}

func TestConfigCenter_Close_Error(t *testing.T) {
	t.Parallel()
	cc := &mockConfigCenter{closeErr: context.Canceled}
	err := cc.Close()
	if err == nil {
		t.Error("Close() should return error")
	}
}

func TestRegisterConfigCenterFactory_NilFactory(t *testing.T) {
	t.Parallel()

	// 保存原始工厂映射
	origFactories := make(map[string]ConfigCenterFactory)
	factoryMutex.RLock()
	for k, v := range configCenterFactories {
		origFactories[k] = v
	}
	factoryMutex.RUnlock()
	t.Cleanup(func() {
		factoryMutex.Lock()
		defer factoryMutex.Unlock()
		configCenterFactories = origFactories
	})

	// 注册 nil 工厂不会 panic，只是存储 nil
	RegisterConfigCenterFactory("nil-factory", nil)
	// 验证工厂已注册
	factoryMutex.RLock()
	_, ok := configCenterFactories["nil-factory"]
	factoryMutex.RUnlock()
	if !ok {
		t.Error("Expected nil-factory to be registered")
	}
}
