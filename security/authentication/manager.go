package authentication

import (
	"context"
	"sync"
)

// ProviderManager 认证提供者管理器。
//
// 管理多个 AuthenticationProvider，按顺序尝试认证。
// 执行流程：遍历所有支持的提供者，返回第一个成功认证的结果。
type ProviderManager struct {
	mu        sync.Mutex
	providers []AuthenticationProvider
}

// NewProviderManager 创建认证提供者管理器。
//
// 参数:
//   - providers: 认证提供者列表，按顺序尝试认证
//
// 返回:
//   - AuthenticationManager: 认证管理器接口
func NewProviderManager(providers ...AuthenticationProvider) AuthenticationManager {
	return &ProviderManager{
		providers: providers,
	}
}

// Authenticate 尝试通过配置的提供者进行认证。
//
// 遍历所有提供者，返回第一个成功认证的结果。
func (m *ProviderManager) Authenticate(ctx context.Context, token AuthenticationToken) (Authentication, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error

	for _, provider := range m.providers {
		if !provider.Supports(token) {
			continue
		}
		result, err := provider.Authenticate(ctx, token)
		if err != nil {
			lastErr = err
			continue
		}
		if result != nil {
			return result, nil
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, ErrAuthenticationFailed
}

// AddProvider 向管理器中添加认证提供者。
//
// 添加的提供者将按添加顺序参与后续的认证尝试。
func (m *ProviderManager) AddProvider(provider AuthenticationProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers = append(m.providers, provider)
}
