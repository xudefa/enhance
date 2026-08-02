package environment

import (
	"os"
	"strings"
)

// NewMapPropertySource 创建 MapPropertySource
func NewMapPropertySource(name string, priority Priority, data map[string]any) *MapPropertySource {
	if data == nil {
		data = make(map[string]any)
	}
	return &MapPropertySource{name: name, data: data, priority: priority}
}

// NewDefaultPropertySource 创建最低优先级的默认配置源
//
// 使用 PriorityFallback 优先级，被所有其他配置源（含文件配置源）覆盖。
func NewDefaultPropertySource(name string, data map[string]any) *MapPropertySource {
	return NewMapPropertySource(name, PriorityFallback, data)
}

// Name 返回 MapPropertySource 的名称。
func (m *MapPropertySource) Name() string {
	return m.name
}

// GetProperty 从 MapPropertySource 中获取指定键的值。
func (m *MapPropertySource) GetProperty(key string) (any, bool) {
	val, ok := m.data[key]
	return val, ok
}

// Priority 返回 MapPropertySource 的优先级。
func (m *MapPropertySource) Priority() Priority {
	return m.priority
}

// Contains 检查 MapPropertySource 中是否存在指定键。
func (m *MapPropertySource) Contains(key string) bool {
	_, ok := m.data[key]
	return ok
}

// Keys 返回所有键
func (m *MapPropertySource) Keys() []string {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

// NewEnvPropertySource 创建环境变量配置源
//
// prefix 是环境变量前缀，例如 "GO_BOOT"。
// 环境变量 GO_BOOT_SERVER_PORT=9090 将映射为键 "server.port"。
func NewEnvPropertySource(name, prefix string) *EnvPropertySource {
	return &EnvPropertySource{name: name, prefix: prefix, priority: PriorityHigh}
}

// Name 返回 EnvPropertySource 的名称。
func (e *EnvPropertySource) Name() string {
	return e.name
}

// Priority 返回 EnvPropertySource 的优先级。
func (e *EnvPropertySource) Priority() Priority {
	return e.priority
}

// Contains 检查环境变量配置源中是否存在指定键。
func (e *EnvPropertySource) Contains(key string) bool {
	_, ok := e.GetProperty(key)
	return ok
}

// GetProperty 从环境变量获取配置
//
// 键 "server.port" 将转换为环境变量 "GO_BOOT_SERVER_PORT"（如果 prefix="GO_BOOT"）。
func (e *EnvPropertySource) GetProperty(key string) (any, bool) {
	envKey := toEnvKey(key)
	if e.prefix != "" {
		envKey = e.prefix + "_" + envKey
	}
	val, ok := lookupEnv(envKey)
	if !ok {
		return nil, false
	}
	return val, true
}

// lookupEnv 可被测试替换
var lookupEnv = os.LookupEnv

// NewArgsPropertySource 解析命令行参数并创建配置源
//
// 支持的格式：--key=value
func NewArgsPropertySource(name string, args []string) *ArgsPropertySource {
	data := make(map[string]string)
	for _, arg := range args {
		if len(arg) > 2 && arg[:2] == "--" {
			kv := arg[2:]
			if key, val, found := strings.Cut(kv, "="); found && key != "" {
				data[key] = val
			}
		}
	}
	return &ArgsPropertySource{name: name, args: data, priority: PriorityHighest}
}

// Name 返回 ArgsPropertySource 的名称。
func (a *ArgsPropertySource) Name() string {
	return a.name
}

// Priority 返回 ArgsPropertySource 的优先级。
func (a *ArgsPropertySource) Priority() Priority {
	return a.priority
}

// Contains 检查命令行参数配置源中是否存在指定键。
func (a *ArgsPropertySource) Contains(key string) bool {
	_, ok := a.args[key]
	return ok
}

// GetProperty 从命令行参数配置源中获取指定键的值。
func (a *ArgsPropertySource) GetProperty(key string) (any, bool) {
	val, ok := a.args[key]
	return val, ok
}

// toEnvKey 将 "server.port" 转换为 "SERVER_PORT"
func toEnvKey(key string) string {
	result := make([]byte, 0, len(key))
	for i := 0; i < len(key); i++ {
		c := key[i]
		if c == '.' || c == '-' {
			result = append(result, '_')
			continue
		}
		if c >= 'a' && c <= 'z' {
			c = c - 'a' + 'A'
		}
		result = append(result, c)
	}
	return string(result)
}
