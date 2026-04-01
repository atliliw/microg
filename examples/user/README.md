# User Service 示例

这是一个完整的微服务示例，展示了如何使用 Microg 框架构建生产级的用户服务。

## 📋 项目结构

```
user/
├── docker-compose.yaml    # 基础设施配置
├── README.md              # 使用文档
├── api/v1/                # API 定义
│   ├── user.proto         # Proto 定义
│   ├── user.pb.go         # Protobuf 生成代码
│   └── user_grpc.pb.go    # gRPC 生成代码
├── code/                  # 错误码定义
│   ├── base.go            # 基础错误码
│   ├── code.go            # 用户模块错误码
│   └── code_generated.go  # 自动生成的注册代码
│   └── error_code_generated.md  # 错误码文档
├── srv/                   # gRPC 服务端
│   ├── main.go            # 入口
│   ├── config/            # 配置
│   │   ├── config.yaml    # 配置文件
│   │   ├── config.go      # 配置结构
│   │   └── provider.go    # Wire Provider
│   ├── controller/        # gRPC 控制器
│   │   ├── user.go        # 用户控制器实现
│   │   └── provider.go    # Wire Provider
│   ├── service/           # 业务逻辑层
│   │   ├── user.go        # 用户服务实现
│   │   └── provider.go    # Wire Provider
│   ├── data/              # 数据访问层
│   │   ├── user.go        # 用户数据模型
│   │   ├── mysql.go       # MySQL 实现
│   │   └── provider.go    # Wire Provider
│   ├── server/            # 服务器配置
│   │   └── provider.go    # Wire Provider
│   ├── wire.go            # Wire 定义
│   └── wire_gen.go        # Wire 生成代码
└── client/                # HTTP 客户端（API 网关）
    ├── main.go            # 入口
    ├── config/            # 配置
    │   ├── config.yaml    # 配置文件
    │   ├── config.go      # 配置结构
    │   └── provider.go    # Wire Provider
    ├── controller/        # HTTP 控制器
    │   ├── user.go        # 用户控制器实现
    │   └── provider.go    # Wire Provider
    ├── service/           # 业务逻辑层
    │   ├── user.go        # 用户服务实现
    │   └── provider.go    # Wire Provider
    ├── data/              # gRPC 客户端封装
    │   ├── user.go        # 用户 gRPC 客户端
    │   └── provider.go    # Wire Provider
    ├── server/            # HTTP 服务器配置
    │   └── provider.go    # Wire Provider
    ├── wire.go            # Wire 定义
    └── wire_gen.go        # Wire 生成代码
```

## 🚀 快速开始

### 1. 启动基础设施

使用 Docker Compose 启动所需的基础设施服务：

```bash
# 启动必需服务（Consul + MySQL）
docker-compose up -d

# 启动可选服务（Jaeger 链路追踪）
docker-compose --profile tracing up -d

# 启动可选服务（Redis 缓存）
docker-compose --profile cache up -d

# 启动所有服务
docker-compose --profile tracing --profile cache up -d
```

**服务访问地址：**
- Consul UI: http://localhost:8500
- MySQL: localhost:3306 (root/root)
- Jaeger UI: http://localhost:16686
- Redis: localhost:6379

### 2. 配置调整

根据实际情况修改配置文件：

**srv/config/config.yaml**（gRPC 服务端配置）
```yaml
consul:
  address: "localhost:8500"  # Docker: localhost, 或 consul

mysql:
  host: "localhost"          # Docker: localhost, 或 mysql
  port: 3306
  username: "root"
  password: "root"
  database: "microg_user"

tracing:
  enabled: true              # 启用链路追踪
  endpoint: "http://localhost:14268/api/traces"
```

**client/config/config.yaml**（HTTP 客户端配置）
```yaml
consul:
  address: "localhost:8500"
```

### 3. 运行服务

**方式一：直接运行**

```bash
# 运行 gRPC 服务端
cd srv
go run .

# 运行 HTTP 客户端（另一个终端）
cd client
go run .
```

**方式二：构建后运行**

```bash
# 构建 gRPC 服务端
cd srv
go build -o user-srv .
./user-srv

# 构建 HTTP 客户端
cd client
go build -o user-client .
./user-client
```

### 4. 测试 API

**使用 curl 测试：**

```bash
# 查询用户列表
curl http://localhost:8081/api/v1/users

# 创建用户
curl -X POST http://localhost:8081/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "nickName": "John Doe",
    "passWord": "password123",
    "mobile": "13800138000"
  }'

# 查询指定用户
curl http://localhost:8081/api/v1/users/1

# 更新用户
curl -X PUT http://localhost:8081/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "nickName": "Jane Doe",
    "gender": "female"
  }'

# 删除用户
curl -X DELETE http://localhost:8081/api/v1/users/1
```

**使用 Postman/Insomnia：**
- 导入 API 定义或手动创建请求
- Base URL: http://localhost:8081

### 5. 验证服务注册

访问 Consul UI：http://localhost:8500

在 Services 列表中应该看到：
- `user-srv` (gRPC 服务端)
- `user-client` (HTTP API 网关)

### 6. 查看链路追踪（可选）

如果启用了 Jaeger，访问：http://localhost:16686

- Service: `user-srv` 或 `user-client`
- 可以查看请求的完整调用链路

## 🛠️ 开发指南

### Wire 依赖注入

项目使用 Wire 进行依赖注入管理。如果修改了依赖关系，需要重新生成：

