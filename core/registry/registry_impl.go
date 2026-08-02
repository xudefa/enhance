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
	definitions     sync.Map     // beanID -> *BeanDef
	instances       sync.Map     // beanID -> any (Singleton 实例缓存)
	typeIndex       sync.Map     // reflect.Type -> []string (类型到 Bean ID 列表)
	primaryIndex    sync.Map     // reflect.Type -> string (类型到首选 Bean ID)
	customNameIndex sync.Map     // customName -> full beanID (自定义名称到完整 BeanID 的索引)
	insertOrder     []string     // 记录注册顺序，用于保证销毁时的逆序
	mu              sync.RWMutex // 保护 insertOrder 和 typeIndex 更新
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
	// 重复注册检测：相同 ID 已存在
	if existing, exists := r.definitions.Load(beanID); exists {
		if def.Primary {
			r.primaryIndex.Store(def.Type, beanID)
		}
		// 等价定义重复注册保持幂等，真正不同的定义返回错误
		if sameBeanDefinition(existing.(*BeanDef), &def) {
			return nil
		}
		return fmt.Errorf("%w: bean id %q already registered", ErrBeanAlreadyExists, beanID)
	}

	// 先验证自定义名称冲突，避免注册失败后留下半污染状态
	if def.Name != "" {
		if _, loaded := r.customNameIndex.Load(def.Name); loaded {
			return fmt.Errorf("custom name %q already registered", def.Name)
		}
		if idx := strings.LastIndex(def.Name, "#"); idx != -1 {
			suffix := def.Name[idx+1:]
			if _, loaded := r.customNameIndex.Load(suffix); loaded {
				return fmt.Errorf("custom name %q already registered", suffix)
			}
		}
	}

	r.definitions.Store(beanID, &def)
	r.insertOrder = append(r.insertOrder, beanID)

	if def.Name != "" {
		r.customNameIndex.Store(def.Name, beanID)
		if idx := strings.LastIndex(def.Name, "#"); idx != -1 {
			r.customNameIndex.Store(def.Name[idx+1:], beanID)
		}
	}

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

// funcPtr 返回函数指针，nil 函数返回 0。
func funcPtr(fn any) uintptr {
	if fn == nil {
		return 0
	}
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func || v.IsNil() {
		return 0
	}
	return v.Pointer()
}

// sameBeanDefinition 判断两个 Bean 定义是否等价（用于幂等重复注册检测）。
//
// 相同 beanID 的重复注册视为幂等操作，仅当定义存在实质差异时
// （作用域、懒加载、回调存在性不同）才判定为不同的 Bean。
// Name 不参与比较：RegisterBean 会将生成的 beanID 回填到 Name，
// 而 RegisterInstance 的 Name 为空，两者表示同一 Bean。
// Factory 无法可靠比较：闭包指针每次创建都不同（如实例包装工厂），
// 框架约定同 ID 重复注册时以首次注册为准。
func sameBeanDefinition(a, b *BeanDef) bool {
	return a.Type == b.Type &&
		normalizeScope(a.Scope) == normalizeScope(b.Scope) &&
		a.Lazy == b.Lazy &&
		funcPtr(a.Init) == funcPtr(b.Init) &&
		funcPtr(a.Destroy) == funcPtr(b.Destroy)
}

// normalizeScope 规范化作用域：空值等同于 Singleton。
func normalizeScope(s Scope) Scope {
	if s == "" {
		return Singleton
	}
	return s
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

	// 通过自定义名称索引查找 O(1)
	if fullID, ok := r.customNameIndex.Load(beanID); ok {
		fid := fullID.(string)
		if def, ok := r.definitions.Load(fid); ok {
			return def.(*BeanDef), true
		}
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
	r.mu.RLock()
	defer r.mu.RUnlock()

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

	// 通过自定义名称索引查找 O(1)
	if fullID, ok := r.customNameIndex.Load(beanID); ok {
		fid := fullID.(string)
		_, exists := r.definitions.Load(fid)
		return exists
	}

	return false
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
		src := value.(*BeanDef)
		copy := *src
		beanMap[key.(string)] = &copy
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
	r.mu.Lock()
	defer r.mu.Unlock()

	r.definitions = sync.Map{}
	r.instances = sync.Map{}
	r.typeIndex = sync.Map{}
	r.primaryIndex = sync.Map{}
	r.customNameIndex = sync.Map{}
	r.insertOrder = nil
}

// NewBeanRegistry 创建 Bean 注册表实例。
func NewBeanRegistry() BeanRegistry {
	return &defaultBeanRegistry{}
}
