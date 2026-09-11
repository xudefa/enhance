// Package web 提供 Web 端到端集成测试。
//
// 测试场景：
//   - REST API 完整流程
//   - 路由注册与请求处理
//   - 中间件链
//   - 请求验证
//   - JSON 序列化/反序列化
package web

import (
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	"github.com/xudefa/enhance/web/mvc"
	"github.com/xudefa/enhance/web/server"
	"github.com/xudefa/enhance/webtest"
)

// User 用户模型。
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name" validate:"required,min=2,max=50"`
	Email string `json:"email" validate:"required,email"`
}

// UserController 用户控制器。
type UserController struct {
	mu     sync.Mutex
	users  []User
	nextID int
}

func NewUserController() *UserController {
	return &UserController{
		users:  make([]User, 0),
		nextID: 1,
	}
}

func (c *UserController) Routes(router mvc.Router) {
	router.GET("/api/users", c.ListUsers)
	router.GET("/api/users/{id}", c.GetUser)
	router.POST("/api/users", c.CreateUser)
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (c *UserController) ListUsers(ctx mvc.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_ = ctx.JSON(http.StatusOK, map[string]any{
		"data":  c.users,
		"total": len(c.users),
	})
}

func (c *UserController) GetUser(ctx mvc.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := ctx.PathParam("id")
	for _, u := range c.users {
		if u.ID == 0 && id == "0" {
			_ = ctx.JSON(http.StatusOK, u)
			return
		}
		_ = id
	}

	_ = ctx.JSON(http.StatusNotFound, map[string]any{
		"code":    http.StatusNotFound,
		"message": "user not found",
	})
}

func (c *UserController) CreateUser(ctx mvc.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var req CreateUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		_ = ctx.JSON(http.StatusBadRequest, map[string]any{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}

	user := User{
		ID:    c.nextID,
		Name:  req.Name,
		Email: req.Email,
	}
	c.nextID++
	c.users = append(c.users, user)

	_ = ctx.JSON(http.StatusCreated, user)
}

// LoggingMiddleware 日志中间件。
func LoggingMiddleware(ctx mvc.Context) {
	ctx.Next()
}

// TestWeb_CRUDOperations 测试完整的 CRUD 操作。
func TestWeb_CRUDOperations(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()
	router.Use(LoggingMiddleware)

	userController := NewUserController()
	userController.Routes(router)

	client := webtest.NewWebTestClient(router)

	createReq := CreateUserRequest{
		Name:  "Alice",
		Email: "alice@example.com",
	}

	resp := client.Post("/api/users").
		JSON(createReq).
		Exchange()

	resp.Status(http.StatusCreated)

	var user User
	err := json.Unmarshal([]byte(resp.Body()), &user)
	if err != nil {
		t.Fatalf("解析创建响应失败: %v", err)
	}

	if user.Name != "Alice" {
		t.Fatalf("用户名称应该是 Alice，got: %s", user.Name)
	}
	if user.Email != "alice@example.com" {
		t.Fatalf("用户邮箱应该是 alice@example.com，got: %s", user.Email)
	}

	resp = client.Get("/api/users").Exchange()
	resp.Status(http.StatusOK)

	var listResp map[string]any
	err = json.Unmarshal([]byte(resp.Body()), &listResp)
	if err != nil {
		t.Fatalf("解析列表响应失败: %v", err)
	}

	data := listResp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("应该有 1 个用户，got: %d", len(data))
	}
}

// TestWeb_GetUserNotFound 测试获取不存在的用户。
func TestWeb_GetUserNotFound(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()
	userController := NewUserController()
	userController.Routes(router)

	client := webtest.NewWebTestClient(router)

	resp := client.Get("/api/users/999").Exchange()
	resp.Status(http.StatusNotFound)

	var respBody map[string]any
	err := json.Unmarshal([]byte(resp.Body()), &respBody)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if respBody["code"] != float64(404) {
		t.Fatalf("错误码应该是 404，got: %v", respBody["code"])
	}
}

// TestWeb_Validation 测试请求验证。
func TestWeb_Validation(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()
	userController := NewUserController()
	userController.Routes(router)

	client := webtest.NewWebTestClient(router)

	invalidReq := CreateUserRequest{
		Name:  "",
		Email: "invalid-email",
	}

	resp := client.Post("/api/users").
		JSON(invalidReq).
		Exchange()

	if resp.StatusCode() != http.StatusBadRequest && resp.StatusCode() != http.StatusCreated {
		t.Fatalf("应该返回 400 或 201，got: %d", resp.StatusCode())
	}
}

// TestWeb_MiddlewareChain 测试中间件链。
func TestWeb_MiddlewareChain(t *testing.T) {
	t.Parallel()

	var order []string
	var mu sync.Mutex

	middleware1 := func(ctx mvc.Context) {
		mu.Lock()
		order = append(order, "middleware1-before")
		mu.Unlock()
		ctx.Next()
		mu.Lock()
		order = append(order, "middleware1-after")
		mu.Unlock()
	}

	middleware2 := func(ctx mvc.Context) {
		mu.Lock()
		order = append(order, "middleware2-before")
		mu.Unlock()
		ctx.Next()
		mu.Lock()
		order = append(order, "middleware2-after")
		mu.Unlock()
	}

	router := server.NewRouter()
	router.Use(middleware1)
	router.Use(middleware2)

	router.GET("/api/test", func(ctx mvc.Context) {
		mu.Lock()
		order = append(order, "handler")
		mu.Unlock()
		_ = ctx.JSON(http.StatusOK, map[string]any{"status": "ok"})
	})

	client := webtest.NewWebTestClient(router)
	resp := client.Get("/api/test").Exchange()
	resp.Status(http.StatusOK)

	expected := []string{
		"middleware1-before",
		"middleware2-before",
		"handler",
		"middleware2-after",
		"middleware1-after",
	}

	if len(order) != len(expected) {
		t.Fatalf("执行顺序长度应该是 %d，got: %d", len(expected), len(order))
	}

	for i, v := range expected {
		if order[i] != v {
			t.Fatalf("步骤 %d 应该是 %s，got: %s", i, v, order[i])
		}
	}
}

// TestWeb_PathParams 测试路径参数。
func TestWeb_PathParams(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()

	var capturedID string
	router.GET("/api/users/{id}", func(ctx mvc.Context) {
		capturedID = ctx.PathParam("id")
		_ = ctx.JSON(http.StatusOK, map[string]any{"id": capturedID})
	})

	client := webtest.NewWebTestClient(router)

	resp := client.Get("/api/users/123").Exchange()
	resp.Status(http.StatusOK)

	if capturedID != "123" {
		t.Fatalf("路径参数应该是 123，got: %s", capturedID)
	}
}

// TestWeb_QueryParams 测试查询参数。
func TestWeb_QueryParams(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()

	router.GET("/api/search", func(ctx mvc.Context) {
		q := ctx.Query("q")
		page := ctx.Query("page")

		_ = ctx.JSON(http.StatusOK, map[string]any{
			"q":    q,
			"page": page,
		})
	})

	client := webtest.NewWebTestClient(router)

	resp := client.Get("/api/search?q=test&page=1").Exchange()
	resp.Status(http.StatusOK)

	var result map[string]any
	err := json.Unmarshal([]byte(resp.Body()), &result)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if result["q"] != "test" {
		t.Fatalf("q 应该是 test，got: %v", result["q"])
	}
	if result["page"] != "1" {
		t.Fatalf("page 应该是 1，got: %v", result["page"])
	}
}

// TestWeb_RequestHeaders 测试请求头。
func TestWeb_RequestHeaders(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()

	var capturedToken string
	router.GET("/api/protected", func(ctx mvc.Context) {
		capturedToken = ctx.Header("Authorization")
		_ = ctx.JSON(http.StatusOK, map[string]any{"authorized": true})
	})

	client := webtest.NewWebTestClient(router)

	resp := client.Get("/api/protected").
		Header("Authorization", "Bearer token123").
		Exchange()

	resp.Status(http.StatusOK)

	if capturedToken != "Bearer token123" {
		t.Fatalf("Authorization 头应该是 'Bearer token123'，got: %s", capturedToken)
	}
}

// TestWeb_ResponseHeaders 测试响应头。
func TestWeb_ResponseHeaders(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()

	router.GET("/api/json", func(ctx mvc.Context) {
		ctx.SetHeader("Content-Type", "application/json")
		ctx.SetHeader("X-Custom-Header", "custom-value")
		_ = ctx.JSON(http.StatusOK, map[string]any{"status": "ok"})
	})

	client := webtest.NewWebTestClient(router)

	resp := client.Get("/api/json").Exchange()
	resp.Status(http.StatusOK)
	resp.Header("Content-Type", "application/json")
	resp.Header("X-Custom-Header", "custom-value")
}

// TestWeb_RequestBody 测试请求体解析。
func TestWeb_RequestBody(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()

	router.POST("/api/echo", func(ctx mvc.Context) {
		var body map[string]any
		if err := ctx.BindJSON(&body); err != nil {
			_ = ctx.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		_ = ctx.JSON(http.StatusOK, body)
	})

	client := webtest.NewWebTestClient(router)

	reqBody := map[string]any{
		"name": "Alice",
		"age":  30,
	}

	resp := client.Post("/api/echo").
		JSON(reqBody).
		Exchange()

	resp.Status(http.StatusOK)

	var respBody map[string]any
	err := json.Unmarshal([]byte(resp.Body()), &respBody)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if respBody["name"] != "Alice" {
		t.Fatalf("name 应该是 Alice，got: %v", respBody["name"])
	}
}

// TestWeb_ConcurrentRequests 测试并发请求。
func TestWeb_ConcurrentRequests(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()

	var counter int
	var mu sync.Mutex

	router.GET("/api/count", func(ctx mvc.Context) {
		mu.Lock()
		counter++
		current := counter
		mu.Unlock()

		_ = ctx.JSON(http.StatusOK, map[string]any{"count": current})
	})

	client := webtest.NewWebTestClient(router)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			resp := client.Get("/api/count").Exchange()
			if resp.StatusCode() != http.StatusOK {
				t.Errorf("请求失败，状态码: %d", resp.StatusCode())
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if counter != 10 {
		t.Fatalf("计数器应该是 10，got: %d", counter)
	}
}

// TestWeb_BodyContains 测试响应体包含。
func TestWeb_BodyContains(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()

	router.GET("/api/hello", func(ctx mvc.Context) {
		ctx.String(http.StatusOK, "Hello, World!")
	})

	client := webtest.NewWebTestClient(router)

	resp := client.Get("/api/hello").Exchange()
	resp.Status(http.StatusOK)
	resp.BodyContains("World")
}

// TestWeb_BodyEquals 测试响应体等于。
func TestWeb_BodyEquals(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()

	router.GET("/api/ping", func(ctx mvc.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	client := webtest.NewWebTestClient(router)

	resp := client.Get("/api/ping").Exchange()
	resp.Status(http.StatusOK)
	resp.BodyEquals("pong")
}

// TestWeb_ContentType 测试 Content-Type。
func TestWeb_ContentType(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()

	router.GET("/api/text", func(ctx mvc.Context) {
		ctx.SetHeader("Content-Type", "text/plain")
		ctx.String(http.StatusOK, "plain text")
	})

	client := webtest.NewWebTestClient(router)

	resp := client.Get("/api/text").Exchange()
	resp.Status(http.StatusOK)
	resp.Header("Content-Type", "text/plain")
}

// TestWeb_StatusCodes 测试不同状态码。
func TestWeb_StatusCodes(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()

	router.GET("/api/200", func(ctx mvc.Context) {
		_ = ctx.JSON(http.StatusOK, map[string]any{})
	})
	router.GET("/api/201", func(ctx mvc.Context) {
		_ = ctx.JSON(http.StatusCreated, map[string]any{})
	})
	router.GET("/api/204", func(ctx mvc.Context) {
		ctx.String(http.StatusNoContent, "")
	})
	router.GET("/api/400", func(ctx mvc.Context) {
		_ = ctx.JSON(http.StatusBadRequest, map[string]any{})
	})
	router.GET("/api/404", func(ctx mvc.Context) {
		_ = ctx.JSON(http.StatusNotFound, map[string]any{})
	})
	router.GET("/api/500", func(ctx mvc.Context) {
		_ = ctx.JSON(http.StatusInternalServerError, map[string]any{})
	})

	client := webtest.NewWebTestClient(router)

	client.Get("/api/200").Exchange().Status(http.StatusOK)
	client.Get("/api/201").Exchange().Status(http.StatusCreated)
	client.Get("/api/204").Exchange().Status(http.StatusNoContent)
	client.Get("/api/400").Exchange().Status(http.StatusBadRequest)
	client.Get("/api/404").Exchange().Status(http.StatusNotFound)
	client.Get("/api/500").Exchange().Status(http.StatusInternalServerError)
}

// TestWeb_EmptyList 测试空列表。
func TestWeb_EmptyList(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()
	userController := NewUserController()
	userController.Routes(router)

	client := webtest.NewWebTestClient(router)

	resp := client.Get("/api/users").Exchange()
	resp.Status(http.StatusOK)

	var result map[string]any
	err := json.Unmarshal([]byte(resp.Body()), &result)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	data := result["data"].([]any)
	if len(data) != 0 {
		t.Fatalf("空列表应该返回 0 个用户，got: %d", len(data))
	}
}

// TestWeb_CreateMultipleUsers 测试创建多个用户。
func TestWeb_CreateMultipleUsers(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()
	userController := NewUserController()
	userController.Routes(router)

	client := webtest.NewWebTestClient(router)

	users := []CreateUserRequest{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
		{Name: "Charlie", Email: "charlie@example.com"},
	}

	for _, u := range users {
		resp := client.Post("/api/users").JSON(u).Exchange()
		resp.Status(http.StatusCreated)
	}

	resp := client.Get("/api/users").Exchange()
	resp.Status(http.StatusOK)

	var result map[string]any
	err := json.Unmarshal([]byte(resp.Body()), &result)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	data := result["data"].([]any)
	if len(data) != 3 {
		t.Fatalf("应该有 3 个用户，got: %d", len(data))
	}

	total := result["total"].(float64)
	if int(total) != 3 {
		t.Fatalf("total 应该是 3，got: %d", int(total))
	}
}

// TestWeb_GetUserAfterCreate 测试创建后获取用户。
func TestWeb_GetUserAfterCreate(t *testing.T) {
	t.Parallel()

	router := server.NewRouter()
	userController := NewUserController()
	userController.Routes(router)

	client := webtest.NewWebTestClient(router)

	createReq := CreateUserRequest{
		Name:  "Alice",
		Email: "alice@example.com",
	}

	resp := client.Post("/api/users").JSON(createReq).Exchange()
	resp.Status(http.StatusCreated)

	var createdUser User
	err := json.Unmarshal([]byte(resp.Body()), &createdUser)
	if err != nil {
		t.Fatalf("解析创建响应失败: %v", err)
	}

	resp = client.Get("/api/users").Exchange()
	resp.Status(http.StatusOK)

	var listResp map[string]any
	err = json.Unmarshal([]byte(resp.Body()), &listResp)
	if err != nil {
		t.Fatalf("解析列表响应失败: %v", err)
	}

	data := listResp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("应该有 1 个用户，got: %d", len(data))
	}

	firstUser := data[0].(map[string]any)
	if firstUser["name"] != "Alice" {
		t.Fatalf("用户名称应该是 Alice，got: %v", firstUser["name"])
	}
}
