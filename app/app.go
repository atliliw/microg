// Package app 提供微服务应用生命周期管理
// 负责服务启动、注册、停止等核心流程
package app

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	"microg/pkg/log"
	"microg/registry"
	gs "microg/server"
)

// App 是微服务应用的核心结构体
// 管理服务的完整生命周期：启动 → 注册 → 运行 → 注销 → 停止
type App struct {
	opts options // 配置项

	lk       sync.Mutex                // 保护 instance 的并发访问
	instance *registry.ServiceInstance // 服务实例信息，用于注册到 Consul

	cancel func() // 取消函数，用于停止所有 goroutine
}

// New 创建微服务应用实例
// 使用 Options Pattern 进行配置
func New(opts ...Option) *App {
	// 设置默认值
	o := options{
		sigs:             []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT}, // 默认监听的退出信号
		registrarTimeout: 10 * time.Second,                                              // 注册超时时间
		stopTimeout:      10 * time.Second,                                              // 停止超时时间
	}

	// 生成唯一服务 ID
	if id, err := uuid.NewUUID(); err == nil {
		o.id = id.String()
	}

	// 应用所有配置项（Options Pattern）
	for _, opt := range opts {
		opt(&o)
	}

	return &App{
		opts: o,
	}
}

// Run 启动服务
// 完整流程：构建实例 → 启动 Server → 注册到 Consul → 监听退出信号
func (a *App) Run() error {
	// 1. 构建服务实例信息（ID、名称、端点）
	instance, err := a.buildInstance()
	if err != nil {
		return err
	}

	// 保存实例信息（并发安全）
	a.lk.Lock()
	a.instance = instance
	a.lk.Unlock()

	// 2. 收集所有需要启动的 Server
	var servers []gs.Server
	if a.opts.restServer != nil {
		servers = append(servers, a.opts.restServer)
	}
	if a.opts.rpcServer != nil {
		servers = append(servers, a.opts.rpcServer)
	}

	// 3. 使用 errgroup 并发启动所有 Server
	//    errgroup 特性：任意一个 goroutine 返回错误，其他都会被取消
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	eg, ctx := errgroup.WithContext(ctx)
	wg := sync.WaitGroup{}

	for _, srv := range servers {
		srv := srv // 捕获循环变量

		// goroutine 1: 监听停止信号，执行优雅关闭
		eg.Go(func() error {
			<-ctx.Done() // 等待取消信号
			// 带超时的停止，防止无限等待
			sctx, cancel := context.WithTimeout(context.Background(), a.opts.stopTimeout)
			defer cancel()
			return srv.Stop(sctx)
		})

		// goroutine 2: 启动 Server
		wg.Add(1)
		eg.Go(func() error {
			wg.Done()
			log.Info("start server")
			return srv.Start(ctx)
		})
	}

	// 等待所有 Server 启动
	wg.Wait()

	// 4. 服务启动成功后，注册到 Consul
	if a.opts.registrar != nil {
		rctx, rcancel := context.WithTimeout(context.Background(), a.opts.registrarTimeout)
		defer rcancel()
		err := a.opts.registrar.Register(rctx, instance)
		if err != nil {
			log.Errorf("register service error: %s", err)
			return err
		}
	}

	// 5. 监听系统退出信号（SIGTERM、SIGINT、SIGQUIT）
	c := make(chan os.Signal, 1)
	signal.Notify(c, a.opts.sigs...)
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c:
			// 收到退出信号，执行停止流程
			return a.Stop()
		}
	})

	// 等待所有 goroutine 结束
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

// Stop 停止服务
// 流程：从 Consul 注销 → 取消所有 goroutine
func (a *App) Stop() error {
	a.lk.Lock()
	instance := a.instance
	a.lk.Unlock()

	log.Info("start deregister service")

	// 1. 从 Consul 注销服务
	if a.opts.registrar != nil && instance != nil {
		rctx, rcancel := context.WithTimeout(context.Background(), a.opts.stopTimeout)
		defer rcancel()
		if err := a.opts.registrar.Deregister(rctx, instance); err != nil {
			log.Errorf("deregister service error: %s", err)
			return err
		}
	}

	// 2. 取消所有 goroutine（触发 Server 停止）
	if a.cancel != nil {
		a.cancel()
	}

	return nil
}

// buildInstance 构建服务实例信息
// 用于注册到服务发现中心（Consul）
func (a *App) buildInstance() (*registry.ServiceInstance, error) {
	endpoints := make([]string, 0)

	// 添加用户自定义的端点
	for _, e := range a.opts.endpoints {
		endpoints = append(endpoints, e.String())
	}

	// 从 RPC Server 自动获取端点
	if a.opts.rpcServer != nil {
		if a.opts.rpcServer.Endpoint() != nil {
			endpoints = append(endpoints, a.opts.rpcServer.Endpoint().String())
		} else {
			// 如果没有 Endpoint，手动构建
			u := &url.URL{
				Scheme: "grpc",
				Host:   a.opts.rpcServer.Address(),
			}
			endpoints = append(endpoints, u.String())
		}
	}

	return &registry.ServiceInstance{
		ID:        a.opts.id,   // 服务唯一 ID
		Name:      a.opts.name, // 服务名称
		Endpoints: endpoints,   // 端点列表，如 ["grpc://192.168.1.1:9000"]
	}, nil
}
