package controller

import (
	"github.com/google/wire"
)

// ProviderSet Wire 提供者集合
// NewUserServer 在 user.go 中已定义
var ProviderSet = wire.NewSet(NewUserServer)