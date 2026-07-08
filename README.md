[English](#-feng) | [中文](#-feng中文)

# 🍃 Feng

`feng` is a game communication framework based on **Gin** and **WebSocket**. It aims to provide an **easy-to-use and concise** communication solution to help developers quickly build game servers, rather than chasing extreme performance.

> ⚠️ **Note**: This project is currently under active development. Significant API changes may occur before the first official release.

---

## 🎮 Multi-Platform Support

`feng` ships with a Go server/client and is designed to work with mainstream game engines:

- **Cocos**: [feng-cocos](https://github.com/zmhuanf/feng-cocos)
- **Unity**: [feng-unity](https://github.com/zmhuanf/feng-unity)

---

## 🚀 Quick Start

Creating an Echo service with `feng` is straightforward.

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
*(Note: error handling is omitted above for brevity)*

---

## 📖 Basic Concepts

### 🖥️ Server

#### Handle（注册处理器）
Register business logic with the `Handle` method.

*   **Path**: First parameter. Can be any string; not strictly required to start with `/`.
*   **Handler Function**: Second parameter.
    *   **Arguments**:
        1.  `ctx feng.ServerContext` (Required): The server context.
        2.  `Request Data` (Optional): `string`, `[]byte`, struct, slice, map, bool, or any numeric type deserializable by `Codec`.
    *   **Return Values**:
        1.  `error` (Required): The processing result. Returned alone when there is no response body.
        2.  `Response Data` (Optional): Returned as the first value alongside `error`. Can be `string`, `[]byte`, struct, slice, map, bool, or any numeric type serializable by `Codec`.

#### Use（中间件）
Register middleware with the `Use` method.

*   **Scope**: First parameter is the path prefix. Middleware applies to all handlers matching that prefix, in the order they were added.
*   **Signature**: First argument must be `feng.ServerContext`; an optional payload argument follows the same rules as handlers. The **return value must be `error` only**.
*   **Interception**: Returning a non-`nil` error stops the route handler and short-circuits the call.

#### Server Context（上下文）
Useful methods on `feng.ServerContext`:

*   `ctx.Server()` returns the current `feng.Server`.
*   `ctx.User()` returns the connected `feng.User`.
*   `ctx.Room()` returns the current `feng.Room`.
*   `ctx.Get(key)` and `ctx.Set(key, value)` store per-connection data.
*   `ctx.GinContext()` returns the underlying `*gin.Context`.

#### Core Types（核心类型）
The root package exposes the following stable types:

*   `feng.Server` — created by `feng.NewServer(feng.ServerConfig)`.
*   `feng.ServerConfig` — built from `feng.NewDefaultServerConfig()`; `Codec` and `Logger` are pluggable.
*   `feng.ServerContext` — passed into handlers and middleware.
*   `feng.User` / `feng.Room` — connection and room abstractions (`ID`, `JoinRoom`, `LeaveRoom`, `Push`, `Request`, etc.).
*   `feng.Codec` — serialization interface. Use `feng.NewJSONCodec()` (or `NewJsonCodec`) / `feng.NewProtoCodec()`.
*   `feng.Logger` — logging interface. Use `feng.NewSlogLogger()`.

> Implementation details live under `internal/` and should not be imported by user code.

### 📱 Client

#### Request（请求）
Send a bidirectional request via the `Request` method.

*   **Arguments**:
    1.  `context.Context` — request context.
    2.  `Path` — handler path on the peer.
    3.  `Data` — request payload.
    4.  `Callback` — invoked after the response arrives, with the same argument shape as a handler and no return value.

#### RequestAsync（异步请求）
Same as `Request` but without a `context.Context`. Use it when the caller does not need to wait for completion directly.

#### Push（推送）
Send a one-way message via the `Push` method.

*   **Arguments**: `Path`, `Data`.

#### Client Handler（客户端处理器）
Clients can register handlers and middleware for the server to call via `Handle` and `Use`. The usage is identical to the server side, except the first argument must be `feng.ClientContext`.

#### Core Types（核心类型）
*   `feng.Client` — created by `feng.NewClient(feng.ClientConfig)`.
*   `feng.ClientConfig` — built from `feng.NewDefaultClientConfig()`; `Codec`, `Logger`, `EnableTLS`, and `DirectConnect` are configurable.
*   `feng.ClientContext` — passed into client-side handlers and middleware.

---

## 💡 Practical Example

A more production-shaped example: simple login validation followed by a data fetch.

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

// Handler: login
func LoginHandler(ctx feng.ServerContext, req LoginReq) error {
	if req.Token == "" {
		return errors.New("empty token")
	}

	// Stash the user id on the connection context for later handlers.
	ctx.Set("uuid", "demo-user-id")
	return nil
}

// Handler: get account info
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

---

<br/>

# 🍃 Feng（中文）

`feng` 是一个基于 **Gin** 和 **WebSocket** 的游戏通信框架。它不追求极致性能，而是致力于提供一种**简单易用**的通信方案，帮助开发者快速搭建游戏服务器。

> ⚠️ **注意**：本项目仍在积极开发中，在第一个正式版本发布前，API 可能会出现较大变动。

---

## 🎮 多平台支持

`feng` 自带 Go 服务端/客户端，同时面向主流游戏引擎提供配套 SDK：

- **Cocos**: [feng-cocos](https://github.com/zmhuanf/feng-cocos)
- **Unity**: [feng-unity](https://github.com/zmhuanf/feng-unity)

---

## 🚀 快速开始

用 `feng` 写一个回显（Echo）服务非常简单。

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

调用接口同样直观：

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
*（注：以上示例为突出重点，省略了部分错误处理）*

---

## 📖 基本概念

### 🖥️ 服务端（Server）

#### Handle（注册处理器）
通过 `Handle` 方法注册业务逻辑。

*   **路径（Path）**：第一个参数，可以是任意字符串，不强制要求以 `/` 开头。
*   **处理器函数**：第二个参数。
    *   **参数**：
        1.  `ctx feng.ServerContext`（必须）：服务端上下文。
        2.  `请求数据`（可选）：支持 `string`、`[]byte`、结构体、切片、map、bool 以及任意数字类型，只要能被 `Codec` 反序列化即可。
    *   **返回值**：
        1.  `error`（必须）：表示处理结果。无响应体时只返回它即可。
        2.  `响应数据`（可选）：与 `error` 一同返回，作为第一个返回值。同样支持 `string`、`[]byte`、结构体、切片、map、bool 和数字类型。

#### Use（中间件）
通过 `Use` 方法注册中间件。

*   **作用范围**：第一个参数是路径前缀。中间件会按注册顺序作用于所有匹配该前缀的处理器。
*   **函数签名**：第一个参数必须是 `feng.ServerContext`，后续的请求参数与处理器一致。**返回值只能是 `error`**。
*   **拦截**：返回非 `nil` 的 error 会中断后续处理器的执行，并把错误回传给调用方。

#### Server Context（上下文）
`feng.ServerContext` 上常用的方法：

*   `ctx.Server()` 获取当前 `feng.Server`。
*   `ctx.User()` 获取当前连接对应的 `feng.User`。
*   `ctx.Room()` 获取当前连接所在的 `feng.Room`。
*   `ctx.Get(key)` / `ctx.Set(key, value)` 在连接维度存取数据。
*   `ctx.GinContext()` 获取底层 `*gin.Context`。

#### 核心类型
根包对外暴露的稳定类型如下：

*   `feng.Server`：通过 `feng.NewServer(feng.ServerConfig)` 创建。
*   `feng.ServerConfig`：使用 `feng.NewDefaultServerConfig()` 构造，`Codec` 和 `Logger` 都可以替换。
*   `feng.ServerContext`：传给处理器和中间件。
*   `feng.User` / `feng.Room`：连接与房间抽象，提供 `ID`、`JoinRoom`、`LeaveRoom`、`Push`、`Request` 等方法。
*   `feng.Codec`：序列化接口。内置 `feng.NewJSONCodec()`（也可用 `NewJsonCodec`）和 `feng.NewProtoCodec()`。
*   `feng.Logger`：日志接口。内置 `feng.NewSlogLogger()`。

> 具体实现位于 `internal/` 目录，业务代码不要直接导入。

### 📱 客户端（Client）

#### Request（请求）
通过 `Request` 方法发起双向请求。

*   **参数**：
    1.  `context.Context`：请求上下文。
    2.  `Path`：对端的处理器路径。
    3.  `Data`：请求数据。
    4.  `Callback`：收到响应后被调用，参数形态与处理器一致，没有返回值。

#### RequestAsync（异步请求）
与 `Request` 相同，只是不需要传 `context.Context`。当调用方不关心是否同步等待完成时使用。

#### Push（推送）
通过 `Push` 方法发送单向消息。

*   **参数**：`Path`、`Data`。

#### 客户端处理器
客户端也可以通过 `Handle` 和 `Use` 注册供服务端调用的接口和中间件，用法与服务端一致，只是第一个参数必须是 `feng.ClientContext`。

#### 核心类型
*   `feng.Client`：通过 `feng.NewClient(feng.ClientConfig)` 创建。
*   `feng.ClientConfig`：使用 `feng.NewDefaultClientConfig()` 构造，可配置 `Codec`、`Logger`、`EnableTLS`、`DirectConnect` 等。
*   `feng.ClientContext`：传给客户端侧的处理器和中间件。

---

## 💡 实战示例

下面是一个更贴近真实业务的小例子：先登录校验，再读取数据。

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

// Handler：登录
func LoginHandler(ctx feng.ServerContext, req LoginReq) error {
	if req.Token == "" {
		return errors.New("empty token")
	}

	// 把登录后的用户 id 暂存到连接上下文，后续接口可以直接读取。
	ctx.Set("uuid", "demo-user-id")
	return nil
}

// Handler：获取账户信息
func GetAccountInfoHandler(ctx feng.ServerContext) (AccountResp, error) {
	uuid, ok := ctx.Get("uuid")
	if !ok {
		return AccountResp{}, errors.New("未登录")
	}

	return AccountResp{
		Name:     uuid.(string),
		IsTester: true,
	}, nil
}
```
