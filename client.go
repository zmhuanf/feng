package feng

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
	GetConfig() *clientConfig
	// 关闭连接
	Close()
}

type client struct {
	// 配置
	config *clientConfig
	// 连接
	conn    *websocket.Conn
	connSys *websocket.Conn
	// 路由
	route    map[string]any
	routeSys map[string]any
	// 路由锁
	routeLock    sync.RWMutex
	routeSysLock sync.RWMutex
	// 中间层
	middlewares    []*middleware
	middlewaresSys []*middleware
	// 中间层锁
	middlewaresLock    sync.RWMutex
	middlewaresSysLock sync.RWMutex
	// 响应
	responses    map[string]*response
	responsesSys map[string]*response
	// 响应锁
	responsesLock    sync.RWMutex
	responsesSysLock sync.RWMutex
	// 关闭控制
	ctx           context.Context
	cancel        context.CancelFunc
	closed        bool
	closedLock    sync.RWMutex
	ctxSys        context.Context
	cancelSys     context.CancelFunc
	closedSys     bool
	closedSysLock sync.RWMutex
}

func NewClient(config *clientConfig) IClient {
	if config == nil {
		config = NewDefaultClientConfig()
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctxSys, cancelSys := context.WithCancel(context.Background())
	return &client{
		config:         config,
		route:          make(map[string]any),
		routeSys:       make(map[string]any),
		middlewares:    make([]*middleware, 0),
		middlewaresSys: make([]*middleware, 0),
		responses:      make(map[string]*response),
		responsesSys:   make(map[string]*response),
		ctx:            ctx,
		ctxSys:         ctxSys,
		cancel:         cancel,
		cancelSys:      cancelSys,
		closed:         false,
		closedSys:      false,
	}
}

func (c *client) send(res *message, isSys bool) error {
	conn := c.conn
	if isSys {
		conn = c.connSys
	}

	if c.isClosed(isSys) {
		return errors.New("client is closed")
	}
	if conn == nil {
		return errors.New("client not connected")
	}
	resByte, err := c.config.Codec.Marshal(res)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, resByte)
}

func (c *client) AddHandler(route string, fn any) error {
	return c.addHandler(route, fn, false)
}

func (c *client) addHandler(route string, fn any, isSys bool) error {
	// 检查函数签名
	err := checkFuncType(fn, false)
	if err != nil {
		return err
	}

	lock := &c.routeLock
	routeMap := c.route
	if isSys {
		lock = &c.routeSysLock
		routeMap = c.routeSys
	}
	lock.Lock()
	defer lock.Unlock()
	// 保存
	routeMap[route] = fn
	return nil
}

func (c *client) AddMiddleware(route string, fn any) error {
	return c.addMiddleware(route, fn, false)
}

func (c *client) addMiddleware(route string, fn any, isSys bool) error {
	// 检查函数签名
	err := checkFuncType(fn, false)
	if err != nil {
		return err
	}
	lock := &c.middlewaresLock
	middlewares := &c.middlewares
	if isSys {
		lock = &c.middlewaresSysLock
		middlewares = &c.middlewaresSys
	}
	lock.Lock()
	defer lock.Unlock()
	// 保存
	*middlewares = append(*middlewares, &middleware{
		route: route,
		fn:    fn,
	})
	return nil
}

func (c *client) getRoute(route string, isSys bool) (any, bool) {
	lock := &c.routeLock
	routeMap := c.route
	if isSys {
		lock = &c.routeSysLock
		routeMap = c.routeSys
	}
	lock.RLock()
	defer lock.RUnlock()
	handler, ok := routeMap[route]
	return handler, ok
}

func (c *client) getResponse(id string, isSys bool) (*response, bool) {
	lock := &c.responsesLock
	responses := c.responses
	if isSys {
		lock = &c.responsesSysLock
		responses = c.responsesSys
	}
	lock.RLock()
	defer lock.RUnlock()
	res, ok := responses[id]
	return res, ok
}

func (c *client) addResponse(id string, resp *response, isSys bool) {
	lock := &c.responsesLock
	responses := c.responses
	if isSys {
		lock = &c.responsesSysLock
		responses = c.responsesSys
	}
	lock.Lock()
	defer lock.Unlock()
	responses[id] = resp
}

func (c *client) deleteResponse(id string, isSys bool) {
	lock := &c.responsesLock
	responses := c.responses
	if isSys {
		lock = &c.responsesSysLock
		responses = c.responsesSys
	}
	lock.Lock()
	defer lock.Unlock()
	delete(responses, id)
}

func (c *client) Connect() error {
	// 协议
	addr := fmt.Sprintf("%s:%d", c.config.Addr, c.config.Port)
	needNew := !c.GetConfig().DirectConnect
	if c.config.mode == tModeServer {
		needNew = false
	}
	return c.connect(addr, needNew)
}

