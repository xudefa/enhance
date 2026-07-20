// Package main 演示如何使用 enhance.Run() 简洁启动 Web 服务
//
// 参考 Spring Boot 的 SpringApplication.run() 设计，
// 一行代码即可启动一个完整的 Web 服务，包含 GORM 数据库支持。
package main

import (
	"fmt"
	"net/http"

	"github.com/xudefa/enhance"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/web/mvc"
)

// HelloController 简单的 Hello 控制器
type HelloController struct{}

// Routes 注册路由
// 使用 mvc.Router 接口而非具体实现，方便后续替换为 gin/hertz 等其他框架
func (c *HelloController) Routes(router mvc.Router) {
	router.GET("/hello", c.Hello)
	router.GET("/hello/{name}", c.HelloName)
}

// Hello 返回 Hello World
func (c *HelloController) Hello(ctx mvc.Context) {
	ctx.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "Hello, World!",
	})
}

// HelloName 返回带名字的 Hello
func (c *HelloController) HelloName(ctx mvc.Context) {
	name := ctx.PathParam("name")
	ctx.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": fmt.Sprintf("Hello, %s!", name),
	})
}

func init() {
	// 注册控制器
	mvc.RegisterController(&HelloController{})
}

func main() {
	// 方式 1：最简用法 - 一行启动 Web 服务
	// 自动加载 config/application.json 配置
	// 自动启动所有 Starter（包括 Web Server）
	// enhance.Run()

	// 方式 2：带配置选项启动
	// enhance.Run(
	//     boot.WithAppName("my-web-app"),
	//     boot.WithVersion("1.0.0"),
	//     boot.WithProfiles("dev"),
	// )

	// 方式 3：带 GORM 自动配置启动 - 连接 MySQL 数据库
	// GORM、Web、Database 自动配置会在 init() 中注册
	enhance.Run(
		boot.WithAppName("enhance-gorm-demo"),
		boot.WithVersion("1.0.0"),
	)
}
