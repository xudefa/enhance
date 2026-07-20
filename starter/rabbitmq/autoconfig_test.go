package rabbitmq

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestRabbitMQConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-rabbitmq", environment.PriorityNormal, map[string]any{
		"rabbitmq.enabled":    "true",
		"rabbitmq.host":       "192.168.1.100",
		"rabbitmq.port":       5673,
		"rabbitmq.queue_name": "test-queue",
	}))

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

	err := env.BindPrefix("rabbitmq", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected rabbitmq.enabled to be true")
	}
	if cfg.Host != "192.168.1.100" {
		t.Errorf("expected host '192.168.1.100', got '%s'", cfg.Host)
	}
	if cfg.Port != 5673 {
		t.Errorf("expected port 5673, got %d", cfg.Port)
	}
	if cfg.QueueName != "test-queue" {
		t.Errorf("expected queue_name 'test-queue', got '%s'", cfg.QueueName)
	}
}

func TestRabbitMQConfig_DefaultValues(t *testing.T) {
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

	if cfg.Host != "localhost" {
		t.Errorf("expected default host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 5672 {
		t.Errorf("expected default port 5672, got %d", cfg.Port)
	}
	if cfg.Username != "guest" {
		t.Errorf("expected default username 'guest', got '%s'", cfg.Username)
	}
}

func TestRabbitMQConfig_BuildURL(t *testing.T) {
	cfg := &RabbitMQConfig{
		Host:     "localhost",
		Port:     5672,
		Username: "guest",
		Password: "guest",
		VHost:    "/",
	}

	expected := "amqp://guest:guest@localhost:5672/"
	url := cfg.buildURL()
	if url != expected {
		t.Errorf("expected URL '%s', got '%s'", expected, url)
	}
}
