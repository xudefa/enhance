package environment

// NewMapEnvironment 创建基于 map 的环境配置
//
// 使用单个 MapPropertySource 作为配置源，优先级为 PriorityNormal。
func NewMapEnvironment(properties map[string]string) *Environment {
	data := make(map[string]any, len(properties))
	for k, v := range properties {
		data[k] = v
	}
	source := NewMapPropertySource("map", PriorityNormal, data)

	env := &Environment{
		sources: []PropertySource{source},
	}
	return env
}

// NewMapEnvironmentWithProfiles 创建带 profiles 的 map 环境配置
//
// 使用单个 MapPropertySource 作为配置源，并设置激活的 profiles。
func NewMapEnvironmentWithProfiles(properties map[string]string, activeProfiles []string) *Environment {
	env := NewMapEnvironment(properties)
	for _, p := range activeProfiles {
		env.AddActiveProfile(p)
	}
	return env
}

// AcceptsProfiles 检查是否接受指定的 profiles（至少一个匹配即返回 true）
func (e *Environment) AcceptsProfiles(profiles ...string) bool {
	for _, p := range profiles {
		if e.AcceptsProfile(p) {
			return true
		}
	}
	return false
}

// GetPropertyWithDefault 获取属性值，不存在时返回默认值
func (e *Environment) GetPropertyWithDefault(key, defaultValue string) string {
	if val, ok := e.GetProperty(key); ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return defaultValue
}
