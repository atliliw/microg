# Microg

一个简单、轻量的 Go 微服务框架，提供服务注册发现、gRPC/HTTP 服务、链路追踪等功能。

## 特性

- 🚀 **简单易用** - API 设计简洁，快速上手
- 🔧 **服务注册发现** - 内置 Consul 支持
- 📡 **gRPC & HTTP** - 同时支持 gRPC 和 HTTP 服务
- 📊 **可观测性** - 内置 OpenTelemetry 链路追踪、Prometheus 指标
- 🛡️ **优雅关闭** - 自动处理服务注销和连接关闭

## 安装

```bash
go get microg
```

## 快速开始

### 1. 创建 gRPC 服务

```go
package main

import (
    "context"
    
    "microg/app"
    "microg/registry/consul"
    "microg/server/rpcserver"
    
    "github.com/hashicorp/consul/api"
)

func main() {
    // 创建 Consul 客户端
    consulClient, _ := api.NewClient(api.DefaultConfig())
    
    // 创建服务注册器
    registrar := consul.New(consulClient, consul.WithHealthCheck(true))
    
    // 创建 gRPC 服务器
    rpcServer := rpcserver.NewServer(
        rpcserver.WithAddress(":9000"),
    )
    
    // 注册你的 gRPC 服务
    // pb.RegisterYourServiceServer(rpcServer.Server, &yourServiceImpl{})
    
    // 创建应用
    application := app.New(
        app.WithName("your-service"),
        app.WithRPCServer(rpcServer),
        app.WithRegistrar(registrar),
    )
    
    // 启动服务
    if err := application.Run(); err != nil {
        panic(err)
    }
}
```

### 2. 创建 HTTP 服务

```go
package main

import (
    "microg/app"
    "microg/server/restserver"
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
    
    // 创建应用
    application := app.New(
        app.WithName("your-api"),
        app.WithRestServer(restServer),
    )
    
    // 启动服务
    application.Run()
}
```

### 3. 同时运行 gRPC 和 HTTP

```go
package main

import (
    "microg/app"
    "microg/registry/consul"
    "microg/server/restserver"
    "microg/server/rpcserver"
)

func main() {
    // 创建注册器
    registrar := consul.New(consulClient)
    
    // 创建 gRPC 服务器
    rpcServer := rpcserver.NewServer(
        rpcserver.WithAddress(":9000"),
        rpcserver.WithMetrics(true),  // 启用指标
    )
    
    // 创建 HTTP 服务器
    restServer := restserver.NewServer(
        restserver.WithPort(8080),
    )
    
    // 创建应用（同时运行两个服务）
    application := app.New(
        app.WithName("your-service"),
        app.WithRPCServer(rpcServer),
        app.WithRestServer(restServer),
        app.WithRegistrar(registrar),
    )
    
    application.Run()
}
```

## 核心模块

### 应用生命周期 (app)

```go
// 创建应用时配置
app.New(
    app.WithName("service-name"),           // 服务名称
    app.WithID("unique-id"),                // 服务ID（可选，自动生成UUID）
    app.WithRPCServer(rpcServer),           // gRPC服务器
    app.WithRestServer(restServer),         // HTTP服务器
    app.WithRegistrar(registrar),           // 服务注册器
    app.WithEndpoints([]*url.URL{...}),     // 自定义端点
)
```

### 服务注册 (registry)

```go
// Consul 注册器
registrar := consul.New(consulClient,
    consul.WithHealthCheck(true),                    // 启用健康检查
    consul.WithHeartbeat(true),                      // 启用心跳
    consul.WithHealthCheckInterval(10),              // 健康检查间隔(秒)
    consul.WithDeregisterCriticalServiceAfter(600),  // 服务注销超时
)
```

### gRPC 服务器 (rpcserver)

```go
rpcServer := rpcserver.NewServer(
    rpcserver.WithAddress(":9000"),                  // 监听地址
    rpcserver.WithTimeout(30 * time.Second),         // 请求超时
    rpcserver.WithMetrics(true),                     // 启用Prometheus指标
    rpcserver.WithUnaryInterceptor(yourInterceptor), // 自定义拦截器
)
```

### HTTP 服务器 (restserver)

```go
restServer := restserver.NewServer(
    restserver.WithPort(8080),                       // 端口
    restserver.WithMode("release"),                  // 运行模式
    restserver.WithHealthz(true),                    // 健康检查
    restserver.WithPprof(true),                      // 性能分析
    restserver.WithMetrics(true),                    // Prometheus指标
)
```

### 日志 (log)

```go
import "microg/log"

// 使用默认日志
log.Info("message")
log.Errorf("error: %v", err)

// 自定义日志实现
type MyLogger struct{}
func (l *MyLogger) Info(msg string, fields ...log.Field) { ... }

log.SetLogger(&MyLogger{})
```

### 错误处理 (errors)

```go
import "microg/errors"

// 创建带错误码的错误
err := errors.WithCode(404, "user not found")

// 创建普通错误
err := errors.New("something wrong")

// 包装错误
err := errors.Wrap(err, "failed to process")
```

## 项目结构

```
microg/
├── app/                    # 应用生命周期管理
├── api/                    # gRPC proto 定义
├── code/                   # 错误码定义
├── core/
│   ├── metric/             # Prometheus 指标
│   └── trace/              # OpenTelemetry 链路追踪
├── errors/                 # 错误处理
├── internal/
│   └── host/               # 网络地址提取
├── log/                    # 日志接口
├── registry/
│   └── consul/             # Consul 服务注册发现
└── server/
    ├── restserver/         # HTTP 服务器 (Gin)
    └── rpcserver/          # gRPC 服务器
```

## 示例

完整示例请查看 [examples](./examples) 目录：

- [basic](./examples/basic) - 基础 gRPC 服务
- [http](./examples/http) - HTTP 服务
- [full](./examples/full) - 完整微服务示例

## 依赖

- [google.golang.org/grpc](https://github.com/grpc/grpc-go) - gRPC 框架
- [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin) - HTTP 框架
- [github.com/hashicorp/consul/api](https://github.com/hashicorp/consul) - Consul 客户端
- [go.opentelemetry.io/otel](https://github.com/open-telemetry/opentelemetry-go) - 链路追踪
- [go.uber.org/zap](https://github.com/uber-go/zap) - 日志库

## License

MIT License