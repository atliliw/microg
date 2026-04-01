//go:build wireinject
// +build wireinject

package main

import (
	"github.com/atliliw/microg/examples/user/client/config"
	"github.com/atliliw/microg/examples/user/client/controller"
	"github.com/atliliw/microg/examples/user/client/data"
	"github.com/atliliw/microg/examples/user/client/server"
	"github.com/atliliw/microg/examples/user/client/service"

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
