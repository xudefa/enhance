package environment

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigType 配置文件类型枚举
type ConfigType string

const (
	// ConfigTypeJSON JSON 配置文件类型
	ConfigTypeJSON ConfigType = "json"
)

// ConfigLoader 配置文件加载器
//
// 在预设搜索路径中查找配置文件，支持基础配置和 Profile 配置的分层加载。
// Profile 配置会覆盖基础配置中的同名属性。
//
// 加载流程：
//  1. 如果指定了 configLocation，直接加载该文件
//  2. 否则搜索 configName.{configType} 作为基础配置
//  3. 依次搜索 configName-{profile}.{configType} 作为 Profile 配置
type ConfigLoader struct {
	configName     string     // 配置文件名（不含扩展名）
	configType     ConfigType // 配置文件类型
	configLocation string     // 自定义配置文件路径（优先级最高）
	profiles       []string   // 激活的 Profile 列表
	searchPaths    []string   // 自定义搜索路径（可选，为空时使用默认路径）
}

// NewConfigLoader 创建配置文件加载器
//
// 参数：
//   - configName: 配置文件名（不含扩展名），如 "application"
//   - configType: 配置文件类型，如 ConfigTypeJSON
//   - configLocation: 自定义配置文件路径，为空时自动搜索
//   - profiles: 激活的 Profile 列表
func NewConfigLoader(configName string, configType ConfigType, configLocation string, profiles []string) *ConfigLoader {
	return &ConfigLoader{
		configName:     configName,
		configType:     configType,
		configLocation: configLocation,
		profiles:       profiles,
	}
}

// ConfigLoaderOption 配置加载器的可选参数。
type ConfigLoaderOption func(*configLoaderOptions)

// configLoaderOptions 保存配置加载器的可选参数。
type configLoaderOptions struct {
	configLocation string
	profiles       []string
	searchPaths    []string
}

// WithLoaderLocation 设置自定义配置文件路径，为空时自动搜索。
func WithLoaderLocation(location string) ConfigLoaderOption {
	return func(o *configLoaderOptions) {
		o.configLocation = location
	}
}

// WithLoaderProfiles 设置激活的 Profile 列表。
func WithLoaderProfiles(profiles []string) ConfigLoaderOption {
	return func(o *configLoaderOptions) {
		o.profiles = profiles
	}
}

// WithLoaderSearchPaths 设置自定义搜索路径列表。
func WithLoaderSearchPaths(paths []string) ConfigLoaderOption {
	return func(o *configLoaderOptions) {
		o.searchPaths = paths
	}
}

// NewConfigLoaderWithPaths 创建配置文件加载器并指定搜索路径
//
// 参数：
//   - configName: 配置文件名（不含扩展名），如 "application"
//   - configType: 配置文件类型，如 ConfigTypeJSON
//   - opts: 可选参数（WithLoaderLocation / WithLoaderProfiles / WithLoaderSearchPaths）
func NewConfigLoaderWithPaths(configName string, configType ConfigType, opts ...ConfigLoaderOption) *ConfigLoader {
	options := &configLoaderOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return &ConfigLoader{
		configName:     configName,
		configType:     configType,
		configLocation: options.configLocation,
		profiles:       options.profiles,
		searchPaths:    options.searchPaths,
	}
}

// Load 加载配置文件并返回配置源列表
//
// 返回的配置源按加载顺序排列：基础配置在前，Profile 配置在后。
// Profile 配置的优先级高于基础配置。
func (l *ConfigLoader) Load() ([]PropertySource, error) {
	var sources []PropertySource

	if l.configLocation != "" {
		source, err := l.loadConfigFile(l.configLocation, "custom-config")
		if err != nil {
			return nil, fmt.Errorf("加载自定义配置文件 %s 失败: %w", l.configLocation, err)
		}
		sources = append(sources, source)
		return sources, nil
	}

	baseConfigPath, err := l.findConfigFile(l.configName)
	if err != nil {
		return nil, fmt.Errorf("查找配置文件 %s 失败: %w", l.configName, err)
	}

	if baseConfigPath != "" {
		baseSource, err := l.loadConfigFile(baseConfigPath, "base-config")
		if err != nil {
			return nil, fmt.Errorf("failed to load base config: %w", err)
		}
		sources = append(sources, baseSource)
	}

	for _, profile := range l.profiles {
		profileConfigName := fmt.Sprintf("%s-%s", l.configName, profile)
		profileConfigPath, err := l.findConfigFile(profileConfigName)
		if err != nil {
			return nil, fmt.Errorf("查找 profile 配置文件 %s 失败: %w", profileConfigName, err)
		}

		if profileConfigPath != "" {
			profileSource, err := l.loadConfigFile(profileConfigPath, fmt.Sprintf("profile-config-%s", profile))
			if err != nil {
				return nil, fmt.Errorf("failed to load profile config %s: %w", profile, err)
			}
			sources = append(sources, profileSource)
		}
	}

	return sources, nil
}

// findConfigFile 在预设搜索路径中查找配置文件
//
// 搜索路径优先级：
//  1. 自定义搜索路径（如果指定）
//  2. /etc/config
//  3. 当前目录
//  4. ./config
//  5. 可执行文件所在目录及其 ./config 子目录
func (l *ConfigLoader) findConfigFile(configName string) (string, error) {
	var searchPaths []string

	if len(l.searchPaths) > 0 {
		searchPaths = l.searchPaths
	} else {
		searchPaths = []string{
			"/etc/config",
			".",
			"./config",
		}
		if exePath, err := os.Executable(); err == nil {
			exeDir := filepath.Dir(exePath)
			searchPaths = append(searchPaths, exeDir, filepath.Join(exeDir, "config"), filepath.Join(exeDir, "src", "config"))
		}
	}

	for _, path := range searchPaths {
		configFile := filepath.Join(path, fmt.Sprintf("%s.%s", configName, l.configType))
		if _, err := os.Stat(configFile); err == nil {
			return configFile, nil
		}
	}

	return "", nil
}

// loadConfigFile 加载单个配置文件为 PropertySource
func (l *ConfigLoader) loadConfigFile(filePath, sourceName string) (PropertySource, error) {
	return NewJSONPropertySource(sourceName, filePath)
}

// GetConfigFileExtension 根据配置类型返回文件扩展名
func GetConfigFileExtension(configType ConfigType) string {
	return string(configType)
}

// ParseConfigType 根据文件路径解析配置类型（当前仅支持 JSON）
func ParseConfigType(filePath string) ConfigType {
	return ConfigTypeJSON
}
