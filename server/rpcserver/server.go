// Package rpcserver 提供 gRPC 服务器封装
// 集成了健康检查、链路追踪、指标监控等功能
package rpcserver

import (
	"context"
	"net"
	"net/url"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	apimd "microg/api/metadata"
	"microg/pkg/host"
	"microg/pkg/log"
	srvintc "microg/server/rpcserver/serverinterceptors"
)

// ServerOption 是 gRPC 服务器的配置函数
type ServerOption func(o *Server)

// Server 是 gRPC 服务器的封装
// 内嵌了 grpc.Server，并添加了健康检查、元数据服务等功能
type Server struct {
	*grpc.Server

	address    string                         // 监听地址，如 ":9000" 或 "0.0.0.0:9000"
	unaryInts  []grpc.UnaryServerInterceptor  // 用户自定义的一元拦截器
	streamInts []grpc.StreamServerInterceptor // 用户自定义的流拦截器
	grpcOpts   []grpc.ServerOption            // gRPC 原生配置选项
	lis        net.Listener                   // 网络监听器
	timeout    time.Duration                  // 请求超时时间

	health   *health.Server // gRPC 健康检查服务
	metadata *apimd.Server  // 元数据服务，用于服务发现
	endpoint *url.URL       // 服务端点，如 grpc://192.168.1.1:9000

	enableMetrics bool // 是否启用 Prometheus 指标
}

// Endpoint 返回服务端点
// 用于注册到服务发现中心
func (s *Server) Endpoint() *url.URL {
	return s.endpoint
}

// Address 返回监听地址
func (s *Server) Address() string {
	return s.address
}

// NewServer 创建 gRPC 服务器实例
// 自动添加默认拦截器：crash 恢复、链路追踪、可选的指标监控和超时控制
func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		address: ":0", // 默认随机端口
		health:  health.NewServer(),
	}

	// 应用用户配置（Options Pattern）
	for _, o := range opts {
		o(srv)
	}

	// 构建拦截器链
	// 1. 默认拦截器：crash 恢复 + 链路追踪
	unaryInts := []grpc.UnaryServerInterceptor{
		srvintc.UnaryCrashInterceptor,
		otelgrpc.UnaryServerInterceptor(),
	}

	// 2. 可选：Prometheus 指标
	if srv.enableMetrics {
		unaryInts = append(unaryInts, srvintc.UnaryPrometheusInterceptor)
	}

	// 3. 可选：请求超时
	if srv.timeout > 0 {
		unaryInts = append(unaryInts, srvintc.UnaryTimeoutInterceptor(srv.timeout))
	}

	// 4. 用户自定义拦截器（追加到末尾）
	if len(srv.unaryInts) > 0 {
		unaryInts = append(unaryInts, srv.unaryInts...)
	}

	// 将拦截器链转换为 grpc.ServerOption
	grpcOpts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(unaryInts...)}

	// 添加用户传入的其他 grpc.ServerOption
	if len(srv.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, srv.grpcOpts...)
	}

	// 创建 gRPC Server
	srv.Server = grpc.NewServer(grpcOpts...)

	// 注册元数据服务（用于服务发现）
	srv.metadata = apimd.NewServer(srv.Server)

	// 解析地址并提取端点
	err := srv.listenAndEndpoint()
	if err != nil {
		panic(err)
	}

	// 注册健康检查服务
	grpc_health_v1.RegisterHealthServer(srv.Server, srv.health)
	// 注册元数据服务
	apimd.RegisterMetadataServer(srv.Server, srv.metadata)
	// 注册 gRPC 反射服务（支持 grpcurl 等工具）
	reflection.Register(srv.Server)

	return srv
}

// WithAddress 设置监听地址
// 例如: WithAddress(":9000") 或 WithAddress("0.0.0.0:9000")
func WithAddress(address string) ServerOption {
	return func(s *Server) {
		s.address = address
	}
}

// WithMetrics 启用 Prometheus 指标收集
func WithMetrics(metric bool) ServerOption {
	return func(s *Server) {
		s.enableMetrics = metric
	}
}

// WithTimeout 设置请求超时时间
func WithTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// WithLis 设置自定义的网络监听器
// 用于高级场景，如控制端口绑定
func WithLis(lis net.Listener) ServerOption {
	return func(s *Server) {
		s.lis = lis
	}
}

// WithUnaryInterceptor 添加一元拦截器
func WithUnaryInterceptor(in ...grpc.UnaryServerInterceptor) ServerOption {
	return func(s *Server) {
		s.unaryInts = in
	}
}

// WithStreamInterceptor 添加流拦截器
func WithStreamInterceptor(in ...grpc.StreamServerInterceptor) ServerOption {
	return func(s *Server) {
		s.streamInts = in
	}
}

// WithOptions 添加 gRPC 原生配置选项
func WithOptions(opts ...grpc.ServerOption) ServerOption {
	return func(s *Server) {
		s.grpcOpts = opts
	}
}

// listenAndEndpoint 创建监听器并提取端点地址
// 处理地址为空或 0.0.0.0 的情况，自动获取实际 IP
func (s *Server) listenAndEndpoint() error {
	// 如果没有传入监听器，创建新的
	if s.lis == nil {
		lis, err := net.Listen("tcp", s.address)
		if err != nil {
			return err
		}
		s.lis = lis
	}

	// 提取实际地址（处理 0.0.0.0 的情况）
	addr, err := host.Extract(s.address, s.lis)
	if err != nil {
		_ = s.lis.Close()
		return err
	}

	// 构建 endpoint URL
	s.endpoint = &url.URL{Scheme: "grpc", Host: addr}
	return nil
}

// Start 启动 gRPC 服务器
func (s *Server) Start(ctx context.Context) error {
	log.Infof("[grpc] server listening on: %s", s.lis.Addr().String())
	// 设置健康状态为 serving，开始接收请求
	s.health.Resume()
	return s.Serve(s.lis)
}

// Stop 停止 gRPC 服务器
// 优雅关闭：先设置为 not_serving，再停止接收新请求
func (s *Server) Stop(ctx context.Context) error {
	// 设置健康状态为 not_serving，告诉负载均衡器不再路由请求
	s.health.Shutdown()
	// 优雅关闭：等待现有请求完成
	s.GracefulStop()
	log.Infof("[grpc] server stopped")
	return nil
}
