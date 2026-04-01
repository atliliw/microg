package config

import (
	"fmt"
	"os"

	"github.com/google/wire"
)

// ProviderSet Wire 提供者集合
var ProviderSet = wire.NewSet(LoadConfig)

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	searchPaths := []string{
		"examples/user/srv/config/config.yaml",
		"config.yaml",
	}

	if p := os.Getenv("CONFIG_PATH"); p != "" {
		fmt.Printf("使用环境变量指定的配置: %s\n", p)
		return Load(p)
	}

	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			fmt.Printf("加载配置文件: %s\n", p)
			return Load(p)
		}
	}

	return nil, fmt.Errorf("配置文件未找到")
}