```bash
# 生成 gRPC 服务端依赖注入代码
cd srv
wire ./...

# 生成 HTTP 客户端依赖注入代码
cd client
wire ./...
```

### 错误码生成

项目使用错误码生成工具自动注册错误码：

```bash
# 安装工具（首次）
go install github.com/atliliw/microg/pkg/tools/codegen@latest

# 生成错误码注册代码
cd code
go generate ./...

# 生成错误码文档
codegen -type=int -doc -output error_code_generated.md .
```

### API 更新

如果修改了 Proto 定义，需要重新生成：

```bash
# 安装 protoc 和插件
# 参考: https://grpc.io/docs/languages/go/quickstart/

# 生成代码
cd api/v1
protoc --go_out=. --go-grpc_out=. user.proto
```

## 📊 配置说明

### gRPC 服务端配置 (srv/config/config.yaml)

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| service.name | 服务名称 | user-srv |
| grpc.address | gRPC 监听地址 | :9001 |
| grpc.timeout | 请求超时时间 | 30s |
| grpc.metrics | 启用 Prometheus 指标 | true |
| consul.address | Consul 地址 | localhost:8500 |
| consul.enabled | 启用服务注册 | true |
| consul.health_check | 启用健康检查 | true |
| tracing.enabled | 启用链路追踪 | false |
| tracing.endpoint | Jaeger endpoint | http://localhost:14268 |
| log.level | 日志级别 | debug |
| log.format | 日志格式 | console |
| mysql.host | MySQL 主机 | localhost |
| mysql.database | 数据库名称 | microg_user |

### HTTP 客户端配置 (client/config/config.yaml)

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| service.name | 服务名称 | user-client |
| http.port | HTTP 监听端口 | 8081 |
| http.mode | Gin 运行模式 | debug |
| http.healthz | 健康检查接口 | true |
| http.metrics | Prometheus 指标 | true |
| consul.address | Consul 地址 | localhost:8500 |
| user_srv.name | gRPC 服务名称 | user-srv |

## 🔍 服务架构

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Client    │  HTTP   │  HTTP Client │  gRPC   │  gRPC Srv   │
│  (Browser)  │───────▶ │  (API 网关)   │───────▶ │ (业务服务)   │
└─────────────┘         └──────────────┘         └─────────────┘
                               │                        │
                               │                        │
                               ├────────────────────────┤
                               │                        │
                        ┌──────▼────────────────────────▼──────┐
                        │           Consul (服务发现)            │
                        └──────────────────────────────────────┘
                               │                        │
                               │                        │
                        ┌──────▼──────┐          ┌─────▼──────┐
                        │   MySQL     │          │   Jaeger   │
                        │  (数据库)    │          │ (链路追踪)  │
                        └─────────────┘          └────────────┘
```

**说明：**
1. **HTTP Client (API 网关)**：接收 HTTP 请求，通过 Consul 发现 gRPC 服务，转发请求
2. **gRPC Server (业务服务)**：处理业务逻辑，操作 MySQL 数据库，注册到 Consul
3. **Consul**：服务注册与发现中心，管理服务健康状态
4. **MySQL**：持久化存储用户数据
5. **Jaeger**：分布式链路追踪，监控请求调用链

## 🐛 常见问题

### 1. 服务注册失败

**问题：** Consul 中看不到服务

**解决：**
```bash
# 检查 Consul 是否运行
docker ps | grep consul

# 检查 Consul 连接
curl http://localhost:8500/v1/catalog/services

# 检查服务配置中的 Consul 地址是否正确
```

### 2. 数据库连接失败

**问题：** MySQL 连接报错

**解决：**
```bash
# 检查 MySQL 是否运行
docker ps | grep mysql

# 手动连接测试
mysql -h localhost -u root -p root

# 检查数据库是否存在
mysql -h localhost -u root -p root -e "SHOW DATABASES LIKE 'microg_user';"
```

### 3. gRPC 调用失败

**问题：** HTTP Client 无法调用 gRPC Server

**解决：**
```bash
# 检查 gRPC Server 是否运行并注册到 Consul
curl http://localhost:8500/v1/catalog/service/user-srv

# 检查 HTTP Client 配置中的 user_srv.name 是否与 gRPC Server 的 service.name 一致

# 查看日志中的错误信息
```

### 4. Docker 容器启动失败

**问题：** Docker Compose 服务启动报错

**解决：**
```bash
# 查看容器日志
docker-compose logs consul
docker-compose logs mysql

# 重启服务
docker-compose restart

# 完全重建
docker-compose down
docker-compose up -d
```

## 📚 相关文档

- [Microg 主文档](../../README.md)
- [错误处理文档](../../pkg/errors/README.md)
- [日志文档](../../pkg/log/README.md)
- [Wire 使用指南](https://github.com/google/wire)

## 🎯 最佳实践

1. **配置管理**
   - 生产环境使用环境变量或配置中心
   - 开发环境使用本地配置文件
   - 避免硬编码配置

2. **错误处理**
   - 使用错误码而非字符串
   - 区分业务错误和系统错误
   - 记录完整的错误上下文

3. **日志规范**
   - 使用结构化日志
   - 包含必要的上下文信息（traceID、userID等）
   - 合理设置日志级别

4. **服务治理**
   - 启用健康检查
   - 配置合理的超时时间
   - 使用链路追踪定位问题

## 📝 License

本示例遵循 Microg 项目的 MIT License。