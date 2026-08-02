package admin

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// instanceIDCounter 全局实例 ID 计数器，确保 ID 唯一
var instanceIDCounter atomic.Int64

// Application 应用对象
//
// 锁顺序约束：所有方法必须按 Application.mu -> ApplicationInstance.mu 的顺序获取锁，
// 禁止反向获取以避免死锁。
type Application struct {
	// ID 应用 ID
	ID string
	// Name 应用名称
	Name string
	// Version 应用版本
	Version string
	// Instances 应用实例列表
	Instances []*ApplicationInstance
	// Status 应用状态
	Status ApplicationStatus
	// Metadata 应用元数据
	Metadata map[string]string
	// mu 互斥锁，保护 Instances 和 Status 字段
	mu sync.RWMutex
}

// ApplicationStatus 应用状态
type ApplicationStatus string

const (
	// StatusUp 运行中
	StatusUp ApplicationStatus = "UP"
	// StatusDown 已停止
	StatusDown ApplicationStatus = "DOWN"
	// StatusUnknown 未知
	StatusUnknown ApplicationStatus = "UNKNOWN"
	// StatusOutOfService 服务不可用
	StatusOutOfService ApplicationStatus = "OUT_OF_SERVICE"
)

// ApplicationInstance 应用实例
type ApplicationInstance struct {
	// ID 实例 ID
	ID string `json:"id"`
	// ApplicationID 所属应用 ID
	ApplicationID string `json:"application_id"`
	// Name 实例名称
	Name string `json:"name"`
	// URL 实例 URL
	URL string `json:"url"`
	// Status 实例状态
	Status ApplicationStatus `json:"status"`
	// Health 健康信息
	Health *HealthInfo `json:"health,omitempty"`
	// Metrics 指标信息
	Metrics map[string]float64 `json:"metrics,omitempty"`
	// Metadata 实例元数据
	Metadata map[string]string `json:"metadata,omitempty"`
	// RegisteredAt 注册时间
	RegisteredAt time.Time `json:"registered_at"`
	// LastSeen 最后活跃时间
	LastSeen time.Time `json:"last_seen"`
	// mu 互斥锁
	mu sync.RWMutex `json:"-"`
}

// HealthInfo 健康信息
type HealthInfo struct {
	// Status 健康状态
	Status ApplicationStatus
	// Components 组件健康信息
	Components map[string]*ComponentHealth
	// Details 详细信息
	Details map[string]any
}

// ComponentHealth 组件健康信息
type ComponentHealth struct {
	// Status 组件状态
	Status ApplicationStatus
	// Details 详细信息
	Details map[string]any
}

// ApplicationRegistry 应用注册中心
// 使用 sync.Map 优化读多写少场景的并发性能
type ApplicationRegistry struct {
	applications sync.Map // map[string]*Application
	instances    sync.Map // map[string]*ApplicationInstance
}

// NewApplicationRegistry 创建应用注册中心
func NewApplicationRegistry() *ApplicationRegistry {
	return &ApplicationRegistry{}
}

// Register 注册应用实例
func (r *ApplicationRegistry) Register(instance *ApplicationInstance) {
	// 注册实例（无锁）
	r.instances.Store(instance.ID, instance)

	// 查找或创建应用（使用 LoadOrStore 避免竞态）
	appValue, _ := r.applications.LoadOrStore(instance.ApplicationID, &Application{
		ID:        instance.ApplicationID,
		Name:      instance.Name,
		Instances: make([]*ApplicationInstance, 0),
		Status:    StatusUnknown,
		Metadata:  make(map[string]string),
	})
	app, _ := appValue.(*Application)

	// 添加实例到应用（需要应用级别的锁）
	app.mu.Lock()
	app.Instances = append(app.Instances, instance)
	app.mu.Unlock()

	// 更新应用状态
	r.updateApplicationStatus(app)
}

