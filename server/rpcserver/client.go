package rpcserver

import (
	"context"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	grpcinsecure "google.golang.org/grpc/credentials/insecure"

	"github.com/atliliw/microg/pkg/log"
	"github.com/atliliw/microg/registry"
	"github.com/atliliw/microg/server/rpcserver/clientinterceptors"
	"github.com/atliliw/microg/server/rpcserver/resolver/discovery"
)

// ClientOption gRPC 客户端配置选项函数
type ClientOption func(o *clientOptions)

// clientOptions gRPC 客户端配置结构
type clientOptions struct {
	endpoint      string                         // 服务端点地址，支持 discovery:///服务名 格式
	timeout       time.Duration                  // 请求超时时间
	discovery     registry.Discovery             // 服务发现实例
	unaryInts     []grpc.UnaryClientInterceptor  // 一元拦截器列表
	streamInts    []grpc.StreamClientInterceptor // 流拦截器列表
	rpcOpts       []grpc.DialOption              // 原生 gRPC DialOption 列表
	balancerName  string                         // 负载均衡策略，默认 round_robin
	logger        log.Logger                     // 日志实例
	enableTracing bool                           // 启用链路追踪
	enableMetrics bool                           // 启用 Prometheus 指标
}

// WithEnableTracing 启用或禁用链路追踪
func WithEnableTracing(enable bool) ClientOption {
	return func(o *clientOptions) {
		o.enableTracing = enable
	}
}

// WithEndpoint 设置服务端点地址
// 支持格式：
//   - "127.0.0.1:9000" - 直连地址
//   - "discovery:///user-srv" - 服务发现地址（通过 Consul 等注册中心解析）
func WithEndpoint(endpoint string) ClientOption {
	return func(o *clientOptions) {
		o.endpoint = endpoint
	}
}

// WithClientTimeout 设置请求超时时间
func WithClientTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

// WithDiscovery 设置服务发现实例
// 用于通过服务名发现服务地址，配合 discovery:/// 前缀使用
func WithDiscovery(d registry.Discovery) ClientOption {
	return func(o *clientOptions) {
		o.discovery = d
	}
}

// WithClientUnaryInterceptor 设置一元拦截器
// 拦截器按添加顺序依次执行
func WithClientUnaryInterceptor(in ...grpc.UnaryClientInterceptor) ClientOption {
	return func(o *clientOptions) {
		o.unaryInts = in
	}
}

// WithClientStreamInterceptor 设置流拦截器
func WithClientStreamInterceptor(in ...grpc.StreamClientInterceptor) ClientOption {
	return func(o *clientOptions) {
		o.streamInts = in
	}
}

// WithClientOptions 设置原生 gRPC DialOption
func WithClientOptions(opts ...grpc.DialOption) ClientOption {
	return func(o *clientOptions) {
		o.rpcOpts = opts
	}
}

// WithBalancerName 设置负载均衡策略
// 常用值:
//   - round_robin: 轮询（默认）
//   - pick_first: 选择第一个可用地址
func WithBalancerName(name string) ClientOption {
	return func(o *clientOptions) {
		o.balancerName = name
	}
}

// DialInsecure 创建不安全的 gRPC 连接（无 TLS 加密）
// 适用于内网环境或开发测试
func DialInsecure(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dial(ctx, true, opts...)
}

// Dial 创建安全的 gRPC 连接（需要 TLS）
// 适用于生产环境
func Dial(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dial(ctx, false, opts...)
}

// dial 创建 gRPC 连接的核心实现
//
// 完整流程:
//  1. 初始化默认配置
//  2. 应用用户配置选项
//  3. 组装拦截器链（超时 → 链路追踪 → 指标 → 用户拦截器）
//  4. 配置服务发现解析器（如果使用 discovery:/// 格式）
//  5. 设置传输凭证（安全/不安全）
//  6. 创建 gRPC 连接
//
// 参数:
//   - ctx: 上下文，用于控制连接创建超时
//   - insecure: true=不安全连接(无TLS), false=安全连接(需要TLS)
//   - opts: 客户端配置选项
func dial(ctx context.Context, insecure bool, opts ...ClientOption) (*grpc.ClientConn, error) {
	// ========================================
	// 第一步：初始化默认配置
	// ========================================
	options := clientOptions{
		timeout:       2000 * time.Millisecond, // 默认 2 秒超时
		balancerName:  "round_robin",           // 默认轮询负载均衡
		enableTracing: true,                    // 默认启用链路追踪
	}

	// ========================================
	// 第二步：应用用户配置选项
	// ========================================
	// 按顺序调用每个 ClientOption 函数，覆盖默认值
	for _, o := range opts {
		o(&options)
	}

	// ========================================
	// 第三步：组装一元拦截器链
	// ========================================
	// 拦截器执行顺序：TimeoutInterceptor → TracingInterceptor → MetricsInterceptor → UserInterceptors
	// 每个拦截器可以：修改请求、记录日志、处理错误等

	ints := []grpc.UnaryClientInterceptor{
		clientinterceptors.TimeoutInterceptor(options.timeout), // 超时控制
	}

	// 链路追踪拦截器（OpenTelemetry）
	if options.enableTracing {
		ints = append(ints, otelgrpc.UnaryClientInterceptor())
	}

	// Prometheus 指标拦截器
	if options.enableMetrics {
		ints = append(ints, clientinterceptors.PrometheusInterceptor())
	}

	streamInts := []grpc.StreamClientInterceptor{}

	// 追加用户自定义拦截器
	if len(options.unaryInts) > 0 {
		ints = append(ints, options.unaryInts...)
	}
	if len(options.streamInts) > 0 {
		streamInts = append(streamInts, options.streamInts...)
	}

	// ========================================
	// 第四步：配置 gRPC DialOption
	// ========================================
	grpcOpts := []grpc.DialOption{
		// 负载均衡策略配置
		// round_robin: 轮询所有可用地址
		// pick_first: 选择第一个可用地址
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "` + options.balancerName + `"}`),

		// 拦截器链
		grpc.WithChainUnaryInterceptor(ints...),
		grpc.WithChainStreamInterceptor(streamInts...),
	}

	// ========================================
	// 第五步：配置服务发现解析器
	// ========================================
	// 如果设置了 discovery，则启用服务发现功能
	// 支持 discovery:///服务名 格式的 endpoint
	if options.discovery != nil {
		grpcOpts = append(grpcOpts, grpc.WithResolvers(
			discovery.NewBuilder(
				options.discovery,
				discovery.WithInsecure(insecure),
			),
		))
	}

	// ========================================
	// 第六步：设置传输凭证
	// ========================================
	if insecure {
		// 不安全连接：无 TLS 加密
		// 警告：仅适用于内网或开发环境
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(grpcinsecure.NewCredentials()))
	}
	// 安全连接需要配置 TLS 证书（通过 WithClientOptions 传入）

	// 追加用户自定义 DialOption
	if len(options.rpcOpts) > 0 {
		grpcOpts = append(grpcOpts, options.rpcOpts...)
	}

	// ========================================
	// 第七步：创建 gRPC 连接
	// ========================================
	// grpc.DialContext 会：
	//  1. 解析 endpoint（直连或服务发现）
	//  2. 建立连接
	//  3. 启动后台连接管理
	return grpc.DialContext(ctx, options.endpoint, grpcOpts...)
}
