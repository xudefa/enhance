package kafka

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestKafkaConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-kafka", environment.PriorityNormal, map[string]any{
		"kafka.enabled":  "true",
		"kafka.brokers":  "[\"localhost:9092\"]",
		"kafka.topic":    "test-topic",
		"kafka.group_id": "test-group",
	}))

	cfg := &KafkaConfig{
		Topic:   DefaultTopic,
		GroupID: DefaultGroupID,
	}

	err := env.BindPrefix("kafka", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected kafka.enabled to be true")
	}
	if cfg.Topic != "test-topic" {
		t.Errorf("expected topic 'test-topic', got '%s'", cfg.Topic)
	}
	if cfg.GroupID != "test-group" {
		t.Errorf("expected group_id 'test-group', got '%s'", cfg.GroupID)
	}
}

func TestKafkaConfig_DefaultValues(t *testing.T) {
	cfg := &KafkaConfig{
		Topic:   DefaultTopic,
		GroupID: DefaultGroupID,
	}

	if cfg.Topic != "enhance-events" {
		t.Errorf("expected default topic 'enhance-events', got '%s'", cfg.Topic)
	}
	if cfg.GroupID != "enhance-consumer" {
		t.Errorf("expected default group_id 'enhance-consumer', got '%s'", cfg.GroupID)
	}
}

func TestNewKafkaQueue(t *testing.T) {
	brokers := []string{"localhost:9092"}
	topic := "test-topic"
	groupID := "test-group"

	queue := NewKafkaQueue(brokers, topic, groupID)

	if queue == nil {
		t.Fatal("expected queue to be created")
	}
	if queue.GetTopic() != topic {
		t.Errorf("expected topic '%s', got '%s'", topic, queue.GetTopic())
	}
}
