// Package kafka 提供 Kafka 消息队列自动配置。
package kafka

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&KafkaAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(KafkaEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityBusinessLayer)),
	)
}

// KafkaAutoConfiguration Kafka 自动配置类。
type KafkaAutoConfiguration struct {
	logger log.Logger
	queue  *KafkaQueue
}

// Configure 配置 Kafka 连接。
func (c *KafkaAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Kafka config: %w", err)
	}

	if len(cfg.Brokers) == 0 {
		return errors.New("kafka brokers must not be empty")
	}

	connCtx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	conn, err := kafka.DialLeader(connCtx, "tcp", cfg.Brokers[0], cfg.Topic, 0)
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	err = conn.Close()
	if err != nil {
		return fmt.Errorf("failed to close Kafka connection: %w", err)
	}

	c.queue = NewKafkaQueue(cfg.Brokers, cfg.Topic, cfg.GroupID)

	if err := ctx.Container().RegisterInstance(c.queue, reflect.TypeFor[*KafkaQueue]()); err != nil {
		return fmt.Errorf("failed to register Kafka Queue: %w", err)
	}

	c.logger.Info(ctx.Context(), "Kafka connected successfully",
		log.KeyValue{Key: "brokers", Value: cfg.Brokers},
		log.KeyValue{Key: "topic", Value: cfg.Topic},
	)

	return nil
}

// Stop 关闭 Kafka 连接。
func (c *KafkaAutoConfiguration) Stop() error {
	if c.queue != nil {
		return c.queue.Close()
	}
	return nil
}

// KafkaConfig Kafka 配置。
type KafkaConfig struct {
	Enabled bool     `json:"enabled" mapstructure:"enabled"`
	Brokers []string `json:"brokers" mapstructure:"brokers"`
	Topic   string   `json:"topic" mapstructure:"topic"`
	GroupID string   `json:"group_id" mapstructure:"group_id"`
}

// KafkaQueue Kafka 消息队列实现。
type KafkaQueue struct {
	writer  *kafka.Writer
	reader  *kafka.Reader
	topic   string
	brokers []string
}

// NewKafkaQueue 创建 Kafka 消息队列。
func NewKafkaQueue(brokers []string, topic string, groupID string) *KafkaQueue {
	return &KafkaQueue{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,
		}),
		topic:   topic,
		brokers: brokers,
	}
}

// Publish 发送消息到 Kafka。
func (q *KafkaQueue) Publish(ctx context.Context, message []byte) error {
	return q.writer.WriteMessages(ctx, kafka.Message{
		Value: message,
		Time:  time.Now(),
	})
}

// Subscribe 订阅消息。
func (q *KafkaQueue) Subscribe(ctx context.Context, handler func([]byte) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msgCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
			m, err := q.reader.ReadMessage(msgCtx)
			cancel()

			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					continue
				}
				return fmt.Errorf("failed to read message: %w", err)
			}
			if err := handler(m.Value); err != nil {
				return fmt.Errorf("failed to process message: %w", err)
			}
		}
	}
}

// Close 关闭 Kafka 连接。
func (q *KafkaQueue) Close() error {
	var errs []error
	if err := q.writer.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := q.reader.Close(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to close Kafka connection: %w", errors.Join(errs...))
	}
	return nil
}

// GetTopic 获取主题名称。
func (q *KafkaQueue) GetTopic() string {
	return q.topic
}

// 配置常量。
const (
	KafkaEnabled   = "kafka.enabled"
	DefaultTopic   = "enhance-events"
	DefaultGroupID = "enhance-consumer"
	ConditionTrue  = "true"
)

// loadConfig 从 Environment 加载 Kafka 配置。
func (c *KafkaAutoConfiguration) loadConfig(env *environment.Environment) (*KafkaConfig, error) {
	cfg := &KafkaConfig{
		Topic:   DefaultTopic,
		GroupID: DefaultGroupID,
	}

	if err := env.BindPrefix("kafka", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind Kafka config: %w", err)
	}

	return cfg, nil
}
