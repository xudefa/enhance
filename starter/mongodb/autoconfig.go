// Package mongodb 提供 MongoDB 数据库自动配置。
package mongodb

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&MongoDBAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(MongoDBEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityDataLayer)),
	)
}

// MongoDBAutoConfiguration MongoDB 自动配置类。
type MongoDBAutoConfiguration struct {
	logger log.Logger
	client *mongo.Client
	config *MongoDBConfig
	ctx    context.Context
}

// Configure 配置 MongoDB 连接。
func (c *MongoDBAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load MongoDB config: %w", err)
	}

	c.config = cfg

	// 存储应用上下文
	c.ctx = ctx.Context()

	uri := c.buildURI(cfg)
	clientOpts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(uint64(cfg.MaxPoolSize)).
		SetMinPoolSize(uint64(cfg.MinPoolSize)).
		SetConnectTimeout(time.Duration(cfg.ConnectTimeout) * time.Second).
		SetServerSelectionTimeout(time.Duration(cfg.ServerSelectionTimeout) * time.Second)

	ctx2, cancel := context.WithTimeout(ctx.Context(), time.Duration(cfg.ConnectTimeout)*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx2, clientOpts)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx2, nil); err != nil {
		return fmt.Errorf("MongoDB ping failed: %w", err)
	}

	c.client = client

	if err := ctx.Container().RegisterInstance(c.client, reflect.TypeFor[*mongo.Client]()); err != nil {
		return fmt.Errorf("failed to register MongoDB Client: %w", err)
	}

	c.logger.Info(ctx.Context(), "MongoDB connected successfully",
		log.KeyValue{Key: "host", Value: cfg.Host},
		log.KeyValue{Key: "port", Value: cfg.Port},
		log.KeyValue{Key: "database", Value: cfg.Database},
	)

	return nil
}

// Stop 关闭 MongoDB 连接。
func (c *MongoDBAutoConfiguration) Stop() error {
	if c.client != nil {
		ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
		defer cancel()
		return c.client.Disconnect(ctx)
	}
	return nil
}

// GetClient 获取 MongoDB Client 实例。
func (c *MongoDBAutoConfiguration) GetClient() *mongo.Client {
	return c.client
}

// GetDatabase 获取指定数据库实例。
func (c *MongoDBAutoConfiguration) GetDatabase(name string) *mongo.Database {
	return c.client.Database(name)
}

// GetCollection 获取指定集合实例。
func (c *MongoDBAutoConfiguration) GetCollection(database, collection string) *mongo.Collection {
	return c.client.Database(database).Collection(collection)
}

// MongoDBConfig MongoDB 数据库配置。
type MongoDBConfig struct {
	Enabled                bool   `json:"enabled" mapstructure:"enabled"`
	Host                   string `json:"host" mapstructure:"host"`
	Port                   int    `json:"port" mapstructure:"port"`
	Username               string `json:"username" mapstructure:"username"`
	Password               string `json:"password" mapstructure:"password"`
	Database               string `json:"database" mapstructure:"database"`
	AuthSource             string `json:"auth-source" mapstructure:"auth-source"`
	MaxPoolSize            int    `json:"max-pool-size" mapstructure:"max-pool-size"`
	MinPoolSize            int    `json:"min-pool-size" mapstructure:"min-pool-size"`
	ConnectTimeout         int    `json:"connect-timeout" mapstructure:"connect-timeout"`
	ServerSelectionTimeout int    `json:"server-selection-timeout" mapstructure:"server-selection-timeout"`
}

// 配置常量。
const (
	MongoDBEnabled                = "mongodb.enabled"
	DefaultMongoDBHost            = "localhost"
	DefaultMongoDBPort            = 27017
	DefaultMongoDBDatabase        = "enhance"
	DefaultMongoDBAuthSource      = "admin"
	DefaultMaxPoolSize            = 100
	DefaultMinPoolSize            = 10
	DefaultConnectTimeout         = 10
	DefaultServerSelectionTimeout = 5
	ConditionTrue                 = "true"
)

// buildURI 构建 MongoDB 连接 URI。
func (c *MongoDBAutoConfiguration) buildURI(cfg *MongoDBConfig) string {
	if cfg.Username != "" && cfg.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=%s",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
			cfg.AuthSource,
		)
	}
	return fmt.Sprintf("mongodb://%s:%d/%s",
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
}

// loadConfig 从 Environment 加载 MongoDB 配置。
func (c *MongoDBAutoConfiguration) loadConfig(env *environment.Environment) (*MongoDBConfig, error) {
	cfg := &MongoDBConfig{
		Host:                   DefaultMongoDBHost,
		Port:                   DefaultMongoDBPort,
		Database:               DefaultMongoDBDatabase,
		AuthSource:             DefaultMongoDBAuthSource,
		MaxPoolSize:            DefaultMaxPoolSize,
		MinPoolSize:            DefaultMinPoolSize,
		ConnectTimeout:         DefaultConnectTimeout,
		ServerSelectionTimeout: DefaultServerSelectionTimeout,
	}

	if err := env.BindPrefix("mongodb", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind MongoDB config: %w", err)
	}

	return cfg, nil
}
