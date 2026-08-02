// Package rabbitmq 提供 RabbitMQ 消息队列自动配置。
package rabbitmq

import (
	"context"
	"fmt"
	"reflect"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&RabbitMQAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(RabbitMQEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityBusinessLayer)),
	)
}

// RabbitMQAutoConfiguration RabbitMQ 消息队列自动配置类。
type RabbitMQAutoConfiguration struct {
	logger     log.Logger
	connection *amqp.Connection
	channel    *amqp.Channel
	ctx        context.Context
}

// Configure 配置 RabbitMQ 连接。
func (c *RabbitMQAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load RabbitMQ config: %w", err)
	}

	url := cfg.buildURL()

	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to create RabbitMQ Channel: %w", err)
	}

	c.connection = conn
	c.channel = ch

	// 存储应用上下文
	c.ctx = ctx.Context()

	if err := ctx.Container().RegisterInstance(conn, reflect.TypeFor[*amqp.Connection]()); err != nil {
		return fmt.Errorf("failed to register RabbitMQ Connection: %w", err)
	}

	if err := ctx.Container().RegisterInstance(ch, reflect.TypeFor[*amqp.Channel]()); err != nil {
		return fmt.Errorf("failed to register RabbitMQ Channel: %w", err)
	}

	queue := NewRabbitMQQueue(ch, cfg)

	if err := ctx.Container().RegisterInstance(queue, reflect.TypeFor[*RabbitMQQueue]()); err != nil {
		return fmt.Errorf("failed to register RabbitMQ Queue: %w", err)
	}

	c.logger.Info(ctx.Context(), "RabbitMQ connected successfully",
		log.KeyValue{Key: "host", Value: cfg.Host},
		log.KeyValue{Key: "port", Value: cfg.Port},
		log.KeyValue{Key: "vhost", Value: cfg.VHost},
	)

	return nil
}

// Stop 关闭 RabbitMQ 连接。
func (c *RabbitMQAutoConfiguration) Stop() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			c.logger.Error(c.ctx, "failed to close RabbitMQ Channel",
				log.KeyValue{Key: "error", Value: err.Error()},
			)
		}
	}
	if c.connection != nil {
		return c.connection.Close()
	}
	return nil
}

// RabbitMQConfig RabbitMQ 配置。
type RabbitMQConfig struct {
	Enabled    bool   `json:"enabled" mapstructure:"enabled"`
	Host       string `json:"host" mapstructure:"host"`
	Port       int    `json:"port" mapstructure:"port"`
	Username   string `json:"username" mapstructure:"username"`
	Password   string `json:"password" mapstructure:"password"`
	VHost      string `json:"vhost" mapstructure:"vhost"`
	QueueName  string `json:"queue_name" mapstructure:"queue_name"`
	Exchange   string `json:"exchange" mapstructure:"exchange"`
	RoutingKey string `json:"routing_key" mapstructure:"routing_key"`
	Durable    bool   `json:"durable" mapstructure:"durable"`
	AutoDelete bool   `json:"auto_delete" mapstructure:"auto_delete"`
}

// RabbitMQQueue RabbitMQ 消息队列实现。
type RabbitMQQueue struct {
	channel *amqp.Channel
	config  *RabbitMQConfig
}

// NewRabbitMQQueue 创建 RabbitMQ 消息队列。
func NewRabbitMQQueue(ch *amqp.Channel, cfg *RabbitMQConfig) *RabbitMQQueue {
	return &RabbitMQQueue{
		channel: ch,
		config:  cfg,
	}
}

// Publish 发送消息到 RabbitMQ。
func (q *RabbitMQQueue) Publish(ctx context.Context, message []byte) error {
	return q.channel.PublishWithContext(ctx,
		q.config.Exchange,
		q.config.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/octet-stream",
			Body:        message,
		},
	)
}

// Subscribe 订阅消息。
func (q *RabbitMQQueue) Subscribe(ctx context.Context, handler func([]byte) error) error {
	msgs, err := q.channel.Consume(
		q.config.QueueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to RabbitMQ queue: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-msgs:
			if err := handler(msg.Body); err != nil {
				return fmt.Errorf("failed to process message: %w", err)
			}
		}
	}
}

// DeclareQueue 声明队列。
func (q *RabbitMQQueue) DeclareQueue() (amqp.Queue, error) {
	return q.channel.QueueDeclare(
		q.config.QueueName,
		q.config.Durable,
		q.config.AutoDelete,
		false,
		false,
		nil,
	)
}

// buildURL 构建 RabbitMQ 连接 URL。
func (c *RabbitMQConfig) buildURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		c.Username, c.Password, c.Host, c.Port, c.VHost,
	)
}

// 配置常量。
const (
	RabbitMQEnabled     = "rabbitmq.enabled"
	DefaultRabbitMQHost = "localhost"
	DefaultRabbitMQPort = 5672
	DefaultUsername     = "guest"
	DefaultPassword     = "guest"
	DefaultVHost        = "/"
	DefaultQueueName    = "enhance-queue"
	DefaultExchange     = ""
	DefaultRoutingKey   = "enhance-key"
	DefaultDurable      = true
	DefaultAutoDelete   = false
	ConditionTrue       = "true"
)

// loadConfig 从 Environment 加载 RabbitMQ 配置。
func (c *RabbitMQAutoConfiguration) loadConfig(env *environment.Environment) (*RabbitMQConfig, error) {
	cfg := &RabbitMQConfig{
		Host:       DefaultRabbitMQHost,
		Port:       DefaultRabbitMQPort,
		Username:   DefaultUsername,
		Password:   DefaultPassword,
		VHost:      DefaultVHost,
		QueueName:  DefaultQueueName,
		Exchange:   DefaultExchange,
		RoutingKey: DefaultRoutingKey,
		Durable:    DefaultDurable,
		AutoDelete: DefaultAutoDelete,
	}

	if err := env.BindPrefix("rabbitmq", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind RabbitMQ config: %w", err)
	}

	return cfg, nil
}