func (c *client) connect(addr string, needNew bool) error {
	if c.isClosed(false) || c.isClosed(true) {
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
		err = c.request(
			context.Background(),
			"/get_low_load_server_addr",
			needNew,
			func(ctx IClientContext, data string) {
				serverAddr = data
			},
			true,
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
		return c.connectSys(fmt.Sprintf("%s://%s/system", proto, addr))
	}
	return nil
}

func (c *client) connectSys(url string) error {
	if c.isClosed(true) {
		return errors.New("sys client is closed")
	}
	// 连接
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	c.connSys = conn
	go clientHandle(c, true)
	return nil
}

func (c *client) connectUser(url string) error {
	if c.isClosed(false) {
		return errors.New("client is closed")
	}
	// 连接
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	go clientHandle(c, false)
	return nil
}

func (c *client) Push(route string, data any) error {
	return c.push(route, data, false)
}

func (c *client) push(route string, data any, isSys bool) error {
	if c.isClosed(isSys) {
		return errors.New("client is closed")
	}
	dataBytes, err := c.config.Codec.Marshal(data)
	if err != nil {
		return err
	}
	err = c.send(&message{
		ID:    uuid.New().String(),
		Route: route,
		Type:  messageTypePush,
		Data:  string(dataBytes),
	}, isSys)
	return err
}

func (c *client) RequestAsync(route string, data any, callback any) error {
	return c.requestAsync(route, data, callback, false)
}

func (c *client) requestAsync(route string, data any, callback any, isSys bool) error {
	// 检查函数签名
	err := checkFuncType(callback, false)
	if err != nil {
		return err
	}
	if c.isClosed(isSys) {
		return errors.New("client is closed")
	}
	id := uuid.New().String()
	ch := make(chan chanData)
	c.addResponse(id, &response{
		fn: callback,
		ch: ch,
	}, isSys)
	// 配置一个额外的删除，防止内存泄漏
	go func() {
		timer := time.NewTimer(c.config.Timeout)
		defer timer.Stop()

		select {
		case <-timer.C:
			c.deleteResponse(id, isSys)
		case <-ch:
		}
	}()
	dataBytes, err := c.config.Codec.Marshal(data)
	if err != nil {
		return err
	}
	err = c.send(&message{
		ID:    id,
		Route: route,
		Type:  messageTypeRequest,
		Data:  string(dataBytes),
	}, isSys)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) Request(ctx context.Context, route string, data any, callback any) error {
	return c.request(ctx, route, data, callback, false)
}

func (c *client) request(ctx context.Context, route string, data any, callback any, isSys bool) error {
	// 检查函数签名
	err := checkFuncType(callback, false)
	if err != nil {
		return err
	}
	if c.isClosed(isSys) {
		return errors.New("client is closed")
	}
	id := uuid.New().String()
	ch := make(chan chanData)
	c.addResponse(id, &response{
		fn: callback,
		ch: ch,
	}, isSys)
	// 超时删除
	timer := time.NewTimer(c.config.Timeout)
	defer timer.Stop()
	dataBytes, err := c.config.Codec.Marshal(data)
	if err != nil {
		c.deleteResponse(id, isSys)
		return err
	}
	err = c.send(&message{
		ID:    id,
		Route: route,
		Type:  messageTypeRequest,
		Data:  string(dataBytes),
	}, isSys)
	if err != nil {
		c.deleteResponse(id, isSys)
		return err
	}
	select {
	case data := <-ch:
		if data.Success {
			return nil
		}
		return fmt.Errorf("%v", data.Data)
	case <-ctx.Done():
		c.deleteResponse(id, isSys)
		close(ch)
		return ctx.Err()
	case <-timer.C:
		c.deleteResponse(id, isSys)
		close(ch)
		return fmt.Errorf("request timeout")
	}
}

func (c *client) GetConfig() *clientConfig {
	return c.config
}

func (c *client) Close() {
	c.closeUser()
	c.closeSys()
}

func (c *client) closeSys() {
	c.closedSysLock.Lock()
	if !c.closedSys {
		c.closedSys = true
		// 退出读取协程
		if c.cancelSys != nil {
			c.cancelSys()
		}
		// 关闭WebSocket连接
		if c.connSys != nil {
			c.connSys.Close()
			c.connSys = nil
		}
		// 清理响应通道，防止内存泄漏
		c.responsesSysLock.Lock()
		for id, resp := range c.responsesSys {
			if resp.ch != nil {
				close(resp.ch)
			}
			delete(c.responsesSys, id)
		}
		c.responsesSysLock.Unlock()
	}
	c.closedSysLock.Unlock()
}

func (c *client) closeUser() {
	c.closedLock.Lock()
	if !c.closed {
		c.closed = true
		// 退出读取协程
		if c.cancel != nil {
			c.cancel()
		}
		// 关闭WebSocket连接
		if c.conn != nil {
			c.conn.Close()
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
	}
	c.closedLock.Unlock()
}

func (c *client) isClosed(isSys bool) bool {
	lock := &c.closedLock
	closed := &c.closed
	if isSys {
		lock = &c.closedSysLock
		closed = &c.closedSys
	}

	lock.RLock()
	defer lock.RUnlock()
	return *closed
}
