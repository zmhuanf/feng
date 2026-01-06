# 🍃 Feng

`feng` 是一个基于 **Gin** 和 **Websocket** 的游戏通信框架。

其设计哲学不在于追求极致的性能，而是致力于提供一种**易用、简洁**的通信解决方案，帮助开发者快速构建游戏服务器。

> ⚠️ **注意**：本项目仍在积极开发中。在发布第一个正式版本之前，API 可能会发生重大变化。

---

## 🎮 多平台支持

除了 Go 客户端外，`feng` 还致力于支持所有主流游戏引擎：

- **Cocos**: [feng-cocos](https://github.com/zmhuanf/feng-cocos)
- **Unity**: [feng-unity](https://github.com/zmhuanf/feng-unity)

---

## 🚀 快速开始

使用 `feng` 创建一个回显（Echo）服务十分简单。

### 服务端

```go
server := feng.NewServer(feng.NewDefaultServerConfig())

// 注册简单的回显接口
server.AddHandler("/echo", func(ctx feng.IServerContext, msg string) (string, error) {
    return msg, nil
})

server.Start()
```

### 客户端

调用接口也非常直观：

```go
client := feng.NewClient(feng.NewDefaultClientConfig())
client.Connect()
defer client.Close()

// 发送请求并处理回调
client.Request(context.TODO(), "/echo", "hello, world!", func(ctx feng.IClientContext, msg string) {
    fmt.Println(msg)
})
```
*(注：上述示例为了简洁忽略了错误处理)*

---

## 📖 基本概念

### 🖥️ 服务器 (Server)

#### 添加处理器 (AddHandler)
服务器通过 `AddHandler` 方法注册业务逻辑。

*   **路径 (Path)**: 第一个参数。可以是任意字符串，不强制要求以 `/` 开头。
*   **处理器函数 (Handler)**: 第二个参数。
    *   **参数**:
        1.  `ctx feng.IServerContext` (必须)
        2.  `Request Data` (可选): 可以是 `string`, `[]byte` 或任意可被 `ICodec` 反序列化的结构体。
    *   **返回值**:
        1.  `Response Data` (可选): 当有两个返回值时作为第一个返回。可以是 `string`, `[]byte` 或任意可被 `ICodec` 序列化的结构体。
        2.  `error` (必须): 表示处理结果。如果不返回数据，它是唯一的返回值；如果返回数据，它是第二个返回值。

#### 中间件 (Middleware)
通过 `AddMiddleware` 添加。
*   **作用范围**: 第一个参数为路径前缀。中间件会按添加顺序作用于所有匹配该前缀的处理器。
*   **函数签名**: 参数与处理器函数相同，但 **返回值只能为 `error` 类型**。
*   **拦截**: 返回非 `nil` 的 error 会阻止后续处理器的执行。

### 📱 客户端 (Client)

#### 发送请求 (Request)
通过 `Request` 方法发送双向请求。
*   **参数**:
    1.  `Context`
    2.  `Path`: 处理器路径
    3.  `Data`: 请求数据
    4.  `Callback`: 回调函数（参数与处理器函数相同，无返回值），在收到响应后调用。

#### 发送推送 (Push)
通过 `Push` 方法发送单向消息。
*   **参数**: Path, Data。

#### 客户端处理器
客户端也可以通过 `AddHandler` 和 `AddMiddleware` 注册供服务器调用的接口，用法与服务端一致，只是上下文接口为 `IClientContext`。

---

## 💡 实战示例

下面是一个更接近生产环境的示例，包含简单的登录验证和数据获取。

```go
func main() {
	opt := feng.NewDefaultServerConfig()
	opt.Addr = "0.0.0.0"
	opt.Port = 22002
	server := feng.NewServer(opt)

	// 注册路由
	server.AddHandler("/login", LoginHandler)
	server.AddHandler("/get_account_info", GetAccountInfoHandler)

	// 启动服务
	err := server.Start()
	if err != nil {
		slog.Error("GM服务启动失败", "错误", err)
		return
	}
}

// 定义请求结构体
type LoginReq struct {
	Token string `json:"token"`
}

// Handler: 登录
func LoginHandler(ctx feng.IServerContext, data LoginReq) error {
	configs := config.GetConfig()
	
    // 验证 Token
	uuid, err := tool.ValidateToken(data.Token, configs.JWTKey)
	if err != nil {
		slog.Error("Token验证失败", "错误", err)
		return err
	}
    
    // 将 UUID 存入上下文供后续使用
	ctx.Set("uuid", uuid)
	return nil
}

// Handler: 获取账户信息
func GetAccountInfoHandler(ctx feng.IServerContext) (map[string]any, error) {
	uuid, ok := ctx.Get("uuid")
	if !ok {
		slog.Error("uuid 不存在")
		return nil, errors.New("非法请求")
	}

    // 模拟从数据库获取用户
	user, err := mongodb.GetUser(context.Background(), uuid.(string))
	if err != nil {
		slog.Error("GetAccountInfoHandler GetUser 失败", "uuid", uuid, "err", err)
		return nil, errors.New("服务器内部错误")
	}

	return map[string]any{
		"name":      user.Name,
		"is_tester": user.IsTester,
	}, nil
}
```
        