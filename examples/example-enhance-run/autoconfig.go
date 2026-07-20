package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/web/mvc"
	"github.com/xudefa/enhance/web/server"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormConfig GORM 数据库配置
type GormConfig struct {
	Enabled         bool   `json:"enabled"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	Database        string `json:"database"`
	Charset         string `json:"charset"`
	MaxOpenConns    int    `json:"max-open-conns"`
	MaxIdleConns    int    `json:"max-idle-conns"`
	ConnMaxLifetime int    `json:"conn-max-lifetime"`
}

// GormAutoConfig GORM 自动配置
type GormAutoConfig struct {
}

// Configure 配置 GORM 组件
func (c *GormAutoConfig) Configure(ctx boot.ApplicationContext) error {
	container := ctx.Container()
	env := ctx.Environment()

	// 读取配置
	var cfg GormConfig
	if err := env.BindPrefix("gorm", &cfg); err != nil {
		return fmt.Errorf("绑定 GORM 配置失败: %w", err)
	}

	if !cfg.Enabled {
		fmt.Println("[GORM] GORM 未启用，跳过配置")
		return nil
	}

	// 创建数据库连接
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// 注册到容器
	if err := container.RegisterInstance(db, reflect.TypeOf(db)); err != nil {
		return fmt.Errorf("注册 GORM DB 失败: %w", err)
	}

	fmt.Printf("[GORM] 数据库连接成功: %s:%d/%s\n", cfg.Host, cfg.Port, cfg.Database)
	return nil
}

// WebAutoConfig Web 自动配置
type WebAutoConfig struct {
}

// Configure 配置 Web 组件
func (c *WebAutoConfig) Configure(ctx boot.ApplicationContext) error {
	container := ctx.Container()
	env := ctx.Environment()

	// 读取服务器配置
	var serverConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	if err := env.BindPrefix("server", &serverConfig); err != nil {
		return fmt.Errorf("绑定服务器配置失败: %w", err)
	}

	if serverConfig.Port == 0 {
		serverConfig.Port = 8080
	}
	if serverConfig.Host == "" {
		serverConfig.Host = "0.0.0.0"
	}

	// 创建路由器（使用 server.NewRouter 创建默认路由器）
	router := server.NewRouter()

	// 获取数据库实例（如果存在）
	var db any
	dbBeans, dbErr := container.Get(reflect.TypeOf((*gorm.DB)(nil)))
	if dbErr != nil || len(dbBeans) == 0 {
		fmt.Println("[Web] 数据库实例不存在，跳过控制器注册")
		db = nil
	} else {
		db = dbBeans[0]
	}

	// 注意：控制器通过 mvc.RegisterController 注册，WebStarter.Start() 会自动调用 Routes()
	// 如果数据库可用，注册 UserController（需要数据库依赖注入）
	if db != nil {
		userController := NewUserController(db.(*gorm.DB))
		mvc.RegisterController(userController)
		fmt.Println("[Web] UserController 已注册")
	}

	// 创建 HttpServer 适配器（使用 web/server 包）
	httpServer := server.NewHttpServerAdapter(
		server.WithHost(fmt.Sprintf("%s:%d", serverConfig.Host, serverConfig.Port)),
	)

	// 创建 WebStarter（使用 mvc 包）
	webStarter := mvc.NewWebStarter(
		mvc.WithConfig(mvc.WebConfig{
			Host: serverConfig.Host,
			Port: serverConfig.Port,
		}),
		mvc.WithRouter(router),
		mvc.WithServer(httpServer),
	)

	// 注册 WebStarter 到全局启动器注册表
	boot.RegisterStarter(webStarter)

	fmt.Printf("[Web] Web 服务器配置: %s:%d\n", serverConfig.Host, serverConfig.Port)
	return nil
}

// DatabaseAutoConfig 数据库自动配置
type DatabaseAutoConfig struct {
}

// Configure 配置数据库组件
func (c *DatabaseAutoConfig) Configure(ctx boot.ApplicationContext) error {
	container := ctx.Container()

	// 获取数据库实例
	dbBeans, err := container.Get(reflect.TypeOf((*gorm.DB)(nil)))
	if err != nil || len(dbBeans) == 0 {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}
	db := dbBeans[0].(*gorm.DB)

	// 自动迁移表
	if err := db.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("自动迁移表失败: %w", err)
	}

	fmt.Println("[GORM] 数据库表自动迁移完成")
	return nil
}

func init() {
	// 注册自动配置（注意顺序：GormAutoConfig 必须在 WebAutoConfig 之前）
	boot.RegisterAutoConfig(&GormAutoConfig{})
	boot.RegisterAutoConfig(&WebAutoConfig{})
	boot.RegisterAutoConfig(&DatabaseAutoConfig{})
}
