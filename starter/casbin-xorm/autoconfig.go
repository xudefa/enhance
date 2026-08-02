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
	"github.com/xudefa/enhance/config/environment"
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

// CasbinXormAutoConfiguration Casbin XORM auto-configuration.
//
// Activates when Casbin is enabled in config and policy type is xorm.
// Creates a XORM-based Casbin adapter and integrates it into CasbinEnforcer.
//
// Execution order: Order = -1300 (OrderPriorityAuthorizationGorm), executes before Casbin base config (-1200).
//
// Dependencies:
//  1. Depends on XORM auto-configuration (-2000), requires database connection
//  2. Executes before Casbin base configuration (-1200), provides XORM-based CasbinEnforcer
//  3. Casbin base configuration checks if CasbinEnforcer already exists in container
type CasbinXormAutoConfiguration struct {
	logger log.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

// CasbinXormConfig Casbin XORM config.
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

// Configure configures Casbin XORM integration.
//
// Called during the auto-configuration phase, responsible for:
//  1. Getting *xorm.Engine instance from container
//  2. Creating Casbin XORM adapter
//  3. Creating Casbin Enforcer (using XORM adapter)
//  4. Registering with IoC container
//  5. Starting auto-reload timer if enabled
func (c *CasbinXormAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	container := ctx.Container()

	if logger, err := core.GetByName[log.Logger](container, ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	c.logger.Info(ctx.Context(), "configuring Casbin XORM integration...")

	cfg, err := c.loadConfig(ctx.Environment())
	if err != nil {
		return fmt.Errorf("failed to load Casbin XORM config: %w", err)
	}

	if err := c.validateConfig(cfg); err != nil {
		return fmt.Errorf("Casbin XORM config validation failed: %w", err)
	}

	engineBeans, err := container.Get(reflect.TypeFor[*xorm.Engine]())
	if err != nil || len(engineBeans) == 0 {
		return fmt.Errorf("no *xorm.Engine instance found, please ensure XORM module is enabled")
	}

	engine, _ := engineBeans[0].(*xorm.Engine)

	adapter, err := c.createAdapter(ctx.Context(), engine, cfg)
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
		return fmt.Errorf("failed to register CasbinEnforcer: %w", err)
	}
	c.logger.Info(ctx.Context(), "CasbinEnforcer (XORM) registered")

	voter := security.NewCasbinVoter(xormEnforcer)
	if err := container.RegisterInstance(voter, reflect.TypeFor[security.CasbinVoter]()); err != nil {
		return fmt.Errorf("failed to register CasbinVoter: %w", err)
	}
	c.logger.Info(ctx.Context(), "CasbinVoter registered")

	if cfg.AutoLoad {
		c.ctx = ctx.Context()
		c.startAutoReload(xormEnforcer, cfg)
	}

	c.logger.Info(ctx.Context(), "Casbin XORM integration configured",
		log.KeyValue{Key: "table-name", Value: cfg.TableName},
		log.KeyValue{Key: "auto-create-table", Value: cfg.AutoCreateTable},
	)

	return nil
}

// validateConfig validates Casbin XORM config.
func (c *CasbinXormAutoConfiguration) validateConfig(cfg *CasbinXormConfig) error {
	switch cfg.ModelType {
	case "file":
		if cfg.ModelPath == "" {
			return fmt.Errorf("model-path must not be empty when model-type is file")
		}
	case "string":
		if cfg.ModelText == "" {
			return fmt.Errorf("model-text must not be empty when model-type is string")
		}
	default:
		return fmt.Errorf("unsupported model-type: %s, supported values: file, string", cfg.ModelType)
	}

	if cfg.TableName == "" {
		return fmt.Errorf("table-name must not be empty")
	}

	return nil
}

// createAdapter creates Casbin XORM adapter.
func (c *CasbinXormAutoConfiguration) createAdapter(gctx context.Context, engine *xorm.Engine, cfg *CasbinXormConfig) (*xormadapter.Adapter, error) {
	// xorm-adapter v2 NewAdapter signature:
	// NewAdapter(driverName string, dataSourceName string, dbSpecified bool)
	// dataSourceName must be standard DSN format

	driverName := engine.DriverName()

	dsn := engine.DataSourceName()

	if dsn == "" {
		return nil, fmt.Errorf("failed to get valid DSN from XORM Engine")
	}

	// dbSpecified=true means DSN already contains database name
	adapter, err := xormadapter.NewAdapter(driverName, dsn, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin XORM adapter: %w", err)
	}

	c.logger.Info(gctx, "Casbin XORM adapter created successfully",
		log.KeyValue{Key: "table-name", Value: cfg.TableName},
		log.KeyValue{Key: "driver", Value: driverName},
	)

	return adapter, nil
}

// createEnforcer creates Casbin Enforcer.
func (c *CasbinXormAutoConfiguration) createEnforcer(adapter *xormadapter.Adapter, cfg *CasbinXormConfig) (*casbin.Enforcer, error) {
	if cfg.ModelType == "string" && cfg.ModelText != "" {
		m, err := model.NewModelFromString(cfg.ModelText)
		if err != nil {
			return nil, fmt.Errorf("failed to create Casbin model: %w", err)
		}
		enforcer, err := casbin.NewEnforcer(m, adapter)
		if err != nil {
			return nil, fmt.Errorf("failed to create Casbin Enforcer: %w", err)
		}
		return enforcer, nil
	}

	enforcer, err := casbin.NewEnforcer(cfg.ModelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin Enforcer: %w", err)
	}

	return enforcer, nil
}

// startAutoReload starts automatic policy reload.
func (c *CasbinXormAutoConfiguration) startAutoReload(enforcer *XormCasbinEnforcer, cfg *CasbinXormConfig) {
	interval := cfg.AutoLoadInterval
	if interval <= 0 {
		interval = security.DefaultCasbinAutoLoadInterval
	}

	c.logger.Info(c.ctx, "starting Casbin XORM policy auto-reload",
		log.KeyValue{Key: "interval", Value: fmt.Sprintf("%dmin", interval)},
	)

	go func() {
		defer recoverLog("casbin-xorm policy auto-reload", c.ctx, c.logger)
		ticker := time.NewTicker(time.Duration(interval) * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := enforcer.LoadPolicy(c.ctx); err != nil {
					c.logger.Warn(c.ctx, "Casbin XORM policy auto-reload failed",
						log.KeyValue{Key: "error", Value: err.Error()},
					)
				} else {
					c.logger.Info(c.ctx, "Casbin XORM policy auto-reloaded")
				}
			case <-c.ctx.Done():
				return
			}
		}
	}()
}

// Close stops auto-reload timer and releases goroutine resources.
func (c *CasbinXormAutoConfiguration) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}

// recoverLog recovers from panic and logs the error.
func recoverLog(component string, ctx context.Context, logger log.Logger) {
	if r := recover(); r != nil {
		logger.Error(ctx, fmt.Sprintf("%s panic recovered", component),
			log.KeyValue{Key: "panic", Value: fmt.Sprintf("%v", r)},
		)
	}
}

// loadConfig loads Casbin XORM config from Environment.
func (c *CasbinXormAutoConfiguration) loadConfig(env *environment.Environment) (*CasbinXormConfig, error) {
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
		return nil, fmt.Errorf("failed to bind Casbin XORM config: %w", err)
	}

	return cfg, nil
}
