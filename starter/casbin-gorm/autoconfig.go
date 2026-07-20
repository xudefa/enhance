package casbingorm

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/xudefa/enhance/security"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"gorm.io/gorm"
)

func init() {
	boot.RegisterAutoConfigWith(&CasbinGormAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(security.SecurityEnabled, security.ConditionTrue),
			condition.OnProperty(security.CasbinEnabled, security.ConditionTrue),
			condition.OnProperty(security.CasbinPolicyType, DefaultCasbinGormPolicyType),
		),
		boot.WithOrder(int(boot.OrderPriorityAuthorizationGorm)), // 授权层-GORM，在 Casbin 基础配置之前执行
	)
}

// CasbinGormAutoConfiguration Casbin GORM 自动配置类。
//
// 当配置文件中启用 Casbin 且策略类型为 gorm 时自动生效。
// 负责创建基于 GORM 的 Casbin 适配器，并将其集成到 CasbinEnforcer 中。
//
// 执行顺序：Order = -1300（OrderPriorityAuthorizationGorm），确保在 Casbin 基础配置（-1200）之前执行。
//
// 依赖关系：
//  1. 依赖 GORM 自动配置（-2000），需要数据库连接
//  2. 在 Casbin 基础配置（-1200）之前执行，提供 GORM 版本的 CasbinEnforcer
//  3. Casbin 基础配置会检测容器中是否已有 CasbinEnforcer，如果有则直接使用
type CasbinGormAutoConfiguration struct {
	logger log.Logger
}

