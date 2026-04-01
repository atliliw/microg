# errors

基于 `github.com/pkg/errors` 包的增强版，增加对 **错误码 (Error Code)** 的支持，完全兼容 `github.com/pkg/errors`。

[![GoDoc](https://godoc.org/microg/pkg/errors?status.svg)](https://godoc.org/microg/pkg/errors)
[![Go Report Card](https://goreportcard.com/badge/microg/pkg/errors)](https://goreportcard.com/report/microg/pkg/errors)

## 特性

- ✅ **完全兼容** `github.com/pkg/errors` - 所有 API 保持一致
- ✅ **错误码支持** - 每个错误可关联唯一错误码
- ✅ **HTTP 状态码** - 错误码自动映射 HTTP 状态码
- ✅ **堆栈追踪** - 自动记录错误发生位置
- ✅ **错误链** - 支持 `Cause()` 和 Go 1.13+ `Unwrap()`
- ✅ **gRPC 集成** - 支持 gRPC 错误转换
- ✅ **代码生成** - 自动生成错误码注册代码
- ✅ **高性能** - 性能与 `pkg/errors` 基本持平

## 安装

```bash
go get microg/pkg/errors
```

## 快速开始

### 基本用法

```go
package main

import (
    "fmt"
    "microg/pkg/errors"
)

func main() {
    // 创建新错误（带堆栈）
    err := errors.New("something went wrong")
    
    // 格式化创建错误
    err = errors.Errorf("user %s not found", "john")
    
    // 包装错误（添加上下文）
    original := errors.New("database connection failed")
    err = errors.Wrap(original, "failed to query users")
    
    // 获取原始错误
    cause := errors.Cause(err)
    fmt.Println(cause) // database connection failed
}
```

### 错误码用法

```go
package main

import (
    "fmt"
    "microg/pkg/errors"
)

// 定义错误码常量
const (
    // ErrUserNotFound - 404: User not found.
    ErrUserNotFound = 100001
    
    // ErrInvalidParam - 400: Invalid parameter.
    ErrInvalidParam = 100002
    
    // ErrDatabase - 500: Database error.
    ErrDatabase = 100003
)

func main() {
    // 创建带错误码的错误
    err := errors.WithCode(ErrUserNotFound, "user %s not found", "john")
    
    // 包装错误（保持原错误码）
    err = errors.Wrap(err, "query failed")
    
    // 包装错误（更改错误码）
    err = errors.WrapC(err, ErrDatabase, "database operation failed")
    
    // 解析错误码
    coder := errors.ParseCoder(err)
    fmt.Println(coder.Code())        // 100003
    fmt.Println(coder.HTTPStatus())  // 500
    fmt.Println(coder.String())      // Database error
    
    // 判断是否包含特定错误码
    if errors.IsCode(err, ErrDatabase) {
        fmt.Println("database error occurred")
    }
}
```

## API 文档

### 基础错误函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `New(message string) error` | 创建新错误（带堆栈） | `errors.New("failed")` |
| `Errorf(format string, args ...interface{}) error` | 格式化创建错误 | `errors.Errorf("user %s not found", "john")` |

### 错误包装函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `Wrap(err error, message string) error` | 包装错误，添加消息和堆栈 | `errors.Wrap(err, "failed")` |
| `Wrapf(err error, format string, args ...interface{}) error` | 格式化包装错误 | `errors.Wrapf(err, "query %s failed", "users")` |
| `WithStack(err error) error` | 仅添加堆栈（不添加消息） | `errors.WithStack(err)` |
| `WithMessage(err error, message string) error` | 仅添加消息（不添加堆栈） | `errors.WithMessage(err, "context")` |
| `WithMessagef(err error, format string, args ...interface{}) error` | 格式化添加消息 | `errors.WithMessagef(err, "context %d", 1)` |

### 错误码函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `WithCode(code int, format string, args ...interface{}) error` | 创建带错误码的错误 | `errors.WithCode(100001, "not found")` |
| `WrapC(err error, code int, format string, args ...interface{}) error` | 包装错误并设置新错误码 | `errors.WrapC(err, 100002, "wrapped")` |
| `ParseCoder(err error) Coder` | 解析错误的 Coder 信息 | `errors.ParseCoder(err)` |
| `IsCode(err error, code int) bool` | 判断错误链中是否包含指定错误码 | `errors.IsCode(err, 100001)` |
| `Cause(err error) error` | 获取错误链的原始错误 | `errors.Cause(err)` |

### 错误码注册

```go
package myapp

import (
    "microg/pkg/errors"
    "net/http"
)

// 实现 Coder 接口
type ErrCode struct {
    C    int    // 错误码
    HTTP int    // HTTP 状态码
    Ext  string // 外部错误消息
    Ref  string // 参考文档链接
}

func (e ErrCode) Code() int       { return e.C }
func (e ErrCode) HTTPStatus() int { return e.HTTP }
func (e ErrCode) String() string  { return e.Ext }
func (e ErrCode) Reference() string { return e.Ref }

// 注册错误码
func init() {
    coder := ErrCode{
        C:    100001,
        HTTP: http.StatusNotFound,
        Ext:  "User not found",
        Ref:  "https://docs.example.com/errors/100001",
    }
    errors.MustRegister(coder)
}
```

### gRPC 错误转换

```go
package main

import (
    "microg/pkg/errors"
)

func main() {
    // 创建带错误码的错误
    err := errors.WithCode(100001, "user not found")
    
    // 转换为 gRPC 错误（用于 gRPC 服务端返回）
    grpcErr := errors.ToGrpcError(err)
    
    // 从 gRPC 错误解析（用于 gRPC 客户端接收）
    parsedErr := errors.FromGrpcError(grpcErr)
    coder := errors.ParseCoder(parsedErr)
    fmt.Println(coder.Code()) // 100001
}
```

### 格式化输出

所有错误支持 `fmt.Formatter` 接口：

```go
err := errors.New("something failed")

// 基本输出
fmt.Printf("%s", err)  // something failed
fmt.Printf("%v", err)  // something failed

// 详细输出（带堆栈）
fmt.Printf("%+v", err)
// Output:
// something failed
// github.com/me/myapp/main.go:12 (0x1234567)
// main.myFunction
// ...
```

## 错误码设计规范

### 错误码格式

建议使用 6 位数字错误码，按模块分组：

```
100001 - 基础错误
100101 - 数据库错误
100201 - 认证授权错误
100301 - 编解码错误
200001 - 用户模块错误
300001 - 订单模块错误
```

### HTTP 状态码映射

建议错误码关联的 HTTP 状态码：

| HTTP Code | 适用场景 |
|-----------|---------|
| 200 | 成功 |
| 400 | 参数错误、请求格式错误 |
| 401 | 认证失败、Token 无效 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 内部错误、数据库错误 |

## 代码生成工具

`codegen` 工具可自动生成错误码注册代码，避免手动编写。

### 安装

```bash
go install microg/pkg/tools/codegen@latest
```

### 使用方法

**1. 定义错误码常量**

```go
//go:generate codegen -type=int

// 用户模块错误码
const (
    // ErrUserNotFound - 404: User not found.
    ErrUserNotFound int = iota + 100001
    
    // ErrUserAlreadyExists - 400: User already exists.
    ErrUserAlreadyExists
    
    // ErrUserDisabled - 403: User is disabled.
    ErrUserDisabled
)
```

**2. 生成代码**

```bash
# 在定义文件目录下执行
go generate

# 或手动执行
codegen -type=int ./myapp/code/
```

**3. 生成结果**

工具会自动生成 `code_generated.go`：

```go
// Code generated by "codegen -type=int"; DO NOT EDIT.

package myapp

func init() {
    register(ErrUserNotFound, 404, "User not found")
    register(ErrUserAlreadyExists, 400, "User already exists")
    register(ErrUserDisabled, 403, "User is disabled")
}
```

**4. 生成错误码文档**

```bash
# 生成 Markdown 文档
codegen -type=int -doc -output error_code.md ./myapp/code/
```

### 注释格式要求

错误码常量的注释必须遵循格式：

```go
// ErrUserNotFound - 404: User not found.
ErrUserNotFound int = iota + 100001
```

格式：`// <错误名> - <HTTP状态码>: <错误描述>.`

## 最佳实践

### 1. 错误处理层次

```go
// 数据层 - 返回原始错误
func GetUser(id int) (*User, error) {
    user, err := db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        return nil, errors.WrapC(err, ErrDatabase, "query user failed")
    }
    return user, nil
}

// 业务层 - 添加业务上下文
func GetUserInfo(id int) (*UserInfo, error) {
    user, err := GetUser(id)
    if err != nil {
        return nil, errors.Wrapf(err, "get user info for id %d", id)
    }
    return &UserInfo{User: user}, nil
}

// 控制层 - 返回给用户
func HandleGetUser(c *gin.Context) {
    id := c.GetInt("id")
    info, err := GetUserInfo(id)
    if err != nil {
        coder := errors.ParseCoder(err)
        c.JSON(coder.HTTPStatus(), gin.H{
            "code":    coder.Code(),
            "message": coder.String(),
        })
        return
    }
    c.JSON(200, info)
}
```

### 2. 错误码分组管理

```go
// internal/code/base.go - 基础错误
const (
    ErrSuccess     = 100000
    ErrUnknown     = 100001
    ErrBind        = 100002
)

// internal/code/user.go - 用户模块错误
const (
    ErrUserNotFound       int = iota + 200001
    ErrUserAlreadyExists
)

// internal/code/order.go - 订单模块错误  
const (
    ErrOrderNotFound      int = iota + 300001
    ErrOrderStatusInvalid
)
```

### 3. HTTP API 错误响应

```go
{
    "code": 100001,
    "message": "User not found",
    "reference": "https://docs.example.com/errors/100001"
}
```

## 与 pkg/errors 兼容性

| 功能 | pkg/errors | microg/pkg/errors |
|------|-----------|-------------------|
| `New` | ✅ | ✅ |
| `Errorf` | ✅ | ✅ |
| `Wrap` | ✅ | ✅ |
| `Wrapf` | ✅ | ✅ |
| `WithStack` | ✅ | ✅ |
| `WithMessage` | ✅ | ✅ |
| `WithMessagef` | ✅ | ✅ |
| `Cause` | ✅ | ✅ |
| 堆栈追踪 | ✅ | ✅ |
| `Unwrap` (Go 1.13+) | ✅ | ✅ |
| 错误码 | ❌ | ✅ |
| HTTP 状态码 | ❌ | ✅ |
| gRPC 转换 | ❌ | ✅ |

**迁移指南：** 直接替换 import 即可：

```go
// 之前
import "github.com/pkg/errors"

// 之后
import "microg/pkg/errors"
```

## 性能

基准测试结果与 `pkg/errors` 基本持平：

```
BenchmarkNew-8                  5000000    230 ns/op    96 B/op    1 allocs/op
BenchmarkWrap-8                 3000000    350 ns/op   128 B/op    2 allocs/op
BenchmarkWithCode-8             3000000    380 ns/op   144 B/op    2 allocs/op
```

## 参考设计

错误码设计参考：[marmotedu/sample-code](https://github.com/marmotedu/sample-code)

## License

MIT License