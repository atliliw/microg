package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Service ServiceConfig `yaml:"service"`
	GRPC    GRPCConfig    `yaml:"grpc"`
	Consul  ConsulConfig  `yaml:"consul"`
	Tracing TracingConfig `yaml:"tracing"`
	Log     LogConfig     `yaml:"log"`
	MySQL   MySQLConfig   `yaml:"mysql"`
}

type ServiceConfig struct {
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

type GRPCConfig struct {
	Address string        `yaml:"address"`
	Timeout time.Duration `yaml:"timeout"`
	Metrics bool          `yaml:"metrics"`
}

type ConsulConfig struct {
	Address                 string `yaml:"address"`
	Enabled                 bool   `yaml:"enabled"`
	HealthCheck             bool   `yaml:"health_check"`
	Heartbeat               bool   `yaml:"heartbeat"`
	HealthCheckInterval     int    `yaml:"health_check_interval"`
	DeregisterCriticalAfter int    `yaml:"deregister_critical_after"`
}

type TracingConfig struct {
	Enabled  bool    `yaml:"enabled"`
	Endpoint string  `yaml:"endpoint"`
	Sampler  float64 `yaml:"sampler"`
	Batcher  string  `yaml:"batcher"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type MySQLConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	MaxOpenConns int    `yaml:"max_open_conns"`
}

func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

func (c *MySQLConfig) NoDatabaseDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port)
}

func Load(path string) (*Config, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := parseYAML(data, &cfg); err != nil {
		return nil, err
	}

	setDefaults(&cfg)
	return &cfg, nil
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func parseYAML(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

func setDefaults(cfg *Config) {
	if cfg.Service.Name == "" {
		cfg.Service.Name = "microg-service"
	}
	if cfg.GRPC.Address == "" {
		cfg.GRPC.Address = ":9000"
	}
	if cfg.GRPC.Timeout == 0 {
		cfg.GRPC.Timeout = 30 * time.Second
	}
	if cfg.Consul.Address == "" {
		cfg.Consul.Address = "127.0.0.1:8500"
	}
	if cfg.Consul.HealthCheckInterval == 0 {
		cfg.Consul.HealthCheckInterval = 10
	}
}
