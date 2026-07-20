package main

import (
	"fmt"
	"reflect"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/security"
	"github.com/xudefa/enhance/web/mvc"
	"github.com/xudefa/enhance/web/server"
)

// WebAutoConfig Web 自动配置
type WebAutoConfig struct {
}

// Configure 配置 Web 组件
func (c *WebAutoConfig) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	// 读取服务器配置
	var serverConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	if err := env.BindPrefix("server", &serverConfig); err != nil {
		return fmt.Errorf("绑定服务器配置失败: %w", err)
	}

	if serverConfig.Port == 0 {
		serverConfig.Port = 8080
	}
	if serverConfig.Host == "" {
		serverConfig.Host = "0.0.0.0"
	}

	// 创建路由器（使用 server.NewRouter 创建默认路由器）
	router := server.NewRouter()

	// 创建 HttpServer 适配器
	httpServer := server.NewHttpServerAdapter(
		server.WithHost(fmt.Sprintf("%s:%d", serverConfig.Host, serverConfig.Port)),
	)

	// 创建 WebStarter
	webStarter := mvc.NewWebStarter(
		mvc.WithConfig(mvc.WebConfig{
			Host: serverConfig.Host,
			Port: serverConfig.Port,
		}),
		mvc.WithRouter(router),
		mvc.WithServer(httpServer),
	)

	// 从容器获取 SecurityFilterChain 并设置为 handler
	container := ctx.Container()
	filterChainBeans, err := container.Get(reflect.TypeFor[security.SecurityFilterChain]())
	if err != nil || len(filterChainBeans) == 0 {
		fmt.Printf("[Web] 警告: 未找到 SecurityFilterChain (error: %v)，将使用默认路由器\n", err)
	} else {
		if filterChain, ok := filterChainBeans[0].(security.SecurityFilterChain); ok {
			// 创建安全过滤器链处理器
			filterChainHandler := security.NewSecurityFilterChainHandler(filterChain, router)
			webStarter.SetHandler(filterChainHandler)
			fmt.Println("[Web] SecurityFilterChain 已设置为处理器")
		}
	}

	// 注册 WebStarter 到全局启动器注册表
	fmt.Println("[Web] 正在注册 WebStarter...")
	boot.RegisterStarter(webStarter)
	fmt.Println("[Web] WebStarter 已注册")

	fmt.Printf("[Web] Web 服务器配置: %s:%d\n", serverConfig.Host, serverConfig.Port)
	return nil
}

func init() {
	// 注册自动配置
	fmt.Println("[Web] init() 执行，注册 WebAutoConfig...")
	boot.RegisterAutoConfig(&WebAutoConfig{})
	fmt.Println("[Web] WebAutoConfig 已注册")
}
