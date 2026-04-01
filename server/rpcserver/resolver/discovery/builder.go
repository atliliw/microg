// Package discovery 实现基于服务发现的 gRPC 解析器
//
// 使用方式:
//
//	conn, _ := rpcserver.DialInsecure(ctx,
//	    rpcserver.WithEndpoint("discovery:///user-srv"),
//	    rpcserver.WithDiscovery(consulRegistry),
//	)
//
// 工作原理:
//  1. 解析 endpoint 中的服务名 (discovery:///user-srv -> user-srv)
//  2. 通过 Discovery.Watch() 监听服务变化
//  3. 将服务实例转换为 gRPC resolver.Address
//  4. 调用 ClientConn.UpdateState() 更新连接地址
//  5. 后台持续监听服务变化并更新连接
package discovery

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc/resolver"

	"microg/registry"
)

// name 解析器名称，对应 endpoint 中的 scheme
const name = "discovery"

// Option 解析器构建选项
type Option func(o *builder)

// WithTimeout 设置创建 watcher 的超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(b *builder) {
		b.timeout = timeout
	}
}

// WithInsecure 设置是否使用不安全连接
func WithInsecure(insecure bool) Option {
	return func(b *builder) {
		b.insecure = insecure
	}
}

// builder 实现 resolver.Builder 接口
// 用于创建 discovery 解析器
type builder struct {
	discoverer registry.Discovery // 服务发现实例
	timeout    time.Duration      // 创建 watcher 超时时间
	insecure   bool               // 是否不安全连接
}

// NewBuilder 创建解析器构建器
// 参数:
//   - d: 服务发现实例 (如 Consul Registry)
//   - opts: 可选配置
func NewBuilder(d registry.Discovery, opts ...Option) resolver.Builder {
	b := &builder{
		discoverer: d,
		timeout:    time.Second * 10,
		insecure:   false,
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

// Build 创建解析器实例
// 这是 resolver.Builder 接口的核心方法
//
// 参数:
//   - target: 解析目标，包含服务名 (如 discovery:///user-srv)
//   - cc: gRPC 客户端连接，用于更新地址状态
//   - opts: 构建选项
//
// 返回:
//   - resolver.Resolver: 解析器实例
//   - error: 错误信息
func (b *builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var (
		err error
		w   registry.Watcher
	)

	// 创建带超时的上下文
	done := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())

	// 异步创建 watcher，避免阻塞
	go func() {
		// 从 target 中提取服务名: discovery:///user-srv -> user-srv
		w, err = b.discoverer.Watch(ctx, strings.TrimPrefix(target.URL.Path, "/"))
		close(done)
	}()

	// 等待 watcher 创建完成或超时
	select {
	case <-done:
	case <-time.After(b.timeout):
		err = errors.New("discovery create watcher overtime")
	}

	if err != nil {
		cancel()
		return nil, err
	}

	// 创建解析器
	r := &discoveryResolver{
		w:        w,
		cc:       cc,
		ctx:      ctx,
		cancel:   cancel,
		insecure: b.insecure,
	}

	// 启动后台监听 goroutine
	go r.watch()

	return r, nil
}

// Scheme 返回解析器协议名
// gRPC 根据此名称匹配对应的 resolver.Builder
func (*builder) Scheme() string {
	return name
}
