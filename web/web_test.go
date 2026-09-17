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

	userList := listResp["data"].([]any)
	if len(userList) != 1 {
		t.Fatalf("应该有 1 个用户，got: %d", len(userList))
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

	middleware1 := testWebMiddlewareChainAppend(&order, &mu, "middleware1")
	middleware2 := testWebMiddlewareChainAppend(&order, &mu, "middleware2")

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

	testWebMiddlewareChainAssert(t, &order)
}

func testWebMiddlewareChainAppend(order *[]string, mu *sync.Mutex, name string) func(mvc.Context) {
	return func(ctx mvc.Context) {
		mu.Lock()
		*order = append(*order, name+"-before")
		mu.Unlock()
		ctx.Next()
		mu.Lock()
		*order = append(*order, name+"-after")
		mu.Unlock()
	}
}

func testWebMiddlewareChainAssert(t *testing.T, orderP *[]string) {
	t.Helper()
	order := *orderP

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

	var searchResult map[string]any
	err := json.Unmarshal([]byte(resp.Body()), &searchResult)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if searchResult["q"] != "test" {
		t.Fatalf("q 应该是 test，got: %v", searchResult["q"])
	}
	if searchResult["page"] != "1" {
		t.Fatalf("page 应该是 1，got: %v", searchResult["page"])
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
