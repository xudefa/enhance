package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// instanceIDCounter 全局实例 ID 计数器，确保 ID 唯一
var instanceIDCounter atomic.Int64

// Application 应用对象
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
	// mu 互斥锁
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
	app := appValue.(*Application)

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

	instance := instanceValue.(*ApplicationInstance)

	// 从实例列表中移除
	r.instances.Delete(instanceID)

	// 从应用中移除实例
	appValue, exists := r.applications.Load(instance.ApplicationID)
	if !exists {
		return
	}
	app := appValue.(*Application)

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
	if len(app.Instances) == 0 {
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
	instance := instanceValue.(*ApplicationInstance)

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
	instance := instanceValue.(*ApplicationInstance)

	instance.mu.Lock()
	instance.Metrics = metrics
	instance.LastSeen = time.Now()
	instance.mu.Unlock()

	return nil
}

// updateApplicationStatus 更新应用状态
func (r *ApplicationRegistry) updateApplicationStatus(app *Application) {
	if len(app.Instances) == 0 {
		app.Status = StatusUnknown
		return
	}

	upCount := 0
	downCount := 0

	for _, inst := range app.Instances {
		inst.mu.RLock()
		switch inst.Status {
		case StatusUp:
			upCount++
		case StatusDown:
			downCount++
		}
		inst.mu.RUnlock()
	}

	if upCount == len(app.Instances) {
		app.Status = StatusUp
	} else if downCount == len(app.Instances) {
		app.Status = StatusDown
	} else if upCount > 0 {
		app.Status = StatusOutOfService
	} else {
		app.Status = StatusUnknown
	}
}

// AdminServer Admin 服务器
type AdminServer struct {
	registry *ApplicationRegistry
	mux      *http.ServeMux
}

// NewAdminServer 创建 Admin 服务器
func NewAdminServer(registry *ApplicationRegistry) *AdminServer {
	server := &AdminServer{
		registry: registry,
		mux:      http.NewServeMux(),
	}

	server.registerRoutes()

	return server
}

// Handler 获取 HTTP handler
func (s *AdminServer) Handler() http.Handler {
	return s.mux
}

// registerRoutes 注册路由
func (s *AdminServer) registerRoutes() {
	s.mux.HandleFunc("/admin/applications", s.handleListApplications)
	s.mux.HandleFunc("/admin/applications/", s.handleGetApplication)
	s.mux.HandleFunc("/admin/instances", s.handleListInstances)
	s.mux.HandleFunc("/admin/instances/", s.handleGetInstance)
	s.mux.HandleFunc("/admin/instances/{id}/health", s.handleGetHealth)
	s.mux.HandleFunc("/admin/instances/{id}/metrics", s.handleGetMetrics)
	s.mux.HandleFunc("/admin/health", s.handleOverallHealth)
	s.mux.HandleFunc("/admin/register", s.handleRegister)
	s.mux.HandleFunc("/admin/deregister", s.handleDeregister)
}

// handleListApplications 处理列出应用请求
func (s *AdminServer) handleListApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apps := s.registry.ListApplications()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(apps)
}

// handleGetApplication 处理获取应用请求
func (s *AdminServer) handleGetApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appID := r.URL.Query().Get("id")
	if appID == "" {
		http.Error(w, "Application ID required", http.StatusBadRequest)
		return
	}

	app, err := s.registry.GetApplication(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(app)
}

// handleListInstances 处理列出实例请求
func (s *AdminServer) handleListInstances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	instances := s.registry.ListInstances()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(instances)
}

// handleGetInstance 处理获取实例请求
func (s *AdminServer) handleGetInstance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	instanceID := r.URL.Query().Get("id")
	if instanceID == "" {
		http.Error(w, "Instance ID required", http.StatusBadRequest)
		return
	}

	instance, err := s.registry.GetInstance(instanceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(instance)
}

