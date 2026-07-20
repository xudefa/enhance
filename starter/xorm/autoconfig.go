package xorm

import (
	"context"
	"fmt"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"xorm.io/xorm"
)

func init() {
	boot.RegisterAutoConfigWith(&XormAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(XORMEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityDataLayer)),
	)
}

// XormAutoConfiguration XORM 自动配置类。
//
// 当配置文件中启用 XORM 时自动生效（db.xorm.enabled=true）。
// 负责创建数据库连接并注册 *xorm.Engine 到 IoC 容器中。
//
// 执行顺序：Order = -2000，确保在其他需要数据库的组件之前执行。
type XormAutoConfiguration struct {
	logger log.Logger
}

// Configure 配置 XORM 数据库连接。
//
// 该方法在自动配置阶段调用，负责：
//  1. 从 Environment 中读取数据库配置（主机、端口、用户名、密码等）
//  2. 构建 DSN 连接字符串
//  3. 创建 XORM 数据库连接
//  4. 配置连接池参数
//  5. 注册 *xorm.Engine 到 IoC 容器
func (c *XormAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}
	c.logger.Info(context.Background(), "开始配置 XORM 数据库连接...")

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("加载 XORM 配置失败: %w", err)
	}

	dsn := c.buildDSN(cfg)

	engine, err := xorm.NewEngine(cfg.Type, dsn)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	engine.SetMaxOpenConns(cfg.MaxOpenConns)
	engine.SetMaxIdleConns(cfg.MaxIdleConns)
	engine.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if cfg.ShowSQL {
		engine.ShowSQL(true)
	}

	if err := ctx.Container().RegisterInstance(engine, reflect.TypeFor[*xorm.Engine]()); err != nil {
		return fmt.Errorf("注册 XORM Engine 失败: %w", err)
	}

	c.logger.Info(context.Background(), "XORM 数据库连接成功",
		log.KeyValue{Key: LogFieldHost, Value: cfg.Host},
		log.KeyValue{Key: LogFieldPort, Value: cfg.Port},
		log.KeyValue{Key: LogFieldDatabase, Value: cfg.Database},
		log.KeyValue{Key: LogFieldType, Value: cfg.Type},
	)

	return nil
}

// XormConfig XORM 数据库配置。
type XormConfig struct {
	Enabled         bool   `json:"enabled" mapstructure:"enabled"`
	Type            string `json:"type" mapstructure:"type"`
	Host            string `json:"host" mapstructure:"host"`
	Port            int    `json:"port" mapstructure:"port"`
	Username        string `json:"username" mapstructure:"username"`
	Password        string `json:"password" mapstructure:"password"`
	Database        string `json:"database" mapstructure:"database"`
	Charset         string `json:"charset" mapstructure:"charset"`
	MaxOpenConns    int    `json:"max-open-conns" mapstructure:"max-open-conns"`
	MaxIdleConns    int    `json:"max-idle-conns" mapstructure:"max-idle-conns"`
	ConnMaxLifetime int    `json:"conn-max-lifetime" mapstructure:"conn-max-lifetime"`
	ShowSQL         bool   `json:"show-sql" mapstructure:"show-sql"`
}

// loadConfig 从 Environment 加载 XORM 配置。
func (c *XormAutoConfiguration) loadConfig(env *environment.Environment) (*XormConfig, error) {
	cfg := &XormConfig{
		Type:            DefaultXORMType,
		Host:            DefaultXORMHost,
		Port:            DefaultXORMPort,
		Username:        DefaultXORMUsername,
		Password:        DefaultXORMPassword,
		Database:        DefaultXORMDatabase,
		Charset:         DefaultXORMCharset,
		MaxOpenConns:    DefaultXORMMaxOpenConns,
		MaxIdleConns:    DefaultXORMMaxIdleConns,
		ConnMaxLifetime: DefaultXORMConnMaxLifetime,
		ShowSQL:         DefaultXORMShowSQL,
	}

	if err := env.BindPrefix("db.xorm", cfg); err != nil {
		return nil, fmt.Errorf("绑定 XORM 配置失败: %w", err)
	}

	if cfg.Host == "" || cfg.Database == "" {
		return nil, fmt.Errorf("%s 和 %s 配置不能为空", XORMHost, XORMDatabase)
	}

	return cfg, nil
}

// buildDSN 构建数据库连接字符串。
func (c *XormAutoConfiguration) buildDSN(cfg *XormConfig) string {
	switch cfg.Type {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
			cfg.Charset,
		)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host,
			cfg.Port,
			cfg.Username,
			cfg.Password,
			cfg.Database,
		)
	case "sqlite3":
		return cfg.Database
	default:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
			cfg.Charset,
		)
	}
}
