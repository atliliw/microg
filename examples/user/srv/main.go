package main

import (
	"fmt"
	"microg/pkg/log"

	_ "microg/examples/user/code"
)

func main() {
	app, cleanup, err := InitApp()
	if err != nil {
		log.Fatalf("初始化应用失败: %v", err)
	}
	defer cleanup()

	fmt.Printf("启动用户 gRPC 服务 %s...\n", app.Cfg.Service.Name)
	fmt.Printf("gRPC 监听: %s\n", app.Cfg.GRPC.Address)
	fmt.Printf("Consul 地址: %s\n", app.Cfg.Consul.Address)

	if err := app.Application.Run(); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
