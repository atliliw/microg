package service

import (
	"github.com/google/wire"
)

// ProviderSet Wire 提供者集合
var ProviderSet = wire.NewSet(NewUserService)