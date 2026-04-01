package data

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "microg/examples/user/api/v1"
	"microg/examples/user/client/config"
	"microg/registry/consul"
	"microg/server/rpcserver"

	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
)

// ProviderSet Wire 提供者集合
var ProviderSet = wire.NewSet(
	NewDiscovery,
	NewUserClient,
	NewUserRepo,
)

// NewDiscovery 创建 Consul 服务发现实例
func NewDiscovery(cfg *config.Config) *consul.Registry {
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
	return consul.New(client, consul.WithHealthCheck(true))
}

// NewUserClient 创建 gRPC 客户端
func NewUserClient(cfg *config.Config, r *consul.Registry) pb.UserClient {
	serviceName := fmt.Sprintf("discovery:///%s", cfg.UserSrv.Name)
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithEndpoint(serviceName),
		rpcserver.WithDiscovery(r),
		rpcserver.WithClientTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("连接 gRPC 服务失败: %v", err)
	}
	return pb.NewUserClient(conn)
}