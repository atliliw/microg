package app

import (
	"net/url"
	"os"
	"time"

	"microg/registry"
	"microg/server/restserver"
	"microg/server/rpcserver"
)

// Option 是配置函数，用于修改 options 结构体
// 采用 Options Pattern，支持链式配置
type Option func(o *options)

// options 是 App 的内部配置结构体
// 包含服务运行所需的所有配置项
type options struct {
	id        string     // 服务唯一 ID，自动生成 UUID
	endpoints []*url.URL // 服务端点列表，如 grpc://192.168.1.1:9000
	name      string     // 服务名称，用于注册到 Consul

	sigs []os.Signal // 监听的退出信号，默认 [SIGTERM, SIGQUIT, SIGINT]

	registrar        registry.Registrar // 服务注册器（如 Consul），允许用户自定义实现
	registrarTimeout time.Duration      // 注册超时时间，默认 10 秒

	stopTimeout time.Duration // 停止超时时间，默认 10 秒

	restServer *restserver.Server // HTTP 服务器
	rpcServer  *rpcserver.Server  // gRPC 服务器
}

// WithRegistrar 设置服务注册器
// 用于将服务注册到服务发现中心（如 Consul）
func WithRegistrar(registrar registry.Registrar) Option {
	return func(o *options) {
		o.registrar = registrar
	}
}

// WithEndpoints 设置服务端点
// 例如: WithEndpoints([]*url.URL{{Scheme: "grpc", Host: "192.168.1.1:9000"}})
func WithEndpoints(endpoints []*url.URL) Option {
	return func(o *options) {
		o.endpoints = endpoints
	}
}

// WithRPCServer 设置 gRPC 服务器
func WithRPCServer(server *rpcserver.Server) Option {
	return func(o *options) {
		o.rpcServer = server
	}
}

// WithRestServer 设置 HTTP 服务器
func WithRestServer(server *restserver.Server) Option {
	return func(o *options) {
		o.restServer = server
	}
}

// WithID 设置服务 ID
// 不设置则自动生成 UUID
func WithID(id string) Option {
	return func(o *options) {
		o.id = id
	}
}

// WithName 设置服务名称
// 用于注册到 Consul 的服务标识
func WithName(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

// WithSigs 设置监听的退出信号
// 默认监听 [SIGTERM, SIGQUIT, SIGINT]
func WithSigs(sigs []os.Signal) Option {
	return func(o *options) {
		o.sigs = sigs
	}
}
