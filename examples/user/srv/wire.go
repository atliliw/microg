//go:build wireinject
// +build wireinject

package main

import (
	"microg/examples/user/srv/config"
	"microg/examples/user/srv/data"
	"microg/examples/user/srv/server"
	"microg/examples/user/srv/service"
	"microg/examples/user/srv/controller"

	"github.com/google/wire"
)

// InitApp Wire 注入器入口
func InitApp() (*server.App, func(), error) {
	wire.Build(ProviderSet)
	return nil, nil, nil
}

// ProviderSet 依赖提供者集合
var ProviderSet = wire.NewSet(
	config.ProviderSet,
	data.ProviderSet,
	service.ProviderSet,
	controller.ProviderSet,
	server.ProviderSet,
)