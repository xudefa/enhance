package registry

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// defaultBeanRegistry Bean 注册表实现。
// 使用 sync.Map 优化读多写少场景：
// - 注册阶段（启动时）：写入 Bean 定义
// - 运行阶段：高频读取 Bean 实例，极少写入
type defaultBeanRegistry struct {
	definitions  sync.Map   // beanID -> *BeanDef
	instances    sync.Map   // beanID -> any (Singleton 实例缓存)
	typeIndex    sync.Map   // reflect.Type -> []string (类型到 Bean ID 列表)
	primaryIndex sync.Map   // reflect.Type -> string (类型到首选 Bean ID)
	insertOrder  []string   // 记录注册顺序，用于保证销毁时的逆序
	mu           sync.Mutex // 保护 insertOrder 和 typeIndex 更新
}

// Register 注册 Bean 定义。
func (r *defaultBeanRegistry) Register(def BeanDef, beanID string) error {
	if def.Type == nil {
		return fmt.Errorf("type is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.registerInternal(def, beanID)
}

// registerInternal 注册 Bean 定义（内部方法，调用方必须已持有锁）。
func (r *defaultBeanRegistry) registerInternal(def BeanDef, beanID string) error {
	if _, exists := r.definitions.Load(beanID); exists {
		if def.Primary {
			r.primaryIndex.Store(def.Type, beanID)
		}
		return nil
	}

	r.definitions.Store(beanID, &def)
	r.insertOrder = append(r.insertOrder, beanID)

	var ids []string
	if existing, ok := r.typeIndex.Load(def.Type); ok {
		ids = existing.([]string)
	}
	ids = append(ids, beanID)
	r.typeIndex.Store(def.Type, ids)

	if def.Primary {
		r.primaryIndex.Store(def.Type, beanID)
	} else if _, exists := r.primaryIndex.Load(def.Type); !exists {
		r.primaryIndex.Store(def.Type, beanID)
	}

	return nil
}

// RegisterInstance 注册一个 Bean 实例，并设置为首选 Bean，作用域为 Singleton。
func (r *defaultBeanRegistry) RegisterInstance(instance any, typ reflect.Type, beanID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.definitions.Load(beanID); exists {
		return fmt.Errorf("instance already registered")
	}

	return r.registerInternal(BeanDef{
		Type:    typ,
		Primary: true,
		Scope:   Singleton,
		Factory: func(c ...any) (any, error) {
			return instance, nil
		},
	}, beanID)
}

// GetDefinition 获取 Bean 定义。
func (r *defaultBeanRegistry) GetDefinition(beanID string) (*BeanDef, bool) {
	// 直接查找
	if def, ok := r.definitions.Load(beanID); ok {
		return def.(*BeanDef), true
	}

	// 尝试通过自定义名称模糊匹配
	var foundDef *BeanDef
	r.definitions.Range(func(key, value any) bool {
		id := key.(string)
		if strings.HasSuffix(id, "#"+beanID) {
			foundDef = value.(*BeanDef)
			return false
		}
		return true
	})

	if foundDef != nil {
		return foundDef, true
	}

	return nil, false
}

// GetInstance 获取 Bean 实例（从缓存）。
func (r *defaultBeanRegistry) GetInstance(beanID string) (any, bool) {
	return r.instances.Load(beanID)
}

// SetInstance 设置 Bean 实例（用于缓存）。
func (r *defaultBeanRegistry) SetInstance(beanID string, instance any) {
	r.instances.Store(beanID, instance)
}

// GetByType 根据类型获取 Bean ID 列表。
//
// 返回切片的副本，避免调用者修改内部数据结构。
func (r *defaultBeanRegistry) GetByType(typ reflect.Type) []string {
	if ids, ok := r.typeIndex.Load(typ); ok {
		src := ids.([]string)
		dst := make([]string, len(src))
		copy(dst, src)
		return dst
	}
	return nil
}

// GetPrimaryByType 根据类型获取首选 Bean ID。
func (r *defaultBeanRegistry) GetPrimaryByType(typ reflect.Type) (string, bool) {
	if id, ok := r.primaryIndex.Load(typ); ok {
		return id.(string), true
	}
	return "", false
}

// HasBean 检查 Bean 是否存在。
func (r *defaultBeanRegistry) HasBean(beanID string) bool {
	// 直接查找
	if _, exists := r.definitions.Load(beanID); exists {
		return true
	}

	// 尝试通过自定义名称模糊匹配
	found := false
	r.definitions.Range(func(key, _ any) bool {
		id := key.(string)
		// 检查是否以 "#name" 结尾
		if strings.HasSuffix(id, "#"+beanID) {
			found = true
			return false
		}
		return true
	})
	return found
}

// HasType 检查类型是否存在。
func (r *defaultBeanRegistry) HasType(typ reflect.Type) bool {
	_, exists := r.typeIndex.Load(typ)
	return exists
}

// Count 返回已注册的 Bean 数量。
func (r *defaultBeanRegistry) Count() int {
	count := 0
	r.definitions.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// CountByType 返回指定类型的 Bean 数量。
func (r *defaultBeanRegistry) CountByType(typ reflect.Type) int {
	if ids, ok := r.typeIndex.Load(typ); ok {
		return len(ids.([]string))
	}
	return 0
}

// Types 返回所有已注册的 Bean 类型。
func (r *defaultBeanRegistry) Types() []reflect.Type {
	var types []reflect.Type
	r.typeIndex.Range(func(key, _ any) bool {
		types = append(types, key.(reflect.Type))
		return true
	})
	return types
}

// BeanIDs 返回所有已注册的 Bean ID（按注册顺序）。
func (r *defaultBeanRegistry) BeanIDs() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 返回注册顺序副本
	ids := make([]string, len(r.insertOrder))
	copy(ids, r.insertOrder)
	return ids
}

// ListBeans 列出所有已注册的Bean类型信息
func (r *defaultBeanRegistry) ListBeans() map[string]*BeanDef {
	beanMap := make(map[string]*BeanDef)
	r.definitions.Range(func(key, value any) bool {
		beanMap[key.(string)] = value.(*BeanDef)
		return true
	})
	return beanMap
}

// ListInstances 列出所有已注册的Bean实例
func (r *defaultBeanRegistry) ListInstances() map[string]any {
	beanMap := make(map[string]any)
	r.instances.Range(func(key, value any) bool {
		beanMap[key.(string)] = value
		return true
	})
	return beanMap
}

// Clear 清空注册表。
func (r *defaultBeanRegistry) Clear() {
	r.definitions.Range(func(key, _ any) bool {
		r.definitions.Delete(key)
		return true
	})
	r.instances.Range(func(key, _ any) bool {
		r.instances.Delete(key)
		return true
	})
	r.typeIndex.Range(func(key, _ any) bool {
		r.typeIndex.Delete(key)
		return true
	})
	r.primaryIndex.Range(func(key, _ any) bool {
		r.primaryIndex.Delete(key)
		return true
	})
	r.mu.Lock()
	r.insertOrder = nil
	r.mu.Unlock()
}

// defaultBeanIDGenerator Bean ID 生成器实现。
type defaultBeanIDGenerator struct{}

// Generate 生成 Bean ID。
// 格式：包路径.类型名#自定义名称
func (g *defaultBeanIDGenerator) Generate(pkgPath, typeName, customName string) string {
	var sb strings.Builder
	sb.WriteString(pkgPath)
	sb.WriteString(".")
	sb.WriteString(typeName)
	if customName != "" {
		sb.WriteString("#")
		sb.WriteString(customName)
	}
	return sb.String()
}

// Parse 解析 Bean ID。
func (g *defaultBeanIDGenerator) Parse(beanID string) (pkgPath, typeName, customName string) {
	// 解析自定义名称
	parts := strings.SplitN(beanID, "#", 2)
	mainPart := parts[0]
	if len(parts) > 1 {
		customName = parts[1]
	}

	// 解析包路径和类型名
	lastDot := strings.LastIndex(mainPart, ".")
	if lastDot == -1 {
		return "", mainPart, customName
	}

	pkgPath = mainPart[:lastDot]
	typeName = mainPart[lastDot+1:]
	return pkgPath, typeName, customName
}

// String 返回 Bean ID 的可读表示。
func (g *defaultBeanIDGenerator) String(beanID string) string {
	pkg, typ, name := g.Parse(beanID)
	if name != "" {
		return fmt.Sprintf("%s.%s#%s", pkg, typ, name)
	}
	return fmt.Sprintf("%s.%s", pkg, typ)
}

// NewBeanRegistry 创建 Bean 注册表实例。
func NewBeanRegistry() BeanRegistry {
	return &defaultBeanRegistry{}

}
