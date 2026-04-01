// Package config 提供 YAML 配置文件加载功能
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 完整的服务配置结构
type Config struct {
	Service ServiceConfig `yaml:"service"`
	GRPC    GRPCConfig    `yaml:"grpc"`
	HTTP    HTTPConfig    `yaml:"http"`
	Consul  ConsulConfig  `yaml:"consul"`
	Tracing TracingConfig `yaml:"tracing"`
	Log     LogConfig     `yaml:"log"`
	MySQL   MySQLConfig   `yaml:"mysql"`
}

// ServiceConfig 服务基础配置
type ServiceConfig struct {
	Name string `yaml:"name"` // 服务名称
	ID   string `yaml:"id"`   // 服务ID（空则自动生成）
}

// GRPCConfig gRPC 服务器配置
type GRPCConfig struct {
	Address string        `yaml:"address"` // 监听地址，如 ":9000"
	Timeout time.Duration `yaml:"timeout"` // 请求超时
	Metrics bool          `yaml:"metrics"` // 启用 Prometheus 指标
}

// HTTPConfig HTTP 服务器配置
type HTTPConfig struct {
	Port    int    `yaml:"port"`    // 监听端口
	Mode    string `yaml:"mode"`    // 运行模式: debug/release
	Healthz bool   `yaml:"healthz"` // 启用健康检查
	Pprof   bool   `yaml:"pprof"`   // 启用性能分析
	Metrics bool   `yaml:"metrics"` // 启用 Prometheus 指标
}

// ConsulConfig Consul 服务注册配置
type ConsulConfig struct {
	Address                 string `yaml:"address"`                   // Consul 地址
	Enabled                 bool   `yaml:"enabled"`                   // 是否启用
	HealthCheck             bool   `yaml:"health_check"`              // 健康检查
	Heartbeat               bool   `yaml:"heartbeat"`                 // 心跳
	HealthCheckInterval     int    `yaml:"health_check_interval"`     // 健康检查间隔(秒)
	DeregisterCriticalAfter int    `yaml:"deregister_critical_after"` // 失效后注销时间(秒)
}

// TracingConfig 链路追踪配置
type TracingConfig struct {
	Enabled  bool    `yaml:"enabled"`  // 是否启用
	Endpoint string  `yaml:"endpoint"` // Jaeger/Zipkin 地址
	Sampler  float64 `yaml:"sampler"`  // 采样率
	Batcher  string  `yaml:"batcher"`  // 采集器类型
}

// LogConfig 日志配置
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

// Load 从 YAML 文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// 设置默认值
	setDefaults(&cfg)

	return &cfg, nil
}

// LoadFromEnv 从环境变量指定的路径加载配置
// 环境变量 CONFIG_PATH 指定配置文件路径，默认为 "config/config.yaml"
func LoadFromEnv() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path != "" {
		return Load(path)
	}
	return findAndLoad()
}

func findAndLoad() (*Config, error) {
	searchPaths := []string{
		"config.yaml",
		"config/config.yaml",
	}
	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			return Load(p)
		}
	}
	return nil, fmt.Errorf("config file not found in: %v", searchPaths)
}

// setDefaults 设置默认值
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
	if cfg.HTTP.Port == 0 {
		cfg.HTTP.Port = 8080
	}
	if cfg.HTTP.Mode == "" {
		cfg.HTTP.Mode = "release"
	}
	if cfg.Consul.Address == "" {
		cfg.Consul.Address = "127.0.0.1:8500"
	}
	if cfg.Consul.HealthCheckInterval == 0 {
		cfg.Consul.HealthCheckInterval = 10
	}
	if cfg.Tracing.Endpoint == "" {
		cfg.Tracing.Endpoint = "http://127.0.0.1:14268/api/traces"
	}
	if cfg.Tracing.Sampler == 0 {
		cfg.Tracing.Sampler = 1.0
	}
	if cfg.Tracing.Batcher == "" {
		cfg.Tracing.Batcher = "jaeger"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "json"
	}
}
