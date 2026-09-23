# enhance 测试指南

本文档详细说明 enhance 项目的测试规范和最佳实践。

## 1. 基本原则

### 1.1 表驱动测试
所有测试必须使用表驱动风格，标准模板：

```go
func TestService_DoSomething(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "hello", "expected", false},
        {"empty input", "", "", true},
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got, err := DoSomething(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("DoSomething(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("DoSomething(%q) = %v, want %v", tt.input, got, tt.want)
            }
        })
    }
}
```

### 1.2 并发测试
- 独立测试函数：开头调用 `t.Parallel()`
- 表驱动子测试：内部调用 `t.Parallel()`
- 并发场景测试：使用 goroutine + channel 验证并发安全性

### 1.3 测试命名
- 测试函数：`Test{Type}_{Method}_{Scenario}`
- 子测试：`{scenario}` (简洁描述)

## 2. 测试覆盖要求

### 2.1 必须覆盖的场景
- **正常路径**：典型输入，期望输出
- **错误路径**：无效输入、边界条件、nil/空值
- **边界**：零值、最大值、并发场景（如适用）
- **上下文**：context 取消、超时

### 2.2 覆盖率目标
- 整体覆盖率：≥ 80%
- 核心接口覆盖率：≥ 90%
- 错误处理覆盖率：100%

## 3. 特殊测试场景

### 3.1 并发安全测试
```go
func TestService_Concurrent(t *testing.T) {
    t.Parallel()

    instance := NewService()
    ctx := context.Background()

    const goroutines = 100
    errs := make(chan error, goroutines)

    for i := 0; i < goroutines; i++ {
        go func() {
            _, err := instance.DoSomething(ctx, "test")
            errs <- err
        }()
    }

    for i := 0; i < goroutines; i++ {
        if err := <-errs; err != nil {
            t.Errorf("Concurrent DoSomething() error = %v", err)
        }
    }
}
```

### 3.2 Context 测试
```go
func TestService_ContextCancellation(t *testing.T) {
    t.Parallel()

    instance := NewService()
    ctx, cancel := context.WithCancel(context.Background())
    cancel()

    _, err := instance.DoSomething(ctx, "test")
    if err == nil {
        t.Error("DoSomething() expected error on cancelled context")
    }
}
```

### 3.3 超时测试
```go
func TestService_Timeout(t *testing.T) {
    t.Parallel()

    instance := NewService()
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
    defer cancel()

    time.Sleep(10 * time.Millisecond) // 确保超时

    _, err := instance.DoSomething(ctx, "test")
    if err == nil {
        t.Error("DoSomething() expected timeout error")
    }
}
```

## 4. 测试工具

### 4.1 使用 testing 包
```go
import "testing"

func TestSomething(t *testing.T) {
    t.Parallel()
    // 测试逻辑
}
```

### 4.2 使用 testify (可选)
```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSomething(t *testing.T) {
    t.Parallel()

    result, err := DoSomething()
    
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
    
    // 或使用 require (失败时立即停止)
    require.NoError(t, err)
    require.Equal(t, expected, result)
}
```

## 5. 测试文件组织

### 5.1 文件命名
- 单元测试：`{package}_test.go`
- 基准测试：`{package}_benchmark_test.go`
- 示例测试：`{package}_example_test.go`

### 5.2 文件位置
- 与源码同目录
- 不分散测试文件（一个包一个测试文件）

### 5.3 导入分组
```go
import (
    // 标准库
    "context"
    "testing"
    "time"

    // 项目内部包
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/testing"
)
```

## 6. 常见错误

### 6.1 忘记 t.Parallel()
❌ 错误：
```go
func TestSomething(t *testing.T) {
    // 缺少 t.Parallel()
    // 测试逻辑
}
```

✅ 正确：
```go
func TestSomething(t *testing.T) {
    t.Parallel()
    // 测试逻辑
}
```

### 6.2 子测试未并发
❌ 错误：
```go
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // 缺少 t.Parallel()
        // 测试逻辑
    })
}
```

✅ 正确：
```go
for _, tt := range tests {
    tt := tt
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // 测试逻辑
    })
}
```

### 6.3 使用全局变量
❌ 错误：
```go
var globalVar = "test"

func TestSomething(t *testing.T) {
    globalVar = "modified"
    // 测试逻辑
}
```

✅ 正确：
```go
func TestSomething(t *testing.T) {
    t.Parallel()
    localVar := "test"
    // 使用 localVar 进行测试
}
```

## 7. 运行测试

### 7.1 运行所有测试
```bash
go test ./...
```

### 7.2 运行特定包测试
```bash
go test ./core/...
```

### 7.3 运行特定测试
```bash
go test -run TestService_DoSomething ./...
```

### 7.4 并发测试
```bash
go test -parallel 4 ./...
```

### 7.5 竞态检测
```bash
go test -race ./...
```

### 7.6 覆盖率报告
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 8. 最佳实践

### 8.1 测试隔离
- 每个测试独立，不依赖其他测试
- 使用 `t.Parallel()` 提高执行速度
- 避免共享状态

### 8.2 测试数据
- 使用表驱动测试组织测试数据
- 测试数据清晰、易于理解
- 覆盖正常和异常场景

### 8.3 断言
- 使用明确的断言消息
- 检查错误类型和消息
- 验证返回值和状态

### 8.4 清理
- 使用 `t.Cleanup()` 清理资源
- 避免测试间污染
- 及时释放资源

## 9. 参考资料

- [AGENTS.md §2.3](../AGENTS.md#23-单元测试并发规范)
- [CODING_STYLE.md §12](../CODING_STYLE.md#12-测试风格)
- [Go 测试文档](https://pkg.go.dev/testing)
