package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"golang.org/x/time/rate"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

// 定义任务类型常量
const (
	TaskTypeEmailSend = "email:send"
	TaskTypeReportGen = "report:generate"
)

// UserRequest 用户请求结构体（带验证标签）
type UserRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,min=1,max=130"`
	Phone string `json:"phone" validate:"required,phone"`
}

// UserResponse 用户响应结构体
type UserResponse struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

func main() {
	// 创建应用
	app, err := boot.NewApplication(
		boot.WithAppName("enhance-utilities-demo"),
	)
	if err != nil {
		panic(err)
	}
	defer app.Stop()

	// 获取所有服务实例
	validator := core.MustGet[*validator.Validate](app.Container(), "validator")
	limiter := core.MustGet[*rate.Limiter](app.Container(), "limiter")
	cronMgr := core.MustGet[*cron.Cron](app.Container(), "cronMgr")
	asynqClient := core.MustGet[*asynq.Client](app.Container(), "asynqClient")

	// 创建 Gin 路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// 限流中间件
	r.Use(func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后重试",
			})
			return
		}
		c.Next()
	})

	// 注册路由
	r.POST("/api/users", createUserHandler(validator))
	r.GET("/api/health", healthHandler())
	r.POST("/api/tasks/email", sendEmailHandler(asynqClient))

	// 创建日志记录器
	logger := log.Build()

	// 配置定时任务
	setupCronJobs(cronMgr, logger)

	// 启动 HTTP 服务器
	go func() {
		logger.Info(context.Background(), "HTTP 服务器启动在 :8080")
		if err := r.Run(":8080"); err != nil {
			logger.Error(context.Background(), "HTTP 服务器启动失败", log.KeyValue{Key: "error", Value: err.Error()})
		}
	}()

	// 启动应用
	app.Start()
}

// createUserHandler 创建用户处理器（带验证）
func createUserHandler(v *validator.Validate) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 验证请求数据
		if err := v.Struct(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "数据验证失败",
				"details": formatValidationErrors(err),
			})
			return
		}

		// 处理业务逻辑
		resp := UserResponse{
			ID:      1,
			Name:    req.Name,
			Email:   req.Email,
			Message: "用户创建成功",
		}

		c.JSON(http.StatusCreated, resp)
	}
}

// healthHandler 健康检查处理器
func healthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	}
}

// sendEmailHandler 发送邮件任务处理器
func sendEmailHandler(client *asynq.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			To      string `json:"to" validate:"required,email"`
			Subject string `json:"subject" validate:"required"`
			Body    string `json:"body" validate:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 创建异步任务
		payload, _ := json.Marshal(req)
		task := asynq.NewTask(TaskTypeEmailSend, payload)

		// 添加到队列
		info, err := client.Enqueue(task,
			asynq.MaxRetry(3),
			asynq.Timeout(5*time.Minute),
			asynq.Queue("default"),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "任务创建失败",
			})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"message": "邮件发送任务已创建",
			"task_id": info.ID,
			"queue":   info.Queue,
		})
	}
}

// setupCronJobs 配置定时任务
func setupCronJobs(c *cron.Cron, logger log.Logger) {
	// 每分钟执行一次
	c.AddFunc("0 * * * * *", func() {
		logger.Info(context.Background(), "定时任务：每分钟执行")
	})

	// 每小时 30 分执行
	c.AddFunc("0 30 * * * *", func() {
		logger.Info(context.Background(), "定时任务：每小时 30 分执行")
	})

	// 每天凌晨 2 点执行
	c.AddFunc("0 0 2 * * *", func() {
		logger.Info(context.Background(), "定时任务：每天凌晨 2 点执行")
	})

	logger.Info(context.Background(), "定时任务已配置")
}

// formatValidationErrors 格式化验证错误
func formatValidationErrors(err error) []string {
	if errs, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, e := range errs {
			messages = append(messages, fmt.Sprintf(
				"字段 '%s' 验证失败: 规则=%s",
				e.Field(),
				e.Tag(),
			))
		}
		return messages
	}
	return []string{err.Error()}
}
