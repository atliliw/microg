package main

import (
	"github.com/atliliw/microg/pkg/log"

	_ "github.com/atliliw/microg/examples/user/code"
)

func main() {
	app, cleanup, err := InitApp()
	if err != nil {
		log.Fatalf("初始化应用失败: %v", err)
	}
	defer cleanup()

	log.Info("启动用户 gRPC 服务",
		log.String("service", app.Cfg.Service.Name),
		log.String("grpc_address", app.Cfg.GRPC.Address),
		log.String("consul_address", app.Cfg.Consul.Address),
	)

	if err := app.Application.Run(); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
