package feng

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type IClient interface {
	// 添加路由处理器
	AddHandler(string, func(IContext, any) any)
	// 添加中间件
	AddMiddleware(string, func(IContext, any) error)
	// 连接服务器
	Connect() error
	// 推送消息
	Push(string, any) error
	// 异步请求
	RequestAsync(string, any, func(IContext, any)) error
	// 同步请求
	Request(string, any, func(IContext, any), time.Duration) error
}

type client struct {
	// 配置
	config *clientConfig
	// 连接
	conn *websocket.Conn
	// 路由
	route map[string]func(IContext, any) any
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
}

func NewClient(config *clientConfig) IClient {
	return &client{
		config:    config,
		route:     make(map[string]func(IContext, any) any),
		responses: make(map[string]*response),
	}
}

func (c *client) send(res *requestTmp) error {
	if c.conn == nil {
		return errors.New("client not connected")
	}
	resByte, err := c.config.Codec.Marshal(res)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, resByte)
}

func (c *client) AddHandler(route string, handler func(IContext, any) any) {
	c.routeLock.Lock()
	defer c.routeLock.Unlock()
	c.route[route] = handler
}

func (c *client) AddMiddleware(route string, fn func(IContext, any) error) {
	c.middlewaresLock.Lock()
	defer c.middlewaresLock.Unlock()
	c.middlewares = append(c.middlewares, &middleware{
		route: route,
		fn:    fn,
	})
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
	proto := "ws"
	if c.config.EnableTLS {
		proto = "wss"
	}
	// url
	url := fmt.Sprintf("%s://%s:%d", proto, c.config.Addr, c.config.Port)
	// 连接
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *client) Push(route string, data any) error {
	err := c.send(&requestTmp{
		ID:    uuid.New().String(),
		Route: route,
		Type:  requestTypePush,
		Data:  data,
	})
	return err
}

func (c *client) RequestAsync(route string, data any, handler func(IContext, any)) error {
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
	err := c.send(&requestTmp{
		ID:    id,
		Route: route,
		Type:  requestTypeRequest,
		Data:  data,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *client) Request(route string, data any, handler func(IContext, any), timeout time.Duration) error {
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
	err := c.send(&requestTmp{
		ID:    id,
		Route: route,
		Type:  requestTypeRequest,
		Data:  data,
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
	case <-time.After(timeout):
		c.responsesLock.Lock()
		delete(c.responses, id)
		c.responsesLock.Unlock()
		return errors.New("request timeout")
	}
}
