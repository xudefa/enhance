// Package grpc 提供 gRPC 服务自动配置。
package grpc

import (
	"context"
	"fmt"
	"net"
	"reflect"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func init() {
	boot.RegisterAutoConfigWith(&GrpcAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(GrpcEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityWebLayer)),
	)
}

// GrpcAutoConfiguration gRPC 服务自动配置类。
type GrpcAutoConfiguration struct {
	logger   log.Logger
	server   *grpc.Server
	listener net.Listener
	config   *GrpcConfig
}

// Configure 配置 gRPC 服务器。
func (c *GrpcAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("加载 gRPC 配置失败: %w", err)
	}

	c.config = cfg

	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.MaxSendMsgSize),
	}

	c.server = grpc.NewServer(opts...)

	if cfg.EnableReflection {
		reflection.Register(c.server)
	}

	if err := ctx.Container().RegisterInstance(c.server, reflect.TypeFor[*grpc.Server]()); err != nil {
		return fmt.Errorf("注册 gRPC Server 失败: %w", err)
	}

	c.logger.Info(context.Background(), "gRPC 服务器已配置",
		log.KeyValue{Key: "port", Value: cfg.Port},
	)

	return nil
}

// Start 启动 gRPC 服务器。
func (c *GrpcAutoConfiguration) Start() error {
	addr := fmt.Sprintf(":%d", c.config.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("gRPC 监听失败: %w", err)
	}
	c.listener = listener

	c.logger.Info(context.Background(), "gRPC 服务器启动中",
		log.KeyValue{Key: "addr", Value: addr},
	)

	go func() {
		if err := c.server.Serve(listener); err != nil {
			c.logger.Error(context.Background(), "gRPC 服务器错误",
				log.KeyValue{Key: "error", Value: err.Error()},
			)
		}
	}()

	return nil
}

// Stop 停止 gRPC 服务器。
func (c *GrpcAutoConfiguration) Stop() {
	if c.server != nil {
		c.server.GracefulStop()
	}
}

// GetServer 获取 gRPC 服务器实例。
func (c *GrpcAutoConfiguration) GetServer() *grpc.Server {
	return c.server
}

// RegisterService 注册 gRPC 服务。
func (c *GrpcAutoConfiguration) RegisterService(desc *grpc.ServiceDesc, impl any) {
	c.server.RegisterService(desc, impl)
}

// GrpcConfig gRPC 服务器配置。
type GrpcConfig struct {
	Enabled          bool `json:"enabled" mapstructure:"enabled"`
	Port             int  `json:"port" mapstructure:"port"`
	EnableReflection bool `json:"enable_reflection" mapstructure:"enable_reflection"`
	MaxRecvMsgSize   int  `json:"max_recv_msg_size" mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize   int  `json:"max_send_msg_size" mapstructure:"max_send_msg_size"`
}

// 配置常量。
const (
	GrpcEnabled             = "grpc.enabled"
	DefaultGrpcPort         = 9090
	DefaultEnableReflection = true
	DefaultMaxRecvMsgSize   = 1024 * 1024 * 4
	DefaultMaxSendMsgSize   = 1024 * 1024 * 4
	ConditionTrue           = "true"
)

// loadConfig 从 Environment 加载 gRPC 配置。
func (c *GrpcAutoConfiguration) loadConfig(env *environment.Environment) (*GrpcConfig, error) {
	cfg := &GrpcConfig{
		Port:             DefaultGrpcPort,
		EnableReflection: DefaultEnableReflection,
		MaxRecvMsgSize:   DefaultMaxRecvMsgSize,
		MaxSendMsgSize:   DefaultMaxSendMsgSize,
	}

	if err := env.BindPrefix("grpc", cfg); err != nil {
		return nil, fmt.Errorf("绑定 gRPC 配置失败: %w", err)
	}

	return cfg, nil
}
