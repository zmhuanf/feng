package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/google/uuid"
	"github.com/zmhuanf/feng/internal/core"
	"github.com/zmhuanf/feng/internal/pending"
	"github.com/zmhuanf/feng/internal/protocol"
	"github.com/zmhuanf/feng/internal/router"
	"github.com/zmhuanf/feng/internal/transport"
)

type channel struct {
	conn    *transport.Conn
	router  *router.Router
	pending *pending.Store
	cancel  context.CancelFunc
	closed  bool
	lock    sync.RWMutex
}

type Client struct {
	config core.ClientConfig
	user   *channel
	system *channel
}

func New(config core.ClientConfig) core.Client {
	config = core.NormalizeClientConfig(config)
	return &Client{
		config: config,
		user:   newChannel(config),
		system: newChannel(config),
	}
}

func newChannel(config core.ClientConfig) *channel {
	return &channel{
		router:  router.New(reflect.TypeFor[core.ClientContext]()),
		pending: pending.New(config.Timeout),
	}
}

func (c *Client) Config() *core.ClientConfig { return &c.config }

func (c *Client) Handle(route string, handler any) error {
	return c.user.router.Handle(route, handler)
}

func (c *Client) Use(route string, middleware any) error {
	return c.user.router.Use(route, middleware)
}

func (c *Client) Connect(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", c.config.Addr, c.config.Port)
	needNew := !c.config.DirectConnect
	if c.config.Mode == core.ModeServer {
		needNew = false
	}
	return c.connect(ctx, addr, needNew)
}

func (c *Client) connect(ctx context.Context, addr string, needNew bool) error {
	proto := "ws"
	if c.config.EnableTLS {
		proto = "wss"
	}
	if c.config.Mode == core.ModeClient {
		if err := c.connectSystem(ctx, fmt.Sprintf("%s://%s/system", proto, addr)); err != nil {
			return err
		}
		var serverAddr string
		if err := c.request(ctx, "/get_low_load_server_addr", needNew, func(_ core.ClientContext, addr string) {
			serverAddr = addr
		}, true); err != nil {
			return err
		}
		if serverAddr == "" {
			return c.connectUser(ctx, fmt.Sprintf("%s://%s/game", proto, addr))
		}
		return c.connect(ctx, serverAddr, false)
	}
	if c.config.Mode == core.ModeServer {
		return c.connectSystem(ctx, fmt.Sprintf("%s://%s/system", proto, addr))
	}
	return errors.New("unknown client mode")
}

func (c *Client) connectSystem(ctx context.Context, url string) error {
	return c.connectChannel(ctx, c.system, url)
}

func (c *Client) connectUser(ctx context.Context, url string) error {
	return c.connectChannel(ctx, c.user, url)
}

func (c *Client) connectChannel(ctx context.Context, ch *channel, url string) error {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	if ch.closed {
		return errors.New("client is closed")
	}
	conn, err := transport.Dial(url, c.config.Codec)
	if err != nil {
		return err
	}
	readCtx, cancel := context.WithCancel(ctx)
	ch.conn = conn
	ch.cancel = cancel
	go c.readLoop(readCtx, ch)
	return nil
}

func (c *Client) Push(route string, data any) error {
	return c.push(route, data, false)
}

func (c *Client) push(route string, data any, isSystem bool) error {
	bytes, err := c.config.Codec.Marshal(data)
	if err != nil {
		return err
	}
	return c.send(&protocol.Message{ID: uuid.New().String(), Route: route, Type: protocol.MessageTypePush, Data: string(bytes)}, isSystem)
}

func (c *Client) RequestAsync(route string, data any, callback any) error {
	return c.requestAsync(route, data, callback, false)
}

func (c *Client) requestAsync(route string, data any, callback any, isSystem bool) error {
	if err := router.CheckHandler(callback, reflect.TypeFor[core.ClientContext]()); err != nil {
		return err
	}
	req := c.channel(isSystem).pending.Add(callback)
	go c.channel(isSystem).pending.AutoDelete(req)
	return c.sendRequest(req.ID, route, data, isSystem)
}

func (c *Client) Request(ctx context.Context, route string, data any, callback any) error {
	return c.request(ctx, route, data, callback, false)
}

func (c *Client) request(ctx context.Context, route string, data any, callback any, isSystem bool) error {
	if err := router.CheckHandler(callback, reflect.TypeFor[core.ClientContext]()); err != nil {
		return err
	}
	store := c.channel(isSystem).pending
	req := store.Add(callback)
	if err := c.sendRequest(req.ID, route, data, isSystem); err != nil {
		store.Delete(req.ID)
		return err
	}
	return store.Wait(ctx, req)
}

func (c *Client) sendRequest(id, route string, data any, isSystem bool) error {
	bytes, err := c.config.Codec.Marshal(data)
	if err != nil {
		return err
	}
	return c.send(&protocol.Message{ID: id, Route: route, Type: protocol.MessageTypeRequest, Data: string(bytes)}, isSystem)
}

func (c *Client) send(msg *protocol.Message, isSystem bool) error {
	ch := c.channel(isSystem)
	ch.lock.RLock()
	conn := ch.conn
	closed := ch.closed
	ch.lock.RUnlock()
	if closed {
		return errors.New("client is closed")
	}
	if conn == nil {
		return errors.New("client not connected")
	}
	return conn.Send(msg)
}

func (c *Client) Close() error {
	errUser := c.closeChannel(c.user)
	errSystem := c.closeChannel(c.system)
	if errUser != nil {
		return errUser
	}
	return errSystem
}

func (c *Client) closeChannel(ch *channel) error {
	ch.lock.Lock()
	if ch.closed {
		ch.lock.Unlock()
		return nil
	}
	ch.closed = true
	if ch.cancel != nil {
		ch.cancel()
	}
	conn := ch.conn
	ch.conn = nil
	ch.lock.Unlock()
	ch.pending.Close()
	if conn != nil {
		return conn.Close()
	}
	return nil
}

func (c *Client) channel(isSystem bool) *channel {
	if isSystem {
		return c.system
	}
	return c.user
}
