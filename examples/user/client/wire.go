//go:build wireinject
// +build wireinject

package main

import (
	"microg/examples/user/client/config"
	"microg/examples/user/client/data"
	"microg/examples/user/client/server"
	"microg/examples/user/client/service"
	"microg/examples/user/client/controller"

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