// handleGetHealth 处理获取健康信息请求
func (s *AdminServer) handleGetHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	instanceID := r.URL.Query().Get("id")
	if instanceID == "" {
		http.Error(w, "Instance ID required", http.StatusBadRequest)
		return
	}

	instance, err := s.registry.GetInstance(instanceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	instance.mu.RLock()
	health := instance.Health
	instance.mu.RUnlock()

	if health == nil {
		http.Error(w, "Health information not available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(health)
}

// handleGetMetrics 处理获取指标信息请求
func (s *AdminServer) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	instanceID := r.URL.Query().Get("id")
	if instanceID == "" {
		http.Error(w, "Instance ID required", http.StatusBadRequest)
		return
	}

	instance, err := s.registry.GetInstance(instanceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	instance.mu.RLock()
	metrics := instance.Metrics
	instance.mu.RUnlock()

	if metrics == nil {
		http.Error(w, "Metrics information not available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(metrics)
}

// handleOverallHealth 处理整体健康状态请求
func (s *AdminServer) handleOverallHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apps := s.registry.ListApplications()

	totalApps := len(apps)
	upApps := 0
	downApps := 0

	for _, app := range apps {
		switch app.Status {
		case StatusUp:
			upApps++
		case StatusDown:
			downApps++
		}
	}

	overall := map[string]any{
		"status": map[string]any{
			"total":   totalApps,
			"up":      upApps,
			"down":    downApps,
			"unknown": totalApps - upApps - downApps,
		},
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(overall)
}

// handleRegister 处理注册请求
func (s *AdminServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var instance ApplicationInstance
	if err := json.NewDecoder(r.Body).Decode(&instance); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if instance.ID == "" || instance.ApplicationID == "" || instance.URL == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	instance.RegisteredAt = time.Now()
	instance.LastSeen = time.Now()

	if instance.Status == "" {
		instance.Status = StatusUnknown
	}

	if instance.Metrics == nil {
		instance.Metrics = make(map[string]float64)
	}

	s.registry.Register(&instance)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Instance registered successfully",
		"id":      instance.ID,
	})
}

// handleDeregister 处理注销请求
func (s *AdminServer) handleDeregister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		InstanceID string `json:"instance_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.InstanceID == "" {
		http.Error(w, "Instance ID required", http.StatusBadRequest)
		return
	}

	s.registry.Deregister(req.InstanceID)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Instance deregistered successfully",
	})
}

// NewApplicationInstance 创建应用实例
func NewApplicationInstance(appID, url string) *ApplicationInstance {
	return &ApplicationInstance{
		ID:            fmt.Sprintf("%s-%d-%d", appID, time.Now().UnixNano(), instanceIDCounter.Add(1)),
		ApplicationID: appID,
		Name:          appID,
		URL:           url,
		Status:        StatusUnknown,
		Metrics:       make(map[string]float64),
		Metadata:      make(map[string]string),
		RegisteredAt:  time.Now(),
		LastSeen:      time.Now(),
	}
}

// NewHealthInfo 创建健康信息
func NewHealthInfo(status ApplicationStatus) *HealthInfo {
	return &HealthInfo{
		Status:     status,
		Components: make(map[string]*ComponentHealth),
		Details:    make(map[string]any),
	}
}

// AddComponent 添加组件健康信息
func (h *HealthInfo) AddComponent(name string, status ApplicationStatus, details map[string]any) {
	h.Components[name] = &ComponentHealth{
		Status:  status,
		Details: details,
	}
}

// AddDetail 添加详细信息
func (h *HealthInfo) AddDetail(key string, value any) {
	h.Details[key] = value
}

// AddMetric 添加指标
func (i *ApplicationInstance) AddMetric(name string, value float64) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.Metrics == nil {
		i.Metrics = make(map[string]float64)
	}

	i.Metrics[name] = value
}

// GetMetric 获取指标
func (i *ApplicationInstance) GetMetric(name string) (float64, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	value, exists := i.Metrics[name]
	return value, exists
}

// IsHealthy 检查实例是否健康
func (i *ApplicationInstance) IsHealthy() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.Status == StatusUp
}

// GetApplicationCount 获取应用数量
func (r *ApplicationRegistry) GetApplicationCount() int {
	count := 0
	r.applications.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

// GetInstanceCount 获取实例数量
func (r *ApplicationRegistry) GetInstanceCount() int {
	count := 0
	r.instances.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

// GetUpInstanceCount 获取运行中的实例数量
func (r *ApplicationRegistry) GetUpInstanceCount() int {
	count := 0
	r.instances.Range(func(key, value any) bool {
		instance := value.(*ApplicationInstance)
		instance.mu.RLock()
		if instance.Status == StatusUp {
			count++
		}
		instance.mu.RUnlock()
		return true
	})
	return count
}

// GetDownInstanceCount 获取停止的实例数量
func (r *ApplicationRegistry) GetDownInstanceCount() int {
	count := 0
	r.instances.Range(func(key, value any) bool {
		instance := value.(*ApplicationInstance)
		instance.mu.RLock()
		if instance.Status == StatusDown {
			count++
		}
		instance.mu.RUnlock()
		return true
	})
	return count
}
