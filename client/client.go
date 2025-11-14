package client

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type IClient interface {
	// 添加路由处理器
	AddHandler(string, any) error
	// 添加中间件
	AddMiddleware(string, any) error
	// 连接服务器
	Connect() error
	// 推送消息
	Push(string, any) error
	// 异步请求
	RequestAsync(string, any, any) error
	// 同步请求
	Request(context.Context, string, any, any) error
	// 获取配置
	GetConfig() *Config
	// 关闭连接
	Close() error
}

type client struct {
	// 配置
	config *Config
	// 连接
	conn *websocket.Conn
	// 路由
	route map[string]any
	// 路由锁
	routeLock sync.RWMutex
	// 中间层
	middlewares []*middleware
	// 中间层锁
	middlewaresLock sync.RWMutex
	// 响应
	responses map[string]*response
	// 响应锁
	responsesLock sync.RWMutex
	// 关闭控制
	ctx        context.Context
	cancel     context.CancelFunc
	closed     bool
	closedLock sync.RWMutex
	// 系统通信客户端
	sysClient *client
}

type middleware struct {
	route string
	fn    any
}

type response struct {
	fn any
	ch chan chanData
}

type chanData struct {
	Success bool
	Data    any
}

func NewClient(config *Config) IClient {
	ctx, cancel := context.WithCancel(context.Background())
	var sysClient *client
	if config.mode == tModeClient {
		sysClient = &client{
			config:      config,
			route:       make(map[string]any),
			middlewares: make([]*middleware, 0),
			responses:   make(map[string]*response),
			ctx:         ctx,
			cancel:      cancel,
			closed:      false,
		}
	}
	return &client{
		config:      config,
		route:       make(map[string]any),
		middlewares: make([]*middleware, 0),
		responses:   make(map[string]*response),
		ctx:         ctx,
		cancel:      cancel,
		closed:      false,
		sysClient:   sysClient,
	}
}

func (c *client) send(res *request) error {
	if c.isClosed() {
		return errors.New("client is closed")
	}
	if c.conn == nil {
		return errors.New("client not connected")
	}
	resByte, err := c.config.Codec.Marshal(res)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, resByte)
}

func (c *client) AddHandler(route string, fn any) error {
	// 检查函数签名
	err := checkFuncType(fn)
	if err != nil {
		return err
	}
	c.routeLock.Lock()
	defer c.routeLock.Unlock()
	// 保存
	c.route[route] = fn
	return nil
}

func (c *client) AddMiddleware(route string, fn any) error {
	// 检查函数签名
	err := checkFuncType(fn)
	if err != nil {
		return err
	}
	c.middlewaresLock.Lock()
	defer c.middlewaresLock.Unlock()
	// 保存
	c.middlewares = append(c.middlewares, &middleware{
		route: route,
		fn:    fn,
	})
	return nil
}

func (c *client) getRoute(route string) (any, bool) {
	c.routeLock.RLock()
	defer c.routeLock.RUnlock()
	handler, ok := c.route[route]
	return handler, ok
}

func (c *client) getResponse(id string) (*response, bool) {
	c.responsesLock.RLock()
	defer c.responsesLock.RUnlock()
	res, ok := c.responses[id]
	return res, ok
}

func (c *client) addResponse(id string, resp *response) {
	c.responsesLock.Lock()
	defer c.responsesLock.Unlock()
	c.responses[id] = resp
}

func (c *client) deleteResponse(id string) {
	c.responsesLock.Lock()
	defer c.responsesLock.Unlock()
	delete(c.responses, id)
}

func (c *client) Connect() error {
	// 协议
	addr := fmt.Sprintf("%s:%d", c.config.Addr, c.config.Port)
	return c.connect(addr, true)
}

