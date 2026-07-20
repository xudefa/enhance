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

// SecurityConfig 安全配置打印
type SecurityConfig struct {
	Enabled bool `json:"enabled"`
	JWT     JWTConfig
	Casbin  CasbinConfig
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Enabled      bool   `json:"enabled"`
	SecretKey    string `json:"secret-key"`
	Expiration   int    `json:"expiration"`
	ExcludePaths string `json:"exclude-paths"`
}

// CasbinConfig Casbin 配置
type CasbinConfig struct {
	Enabled          bool   `json:"enabled"`
	ModelType        string `json:"model-type"`
	ModelPath        string `json:"model-path"`
	PolicyType       string `json:"policy-type"`
	TableName        string `json:"table-name"`
	AutoCreateTable  bool   `json:"auto-create-table"`
	AutoLoad         bool   `json:"auto-load"`
	AutoLoadInterval int    `json:"auto-load-interval"`
}

// ZerologConfig Zerolog 日志配置
type ZerologConfig struct {
	Enabled    bool   `json:"enabled"`
	Level      string `json:"level"`
	Format     string `json:"format"`
	TimeFormat string `json:"time-format"`
	AddSource  bool   `json:"add-source"`
	OutputPath string `json:"output-path"`
}

// SecurityAutoConfig 安全自动配置（打印配置信息）
type SecurityAutoConfig struct {
}

// Configure 打印安全配置信息
func (c *SecurityAutoConfig) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	// 打印 JWT 配置
	var jwtCfg JWTConfig
	if err := env.BindPrefix("security.jwt", &jwtCfg); err == nil {
		fmt.Printf("[Security] JWT 配置: enabled=%v, secret-key=%s, expiration=%d\n",
			jwtCfg.Enabled, maskSecret(jwtCfg.SecretKey), jwtCfg.Expiration)
		fmt.Printf("[Security] JWT 排除路径: %s\n", jwtCfg.ExcludePaths)
	}

	// 打印 Casbin 配置
	var casbinCfg CasbinConfig
	if err := env.BindPrefix("security.casbin", &casbinCfg); err == nil {
		fmt.Printf("[Security] Casbin 配置: enabled=%v, model-type=%s, model-path=%s\n",
			casbinCfg.Enabled, casbinCfg.ModelType, casbinCfg.ModelPath)
		fmt.Printf("[Security] Casbin 策略存储: policy-type=%s, table-name=%s, auto-create-table=%v\n",
			casbinCfg.PolicyType, casbinCfg.TableName, casbinCfg.AutoCreateTable)
		fmt.Printf("[Security] Casbin 自动刷新: auto-load=%v, interval=%d分钟\n",
			casbinCfg.AutoLoad, casbinCfg.AutoLoadInterval)
	}

	// 打印 Zerolog 配置
	var zerologCfg ZerologConfig
	if err := env.BindPrefix("log.zerolog", &zerologCfg); err == nil {
		fmt.Printf("[Zerolog] 配置: enabled=%v, level=%s, format=%s, add-source=%v\n",
			zerologCfg.Enabled, zerologCfg.Level, zerologCfg.Format, zerologCfg.AddSource)
		if zerologCfg.OutputPath != "" {
			fmt.Printf("[Zerolog] 日志输出路径: %s\n", zerologCfg.OutputPath)
		}
	}

	return nil
}

// maskSecret 隐藏密钥
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

func init() {
	// 注册自动配置（注意顺序：SecurityAutoConfig 必须在 WebAutoConfig 之前）
	// 使用安全核心层优先级，在认证和授权之后执行
	boot.RegisterAutoConfigWith(&SecurityAutoConfig{},
		boot.WithOrder(int(boot.OrderPrioritySecurityCore)),
		boot.WithAfter("github.com/xudefa/enhance/security.SecurityAutoConfiguration"),
	)
	// 使用 Web 层优先级，在安全配置之后执行
	boot.RegisterAutoConfigWith(&WebAutoConfig{},
		boot.WithOrder(int(boot.OrderPriorityWebLayer)),
		boot.WithAfter("github.com/xudefa/enhance/security.SecurityAutoConfiguration"),
	)
}
