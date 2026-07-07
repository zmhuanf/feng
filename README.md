[English](#-feng) | [中文](#-feng-中文)

# Feng

`feng` is a game communication framework based on **Gin** and **WebSocket**. It focuses on a small API surface for quickly building request/response and push-style game services.

> Note: this project is still under active development. Breaking API changes may happen before the first stable release.

## Multi-Platform Support

`feng` provides a Go server/client and is designed to work with game-engine clients:

- Cocos: [feng-cocos](https://github.com/zmhuanf/feng-cocos)
- Unity: [feng-unity](https://github.com/zmhuanf/feng-unity)

## Quick Start

### Server

```go
package main

import (
	"context"
	"log/slog"

	"github.com/zmhuanf/feng"
)

func main() {
	server := feng.NewServer(feng.NewDefaultServerConfig())

	// Register an echo route.
	if err := server.Handle("/echo", func(ctx feng.ServerContext, msg string) (string, error) {
		return msg, nil
	}); err != nil {
		slog.Error("register route failed", "error", err)
		return
	}

	if err := server.ListenAndServe(context.Background()); err != nil {
		slog.Error("server stopped", "error", err)
	}
}
```

### Client

```go
package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/zmhuanf/feng"
)

func main() {
	client := feng.NewClient(feng.NewDefaultClientConfig())
	if err := client.Connect(context.Background()); err != nil {
		slog.Error("connect failed", "error", err)
		return
	}
	defer client.Close()

	if err := client.Request(context.Background(), "/echo", "hello, world!", func(ctx feng.ClientContext, msg string) {
		fmt.Println(msg)
	}); err != nil {
		slog.Error("request failed", "error", err)
	}
}
```

## Basic Concepts

### Public API

The root package exposes the stable API:

- `feng.Server`, created by `feng.NewServer(feng.ServerConfig)`
- `feng.Client`, created by `feng.NewClient(feng.ClientConfig)`
- `feng.ServerContext` and `feng.ClientContext`
- `feng.User` and `feng.Room`
- `feng.Codec` and `feng.Logger`

Implementation details live under `internal/` and should not be imported by user code.

### Server

Register business handlers with `Handle`:

```go
server.Handle("/login", func(ctx feng.ServerContext, req LoginReq) (LoginResp, error) {
	return LoginResp{OK: true}, nil
})
```

Register middleware with `Use`:

```go
server.Use("/api", func(ctx feng.ServerContext, token string) error {
	if token == "" {
		return errors.New("empty token")
	}
	ctx.Set("token", token)
	return nil
})
```

Handler and middleware rules:

- The first argument must be `feng.ServerContext`.
- The second argument is optional and can be `string`, `[]byte`, struct, slice, map, bool, or number.
- A handler may return `error` or `(response, error)`.
- A middleware should return `error`; returning a non-nil error stops the route handler.

Useful server context methods:

- `ctx.Server()` returns the current `feng.Server`.
- `ctx.User()` returns the connected `feng.User`.
- `ctx.Room()` returns the current `feng.Room`.
- `ctx.Get(key)` and `ctx.Set(key, value)` store per-connection data.
- `ctx.GinContext()` returns the underlying `*gin.Context`.

### Client

Send a bidirectional request with `Request`:

```go
err := client.Request(context.Background(), "/profile", ProfileReq{ID: "1001"}, func(ctx feng.ClientContext, resp ProfileResp) {
	fmt.Println(resp.Name)
})
```

Send a one-way message with `Push`:

```go
err := client.Push("/ping", "hello")
```

The client can also receive server-initiated calls:

```go
client.Handle("/notice", func(ctx feng.ClientContext, msg string) error {
	fmt.Println(msg)
	return nil
})

client.Use("/notice", func(ctx feng.ClientContext) error {
	return nil
})
```

Handler and middleware rules are the same as the server side, except the first argument must be `feng.ClientContext`.

## Practical Example

```go
package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/zmhuanf/feng"
)

type LoginReq struct {
	Token string `json:"token"`
}

type AccountResp struct {
	Name     string `json:"name"`
	IsTester bool   `json:"is_tester"`
}

func main() {
	config := feng.NewDefaultServerConfig()
	config.Addr = "0.0.0.0"
	config.Port = 22002

	server := feng.NewServer(config)
	_ = server.Handle("/login", LoginHandler)
	_ = server.Handle("/get_account_info", GetAccountInfoHandler)

	if err := server.ListenAndServe(context.Background()); err != nil {
		slog.Error("server failed", "error", err)
	}
}

func LoginHandler(ctx feng.ServerContext, req LoginReq) error {
	if req.Token == "" {
		return errors.New("empty token")
	}

	// Store values on the connection context for later handlers.
	ctx.Set("uuid", "demo-user-id")
	return nil
}

func GetAccountInfoHandler(ctx feng.ServerContext) (AccountResp, error) {
	uuid, ok := ctx.Get("uuid")
	if !ok {
		return AccountResp{}, errors.New("not logged in")
	}

	return AccountResp{
		Name:     uuid.(string),
		IsTester: true,
	}, nil
}
```

<br/>

# Feng (中文)

`feng` 是一个基于 **Gin** 和 **WebSocket** 的游戏通信框架。它专注于用较小的 API 快速构建游戏服务里的请求/响应和推送通信。

> 注意：本项目仍在积极开发中，在第一个稳定版本发布前可能会出现破坏性 API 变更。

## 多平台支持

`feng` 提供 Go 服务端/客户端，并面向游戏引擎客户端设计：

- Cocos: [feng-cocos](https://github.com/zmhuanf/feng-cocos)
- Unity: [feng-unity](https://github.com/zmhuanf/feng-unity)

## 快速开始

### 服务端

```go
package main

import (
	"context"
	"log/slog"

	"github.com/zmhuanf/feng"
)

func main() {
	server := feng.NewServer(feng.NewDefaultServerConfig())

	// 注册一个 echo 路由。
	if err := server.Handle("/echo", func(ctx feng.ServerContext, msg string) (string, error) {
		return msg, nil
	}); err != nil {
		slog.Error("注册路由失败", "error", err)
		return
	}

	if err := server.ListenAndServe(context.Background()); err != nil {
		slog.Error("服务停止", "error", err)
	}
}
```

### 客户端

```go
package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/zmhuanf/feng"
)

func main() {
	client := feng.NewClient(feng.NewDefaultClientConfig())
	if err := client.Connect(context.Background()); err != nil {
		slog.Error("连接失败", "error", err)
		return
	}
	defer client.Close()

	if err := client.Request(context.Background(), "/echo", "hello, world!", func(ctx feng.ClientContext, msg string) {
		fmt.Println(msg)
	}); err != nil {
		slog.Error("请求失败", "error", err)
	}
}
```

## 基本概念

### 公开 API

根包只暴露稳定 API：

- `feng.Server`，通过 `feng.NewServer(feng.ServerConfig)` 创建
- `feng.Client`，通过 `feng.NewClient(feng.ClientConfig)` 创建
- `feng.ServerContext` 和 `feng.ClientContext`
- `feng.User` 和 `feng.Room`
- `feng.Codec` 和 `feng.Logger`

具体实现位于 `internal/` 下，业务代码不应直接导入。

### 服务端

使用 `Handle` 注册业务处理器：

```go
server.Handle("/login", func(ctx feng.ServerContext, req LoginReq) (LoginResp, error) {
	return LoginResp{OK: true}, nil
})
```

使用 `Use` 注册中间件：

```go
server.Use("/api", func(ctx feng.ServerContext, token string) error {
	if token == "" {
		return errors.New("empty token")
	}
	ctx.Set("token", token)
	return nil
})
```

处理器和中间件规则：

- 第一个参数必须是 `feng.ServerContext`。
- 第二个参数可选，支持 `string`、`[]byte`、结构体、切片、map、bool、数字。
- 处理器可以返回 `error` 或 `(response, error)`。
- 中间件应返回 `error`，返回非 nil error 会阻止路由处理器继续执行。

常用服务端上下文方法：

- `ctx.Server()` 返回当前 `feng.Server`。
- `ctx.User()` 返回当前连接的 `feng.User`。
- `ctx.Room()` 返回当前连接的 `feng.Room`。
- `ctx.Get(key)` 和 `ctx.Set(key, value)` 存取连接级上下文数据。
- `ctx.GinContext()` 返回底层 `*gin.Context`。

### 客户端

使用 `Request` 发送双向请求：

```go
err := client.Request(context.Background(), "/profile", ProfileReq{ID: "1001"}, func(ctx feng.ClientContext, resp ProfileResp) {
	fmt.Println(resp.Name)
})
```

使用 `Push` 发送单向消息：

```go
err := client.Push("/ping", "hello")
```

客户端也可以注册供服务端调用的接口：

```go
client.Handle("/notice", func(ctx feng.ClientContext, msg string) error {
	fmt.Println(msg)
	return nil
})

client.Use("/notice", func(ctx feng.ClientContext) error {
	return nil
})
```

客户端处理器和中间件规则与服务端一致，只是第一个参数必须是 `feng.ClientContext`。

## 实战示例

```go
package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/zmhuanf/feng"
)

type LoginReq struct {
	Token string `json:"token"`
}

type AccountResp struct {
	Name     string `json:"name"`
	IsTester bool   `json:"is_tester"`
}

func main() {
	config := feng.NewDefaultServerConfig()
	config.Addr = "0.0.0.0"
	config.Port = 22002

	server := feng.NewServer(config)
	_ = server.Handle("/login", LoginHandler)
	_ = server.Handle("/get_account_info", GetAccountInfoHandler)

	if err := server.ListenAndServe(context.Background()); err != nil {
		slog.Error("服务启动失败", "error", err)
	}
}

func LoginHandler(ctx feng.ServerContext, req LoginReq) error {
	if req.Token == "" {
		return errors.New("empty token")
	}

	// 将数据存入连接上下文，后续 handler 可以继续读取。
	ctx.Set("uuid", "demo-user-id")
	return nil
}

func GetAccountInfoHandler(ctx feng.ServerContext) (AccountResp, error) {
	uuid, ok := ctx.Get("uuid")
	if !ok {
		return AccountResp{}, errors.New("not logged in")
	}

	return AccountResp{
		Name:     uuid.(string),
		IsTester: true,
	}, nil
}
```
