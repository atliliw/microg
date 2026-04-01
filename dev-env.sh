#!/bin/bash
# Microg 开发环境一键部署脚本
# 使用方法: chmod +x dev-env.sh && ./dev-env.sh

set -e

# 创建配置目录
mkdir -p config

# 创建 prometheus.yml
cat > config/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'microg-services'
    static_configs:
      - targets:
          - 'host.docker.internal:8080'
          - 'host.docker.internal:8081'
          - 'host.docker.internal:9000'
          - 'host.docker.internal:9001'
EOF

# 创建 docker-compose.yml
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  # ==================== 服务注册与发现 ====================
  consul:
    image: consul:1.15
    container_name: consul
    hostname: consul
    ports:
      - "8500:8500"
      - "8600:8600/udp"
    command: agent -server -ui -bootstrap-expect=1 -client=0.0.0.0
    healthcheck:
      test: ["CMD", "consul", "members"]
      interval: 10s
      timeout: 5s
      retries: 3
    restart: unless-stopped

  # ==================== 链路追踪 ====================
  jaeger:
    image: jaegertracing/all-in-one:1.50
    container_name: jaeger
    hostname: jaeger
    ports:
      - "14268:14268"
      - "16686:16686"
      - "6831:6831/udp"
    environment:
      - COLLECTOR_ZIPKIN_HOST_PORT=:9411
    restart: unless-stopped

  # ==================== 指标收集 ====================
  prometheus:
    image: prom/prometheus:v2.47.0
    container_name: prometheus
    hostname: prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./config/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.enable-lifecycle'
    restart: unless-stopped

  # ==================== MySQL 数据库 ====================
  mysql:
    image: mysql:8.0
    container_name: mysql
    hostname: mysql
    ports:
      - "3306:3306"
    environment:
      - MYSQL_ROOT_PASSWORD=root
      - MYSQL_DATABASE=microg
      - TZ=Asia/Shanghai
    volumes:
      - mysql_data:/var/lib/mysql
    command: --default-authentication-plugin=mysql_native_password --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
    restart: unless-stopped

  # ==================== Redis (可选) ====================
  redis:
    image: redis:7-alpine
    container_name: redis
    hostname: redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    restart: unless-stopped

volumes:
  prometheus_data:
  mysql_data:
  redis_data:
EOF

echo "=========================================="
echo "配置文件已生成:"
echo "  - docker-compose.yml"
echo "  - config/prometheus.yml"
echo "=========================================="
echo ""
echo "启动服务: docker-compose up -d"
echo "停止服务: docker-compose down"
echo "查看日志: docker-compose logs -f"
echo ""
echo "服务地址:"
echo "  Consul:     http://localhost:8500"
echo "  Jaeger:     http://localhost:16686"
echo "  Prometheus: http://localhost:9090"
echo "  MySQL:      localhost:3306 (root/root)"
echo "  Redis:      localhost:6379"
echo ""

# 启动服务
read -p "是否立即启动所有服务? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker-compose up -d
    echo ""
    echo "服务启动中... 等待健康检查..."
    sleep 5
    docker-compose ps
fi