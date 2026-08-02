// Package redis 提供 Redis 缓存自动配置。
package redis

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/cache"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

var redisAutoConfig = &RedisAutoConfiguration{}

func init() {
	boot.RegisterAutoConfigWith(redisAutoConfig,
		boot.WithConditions(
			condition.OnProperty(RedisEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityDataLayer)),
	)
	// 注册为 Starter，使其 Start/Stop 生命周期方法被自动调用
	boot.RegisterStarter(redisAutoConfig)
}

// RedisAutoConfiguration Redis 自动配置类。
type RedisAutoConfiguration struct {
	logger     log.Logger
	client     *redis.Client
	mu         sync.Mutex
	configured bool            // 标记是否已配置，防止同一应用上下文重复配置
	ctx        context.Context // 应用上下文
}

// Configure 配置 Redis 连接。
func (c *RedisAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 同一应用上下文的 AutoConfig 与 Starter 双注册会调用两次 Configure，直接跳过
	if c.configured && c.ctx == ctx.Context() {
		return nil
	}
	// 新应用上下文（应用重启）时重新配置，更新 client/ctx 等状态
	c.configured = false

	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Redis config: %w", err)
	}

	c.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	redisCtx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	if err := c.client.Ping(redisCtx).Err(); err != nil {
		_ = c.client.Close()
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	redisCache := NewRedisCache(c.client, cfg.Prefix)

	if err := ctx.Container().RegisterInstance(c.client, reflect.TypeFor[*redis.Client]()); err != nil {
		return fmt.Errorf("failed to register Redis Client: %w", err)
	}

	if err := ctx.Container().RegisterInstance(redisCache, reflect.TypeFor[cache.Cache]()); err != nil {
		return fmt.Errorf("failed to register Redis Cache: %w", err)
	}

	c.logger.Info(ctx.Context(), "Redis connected successfully",
		log.KeyValue{Key: "host", Value: cfg.Host},
		log.KeyValue{Key: "port", Value: cfg.Port},
	)

	c.configured = true
	c.ctx = ctx.Context()

	return nil
}

// Start 启动阶段调用，无需额外操作。
func (c *RedisAutoConfiguration) Start(ctx boot.ApplicationContext) error {
	return nil
}

// Stop 关闭 Redis 连接。
func (c *RedisAutoConfiguration) Stop(ctx boot.ApplicationContext) error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Name 返回启动器名称。
func (c *RedisAutoConfiguration) Name() string {
	return "RedisStarter"
}

// Dependencies 返回依赖的其他启动器名称。
func (c *RedisAutoConfiguration) Dependencies() []string {
	return nil
}

// GetCondition 返回启动器条件。
func (c *RedisAutoConfiguration) GetCondition() condition.Condition {
	return condition.OnProperty(RedisEnabled, ConditionTrue)
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
	if errors.Is(err, redis.Nil) {
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
	if errors.Is(err, redis.Nil) {
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
		return nil, fmt.Errorf("failed to bind Redis config: %w", err)
	}

	return cfg, nil
}
