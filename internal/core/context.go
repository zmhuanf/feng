package core

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
)

type Server interface {
	Config() *ServerConfig
	Handle(route string, handler any) error
	Use(route string, middleware any) error
	ListenAndServe(context.Context) error
	Stop(context.Context) error
	Room(id string) (Room, error)
	Rooms() []Room
	RoomsByPage(page int) []Room
	User(id string) (User, error)
	Users() []User
	UsersByPage(page int) []User
	Gin() *gin.Engine
}

type Client interface {
	Config() *ClientConfig
	Handle(route string, handler any) error
	Use(route string, middleware any) error
	Connect(context.Context) error
	Push(route string, data any) error
	RequestAsync(route string, data any, callback any) error
	Request(context.Context, string, any, any) error
	Close() error
}

type Room interface {
	ID() string
	RemoveUser(User) error
	User(id string) (User, error)
	Users() []User
	UserCount() int
	Page() int
}

type User interface {
	ID() string
	Room() Room
	JoinRoom(Room) error
	CreateAndJoinRoom() error
	LeaveRoom() error
	Context() ServerContext
	ExtraData(key string) (any, bool)
	SetExtraData(key string, value any)
	Page() int
	Push(route string, data any) error
	Request(context.Context, string, any, any) error
	RequestAsync(route string, data any, callback any) error
}

type ServerContext interface {
	Room() Room
	User() User
	Server() Server
	Get(key string) (any, bool)
	Set(key string, value any)
	GinContext() *gin.Context
}

type ClientContext interface {
	Client() Client
}

type BaseServerContext struct {
	data   sync.Map
	room   Room
	user   User
	server Server
	ginCtx *gin.Context
}

func NewServerContext(server Server, ginCtx *gin.Context) *BaseServerContext {
	return &BaseServerContext{server: server, ginCtx: ginCtx}
}

func (c *BaseServerContext) Bind(room Room, user User) {
	c.room = room
	c.user = user
}

func (c *BaseServerContext) Room() Room { return c.room }

func (c *BaseServerContext) User() User { return c.user }

func (c *BaseServerContext) Server() Server { return c.server }

func (c *BaseServerContext) Get(key string) (any, bool) { return c.data.Load(key) }

func (c *BaseServerContext) Set(key string, value any) { c.data.Store(key, value) }

func (c *BaseServerContext) GinContext() *gin.Context { return c.ginCtx }

type BaseClientContext struct {
	client Client
}

func NewClientContext(client Client) ClientContext {
	return &BaseClientContext{client: client}
}

func (c *BaseClientContext) Client() Client { return c.client }
