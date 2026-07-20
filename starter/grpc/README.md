# gRPC Starter

gRPC 服务自动配置模块，提供高性能 RPC 框架支持。

## 功能特性

- ✅ 自动配置 gRPC 服务器
- ✅ 支持服务反射
- ✅ 消息大小限制配置
- ✅ 优雅启动和关闭
- ✅ 服务注册支持

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/grpc"
)
```

### 2. 配置文件

在 `application.json` 中添加 gRPC 配置：

```json
{
  "grpc": {
    "enabled": true,
    "port": 9090,
    "enable_reflection": true,
    "max_recv_msg_size": 4194304,
    "max_send_msg_size": 4194304
  }
}
```

### 3. 使用示例

```go
package main

import (
    "context"
    
    "google.golang.org/grpc"
    
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/grpc"
    pb "your/proto/package"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("grpc-demo"),
    )
    defer app.Stop()
    
    // 获取 gRPC 服务器
    server := core.MustGetBean[*grpc.GrpcAutoConfiguration](app.Container())
    
    // 注册服务
    pb.RegisterUserServiceServer(server.GetServer(), &UserService{})
    
    // 启动服务器
    app.Start()
    app.WaitForSignal()
}

// UserService 实现
type UserService struct {
    pb.UnimplementedUserServiceServer
}

func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    return &pb.GetUserResponse{
        Id:   req.Id,
        Name: "John",
    }, nil
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `grpc.enabled` | bool | false | 是否启用 gRPC |
| `grpc.port` | int | 9090 | gRPC 服务器端口 |
| `grpc.enable_reflection` | bool | true | 是否启用服务反射 |
| `grpc.max_recv_msg_size` | int | 4194304 | 最大接收消息大小（字节） |
| `grpc.max_send_msg_size` | int | 4194304 | 最大发送消息大小（字节） |

## 高级用法

### 服务反射

启用反射后，可以使用 `grpcurl` 工具进行调试：

```bash
# 列出服务
grpcurl -plaintext localhost:9090 list

# 调用服务
grpcurl -plaintext -d '{"id": "123"}' localhost:9090 your.package.UserService/GetUser
```

### 拦截器

```go
server := core.MustGetBean[*grpc.GrpcAutoConfiguration](app.Container())

// 添加一元拦截器
grpcServer := server.GetServer()
// 注意：拦截器需要在服务器创建时添加
```

### 流式服务

```go
func (s *UserService) StreamUsers(req *pb.StreamRequest, stream pb.UserService_StreamUsersServer) error {
    for _, user := range users {
        if err := stream.Send(user); err != nil {
            return err
        }
    }
    return nil
}
```

## 启动顺序

- **优先级**: `OrderPriorityWebLayer` (0)
- **触发条件**: `grpc.enabled=true`

## 依赖

- `google.golang.org/grpc`
- `google.golang.org/grpc/reflection`