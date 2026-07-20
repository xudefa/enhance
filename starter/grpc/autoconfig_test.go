package grpc

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestGrpcConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-grpc", environment.PriorityNormal, map[string]any{
		"grpc.enabled":           "true",
		"grpc.port":              9091,
		"grpc.enable_reflection": "false",
	}))

	cfg := &GrpcConfig{
		Port:             DefaultGrpcPort,
		EnableReflection: DefaultEnableReflection,
		MaxRecvMsgSize:   DefaultMaxRecvMsgSize,
		MaxSendMsgSize:   DefaultMaxSendMsgSize,
	}

	err := env.BindPrefix("grpc", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected grpc.enabled to be true")
	}
	if cfg.Port != 9091 {
		t.Errorf("expected port 9091, got %d", cfg.Port)
	}
	if cfg.EnableReflection {
		t.Error("expected enable_reflection to be false")
	}
}

func TestGrpcConfig_DefaultValues(t *testing.T) {
	cfg := &GrpcConfig{
		Port:             DefaultGrpcPort,
		EnableReflection: DefaultEnableReflection,
		MaxRecvMsgSize:   DefaultMaxRecvMsgSize,
		MaxSendMsgSize:   DefaultMaxSendMsgSize,
	}

	if cfg.Port != 9090 {
		t.Errorf("expected default port 9090, got %d", cfg.Port)
	}
	if !cfg.EnableReflection {
		t.Error("expected default enable_reflection to be true")
	}
	if cfg.MaxRecvMsgSize != 4194304 {
		t.Errorf("expected default max_recv_msg_size 4194304, got %d", cfg.MaxRecvMsgSize)
	}
}
