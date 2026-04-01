package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Service ServiceConfig `yaml:"service"`
	HTTP    HTTPConfig    `yaml:"http"`
	Consul  ConsulConfig  `yaml:"consul"`
	UserSrv UserSrvConfig `yaml:"user_srv"`
	Log     LogConfig     `yaml:"log"`
}

type ServiceConfig struct {
	Name string `yaml:"name"`
}

type HTTPConfig struct {
	Port    int    `yaml:"port"`
	Mode    string `yaml:"mode"`
	Healthz bool   `yaml:"healthz"`
	Metrics bool   `yaml:"metrics"`
}

type ConsulConfig struct {
	Address string `yaml:"address"`
	Enabled bool   `yaml:"enabled"`
}

type UserSrvConfig struct {
	Name string `yaml:"name"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	setDefaults(&cfg)
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Service.Name == "" {
		cfg.Service.Name = "user-client"
	}
	if cfg.HTTP.Port == 0 {
		cfg.HTTP.Port = 8081
	}
	if cfg.HTTP.Mode == "" {
		cfg.HTTP.Mode = "debug"
	}
	if cfg.UserSrv.Name == "" {
		cfg.UserSrv.Name = "user-srv"
	}
}
