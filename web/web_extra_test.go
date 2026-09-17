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

	var respBody map[string]any
	err := json.Unmarshal([]byte(resp.Body()), &respBody)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	userList := respBody["data"].([]any)
	if len(userList) != 0 {
		t.Fatalf("空列表应该返回 0 个用户，got: %d", len(userList))
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

	var respBody map[string]any
	err := json.Unmarshal([]byte(resp.Body()), &respBody)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	userList := respBody["data"].([]any)
	if len(userList) != 3 {
		t.Fatalf("应该有 3 个用户，got: %d", len(userList))
	}

	total := respBody["total"].(float64)
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

	userList := listResp["data"].([]any)
	if len(userList) != 1 {
		t.Fatalf("应该有 1 个用户，got: %d", len(userList))
	}

	firstUser := userList[0].(map[string]any)
	if firstUser["name"] != "Alice" {
		t.Fatalf("用户名称应该是 Alice，got: %v", firstUser["name"])
	}
}
