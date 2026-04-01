package data

import (
	"fmt"

	"microg/examples/user/srv/config"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewUserStore)

func NewUserStore(cfg *config.Config) (UserStore, error) {
	fmt.Printf("连接 MySQL: %s@%s:%d/%s\n", cfg.MySQL.Username, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)
	return NewMySQLUserStore(cfg.MySQL.DSN())
}