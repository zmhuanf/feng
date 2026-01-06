feng是一个基于Gin和Websocket的游戏通信框架，其目的不在于性能的高效，而是提供一种易用的通信解决方案，方便开发者快速构建游戏服务器。
注意：本项目仍在开发中，在发布第一个正式版本之前，API可能会发生重大变化。
除了go客户端外，feng还支持Cosos（https://github.com/zmhuanf/feng-cocos）和Unity（https://github.com/zmhuanf/feng-unity）的客户端，未来会支持所有主流游戏框架。
使用示例（注意示例中忽略了错误处理）：
创建一个回显接口十分的简单：
    server := feng.NewServer(feng.NewDefaultServerConfig())
	server.AddHandler("/echo", func(ctx feng.IServerContext, msg string) (string, error) {
		return msg, nil
	})
	server.Start()
而要调用这个接口也非常简单：
    client := feng.NewClient(feng.NewDefaultClientConfig())
	client.Connect()
	defer client.Close()
	client.Request(context.TODO(), "/echo", "hello, word!", func(ctx feng.IClientContext, msg string) {
		fmt.Println(msg)
	})
下面的会讲述一些基本概念：
服务器：
  添加处理器：
    服务器通过AddHandler方法添加处理器，处理器是一个函数，接收客户端发送的消息并返回响应消息。
    AddHandler方法接受两个参数，第一个参数是处理器的路径，第二个参数是处理器函数。
    路径可以是任意字符串，不强制要求以斜杠开头。
    处理器函数可以有1个或2个参数，第一个参数必须是IServerContext接口，第二个参数为客户端发送的数据，可以是string、[]byte或能够被ICodec反序列化的任意类型
    处理器函数可以有1个或2个返回值，当只有一个返回值时，必须为error类型，表示处理器执行结果；当有两个返回值时，第一个返回值为响应数据，可以是string、[]byte或能够被ICodec序列化的任意类型，第二个返回值必须为error类型。
  中间件：
    服务器可以通过AddMiddleware方法添加中间件，第一个参数为路径，其会按照添加顺序作用于以该路径开头的所有处理器，第二个参数为中间件函数，参数与处理器函数相同，但返回值只能为error类型，返回非nil值会阻止后续处理器的执行。
客户端：
  发送请求：
    客户端通过Request方法发送请求，Request方法接受4个参数，第一个参数为上下文，第二个参数为处理器路径，第三个参数为请求数据，可以是string、[]byte或能够被ICodec序列化的任意类型。第四个参数为回调函数，参数与处理器函数相同，但没有返回值。
    回调函数会在收到服务器响应后被调用。
    还可以通过Push方法发送单向消息，Push方法接受2个参数，第一个参数为处理器路径，第二个参数为请求数据。
  处理器和中间件：
    客户端也可以通过AddHandler和AddMiddleware方法添加处理器和中间件供服务器调用，其用法与服务器端相同，但处理器函数的第一个参数为IClientContext接口。
下面时一个更实用一点的例子：
func main() {
	opt := feng.NewDefaultServerConfig()
	opt.Addr = "0.0.0.0"
	opt.Port = 22002
	server := feng.NewServer(opt)
	// 登录
	server.AddHandler("/login", LoginHandler)
	// 获取账户信息
	server.AddHandler("/get_account_info", GetAccountInfoHandler)
	// 启动服务
	err := server.Start()
	if err != nil {
		slog.Error("GM服务启动失败", "错误", err)
		return
	}
}

type LoginReq struct {
	Token string `json:"token"`
}

// 登录
func LoginHandler(ctx feng.IServerContext, data LoginReq) error {
	configs := config.GetConfig()
	uuid, err := tool.ValidateToken(data.Token, configs.JWTKey)
	if err != nil {
		slog.Error("Token验证失败", "错误", err)
		return err
	}
	ctx.Set("uuid", uuid)
	return nil
}

// 获取账户信息
func GetAccountInfoHandler(ctx feng.IServerContext) (map[string]any, error) {
	uuid, ok := ctx.Get("uuid")
	if !ok {
		slog.Error("SetHeroLevelHandler uuid 不存在")
		return nil, errors.New("非法请求")
	}
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
        