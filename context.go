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
	data sync.Map
}

func newContext() IContext {
	return &fContext{}
}

func (f *fContext) GetRoom() IRoom {
	room, ok := f.data.Load("room")
	if !ok {
		panic("room not set")
	}
	return room.(IRoom)
}

func (f *fContext) GetUser() IUser {
	user, ok := f.data.Load("user")
	if !ok {
		panic("user not set")
	}
	return user.(IUser)
}

func (f *fContext) GetServer() IServer {
	server, ok := f.data.Load("server")
	if !ok {
		panic("server not set")
	}
	return server.(IServer)
}

func (f *fContext) Get(key string) (any, bool) {
	return f.data.Load(key)
}

func (f *fContext) Set(key string, value any) {
	f.data.Store(key, value)
}
