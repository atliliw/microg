// Package consul 实现基于 Consul 的服务注册与发现
//
// 功能特性:
//   - 服务注册: 将服务实例注册到 Consul
//   - 服务发现: 通过服务名获取服务实例列表
//   - 健康检查: TCP 健康检查 + TTL 心跳
//   - 服务监听: 实时监听服务变化
//
// 使用示例:
//
//	// 创建 Consul 客户端
//	client, _ := api.NewClient(api.DefaultConfig())
//
//	// 创建注册器
//	r := consul.New(client,
//	    consul.WithHealthCheck(true),
//	    consul.WithHeartbeat(true),
//	)
//
//	// 注册服务
//	r.Register(ctx, &registry.ServiceInstance{
//	    ID:        "user-srv-001",
//	    Name:      "user-srv",
//	    Endpoints: []string{"grpc://192.168.1.10:9001"},
//	})
//
//	// 发现服务
//	instances, _ := r.GetService(ctx, "user-srv")
//
//	// 监听服务变化
//	watcher, _ := r.Watch(ctx, "user-srv")
//	for {
//	    instances, _ := watcher.Next()
//	    // 处理服务变化
//	}
package consul

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/atliliw/microg/registry"

	"github.com/hashicorp/consul/api"
)

// 编译时接口检查
var (
	_ registry.Registrar = &Registry{}
	_ registry.Discovery = &Registry{}
)

// Option Consul 注册器配置选项
type Option func(*Registry)

// WithHealthCheck 启用或禁用健康检查
// 启用后，Consul 会定期检查服务是否存活
func WithHealthCheck(enable bool) Option {
	return func(o *Registry) {
		o.enableHealthCheck = enable
	}
}

// WithHeartbeat 启用或禁用心跳
// 启用后，服务会定期发送心跳到 Consul
func WithHeartbeat(enable bool) Option {
	return func(o *Registry) {
		if o.cli != nil {
			o.cli.heartbeat = enable
		}
	}
}

// WithServiceResolver 设置自定义服务解析器
// 用于自定义从 Consul ServiceEntry 到 ServiceInstance 的转换逻辑
func WithServiceResolver(fn ServiceResolver) Option {
	return func(o *Registry) {
		if o.cli != nil {
			o.cli.resolver = fn
		}
	}
}

// WithHealthCheckInterval 设置健康检查间隔（秒）
// 默认 10 秒
func WithHealthCheckInterval(interval int) Option {
	return func(o *Registry) {
		if o.cli != nil {
			o.cli.healthcheckInterval = interval
		}
	}
}

// WithDeregisterCriticalServiceAfter 设置服务失效后自动注销时间（秒）
// 默认 600 秒（10 分钟）
func WithDeregisterCriticalServiceAfter(interval int) Option {
	return func(o *Registry) {
		if o.cli != nil {
			o.cli.deregisterCriticalServiceAfter = interval
		}
	}
}

// WithServiceCheck 设置自定义健康检查
func WithServiceCheck(checks ...*api.AgentServiceCheck) Option {
	return func(o *Registry) {
		if o.cli != nil {
			o.cli.serviceChecks = checks
		}
	}
}

// Config Consul 注册器配置
type Config struct {
	*api.Config
}

// Registry Consul 服务注册器
// 实现 registry.Registrar 和 registry.Discovery 接口
type Registry struct {
	cli               *Client                // Consul 客户端
	enableHealthCheck bool                   // 是否启用健康检查
	registry          map[string]*serviceSet // 服务名 -> 服务实例集合
	lock              sync.RWMutex           // 保护 registry 的并发访问
}

