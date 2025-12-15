package feng

import (
	"sync"
)

type IServerContext interface {
	// 获取房间
	GetRoom() IRoom
	// 获取用户
	GetUser() IUser
	// 获取服务器
	GetServer() IServer
	// 获取上下文值
	Get(key string) (any, bool)
	// 设置上下文值
	Set(key string, value any)
}

type fServerContext struct {
	data   sync.Map
	room   IRoom
	user   IUser
	server IServer
}

func newServerContext(room IRoom, user IUser, server IServer) IServerContext {
	return &fServerContext{
		room:   room,
		user:   user,
		server: server,
	}
}

func (f *fServerContext) GetRoom() IRoom {
	return f.room
}

func (f *fServerContext) GetUser() IUser {
	return f.user
}

func (f *fServerContext) GetServer() IServer {
	return f.server
}

func (f *fServerContext) Get(key string) (any, bool) {
	return f.data.Load(key)
}

func (f *fServerContext) Set(key string, value any) {
	f.data.Store(key, value)
}

type IClientContext interface {
	GetClient() IClient
}

type fClientContext struct {
	client IClient
}

func (c *fClientContext) GetClient() IClient {
	return c.client
}

func newClientContext(client IClient) IClientContext {
	return &fClientContext{
		client: client,
	}
}
