package server

import (
	"fmt"
	"log"

	"github.com/atliliw/microg/app"
	"github.com/atliliw/microg/examples/user/srv/config"
	pb "github.com/atliliw/microg/examples/user/api/v1"
	"github.com/atliliw/microg/registry/consul"
	"github.com/atliliw/microg/server/rpcserver"

	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
)

// App 应用实例
type App struct {
	Cfg         *config.Config
	Application *app.App
}

// ProviderSet Wire 提供者集合
var ProviderSet = wire.NewSet(
	NewRPCServer,
	NewRegistrar,
	NewApp,
	NewAppWrapper,
)

// NewRPCServer 创建 gRPC 服务器
func NewRPCServer(cfg *config.Config, userServer pb.UserServer) *rpcserver.Server {
	server := rpcserver.NewServer(
		rpcserver.WithAddress(cfg.GRPC.Address),
		rpcserver.WithMetrics(cfg.GRPC.Metrics),
		rpcserver.WithTimeout(cfg.GRPC.Timeout),
	)
	pb.RegisterUserServer(server.Server, userServer)
	return server
}

// NewRegistrar 创建 Consul 服务注册器
func NewRegistrar(cfg *config.Config) *consul.Registry {
	if !cfg.Consul.Enabled {
		return nil
	}
	consulCfg := api.DefaultConfig()
	consulCfg.Address = cfg.Consul.Address
	client, err := api.NewClient(consulCfg)
	if err != nil {
		log.Printf("创建 Consul 客户端失败: %v", err)
		return nil
	}
	return consul.New(client,
		consul.WithHealthCheck(cfg.Consul.HealthCheck),
		consul.WithHeartbeat(cfg.Consul.Heartbeat),
		consul.WithHealthCheckInterval(cfg.Consul.HealthCheckInterval),
	)
}

// NewApp 创建应用实例
func NewApp(cfg *config.Config, rpcServer *rpcserver.Server, registrar *consul.Registry) *app.App {
	return app.New(
		app.WithName(cfg.Service.Name),
		app.WithRPCServer(rpcServer),
		app.WithRegistrar(registrar),
	)
}

// NewAppWrapper 包装应用实例
func NewAppWrapper(cfg *config.Config, application *app.App) (*App, func(), error) {
	cleanup := func() { fmt.Println("清理资源...") }
	return &App{Cfg: cfg, Application: application}, cleanup, nil
}