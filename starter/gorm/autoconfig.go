package gorm

import (
	"fmt"
	"reflect"
	"time"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	gormlib "gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	boot.RegisterAutoConfigWith(&GormAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(GORMEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityDataLayer)), // 数据层，在日志之后、安全组件之前执行
	)
}

// GormAutoConfiguration GORM 自动配置类。
//
// 当配置文件中启用 GORM 时自动生效（gorm.enabled=true）。
// 负责创建数据库连接并注册 *gorm.DB 到 IoC 容器中。
//
// 执行顺序：Order = -2000，确保在其他需要数据库的组件之前执行。
type GormAutoConfiguration struct {
	logger log.Logger
}

// Configure 配置 GORM 数据库连接。
//
// 该方法在自动配置阶段调用，负责：
//  1. 从 Environment 中读取数据库配置（主机、端口、用户名、密码等）
//  2. 构建 DSN 连接字符串
//  3. 创建 GORM 数据库连接
//  4. 配置连接池参数
//  5. 注册 *gorm.DB 到 IoC 容器
func (c *GormAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	// 从容器获取日志记录器，如果不存在则使用默认值
	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}
	c.logger.Info(ctx.Context(), "configuring GORM database connection...")

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load GORM config: %w", err)
	}

	dialector, err := c.getDialector(cfg)
	if err != nil {
		return fmt.Errorf("failed to get database dialector: %w", err)
	}

	db, err := gormlib.Open(dialector, &gormlib.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying database connection: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if err := ctx.Container().RegisterInstance(db, reflect.TypeFor[*gormlib.DB]()); err != nil {
		return fmt.Errorf("failed to register GORM DB: %w", err)
	}

	c.logger.Info(ctx.Context(), "GORM database connected successfully",
		log.KeyValue{Key: LogFieldHost, Value: cfg.Host},
		log.KeyValue{Key: LogFieldPort, Value: cfg.Port},
		log.KeyValue{Key: LogFieldDatabase, Value: cfg.Database},
	)

	return nil
}

// GormConfig GORM 数据库配置。
type GormConfig struct {
	Enabled         bool   `json:"enabled" mapstructure:"enabled"`
	Driver          string `json:"driver" mapstructure:"driver"`
	Host            string `json:"host" mapstructure:"host"`
	Port            int    `json:"port" mapstructure:"port"`
	Username        string `json:"username" mapstructure:"username"`
	Password        string `json:"password" mapstructure:"password"`
	Database        string `json:"database" mapstructure:"database"`
	Charset         string `json:"charset" mapstructure:"charset"`
	MaxOpenConns    int    `json:"max-open-conns" mapstructure:"max-open-conns"`
	MaxIdleConns    int    `json:"max-idle-conns" mapstructure:"max-idle-conns"`
	ConnMaxLifetime int    `json:"conn-max-lifetime" mapstructure:"conn-max-lifetime"`
}

// loadConfig 从 Environment 加载 GORM 配置。
func (c *GormAutoConfiguration) loadConfig(env *environment.Environment) (*GormConfig, error) {
	cfg := &GormConfig{
		Driver:          DefaultGORMDriver,
		Host:            DefaultGORMHost,
		Port:            DefaultGORMPort,
		Username:        DefaultGORMUsername,
		Password:        DefaultGORMPassword,
		Database:        DefaultGORMDatabase,
		Charset:         DefaultGORMCharset,
		MaxOpenConns:    DefaultGORMMaxOpenConns,
		MaxIdleConns:    DefaultGORMMaxIdleConns,
		ConnMaxLifetime: DefaultGORMConnMaxLifetime,
	}

	if err := env.BindPrefix("db.gorm", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind GORM config: %w", err)
	}

	if cfg.Database == "" {
		return nil, fmt.Errorf("%s config must not be empty", GORMDatabase)
	}

	return cfg, nil
}

// buildDSN 构建数据库连接字符串。
func (c *GormAutoConfiguration) buildDSN(cfg *GormConfig) string {
	switch cfg.Driver {
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
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Host,
			cfg.Port,
			cfg.Username,
			cfg.Password,
			cfg.Database,
		)
	case "sqlite":
		return cfg.Database
	default:
		return ""
	}
}

// getDialector 获取数据库驱动。
func (c *GormAutoConfiguration) getDialector(cfg *GormConfig) (gormlib.Dialector, error) {
	dsn := c.buildDSN(cfg)
	switch cfg.Driver {
	case "mysql":
		return mysql.Open(dsn), nil
	case "postgres":
		return postgres.Open(dsn), nil
	case "sqlite":
		return sqlite.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported database driver %q, supported drivers: mysql, postgres, sqlite", cfg.Driver)
	}
}