// Deregister 注销应用实例
func (r *ApplicationRegistry) Deregister(instanceID string) {
	instanceValue, exists := r.instances.Load(instanceID)
	if !exists {
		return
	}

	instance, _ := instanceValue.(*ApplicationInstance)

	// 从实例列表中移除
	r.instances.Delete(instanceID)

	// 从应用中移除实例
	appValue, exists := r.applications.Load(instance.ApplicationID)
	if !exists {
		return
	}
	app, _ := appValue.(*Application)

	app.mu.Lock()
	newInstances := make([]*ApplicationInstance, 0, len(app.Instances)-1)
	for _, inst := range app.Instances {
		if inst.ID != instanceID {
			newInstances = append(newInstances, inst)
		}
	}
	app.Instances = newInstances
	app.mu.Unlock()

	// 更新应用状态
	r.updateApplicationStatus(app)

	// 如果没有实例了，删除应用
	app.mu.RLock()
	empty := len(app.Instances) == 0
	app.mu.RUnlock()
	if empty {
		r.applications.Delete(app.ID)
	}
}

// GetApplication 获取应用
func (r *ApplicationRegistry) GetApplication(appID string) (*Application, error) {
	appValue, exists := r.applications.Load(appID)
	if !exists {
		return nil, fmt.Errorf("application %s does not exist", appID)
	}
	return appValue.(*Application), nil
}

// GetInstance 获取实例
func (r *ApplicationRegistry) GetInstance(instanceID string) (*ApplicationInstance, error) {
	instanceValue, exists := r.instances.Load(instanceID)
	if !exists {
		return nil, fmt.Errorf("instance %s does not exist", instanceID)
	}
	return instanceValue.(*ApplicationInstance), nil
}

// ListApplications 列出所有应用
func (r *ApplicationRegistry) ListApplications() []*Application {
	var apps []*Application
	r.applications.Range(func(key, value any) bool {
		apps = append(apps, value.(*Application))
		return true
	})
	return apps
}

// ListInstances 列出所有实例
func (r *ApplicationRegistry) ListInstances() []*ApplicationInstance {
	var instances []*ApplicationInstance
	r.instances.Range(func(key, value any) bool {
		instances = append(instances, value.(*ApplicationInstance))
		return true
	})
	return instances
}

// UpdateHealth 更新实例健康信息
func (r *ApplicationRegistry) UpdateHealth(instanceID string, health *HealthInfo) error {
	instanceValue, exists := r.instances.Load(instanceID)
	if !exists {
		return fmt.Errorf("instance %s does not exist", instanceID)
	}
	instance, _ := instanceValue.(*ApplicationInstance)

	instance.mu.Lock()
	instance.Health = health
	instance.Status = health.Status
	instance.LastSeen = time.Now()
	instance.mu.Unlock()

	// 更新应用状态
	appValue, exists := r.applications.Load(instance.ApplicationID)
	if exists {
		r.updateApplicationStatus(appValue.(*Application))
	}

	return nil
}

// UpdateMetrics 更新实例指标信息
func (r *ApplicationRegistry) UpdateMetrics(instanceID string, metrics map[string]float64) error {
	instanceValue, exists := r.instances.Load(instanceID)
	if !exists {
		return fmt.Errorf("instance %s does not exist", instanceID)
	}
	instance, _ := instanceValue.(*ApplicationInstance)

	instance.mu.Lock()
	instance.Metrics = metrics
	instance.LastSeen = time.Now()
	instance.mu.Unlock()

	return nil
}

// updateApplicationStatus 更新应用状态
func (r *ApplicationRegistry) updateApplicationStatus(app *Application) {
	app.mu.RLock()
	instances := make([]*ApplicationInstance, len(app.Instances))
	copy(instances, app.Instances)
	app.mu.RUnlock()

	if len(instances) == 0 {
		app.mu.Lock()
		app.Status = StatusUnknown
		app.mu.Unlock()
		return
	}

	upCount := 0
	downCount := 0

	for _, inst := range instances {
		inst.mu.RLock()
		switch inst.Status {
		case StatusUp:
			upCount++
		case StatusDown:
			downCount++
		default:
			// 其他状态不计入前端可用性统计
		}
		inst.mu.RUnlock()
	}

	app.mu.Lock()
	if upCount == len(instances) {
		app.Status = StatusUp
	} else if downCount == len(instances) {
		app.Status = StatusDown
	} else if upCount > 0 {
		app.Status = StatusOutOfService
	} else {
		app.Status = StatusUnknown
	}
	app.mu.Unlock()
}
