package main

import (
	"github.com/atliliw/microg/pkg/log"
	"github.com/atliliw/microg/server/rpcserver"
	"github.com/atliliw/microg/server/rpcserver/selector"
	"github.com/atliliw/microg/server/rpcserver/selector/node/ewma"
	"github.com/atliliw/microg/server/rpcserver/selector/p2c"

	_ "github.com/atliliw/microg/examples/user/code"
)

// init 初始化自定义负载均衡器
// 使用 P2C 算法 + EWMA 动态加权节点
// P2C (Power of Two Choices): 从两个节点中选择负载较低的节点
// EWMA (Exponentially Weighted Moving Average): 根据延迟和成功率动态计算权重
func init() {
	// 注意顺序：必须先设置 GlobalSelector，再调用 InitBuilder
	// 因为 InitBuilder 会调用 GlobalSelector() 获取 builder

	// 1. 先设置全局负载均衡策略：P2C + EWMA
	selector.SetGlobalSelector(&selector.DefaultBuilder{
		Node:     &ewma.Builder{}, // EWMA 动态加权节点
		Balancer: &p2c.Builder{},  // P2C 负载均衡算法
	})

	// 2. 再注册自定义 selector balancer 到 gRPC
	rpcserver.InitBuilder()

	log.Info("已启用自定义负载均衡: P2C + EWMA")
}

func main() {
	app, cleanup, err := InitApp()
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	defer cleanup()

	log.Infof("启动 User Client %s...\n", app.Cfg.Service.Name)
	log.Infof("HTTP 监听: :%d\n", app.Cfg.HTTP.Port)
	log.Infof("服务发现: %s -> %s\n", app.Cfg.Consul.Address, app.Cfg.UserSrv.Name)

	if err := app.HTTPServer.Run(); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
