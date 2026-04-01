# Microg

[![GoDoc](https://godoc.org/github.com/atliliw/microg?status.svg)](https://godoc.org/github.com/atliliw/microg)
[![Go Report Card](https://goreportcard.com/badge/github.com/atliliw/microg)](https://goreportcard.com/report/github.com/atliliw/microg)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8.svg)](https://golang.org)

一个简单、轻量的 Go 微服务框架，提供服务注册发现、gRPC/HTTP 服务、链路追踪、结构化日志等功能。

## ✨ 特性

- 🚀 **简单易用** - API 设计简洁，快速上手，最小化依赖
- 🔧 **服务注册发现** - 内置 Consul 支持，自动健康检查
- 📡 **gRPC & HTTP** - 同时支持 gRPC 和 HTTP 服务，可独立或混合部署
- 📊 **可观测性** - 内置 OpenTelemetry 链路追踪、Prometheus 指标
- 📝 **结构化日志** - 基于 zap 的高性能日志，自动关联 traceID
- ⚠️ **错误处理** - 增强版错误包，支持错误码和 HTTP 状态码映射
- 🛡️ **优雅关闭** - 自动处理服务注销和连接关闭
- 🔐 **JWT 认证** - 内置 JWT 中间件，支持多种认证方式
- ⚖️ **负载均衡** - 支持多种负载均衡策略（随机、轮询、加权轮询、P2C）
- 🔧 **依赖注入** - 支持 Wire 代码生成，简化依赖管理

## 📦 安装

```bash
go get github.com/atliliw/microg
```

## 🚀 快速开始

### 创建 gRPC 服务

```go
package main

import (
    "github.com/atliliw/microg/app"
    "github.com/atliliw/microg/registry/consul"
    "github.com/atliliw/microg/server/rpcserver"
    "github.com/atliliw/microg/pkg/log"
    
    "github.com/hashicorp/consul/api"
)

func main() {
    // 创建 Consul 客户端
    consulClient, _ := api.NewClient(api.DefaultConfig())
    
    // 创建服务注册器
    registrar := consul.New(consulClient, 
        consul.WithHealthCheck(true),
    )
    
    // 创建 gRPC 服务器
    rpcServer := rpcserver.NewServer(
        rpcserver.WithAddress(":9000"),
    )
    
    // 注册你的 gRPC 服务
    pb.RegisterYourServiceServer(rpcServer.Server, &yourServiceImpl{})
    
    // 创建并启动应用
    application := app.New(
        app.WithName("your-service"),
        app.WithRPCServer(rpcServer),
        app.WithRegistrar(registrar),
    )
    
    if err := application.Run(); err != nil {
        log.Fatalf("服务启动失败: %v", err)
    }
}
```

### 创建 HTTP 服务

```go
package main

import (
    "github.com/atliliw/microg/app"
    "github.com/atliliw/microg/server/restserver"
    "github.com/atliliw/microg/pkg/log"
)

func main() {
    // 创建 HTTP 服务器
    restServer := restserver.NewServer(
        restserver.WithPort(8080),
    )
    
    // 注册路由
    restServer.GET("/hello", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "hello"})
    })
    
    // 创建并启动应用
    application := app.New(
        app.WithName("your-api"),
        app.WithRestServer(restServer),
    )
    
    if err := application.Run(); err != nil {
        log.Fatalf("服务启动失败: %v", err)
    }
}
```

## 📖 核心模块

### 日志 (pkg/log)

```go
import "github.com/atliliw/microg/pkg/log"

// 结构化日志
log.Info("user action",
    log.String("user", "john"),
    log.Int("action", 1),
)

// Context 日志（自动关联 traceID）
log.InfoC(ctx, "processing order",
    log.String("orderId", "123"),
)

// 错误日志
log.Error("operation failed",
    log.Err(err),
    log.String("operation", "create_user"),
)
```

详细文档：[pkg/log/README.md](pkg/log/README.md)

### 错误处理 (pkg/errors)

```go
import "github.com/atliliw/microg/pkg/errors"

// 创建带错误码的错误
err := errors.WithCode(200001, "user not found")

// 解析错误码
coder := errors.ParseCoder(err)
fmt.Println(coder.Code())        // 200001
fmt.Println(coder.HTTPStatus())  // 404

// gRPC 错误转换
grpcErr := errors.ToGrpcError(err)
```

详细文档：[pkg/errors/README.md](pkg/errors/README.md)

### 链路追踪 (core/trace)

```go
import "github.com/atliliw/microg/core/trace"

// 初始化
trace.Init("user-srv", 
    trace.WithEndpoint("http://localhost:14268/api/traces"),
)

// gRPC/HTTP 服务器自动启用链路追踪
```

## 📁 项目结构

```
microg/
├── app/                # 应用生命周期管理
├── core/               # 核心功能（metric, trace）
├── pkg/                # 公共包（errors, log, storage）
├── registry/           # 服务注册发现（consul）
├── server/             # 服务器（rpcserver, restserver）
├── config/             # 配置管理
└── examples/           # 示例项目
```

## 📚 示例项目

完整示例位于 [examples](./examples) 目录：

- `examples/user/srv` - gRPC 服务端
- `examples/user/client` - HTTP 客户端（API 网关）

### 运行示例

```bash
# 1. 启动 Consul
consul agent -dev

# 2. 启动 MySQL
docker run -d --name mysql -e MYSQL_ROOT_PASSWORD=root -p 3306:3306 mysql:8

# 3. 运行 gRPC 服务端
cd examples/user/srv && go run .

# 4. 运行 HTTP 客户端
cd examples/user/client && go run .

# 5. 测试
curl http://localhost:8080/api/v1/users
```

## ⚙️ 配置

```yaml
# config.yaml
service:
  name: user-srv

grpc:
  address: ":9001"
  timeout: 30s
  metrics: true

consul:
  address: "localhost:8500"
  enabled: true
  health_check: true

tracing:
  enabled: true
  endpoint: "http://localhost:14268/api/traces"

log:
  level: info
  format: json
```

## 🤝 贡献

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add some feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

---

⭐ 如果这个项目对你有帮助，请给一个 Star！