// New 创建 Consul 注册器
// 参数:
//   - apiClient: Consul API 客户端
//   - opts: 可选配置
func New(apiClient *api.Client, opts ...Option) *Registry {
	r := &Registry{
		cli:               NewClient(apiClient),
		registry:          make(map[string]*serviceSet),
		enableHealthCheck: true,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Register 注册服务到 Consul
// 会自动配置健康检查和心跳
func (r *Registry) Register(ctx context.Context, svc *registry.ServiceInstance) error {
	return r.cli.Register(ctx, svc, r.enableHealthCheck)
}

// Deregister 从 Consul 注销服务
func (r *Registry) Deregister(ctx context.Context, svc *registry.ServiceInstance) error {
	return r.cli.Deregister(ctx, svc.ID)
}

// GetService 通过服务名获取服务实例列表
// 优先从本地缓存获取，如果缓存为空则查询 Consul
func (r *Registry) GetService(ctx context.Context, name string) ([]*registry.ServiceInstance, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	set := r.registry[name]

	// 从远程 Consul 查询
	getRemote := func() []*registry.ServiceInstance {
		services, _, err := r.cli.Service(ctx, name, 0, true)
		if err == nil && len(services) > 0 {
			return services
		}
		return nil
	}

	// 本地缓存不存在，直接查询远程
	if set == nil {
		if s := getRemote(); len(s) > 0 {
			return s, nil
		}
		return nil, fmt.Errorf("service %s not resolved in registry", name)
	}

	// 从缓存获取
	ss, _ := set.services.Load().([]*registry.ServiceInstance)
	if ss == nil {
		if s := getRemote(); len(s) > 0 {
			return s, nil
		}
		return nil, fmt.Errorf("service %s not found in registry", name)
	}
	return ss, nil
}

// ListServices 列出所有服务
func (r *Registry) ListServices() (allServices map[string][]*registry.ServiceInstance, err error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	allServices = make(map[string][]*registry.ServiceInstance)
	for name, set := range r.registry {
		var services []*registry.ServiceInstance
		ss, _ := set.services.Load().([]*registry.ServiceInstance)
		if ss == nil {
			continue
		}
		services = append(services, ss...)
		allServices[name] = services
	}
	return
}

// Watch 创建服务监听器
// 返回的 Watcher 会实时推送服务变化
//
// 内部机制:
//  1. 创建 serviceSet 存储服务实例和监听器
//  2. 启动后台 goroutine 定期轮询 Consul
//  3. 检测到变化时，通过 channel 通知所有 watcher
func (r *Registry) Watch(ctx context.Context, name string) (registry.Watcher, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	// 获取或创建 serviceSet
	set, ok := r.registry[name]
	if !ok {
		set = &serviceSet{
			watcher:     make(map[*watcher]struct{}),
			services:    &atomic.Value{},
			serviceName: name,
		}
		r.registry[name] = set
	}

	// 创建 watcher
	w := &watcher{
		event: make(chan struct{}, 1),
	}
	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.set = set

	// 注册 watcher
	set.lock.Lock()
	set.watcher[w] = struct{}{}
	set.lock.Unlock()

	// 如果已有服务实例，立即通知
	ss, _ := set.services.Load().([]*registry.ServiceInstance)
	if len(ss) > 0 {
		w.event <- struct{}{}
	}

	// 首次 Watch 时启动后台轮询
	if !ok {
		err := r.resolve(set)
		if err != nil {
			return nil, err
		}
	}
	return w, nil
}

// resolve 解析服务并启动后台轮询
// 使用 Consul 长轮询机制，高效检测服务变化
func (r *Registry) resolve(ss *serviceSet) error {
	// 首次查询
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	services, idx, err := r.cli.Service(ctx, ss.serviceName, 0, true)
	cancel()
	if err != nil {
		return err
	} else if len(services) > 0 {
		ss.broadcast(services)
	}

	// 启动后台轮询 goroutine
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			<-ticker.C

			// 使用 WaitIndex 实现长轮询
			// Consul 会在服务变化时立即返回，否则等待 120 秒
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
			tmpService, tmpIdx, err := r.cli.Service(ctx, ss.serviceName, idx, true)
			cancel()

			if err != nil {
				time.Sleep(time.Second)
				continue
			}

			// 检测到变化
			if len(tmpService) != 0 && tmpIdx != idx {
				services = tmpService
				ss.broadcast(services) // 通知所有 watcher
			}
			idx = tmpIdx
		}
	}()

	return nil
}
