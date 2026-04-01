.PHONY: all build test lint clean examples

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
BASIC_EXAMPLE=basic
HTTP_EXAMPLE=http
FULL_EXAMPLE=full

all: test build

# 编译所有示例
build: examples

# 编译示例
examples:
	@echo "Building examples..."
	@cd examples/basic && $(GOBUILD) -o ../../bin/$(BASIC_EXAMPLE) .
	@cd examples/http && $(GOBUILD) -o ../../bin/$(HTTP_EXAMPLE) .
	@cd examples/full && $(GOBUILD) -o ../../bin/$(FULL_EXAMPLE) .

# 运行测试
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# 运行测试（带覆盖率）
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# 代码检查
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

# 代码格式化
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# 下载依赖
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# 清理
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html

# 运行基础示例
run-basic:
	@echo "Running basic example..."
	@cd examples/basic && $(GOCMD) run .

# 运行 HTTP 示例
run-http:
	@echo "Running HTTP example..."
	@cd examples/http && $(GOCMD) run .

# 帮助
help:
	@echo "Available targets:"
	@echo "  make build        - 编译所有示例"
	@echo "  make test         - 运行测试"
	@echo "  make test-coverage - 运行测试（带覆盖率）"
	@echo "  make lint         - 代码检查"
	@echo "  make fmt          - 代码格式化"
	@echo "  make deps         - 下载依赖"
	@echo "  make clean        - 清理"
	@echo "  make run-basic    - 运行基础示例"
	@echo "  make run-http     - 运行 HTTP 示例"