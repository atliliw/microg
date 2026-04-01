# log

基于 `go.uber.org/zap` 的高性能日志包，增加对 **OpenTelemetry 链路追踪** 的集成支持，提供结构化日志和格式化日志两种 API。

[![GoDoc](https://godoc.org/github.com/atliliw/microg/pkg/log?status.svg)](https://godoc.org/github.com/atliliw/microg/pkg/log)
[![Go Report Card](https://goreportcard.com/badge/github.com/atliliw/microg/pkg/log)](https://goreportcard.com/report/github.com/atliliw/microg/pkg/log)

## 特性

- ✅ **高性能** - 基于 zap，性能卓越
- ✅ **OpenTelemetry 集成** - 日志自动关联链路追踪
- ✅ **结构化日志** - 支持字段化日志输出
- ✅ **格式化日志** - 提供 printf 风格 API
- ✅ **多级别支持** - Debug、Info、Warn、Error、Panic、Fatal
- ✅ **Context 集成** - 从 context 自动提取 traceID
- ✅ **Gin 集成** - 自动提取 requestID 和 username
- ✅ **灵活配置** - 支持命令行参数配置
- ✅ **多种输出格式** - console、json 格式
- ✅ **SugaredLogger** - 提供更便捷的 API

## 安装

```bash
go get github.com/atliliw/microg/pkg/log
```

## 快速开始

### 基本用法

```go
package main

import (
    "github.com/atliliw/microg/pkg/log"
)

func main() {
    // 使用默认配置初始化日志
    log.Init(nil)
    
    // 不同级别的日志
    log.Debug("debug message")
    log.Info("info message")
    log.Warn("warn message")
    log.Error("error message")
    
    // 格式化日志
    log.Infof("user %s logged in", "john")
    log.Errorf("failed to process: %v", err)
    
    // 结构化日志
    log.Info("user action",
        log.String("user", "john"),
        log.Int("action", 1),
        log.String("ip", "192.168.1.1"),
    )
    
    // 刷新日志缓冲
    log.Flush()
}
```

### 自定义配置

```go
package main

import (
    "github.com/atliliw/microg/pkg/log"
)

func main() {
    // 创建自定义配置
    opts := log.NewOptions()
    opts.Level = "debug"              // 日志级别
    opts.Format = "console"           // 输出格式 (console/json)
    opts.EnableColor = true           // 启用彩色输出
    opts.EnableTraceID = true         // 启用 traceID 输出
    opts.OutputPaths = []string{"stdout", "/var/log/app.log"}
    
    // 初始化日志
    log.Init(opts)
    
    log.Info("logger initialized with custom options")
}
```

### Context 集成（链路追踪）

```go
package main

import (
    "context"
    "github.com/atliliw/microg/pkg/log"
)

func ProcessOrder(ctx context.Context, orderId string) {
    // 日志自动关联到 OpenTelemetry span
    log.InfoC(ctx, "processing order",
        log.String("orderId", orderId),
    )
    
    // 格式化版本
    log.InfofC(ctx, "order %s processed", orderId)
    
    // 结构化版本
    log.DebugwC(ctx, "order details",
        "orderId", orderId,
        "status", "pending",
    )
}
```

### Gin 中间件集成

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/atliliw/microg/pkg/log"
)

func main() {
    r := gin.New()
    
    // 使用中间件注入 requestID 和 username
    r.Use(RequestIDMiddleware())
    r.Use(AuthMiddleware())
    
    r.GET("/api/user", func(c *gin.Context) {
        // 从 gin.Context 自动提取 requestID 和 username
        log.InfoC(c.Request.Context(), "api called",
            log.String("path", c.Request.URL.Path),
        )
        c.JSON(200, gin.H{"message": "ok"})
    })
    
    r.Run(":8080")
}
```

## API 文档

### 全局日志函数

#### 基础日志函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `Debug(msg string, fields ...Field)` | 输出 Debug 级别日志 | `log.Debug("debug")` |
| `Info(msg string, fields ...Field)` | 输出 Info 级别日志 | `log.Info("info")` |
| `Warn(msg string, fields ...Field)` | 输出 Warn 级别日志 | `log.Warn("warn")` |
| `Error(msg string, fields ...Field)` | 输出 Error 级别日志 | `log.Error("error")` |
| `Panic(msg string, fields ...Field)` | 输出 Panic 级别日志并 panic | `log.Panic("panic")` |
| `Fatal(msg string, fields ...Field)` | 输出 Fatal 级别日志并退出 | `log.Fatal("fatal")` |

#### 格式化日志函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `Debugf(format string, v ...interface{})` | 格式化 Debug 日志 | `log.Debugf("id: %d", 1)` |
| `Infof(format string, v ...interface{})` | 格式化 Info 日志 | `log.Infof("user: %s", "john")` |
| `Warnf(format string, v ...interface{})` | 格式化 Warn 日志 | `log.Warnf("warn: %v", err)` |
| `Errorf(format string, v ...interface{})` | 格式化 Error 日志 | `log.Errorf("error: %v", err)` |
| `Panicf(format string, v ...interface{})` | 格式化 Panic 日志 | `log.Panicf("panic: %s", msg)` |
| `Fatalf(format string, v ...interface{})` | 格式化 Fatal 日志 | `log.Fatalf("fatal: %s", msg)` |

#### Key-Value 日志函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `Debugw(msg string, keysAndValues ...interface{})` | KV 格式 Debug 日志 | `log.Debugw("msg", "key", "val")` |
| `Infow(msg string, keysAndValues ...interface{})` | KV 格式 Info 日志 | `log.Infow("msg", "id", 1)` |

#### Context 日志函数

所有日志函数都有带 Context 的版本（后缀 `C`）：

| 函数 | 说明 |
|------|------|
| `DebugC(ctx, msg, fields...)` | 从 context 提取 traceID |
| `InfoC(ctx, msg, fields...)` | 从 context 提取 traceID |
| `DebugfC(ctx, format, v...)` | 格式化 + Context |
| `InfofC(ctx, format, v...)` | 格式化 + Context |
| `DebugwC(ctx, msg, kvs...)` | KV + Context |
| `InfowC(ctx, msg, kvs...)` | KV + Context |

### Logger 对象方法

```go
package main

import (
    "context"
    "github.com/atliliw/microg/pkg/log"
)

func main() {
    // 创建自定义 Logger
    opts := log.NewOptions()
    opts.Level = "debug"
    logger := log.New(opts)
    
    // 基础方法
    logger.Debug("message")
    logger.Info("message")
    
    // Context 方法
    logger.DebugContext(ctx, "message")
    logger.InfoContext(ctx, "message")
    
    // 格式化方法
    logger.DebugfContext(ctx, "format %s", "val")
    
    // SugaredLogger
    sugar := logger.Sugar()
    sugar.Debugf("format %s", "val")
    sugar.Infow("message", "key", "value")
}
```

### 日志字段类型

```go
// 字段类型（zap.Field 别名）
log.String("key", "value")
log.Int("key", 123)
log.Int64("key", 123456)
log.Float64("key", 1.23)
log.Bool("key", true)
log.Duration("key", time.Second)
log.Time("key", time.Now())
log.Err(err)                           // 错误对象
log.Any("key", anyObject)              // 任意类型
log.Object("key", objectMarshaler)     // 对象
log.Array("key", arrayMarshaler)       // 数组
log.Namespace("namespace")             // 嵌套命名空间
log.Stack("stacktrace")                // 堆栈追踪
```

### 日志级别

```go
const (
    DebugLevel  // 详细日志，通常生产环境关闭
    InfoLevel   // 默认级别
    WarnLevel   // 警告日志
    ErrorLevel  // 错误日志
    PanicLevel  // Panic 日志，记录后 panic
    FatalLevel  // Fatal 日志，记录后退出
)
```

## 配置选项

### Options 结构

```go
type Options struct {
    Level             string   // 日志级别: debug/info/warn/error
    Format            string   // 输出格式: console/json
    EnableColor       bool     // 启用彩色输出 (仅 console)
    DisableCaller     bool     // 禁用调用位置输出
    DisableStacktrace bool     // 禁用堆栈追踪
    Development       bool     // 开发模式
    EnableTraceID     bool     // 启用 traceID 输出
    EnableTraceStack  bool     // 启用 trace 堆栈
    OutputPaths       []string // 输出路径: ["stdout", "/var/log/app.log"]
    ErrorOutputPaths  []string // 错误输出路径: ["stderr"]
    Name              string   // Logger 名称
}
```

### 默认配置

```go
opts := log.NewOptions()
// 默认值：
// Level:             "info"
// Format:            "console"
// EnableColor:       false
// DisableCaller:     false
// DisableStacktrace: false
// Development:       false
// OutputPaths:       ["stdout"]
// ErrorOutputPaths:  ["stderr"]
```

### 命令行参数

可通过 `AddFlags` 添加到命令行参数：

```go
package main

import (
    "flag"
    "github.com/spf13/pflag"
    "github.com/atliliw/microg/pkg/log"
)

func main() {
    opts := log.NewOptions()
    
    pflag.CommandLine.AddFlagSet(pflag.NewFlagSet("log", pflag.ExitOnError))
    opts.AddFlags(pflag.CommandLine)
    
    pflag.Parse()
    log.Init(opts)
}
```

支持的命令行参数：

```
--log.level=info                      日志级别
--log.format=console                  输出格式 (console/json)
--log.enable-color=false              启用彩色输出
--log.disable-caller=false            禁用调用位置
--log.disable-stacktrace=false        禁用堆栈追踪
--log.development=false               开发模式
--log.output-paths=["stdout"]         输出路径
--log.error-output-paths=["stderr"]   错误输出路径
--log.name=""                         Logger 名称
```

### Logger 配置函数

```go
// 创建带选项的 Logger
logger := log.New(opts)

// 克隆 Logger 并修改配置
clone := logger.Clone(
    log.WithMinLevel(zapcore.WarnLevel),
    log.WithErrorStatusLevel(zapcore.ErrorLevel),
    log.WithCaller(false),
    log.WithStackTrace(true),
    log.WithTraceIDField(true),
)
```

## OpenTelemetry 集成

### 自动关联 Span

```go
package main

import (
    "context"
    "github.com/atliliw/microg/pkg/log"
    "go.opentelemetry.io/otel/trace"
)

func Process(ctx context.Context) {
    // ctx 包含活跃的 span
    span := trace.SpanFromContext(ctx)
    
    // 日志自动添加到 span events
    log.InfoC(ctx, "processing started",
        log.String("step", "init"),
    )
    
    // 日志字段会添加为 span attributes
    // log.severity: INFO
    // log.message: processing started
    // step: init
    
    // Error 级别会设置 span status 为 Error
    log.ErrorC(ctx, "processing failed",
        log.Err(err),
    )
    // span.SetStatus(codes.Error, "processing failed")
}
```

### TraceID 输出

```go
// 配置启用 traceID 输出
opts := log.NewOptions()
opts.EnableTraceID = true
log.Init(opts)

// 日志中包含 trace_id 字段
log.InfoC(ctx, "message")
// Output: {"level":"INFO","message":"message","trace_id":"abc123..."}
```

### Gin Context 集成

当传入 `*gin.Context` 时，自动提取：

```go
// 自动提取 gin.Context 中的字段
log.InfoC(c.Request.Context(), "message")
// 自动添加：
// - requestID (from gin.Context value "requestID")
// - username (from gin.Context value "username")
// - trace_id (from OpenTelemetry span)
```

## 输出格式

### Console 格式（默认）

```
2026-04-01T10:00:00.000+0800 INFO    user action    caller=main.go:42    user=john    action=1
```

启用彩色输出后：

```
2026-04-01T10:00:00.000+0800 INFO    user action    caller=main.go:42    user=john    action=1
                                    ↑ 绿色
```

### JSON 格式

```json
{
    "level": "INFO",
    "timestamp": "2026-04-01T10:00:00.000+0800",
    "caller": "main.go:42",
    "message": "user action",
    "user": "john",
    "action": 1,
    "trace_id": "abc123..."
}
```

## 最佳实践

### 1. 分层日志

```go
// 数据层 - 详细日志
func GetUser(id int) (*User, error) {
    log.Debug("query user", log.Int("id", id))
    user, err := db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        log.Error("query failed", log.Err(err), log.Int("id", id))
        return nil, err
    }
    return user, nil
}

// 业务层 - 关键操作日志
func GetUserInfo(ctx context.Context, id int) (*UserInfo, error) {
    log.InfoC(ctx, "get user info", log.Int("id", id))
    user, err := GetUser(id)
    if err != nil {
        return nil, err
    }
    return &UserInfo{User: user}, nil
}

// 控制层 - 请求日志
func HandleGetUser(c *gin.Context) {
    id := c.GetInt("id")
    log.InfoC(c.Request.Context(), "handle request",
        log.String("path", c.Request.URL.Path),
        log.Int("id", id),
    )
    // ...
}
```

### 2. 结构化日志字段

```go
// 推荐：使用结构化字段
log.Info("user action",
    log.String("user", username),
    log.Int("action", actionType),
    log.String("ip", clientIP),
    log.Duration("elapsed", elapsed),
)

// 不推荐：格式化字符串
log.Infof("user %s performed action %d from %s in %v", username, actionType, clientIP, elapsed)
```

### 3. 错误日志规范

```go
// 推荐：使用 log.Err(err)
log.Error("operation failed",
    log.Err(err),
    log.String("operation", "create_user"),
)

// 不推荐：字符串拼接
log.Errorf("operation failed: %v", err)
```

### 4. Context 传递

```go
// 推荐：始终传递 context
func ProcessOrder(ctx context.Context, orderId string) {
    log.InfoC(ctx, "processing order", log.String("orderId", orderId))
    ValidateOrder(ctx, orderId)  // context 继续传递
}

// 不推荐：丢失 context
func ProcessOrder(orderId string) {
    log.Info("processing order", log.String("orderId", orderId))  // 无法关联 trace
}
```

### 5. 环境区分

```go
// 开发环境
opts := log.NewOptions()
opts.Level = "debug"
opts.Format = "console"
opts.EnableColor = true
opts.Development = true

// 生产环境
opts := log.NewOptions()
opts.Level = "info"
opts.Format = "json"
opts.EnableColor = false
opts.OutputPaths = []string{"stdout", "/var/log/app.log"}
```

## 性能

基于 zap 的高性能设计：

```
BenchmarkDebug-8          10000000    150 ns/op    0 B/op    0 allocs/op
BenchmarkInfo-8           10000000    150 ns/op    0 B/op    0 allocs/op
BenchmarkInfoFields-8      5000000    300 ns/op   32 B/op    1 allocs/op
BenchmarkInfof-8           3000000    400 ns/op   64 B/op    2 allocs/op
```

**性能建议：**
- 使用结构化字段 API (`log.Info(msg, fields...)`) 而不是格式化 API (`log.Infof`)
- 禁用不需要的调用位置 (`DisableCaller: true`)
- 生产环境使用 JSON 格式

## 与 zap 兼容性

| 功能 | zap | github.com/atliliw/microg/pkg/log |
|------|-----|----------------|
| Logger | ✅ | ✅ |
| SugaredLogger | ✅ | ✅ |
| 结构化字段 | ✅ | ✅ |
| 格式化输出 | ✅ | ✅ |
| 日志级别 | ✅ | ✅ |
| Context 支持 | ❌ | ✅ |
| OpenTelemetry 集成 | ❌ | ✅ |
| Gin Context 集成 | ❌ | ✅ |
| 命令行参数 | ❌ | ✅ |
| traceID 输出 | ❌ | ✅ |

**获取底层 zap.Logger：**

```go
// 获取 zap.Logger 用于第三方库集成
zapLogger := log.ZapLogger()

// 获取标准库 log.Logger
stdLogger := log.StdInfoLogger()
```

## 常见问题

### 如何动态修改日志级别？

```go
// 当前不支持动态修改，需要重新 Init
opts := log.NewOptions()
opts.Level = "debug"
log.Init(opts)
```

### 如何在第三方库中使用？

```go
// 1. 直接使用 ZapLogger
zapLogger := log.ZapLogger()
gormLogger := zapgorm.New(zapLogger)

// 2. 使用 SugaredLogger
sugar := log.New(opts).Sugar()
```

### 如何添加全局字段？

```go
logger := log.New(opts).WithOptions(
    zap.Fields(
        log.String("service", "my-service"),
        log.String("version", "v1.0.0"),
    ),
)
```

## 依赖

- [go.uber.org/zap](https://github.com/uber-go/zap) - 高性能日志库
- [go.opentelemetry.io/otel](https://github.com/open-telemetry/opentelemetry-go) - 链路追踪
- [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin) - HTTP 框架集成
- [github.com/spf13/pflag](https://github.com/spf13/pflag) - 命令行参数