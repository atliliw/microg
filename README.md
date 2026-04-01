# Microg

[![GoDoc](https://godoc.org/github.com/atliliw/microg?status.svg)](https://godoc.org/github.com/atliliw/microg)
[![Go Report Card](https://goreportcard.com/badge/github.com/atliliw/microg)](https://goreportcard.com/report/github.com/atliliw/microg)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8.svg)](https://golang.org)

一个简单、轻量的 Go 微服务框架，提供服务注册发现、gRPC/HTTP 服务、链路追踪、结构化日志等功能。

## 特性

- 🚀 简单易用 - API 设计简洁，快速上手
- 🔧 服务注册发现 - 内置 Consul 支持
- 📡 gRPC & HTTP - 同时支持两种服务
- 📊 可观测性 - OpenTelemetry 链路追踪、Prometheus 指标
- 📝 结构化日志 - 基于 zap，自动关联 traceID
- ⚠️ 错误处理 - 支持错误码和 HTTP 状态码映射
- 🛡️ 优雅关闭 - 自动处理服务注销

## 安装

```bash
go get github.com/atliliw/microg
```

## 快速示例

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
    // 创建 Consul 注册器
    consulClient, _ := api.NewClient(api.DefaultConfig())
    registrar := consul.New(consulClient, consul.WithHealthCheck(true))
    
    // 创建 gRPC 服务器
    rpcServer := rpcserver.NewServer(rpcserver.WithAddress(":9000"))
    
    // 注册服务
    // pb.RegisterYourServiceServer(rpcServer.Server, &yourServiceImpl{})
    
    // 启动应用
    app.New(
        app.WithName("your-service"),
        app.WithRPCServer(rpcServer),
        app.WithRegistrar(registrar),
    ).Run()
}
```

## 文档

📖 [详细文档](doc/README.md)

## 项目结构

```
microg/
├── app/          # 应用生命周期管理
├── core/         # 核心功能（metric, trace）
├── pkg/          # 公共包（errors, log, storage）
├── registry/     # 服务注册发现（consul）
├── server/       # 服务器（rpcserver, restserver）
├── config/       # 配置管理
├── doc/          # 详细文档
└── examples/     # 示例项目
```

## 示例

完整示例见 [examples](./examples) 目录：

- `examples/user/srv` - gRPC 服务端
- `examples/user/client` - HTTP 客户端（API 网关）

## 贡献

欢迎提交 Issue 和 Pull Request。

---

⭐ 如果这个项目对你有帮助，请给一个 Star！