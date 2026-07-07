# Feng AI Integration Guide

This file is written for AI coding agents and IDE assistants. Use it when generating code that imports `github.com/zmhuanf/feng` without reading the source.

## What Feng Is

`feng` is a Go game communication framework built on Gin and WebSocket. It supports request/response calls, push messages, server-side handlers, client-side handlers, middleware, rooms, and users.

Only import the root package:

```go
import "github.com/zmhuanf/feng"
```

Do not import `github.com/zmhuanf/feng/internal/...` from application code.

## Current Public API Names

Use these names:

- `feng.Server`
- `feng.Client`
- `feng.ServerConfig`
- `feng.ClientConfig`
- `feng.ServerContext`
- `feng.ClientContext`
- `feng.User`
- `feng.Room`
- `feng.Codec`
- `feng.Logger`

## Minimal Server

```go
package main

import (
	"context"
	"log/slog"

	"github.com/zmhuanf/feng"
)

func main() {
	config := feng.NewDefaultServerConfig()
	config.Addr = "0.0.0.0"
	config.Port = 22100

	server := feng.NewServer(config)
	if err := server.Handle("/echo", func(ctx feng.ServerContext, msg string) (string, error) {
		return msg, nil
	}); err != nil {
		slog.Error("register handler failed", "error", err)
		return
	}

	if err := server.ListenAndServe(context.Background()); err != nil {
		slog.Error("server stopped", "error", err)
	}
}
```

## Minimal Client

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

	if err := client.Request(context.Background(), "/echo", "hello", func(ctx feng.ClientContext, resp string) {
		fmt.Println(resp)
	}); err != nil {
		slog.Error("request failed", "error", err)
	}
}
```

## Server API

Create a server:

```go
config := feng.NewDefaultServerConfig()
server := feng.NewServer(config)
```

Important methods:

- `Config() *feng.ServerConfig`
- `Handle(route string, handler any) error`
- `Use(route string, middleware any) error`
- `ListenAndServe(ctx context.Context) error`
- `Stop(ctx context.Context) error`
- `Room(id string) (feng.Room, error)`
- `Rooms() []feng.Room`
- `RoomsByPage(page int) []feng.Room`
- `User(id string) (feng.User, error)`
- `Users() []feng.User`
- `UsersByPage(page int) []feng.User`
- `Gin() *gin.Engine`

## Client API

Create a client:

```go
config := feng.NewDefaultClientConfig()
client := feng.NewClient(config)
```

Important methods:

- `Config() *feng.ClientConfig`
- `Handle(route string, handler any) error`
- `Use(route string, middleware any) error`
- `Connect(ctx context.Context) error`
- `Push(route string, data any) error`
- `RequestAsync(route string, data any, callback any) error`
- `Request(ctx context.Context, route string, data any, callback any) error`
- `Close() error`

## Handler Signatures

Server handler first argument must be `feng.ServerContext`:

```go
func(ctx feng.ServerContext) error
func(ctx feng.ServerContext, req string) error
func(ctx feng.ServerContext, req LoginReq) (LoginResp, error)
```

Client handler first argument must be `feng.ClientContext`:

```go
func(ctx feng.ClientContext) error
func(ctx feng.ClientContext, msg string) error
func(ctx feng.ClientContext, req NoticeReq) (NoticeResp, error)
```

Supported payload argument types:

- `string`
- `[]byte`
- struct
- slice
- map
- bool
- numeric types

Return rules:

- Return `error` when no response body is needed.
- Return `(response, error)` when a response body is needed.
- A handler may also return nothing, but prefer returning `error` for explicit failure handling.

## Middleware Signatures

Server middleware first argument must be `feng.ServerContext`:

```go
server.Use("/api", func(ctx feng.ServerContext, token string) error {
	return nil
})
```

Client middleware first argument must be `feng.ClientContext`:

```go
client.Use("/notice", func(ctx feng.ClientContext) error {
	return nil
})
```

Middleware route matching is prefix-based. For example, middleware registered on `/api` applies to `/api/login` and `/api/profile`.

Middleware should return `error`. A non-nil error stops the actual route handler and sends a failed response.

## Request And Push Semantics

Use `Request` for bidirectional calls:

```go
err := client.Request(context.Background(), "/profile", ProfileReq{ID: "1"}, func(ctx feng.ClientContext, resp ProfileResp) {
	// handle response
})
```

Use `Push` for one-way messages:

```go
err := client.Push("/heartbeat", map[string]any{"ts": time.Now().Unix()})
```

Use `RequestAsync` only when the caller does not need to wait for completion directly:

```go
err := client.RequestAsync("/profile", ProfileReq{ID: "1"}, func(ctx feng.ClientContext, resp ProfileResp) {
	// handle response later
})
```

## Context Usage

Server context:

```go
func Handler(ctx feng.ServerContext) error {
	user := ctx.User()
	room := ctx.Room()
	server := ctx.Server()
	ctx.Set("key", "value")
	value, ok := ctx.Get("key")
	_ = user
	_ = room
	_ = server
	_ = value
	_ = ok
	return nil
}
```

Client context:

```go
func Handler(ctx feng.ClientContext) error {
	client := ctx.Client()
	_ = client
	return nil
}
```

## User And Room

`feng.User` methods:

- `ID() string`
- `Room() feng.Room`
- `JoinRoom(room feng.Room) error`
- `CreateAndJoinRoom() error`
- `LeaveRoom() error`
- `Context() feng.ServerContext`
- `ExtraData(key string) (any, bool)`
- `SetExtraData(key string, value any)`
- `Page() int`
- `Push(route string, data any) error`
- `Request(ctx context.Context, route string, data any, callback any) error`
- `RequestAsync(route string, data any, callback any) error`

`feng.Room` methods:

- `ID() string`
- `RemoveUser(user feng.User) error`
- `User(id string) (feng.User, error)`
- `Users() []feng.User`
- `UserCount() int`
- `Page() int`

## Config Defaults

Server defaults:

```go
config := feng.NewDefaultServerConfig()
config.Addr = "0.0.0.0"
config.Port = 22100
```

Client defaults:

```go
config := feng.NewDefaultClientConfig()
config.Addr = "127.0.0.1"
config.Port = 22100
config.DirectConnect = true
```

Use `config.Codec` to change serialization and `config.Logger` to change logging.

## Common Mistakes To Avoid

- Do not import `internal/...` packages.
- Do not hold business state in package globals when `ctx.Set/Get` or `User.SetExtraData` is more appropriate.
- Do not assume middleware exact-matches paths; it uses prefix matching.
- Do not return a response value without also returning `error`; use `(resp, error)`.
