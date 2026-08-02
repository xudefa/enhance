// Package ent 提供 ent ORM 框架自动配置。
package ent

import (
	"fmt"
	"reflect"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&EntAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(EntEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityDataLayer)),
	)
}

// EntAutoConfiguration ent ORM 自动配置类。
type EntAutoConfiguration struct {
	logger log.Logger
	driver *entsql.Driver
}

// Configure 配置 ent ORM 数据库连接。
func (c *EntAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load ent config: %w", err)
	}

	driver, err := entsql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to open ent database connection: %w", err)
	}

	c.driver = driver

	if err := ctx.Container().RegisterInstance(driver, reflect.TypeFor[*entsql.Driver]()); err != nil {
		return fmt.Errorf("failed to register ent Driver: %w", err)
	}

	c.logger.Info(ctx.Context(), "ent ORM database connected successfully",
		log.KeyValue{Key: "driver", Value: cfg.Driver},
		log.KeyValue{Key: "database", Value: cfg.Database},
	)

	return nil
}

// GetDriver 获取 ent 数据库驱动。
func (c *EntAutoConfiguration) GetDriver() *entsql.Driver {
	return c.driver
}

// Close 关闭数据库连接。
func (c *EntAutoConfiguration) Close() error {
	if c.driver != nil {
		return c.driver.Close()
	}
	return nil
}

// EntConfig ent ORM 配置。
type EntConfig struct {
	Enabled  bool   `json:"enabled" mapstructure:"enabled"`
	Driver   string `json:"driver" mapstructure:"driver"`
	DSN      string `json:"dsn" mapstructure:"dsn"`
	Database string `json:"database" mapstructure:"database"`
}

// 配置常量。
const (
	EntEnabled      = "ent.enabled"
	DefaultDriver   = dialect.MySQL
	DefaultDSN      = "root:root@tcp(localhost:3306)/enhance?parseTime=True"
	DefaultDatabase = "enhance"
	ConditionTrue   = "true"
)

// loadConfig 从 Environment 加载 ent 配置。
func (c *EntAutoConfiguration) loadConfig(env *environment.Environment) (*EntConfig, error) {
	cfg := &EntConfig{
		Driver:   DefaultDriver,
		DSN:      DefaultDSN,
		Database: DefaultDatabase,
	}

	if err := env.BindPrefix("ent", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind ent config: %w", err)
	}

	return cfg, nil
}
