package feng

import (
	"sync"
)

type IContext interface {
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

type fContext struct {
	data   sync.Map
	room   IRoom
	user   IUser
	server IServer
}

func newContext(room IRoom, user IUser, server IServer) IContext {
	return &fContext{
		room:   room,
		user:   user,
		server: server,
	}
}

func (f *fContext) GetRoom() IRoom {
	return f.room
}

func (f *fContext) GetUser() IUser {
	return f.user
}

func (f *fContext) GetServer() IServer {
	return f.server
}

func (f *fContext) Get(key string) (any, bool) {
	return f.data.Load(key)
}

func (f *fContext) Set(key string, value any) {
	f.data.Store(key, value)
}
