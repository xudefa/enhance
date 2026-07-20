// Package redis 提供 Redis 缓存自动配置。
package redis

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/cache"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&RedisAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(RedisEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityDataLayer)),
	)
}

// RedisAutoConfiguration Redis 自动配置类。
type RedisAutoConfiguration struct {
	logger log.Logger
	client *redis.Client
}

// Configure 配置 Redis 连接。
func (c *RedisAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("加载 Redis 配置失败: %w", err)
	}

	c.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	redisCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.client.Ping(redisCtx).Err(); err != nil {
		return fmt.Errorf("Redis 连接失败: %w", err)
	}

	redisCache := NewRedisCache(c.client, cfg.Prefix)

	if err := ctx.Container().RegisterInstance(c.client, reflect.TypeFor[*redis.Client]()); err != nil {
		return fmt.Errorf("注册 Redis Client 失败: %w", err)
	}

	if err := ctx.Container().RegisterInstance(redisCache, reflect.TypeFor[cache.Cache]()); err != nil {
		return fmt.Errorf("注册 Redis Cache 失败: %w", err)
	}

	c.logger.Info(context.Background(), "Redis 连接成功",
		log.KeyValue{Key: "host", Value: cfg.Host},
		log.KeyValue{Key: "port", Value: cfg.Port},
	)

	return nil
}

// Stop 关闭 Redis 连接。
func (c *RedisAutoConfiguration) Stop() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// RedisConfig Redis 配置。
type RedisConfig struct {
	Enabled  bool   `json:"enabled" mapstructure:"enabled"`
	Host     string `json:"host" mapstructure:"host"`
	Port     int    `json:"port" mapstructure:"port"`
	Password string `json:"password" mapstructure:"password"`
	DB       int    `json:"db" mapstructure:"db"`
	Prefix   string `json:"prefix" mapstructure:"prefix"`
	PoolSize int    `json:"pool_size" mapstructure:"pool_size"`
}

// RedisCache Redis 缓存实现。
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache 创建 Redis 缓存实例。
func NewRedisCache(client *redis.Client, prefix string) *RedisCache {
	return &RedisCache{client: client, prefix: prefix}
}

// Get 获取缓存值。
func (r *RedisCache) Get(ctx context.Context, key string) (any, error) {
	val, err := r.client.Get(ctx, r.prefix+key).Result()
	if err == redis.Nil {
		return nil, cache.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("redis get error: %w", err)
	}
	return val, nil
}

// Set 设置缓存值。
func (r *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return r.client.Set(ctx, r.prefix+key, value, ttl).Err()
}

// Del 删除缓存键。
func (r *RedisCache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	redisKeys := make([]string, len(keys))
	for i, k := range keys {
		redisKeys[i] = r.prefix + k
	}
	return r.client.Del(ctx, redisKeys...).Err()
}

// Exists 检查键是否存在。
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := r.client.Exists(ctx, r.prefix+key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists error: %w", err)
	}
	return exists > 0, nil
}

// TTL 获取键的剩余过期时间。
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, r.prefix+key).Result()
	if err == redis.Nil {
		return 0, cache.ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("redis ttl error: %w", err)
	}
	return ttl, nil
}

// Close 关闭 Redis 连接。
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// GetClient 获取底层 Redis Client。
func (r *RedisCache) GetClient() *redis.Client {
	return r.client
}

// 配置常量。
const (
	RedisEnabled         = "redis.enabled"
	DefaultRedisHost     = "localhost"
	DefaultRedisPort     = 6379
	DefaultRedisDB       = 0
	DefaultRedisPrefix   = "enhance:"
	DefaultRedisPoolSize = 10
	ConditionTrue        = "true"
)

// loadConfig 从 Environment 加载 Redis 配置。
func (c *RedisAutoConfiguration) loadConfig(env *environment.Environment) (*RedisConfig, error) {
	cfg := &RedisConfig{
		Host:     DefaultRedisHost,
		Port:     DefaultRedisPort,
		DB:       DefaultRedisDB,
		Prefix:   DefaultRedisPrefix,
		PoolSize: DefaultRedisPoolSize,
	}

	if err := env.BindPrefix("redis", cfg); err != nil {
		return nil, fmt.Errorf("绑定 Redis 配置失败: %w", err)
	}

	return cfg, nil
}
