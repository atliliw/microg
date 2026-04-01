package main

import (
	"github.com/atliliw/microg/pkg/log"

	_ "github.com/atliliw/microg/examples/user/code"
)

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