func (c *client) connect(addr string, needNew bool) error {
	if c.isClosed() {
		return errors.New("client is closed")
	}
	proto := "ws"
	if c.config.EnableTLS {
		proto = "wss"
	}
	// 客户端模式
	if c.config.mode == tModeClient {
		// 1.客户端模式应该先连接系统通信，然后获取最低负载服务器地址
		// 2.如果最低负载地址和当前地址相同，直接连接用户通信
		// 3.如果不同，再次连接系统通信，并告知自己是转接过来的，服务器将不再给出新的地址
		// 4.然后进入2的逻辑
		err := c.connectSys(fmt.Sprintf("%s://%s/system", proto, addr))
		if err != nil {
			return err
		}
		var serverAddr string
		err = c.sysClient.Request(
			context.Background(),
			"/get_low_load_server_addr",
			needNew,
			func(ctx IContext, data string) {
				serverAddr = data
			},
		)
		if err != nil {
			return err
		}
		// 相同地址
		if serverAddr == "" {
			return c.connectUser(fmt.Sprintf("%s://%s/game", proto, addr))
		}
		// 不同地址
		return c.connect(serverAddr, false)
	}
	// 服务器模式
	if c.config.mode == tModeServer {
		// 1.服务器模式直接连接用户通信就好
		return c.connectUser(fmt.Sprintf("%s://%s/game", proto, addr))
	}
	return nil
}

func (c *client) connectSys(url string) error {
	if c.sysClient.isClosed() {
		return errors.New("sys client is closed")
	}
	// 连接
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	c.sysClient.conn = conn
	go handle(c.sysClient)
	return nil
}

func (c *client) connectUser(url string) error {
	if c.isClosed() {
		return errors.New("client is closed")
	}
	// 连接
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	go handle(c)
	return nil
}

func (c *client) Push(route string, data any) error {
	if c.isClosed() {
		return errors.New("client is closed")
	}
	dataBytes, err := c.config.Codec.Marshal(data)
	if err != nil {
		return err
	}
	err = c.send(&request{
		ID:    uuid.New().String(),
		Route: route,
		Type:  requestTypePush,
		Data:  string(dataBytes),
	})
	return err
}

func (c *client) RequestAsync(route string, data any, handler any) error {
	// 检查函数签名
	err := checkFuncType(handler)
	if err != nil {
		return err
	}
	if c.isClosed() {
		return errors.New("client is closed")
	}
	id := uuid.New().String()
	ch := make(chan chanData)
	c.addResponse(id, &response{
		fn: handler,
		ch: ch,
	})
	// 配置一个额外的删除，防止内存泄漏
	go func() {
		<-time.After(c.config.Timeout)
		c.deleteResponse(id)
	}()
	dataBytes, err := c.config.Codec.Marshal(data)
	if err != nil {
		return err
	}
	err = c.send(&request{
		ID:    id,
		Route: route,
		Type:  requestTypeRequest,
		Data:  string(dataBytes),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *client) Request(ctx context.Context, route string, data any, handler any) error {
	// 检查函数签名
	err := checkFuncType(handler)
	if err != nil {
		return err
	}
	if c.isClosed() {
		return errors.New("client is closed")
	}
	id := uuid.New().String()
	ch := make(chan chanData)
	c.addResponse(id, &response{
		fn: handler,
		ch: ch,
	})
	// 配置一个额外的删除，防止内存泄漏
	go func() {
		<-time.After(c.config.Timeout)
		c.deleteResponse(id)
	}()
	dataBytes, err := c.config.Codec.Marshal(data)
	if err != nil {
		return err
	}
	err = c.send(&request{
		ID:    id,
		Route: route,
		Type:  requestTypeRequest,
		Data:  string(dataBytes),
	})
	if err != nil {
		return err
	}
	select {
	case data := <-ch:
		if data.Success {
			return nil
		}
		return fmt.Errorf("%v", data.Data)
	case <-ctx.Done():
		c.responsesLock.Lock()
		delete(c.responses, id)
		c.responsesLock.Unlock()
		return ctx.Err()
	}
}

func (c *client) GetConfig() *Config {
	return c.config
}

func (c *client) Close() error {
	c.closedLock.Lock()
	defer c.closedLock.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true
	// 退出读取协程
	if c.cancel != nil {
		c.cancel()
	}
	// 关闭WebSocket连接
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			return err
		}
		c.conn = nil
	}
	// 清理响应通道，防止内存泄漏
	c.responsesLock.Lock()
	for id, resp := range c.responses {
		if resp.ch != nil {
			close(resp.ch)
		}
		delete(c.responses, id)
	}
	c.responsesLock.Unlock()

	return nil
}

func (c *client) isClosed() bool {
	c.closedLock.RLock()
	defer c.closedLock.RUnlock()
	return c.closed
}
