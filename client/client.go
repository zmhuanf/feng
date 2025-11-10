package client

import (
	"errors"
	"fmt"
	"reflect"
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
	Request(string, any, any, time.Duration) error
	// 获取配置
	GetConfig() *Config
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
	return &client{
		config:      config,
		route:       make(map[string]any),
		middlewares: make([]*middleware, 0),
		responses:   make(map[string]*response),
	}
}

func (c *client) send(res *request) error {
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
	c.routeLock.Lock()
	defer c.routeLock.Unlock()
	// 检查函数签名
	ft := reflect.TypeOf(fn)
	if ft.Kind() != reflect.Func {
		return errors.New("f must be func")
	}
	if ft.NumIn() != 2 {
		return errors.New("func must have 2 args")
	}
	if ft.In(0) != reflect.TypeOf((*IContext)(nil)).Elem() {
		return errors.New("first arg must be IContext")
	}
	// 保存
	c.route[route] = fn
	return nil
}

func (c *client) AddMiddleware(route string, fn any) error {
	c.middlewaresLock.Lock()
	defer c.middlewaresLock.Unlock()
	// 检查函数签名
	ft := reflect.TypeOf(fn)
	if ft.Kind() != reflect.Func {
		return errors.New("f must be func")
	}
	if ft.NumIn() != 2 {
		return errors.New("func must have 2 args")
	}
	if ft.In(0) != reflect.TypeOf((*IContext)(nil)).Elem() {
		return errors.New("first arg must be IContext")
	}
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
	go handle(c)
	return nil
}

func (c *client) Push(route string, data any) error {
	err := c.send(&request{
		ID:    uuid.New().String(),
		Route: route,
		Type:  requestTypePush,
		Data:  data,
	})
	return err
}

func (c *client) RequestAsync(route string, data any, handler any) error {
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
	err := c.send(&request{
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

func (c *client) Request(route string, data any, handler any, timeout time.Duration) error {
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
	err := c.send(&request{
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

func (c *client) GetConfig() *Config {
	return c.config
}