// CasbinGormConfig Casbin GORM 配置。
type CasbinGormConfig struct {
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

// Configure 配置 Casbin GORM 集成。
//
// 该方法在自动配置阶段调用，负责：
//  1. 从容器获取 *gorm.DB 实例
//  2. 创建 Casbin GORM 适配器
//  3. 创建 Casbin Enforcer（使用 GORM 适配器）
//  4. 注册到 IoC 容器
//  5. 如果启用自动加载，启动定时刷新策略
func (c *CasbinGormAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	container := ctx.Container()

	if logger, err := core.GetByName[log.Logger](container, ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	c.logger.Info(context.Background(), "开始配置 Casbin GORM 集成...")

	cfg, err := c.loadConfig(ctx)
	if err != nil {
		return fmt.Errorf("加载 Casbin GORM 配置失败: %w", err)
	}

	if err := c.validateConfig(cfg); err != nil {
		return fmt.Errorf("Casbin GORM 配置验证失败: %w", err)
	}

	dbBeans, err := container.Get(reflect.TypeFor[*gorm.DB]())
	if err != nil || len(dbBeans) == 0 {
		return fmt.Errorf("未找到 *gorm.DB 实例，请确保已启用 GORM 模块")
	}

	db := dbBeans[0].(*gorm.DB)

	adapter, err := c.createAdapter(db, cfg)
	if err != nil {
		return err
	}

	enforcer, err := c.createEnforcer(adapter, cfg)
	if err != nil {
		return err
	}

	enforcer.EnableAutoSave(true)

	gormEnforcer := &GormCasbinEnforcer{
		Enforcer: enforcer,
		adapter:  adapter,
	}

	if err := container.RegisterInstance(gormEnforcer, reflect.TypeFor[security.CasbinEnforcer]()); err != nil {
		return fmt.Errorf("注册 CasbinEnforcer 失败: %w", err)
	}
	c.logger.Info(context.Background(), "CasbinEnforcer (GORM) 已注册")

	voter := security.NewCasbinVoter(gormEnforcer)
	if err := container.RegisterInstance(voter, reflect.TypeFor[security.CasbinVoter]()); err != nil {
		return fmt.Errorf("注册 CasbinVoter 失败: %w", err)
	}
	c.logger.Info(context.Background(), "CasbinVoter 已注册")

	if cfg.AutoLoad {
		c.startAutoReload(gormEnforcer, cfg)
	}

	c.logger.Info(context.Background(), "Casbin GORM 集成配置完成",
		log.KeyValue{Key: "table-name", Value: cfg.TableName},
		log.KeyValue{Key: "auto-create-table", Value: cfg.AutoCreateTable},
	)

	return nil
}

// validateConfig 验证 Casbin GORM 配置。
func (c *CasbinGormAutoConfiguration) validateConfig(cfg *CasbinGormConfig) error {
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

// createAdapter 创建 Casbin GORM 适配器。
func (c *CasbinGormAutoConfiguration) createAdapter(db *gorm.DB, cfg *CasbinGormConfig) (*gormadapter.Adapter, error) {
	adapter, err := gormadapter.NewAdapterByDBUseTableName(db, cfg.DatabasePrefix, cfg.TableName)
	if err != nil {
		return nil, fmt.Errorf("创建 Casbin GORM 适配器失败: %w", err)
	}

	if cfg.AutoCreateTable {
		if err := c.autoMigrateTable(db, cfg.TableName); err != nil {
			c.logger.Warn(context.Background(), "Casbin GORM 自动迁移失败",
				log.KeyValue{Key: "error", Value: err.Error()},
			)
		} else {
			c.logger.Info(context.Background(), "Casbin GORM 表自动迁移成功",
				log.KeyValue{Key: "table-name", Value: cfg.TableName},
			)
		}
	}

	return adapter, nil
}

// autoMigrateTable 手动创建 Casbin 策略表。
func (c *CasbinGormAutoConfiguration) autoMigrateTable(db *gorm.DB, tableName string) error {
	type CasbinRule struct {
		ID    uint   `gorm:"primaryKey;autoIncrement"`
		PType string `gorm:"size:100;index"`
		V0    string `gorm:"size:100"`
		V1    string `gorm:"size:100"`
		V2    string `gorm:"size:100"`
		V3    string `gorm:"size:100"`
		V4    string `gorm:"size:100"`
		V5    string `gorm:"size:100"`
	}

	return db.Table(tableName).AutoMigrate(&CasbinRule{})
}

// createEnforcer 创建 Casbin Enforcer。
func (c *CasbinGormAutoConfiguration) createEnforcer(adapter *gormadapter.Adapter, cfg *CasbinGormConfig) (*casbin.Enforcer, error) {
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
func (c *CasbinGormAutoConfiguration) startAutoReload(enforcer *GormCasbinEnforcer, cfg *CasbinGormConfig) {
	interval := cfg.AutoLoadInterval
	if interval <= 0 {
		interval = security.DefaultCasbinAutoLoadInterval
	}

	c.logger.Info(context.Background(), "启动 Casbin GORM 策略自动刷新",
		log.KeyValue{Key: "interval", Value: fmt.Sprintf("%d分钟", interval)},
	)

	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := enforcer.LoadPolicy(context.Background()); err != nil {
				c.logger.Warn(context.Background(), "自动刷新 Casbin GORM 策略失败",
					log.KeyValue{Key: "error", Value: err.Error()},
				)
			} else {
				c.logger.Info(context.Background(), "Casbin GORM 策略已自动刷新")
			}
		}
	}()
}

// loadConfig 从 ApplicationContext 加载 Casbin GORM 配置。
func (c *CasbinGormAutoConfiguration) loadConfig(ctx boot.ApplicationContext) (*CasbinGormConfig, error) {
	env := ctx.Environment()

	cfg := &CasbinGormConfig{
		ModelType:        security.DefaultCasbinModelType,
		ModelPath:        security.DefaultCasbinModelPath,
		PolicyType:       DefaultCasbinGormPolicyType,
		AutoCreateTable:  DefaultCasbinGormAutoCreateTable,
		TableName:        DefaultCasbinGormTableName,
		DatabasePrefix:   DefaultCasbinGormDatabasePrefix,
		AutoLoad:         security.DefaultCasbinAutoLoad,
		AutoLoadInterval: security.DefaultCasbinAutoLoadInterval,
	}

	if err := env.BindPrefix("security.casbin", cfg); err != nil {
		return nil, fmt.Errorf("绑定 Casbin GORM 配置失败: %w", err)
	}

	return cfg, nil
}
