// Package main 演示 Kafka 消息队列与 OpenTelemetry 链路追踪的集成使用。
//
// 该示例展示了如何：
//   - 使用 Kafka 发布消息
//   - 使用 OpenTelemetry 进行分布式链路追踪
//   - 结合缓存服务实现完整的业务场景
//
// 运行前请确保：
//   - Kafka 服务已启动（默认 localhost:9092）
//   - Redis 服务已启动（默认 localhost:6379）
//   - OTLP Collector 已启动（默认 localhost:4317）
package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/cache"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/starter/kafka"
	_ "github.com/xudefa/enhance/starter/otel"
	_ "github.com/xudefa/enhance/starter/redis"
	_ "github.com/xudefa/enhance/starter/zap"
)

// UserService 用户服务，演示缓存、日志、链路追踪和消息队列的集成使用。
type UserService struct {
	cache          cache.Cache
	logger         log.Logger
	tracerProvider *sdktrace.TracerProvider
	queue          *kafka.KafkaQueue
}

// NewUserService 创建用户服务实例。
func NewUserService(c cache.Cache, logger log.Logger, tracerProvider *sdktrace.TracerProvider, queue *kafka.KafkaQueue) *UserService {
	return &UserService{
		cache:          c,
		logger:         logger,
		tracerProvider: tracerProvider,
		queue:          queue,
	}
}

// GetUser 获取用户信息，优先从缓存读取，缓存未命中时模拟从数据库加载。
//
// 该方法会：
//  1. 创建链路追踪 span
//  2. 尝试从缓存获取用户数据
//  3. 缓存未命中时模拟加载数据并写入缓存
//  4. 发布用户访问事件到 Kafka
//
// 参数:
//   - ctx: 上下文，用于链路追踪和缓存操作
//   - userID: 用户ID
//
// 返回用户数据和可能的错误。
func (s *UserService) GetUser(ctx context.Context, userID string) (string, error) {
	tracer := s.tracerProvider.Tracer("UserService")
	ctx, span := tracer.Start(ctx, "UserService.GetUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID))

	val, err := s.cache.Get(ctx, "user:"+userID)
	if err == nil {
		s.logger.Info(ctx, "Cache hit", log.KeyValue{Key: "user_id", Value: userID})
		result, ok := val.(string)
		if !ok {
			return "", fmt.Errorf("invalid cache value type for user: %s", userID)
		}
		return result, nil
	}

	s.logger.Info(ctx, "Cache miss, loading from DB", log.KeyValue{Key: "user_id", Value: userID})

	userData := fmt.Sprintf("User-%s-Data", userID)

	if err := s.cache.Set(ctx, "user:"+userID, userData, 5*time.Minute); err != nil {
		s.logger.Warn(ctx, "Failed to cache user", log.KeyValue{Key: "user_id", Value: userID}, log.KeyValue{Key: "error", Value: err.Error()})
	}

	if err := s.queue.Publish(ctx, []byte(fmt.Sprintf("user_access:%s", userID))); err != nil {
		s.logger.Warn(ctx, "Failed to publish event", log.KeyValue{Key: "user_id", Value: userID}, log.KeyValue{Key: "error", Value: err.Error()})
	}

	return userData, nil
}

func main() {
	app, err := boot.NewApplication(
		boot.WithAppName("kafka-otel-demo"),
		boot.WithVersion("1.0.0"),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create application: %v", err))
	}
	defer app.Stop()

	container := app.Container()

	userService := NewUserService(
		core.MustGet[cache.Cache](container, ""),
		core.MustGet[log.Logger](container, ""),
		core.MustGet[*sdktrace.TracerProvider](container, ""),
		core.MustGet[*kafka.KafkaQueue](container, ""),
	)

	ctx := context.Background()

	user, err := userService.GetUser(ctx, "123")
	if err != nil {
		fmt.Printf("Error getting user: %v\n", err)
		return
	}

	fmt.Printf("User: %s\n", user)

	app.Start()
	app.WaitForSignal()
}
