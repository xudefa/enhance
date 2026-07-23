package casbinxorm

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/xudefa/enhance/security"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	xormadapter "github.com/casbin/xorm-adapter/v2"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"xorm.io/xorm"
)

func init() {
	boot.RegisterAutoConfigWith(&CasbinXormAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(security.SecurityEnabled, security.ConditionTrue),
			condition.OnProperty(security.CasbinEnabled, security.ConditionTrue),
			condition.OnProperty(security.CasbinPolicyType, DefaultCasbinXormPolicyType),
		),
		boot.WithOrder(int(boot.OrderPriorityAuthorizationGorm)),
	)
}

// CasbinXormAutoConfiguration Casbin XORM 自动配置类。
//
// 当配置文件中启用 Casbin 且策略类型为 xorm 时自动生效。
// 负责创建基于 XORM 的 Casbin 适配器，并将其集成到 CasbinEnforcer 中。
//
// 执行顺序：Order = -1300（OrderPriorityAuthorizationGorm），确保在 Casbin 基础配置（-1200）之前执行。
//
// 依赖关系：
//  1. 依赖 XORM 自动配置（-2000），需要数据库连接
//  2. 在 Casbin 基础配置（-1200）之前执行，提供 XORM 版本的 CasbinEnforcer
//  3. Casbin 基础配置会检测容器中是否已有 CasbinEnforcer，如果有则直接使用
type CasbinXormAutoConfiguration struct {
	logger log.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

// CasbinXormConfig Casbin XORM 配置。
type CasbinXormConfig struct {
	Enabled          bool   `json:"enabled" mapstructure:"enabled"`
	ModelType        string `json:"model-type" mapstructure:"model-type"`
	ModelPath        string `json:"model-path" mapstructure:"model-path"`
	ModelText        string `json:"model-text" mapstructure:"model-text"`
	PolicyType       string `json:"policy-type" mapstructure:"policy-type"`
	AutoCreateTable  bool   `json:"auto-create-table" mapstructure:"auto-create-table"`
	TableName        string `json:"table-name" mapstructure:"table-name"`
	DatabasePrefix   string `json:"database-prefix" mapstructure:"database-prefix"`
	AutoLoad         bool   `json:"auto-load" mapstructure:"auto-load"`
	AutoLoadInterval int    `json:"auto-load-interval" mapstructure:"auto-load-interval"`
}

// Configure 配置 Casbin XORM 集成。
//
// 该方法在自动配置阶段调用，负责：
//  1. 从容器获取 *xorm.Engine 实例
//  2. 创建 Casbin XORM 适配器
//  3. 创建 Casbin Enforcer（使用 XORM 适配器）
//  4. 注册到 IoC 容器
//  5. 如果启用自动加载，启动定时刷新策略
func (c *CasbinXormAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	container := ctx.Container()

	if logger, err := core.GetByName[log.Logger](container, ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	c.logger.Info(context.Background(), "开始配置 Casbin XORM 集成...")

	cfg, err := c.loadConfig(ctx)
	if err != nil {
		return fmt.Errorf("加载 Casbin XORM 配置失败: %w", err)
	}

	if err := c.validateConfig(cfg); err != nil {
		return fmt.Errorf("Casbin XORM 配置验证失败: %w", err)
	}

	engineBeans, err := container.Get(reflect.TypeFor[*xorm.Engine]())
	if err != nil || len(engineBeans) == 0 {
		return fmt.Errorf("未找到 *xorm.Engine 实例，请确保已启用 XORM 模块")
	}

	engine := engineBeans[0].(*xorm.Engine)

	adapter, err := c.createAdapter(engine, cfg)
	if err != nil {
		return err
	}

	enforcer, err := c.createEnforcer(adapter, cfg)
	if err != nil {
		return err
	}

	enforcer.EnableAutoSave(true)

	xormEnforcer := &XormCasbinEnforcer{
		Enforcer: enforcer,
		adapter:  adapter,
	}

	if err := container.RegisterInstance(xormEnforcer, reflect.TypeFor[security.CasbinEnforcer]()); err != nil {
		return fmt.Errorf("注册 CasbinEnforcer 失败: %w", err)
	}
	c.logger.Info(context.Background(), "CasbinEnforcer (XORM) 已注册")

	voter := security.NewCasbinVoter(xormEnforcer)
	if err := container.RegisterInstance(voter, reflect.TypeFor[security.CasbinVoter]()); err != nil {
		return fmt.Errorf("注册 CasbinVoter 失败: %w", err)
	}
	c.logger.Info(context.Background(), "CasbinVoter 已注册")

	if cfg.AutoLoad {
		c.ctx, c.cancel = context.WithCancel(context.Background())
		c.startAutoReload(xormEnforcer, cfg)
	}

	c.logger.Info(context.Background(), "Casbin XORM 集成配置完成",
		log.KeyValue{Key: "table-name", Value: cfg.TableName},
		log.KeyValue{Key: "auto-create-table", Value: cfg.AutoCreateTable},
	)

	return nil
}

// validateConfig 验证 Casbin XORM 配置。
func (c *CasbinXormAutoConfiguration) validateConfig(cfg *CasbinXormConfig) error {
	switch cfg.ModelType {
	case "file":
		if cfg.ModelPath == "" {
			return fmt.Errorf("model-type 为 file 时，model-path 不能为空")
		}
	case "string":
		if cfg.ModelText == "" {
			return fmt.Errorf("model-type 为 string 时，model-text 不能为空")
		}
	default:
		return fmt.Errorf("不支持的 model-type: %s，支持的值: file, string", cfg.ModelType)
	}

	if cfg.TableName == "" {
		return fmt.Errorf("table-name 不能为空")
	}

	return nil
}

// createAdapter 创建 Casbin XORM 适配器。
func (c *CasbinXormAutoConfiguration) createAdapter(engine *xorm.Engine, cfg *CasbinXormConfig) (*xormadapter.Adapter, error) {
	// xorm-adapter v2 的 NewAdapter 函数签名:
	// NewAdapter(driverName string, dataSourceName string, dbSpecified bool)
	// 其中 dataSourceName 需要是标准 DSN 格式

	// 从 engine 获取 driverName
	driverName := engine.DriverName()

	// 尝试从 engine 获取 DSN
	// engine.DataSourceName() 返回的可能是内部格式，需要转换
	dsn := engine.DataSourceName()

	// 如果 DSN 为空或格式不正确，尝试从配置重新构建
	if dsn == "" {
		return nil, fmt.Errorf("无法从 XORM Engine 获取有效的 DSN")
	}

	// 创建适配器
	// dbSpecified=true 表示 DSN 已经包含了数据库名
	adapter, err := xormadapter.NewAdapter(driverName, dsn, true)
	if err != nil {
		return nil, fmt.Errorf("创建 Casbin XORM 适配器失败: %w", err)
	}

	c.logger.Info(context.Background(), "Casbin XORM 适配器创建成功",
		log.KeyValue{Key: "table-name", Value: cfg.TableName},
		log.KeyValue{Key: "driver", Value: driverName},
	)

	return adapter, nil
}

// createEnforcer 创建 Casbin Enforcer。
func (c *CasbinXormAutoConfiguration) createEnforcer(adapter *xormadapter.Adapter, cfg *CasbinXormConfig) (*casbin.Enforcer, error) {
	if cfg.ModelType == "string" && cfg.ModelText != "" {
		m, err := model.NewModelFromString(cfg.ModelText)
		if err != nil {
			return nil, fmt.Errorf("创建 Casbin 模型失败: %w", err)
		}
		enforcer, err := casbin.NewEnforcer(m, adapter)
		if err != nil {
			return nil, fmt.Errorf("创建 Casbin Enforcer 失败: %w", err)
		}
		return enforcer, nil
	}

	enforcer, err := casbin.NewEnforcer(cfg.ModelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("创建 Casbin Enforcer 失败: %w", err)
	}

	return enforcer, nil
}

// startAutoReload 启动自动重新加载策略。
func (c *CasbinXormAutoConfiguration) startAutoReload(enforcer *XormCasbinEnforcer, cfg *CasbinXormConfig) {
	interval := cfg.AutoLoadInterval
	if interval <= 0 {
		interval = security.DefaultCasbinAutoLoadInterval
	}

	c.logger.Info(context.Background(), "启动 Casbin XORM 策略自动刷新",
		log.KeyValue{Key: "interval", Value: fmt.Sprintf("%d分钟", interval)},
	)

	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := enforcer.LoadPolicy(context.Background()); err != nil {
					c.logger.Warn(context.Background(), "自动刷新 Casbin XORM 策略失败",
						log.KeyValue{Key: "error", Value: err.Error()},
					)
				} else {
					c.logger.Info(context.Background(), "Casbin XORM 策略已自动刷新")
				}
			case <-c.ctx.Done():
				return
			}
		}
	}()
}

// Close 停止自动刷新定时器，释放 goroutine 资源。
func (c *CasbinXormAutoConfiguration) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}

// loadConfig 从 ApplicationContext 加载 Casbin XORM 配置。
func (c *CasbinXormAutoConfiguration) loadConfig(ctx boot.ApplicationContext) (*CasbinXormConfig, error) {
	env := ctx.Environment()

	cfg := &CasbinXormConfig{
		ModelType:        security.DefaultCasbinModelType,
		ModelPath:        security.DefaultCasbinModelPath,
		PolicyType:       DefaultCasbinXormPolicyType,
		AutoCreateTable:  DefaultCasbinXormAutoCreateTable,
		TableName:        DefaultCasbinXormTableName,
		DatabasePrefix:   DefaultCasbinXormDatabasePrefix,
		AutoLoad:         security.DefaultCasbinAutoLoad,
		AutoLoadInterval: security.DefaultCasbinAutoLoadInterval,
	}

	if err := env.BindPrefix("security.casbin", cfg); err != nil {
		return nil, fmt.Errorf("绑定 Casbin XORM 配置失败: %w", err)
	}

	return cfg, nil
}
