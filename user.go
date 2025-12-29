package feng

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type IUser interface {
	GetID() string
	GetRoom() IRoom
	SetRoom(room IRoom)
	GetContext() IServerContext
	SetContext(ctx IServerContext)

	Push(string, any) error
	Request(context.Context, string, any, any) error
	RequestAsync(string, any, any) error
}

type user struct {
	// 用户ID
	id string
	// 上下文
	ctx IServerContext
	// 房间
	room *room
	// 服务
	server *server
	// 链接
	conn *websocket.Conn
	// 是否是系统用户
	isSys bool
}

func (u *user) send(res *message) error {
	resByte, err := u.server.config.Codec.Marshal(res)
	if err != nil {
		return err
	}
	return u.conn.WriteMessage(websocket.TextMessage, resByte)
}

func (u *user) RequestAsync(route string, data any, fn any) error {
	// 检查函数签名
	err := checkFuncType(fn, true)
	if err != nil {
		return err
	}
	id := uuid.New().String()
	ch := make(chan chanData)
	u.server.addResponse(id, &response{
		fn: fn,
		ch: ch,
	}, u.isSys)
	// 配置一个额外的删除，防止内存泄漏
	go func() {
		<-time.After(u.server.config.Timeout)
		close(ch)
		u.server.deleteResponse(id, u.isSys)
	}()
	dataBytes, err := u.server.config.Codec.Marshal(data)
	if err != nil {
		return err
	}
	err = u.send(&message{
		ID:    id,
		Route: route,
		Type:  messageTypeRequest,
		Data:  string(dataBytes),
	})
	if err != nil {
		return err
	}
	return nil
}

func (u *user) Request(ctx context.Context, route string, data any, fn any) error {
	// 检查函数签名
	err := checkFuncType(fn, true)
	if err != nil {
		return err
	}
	id := uuid.New().String()
	ch := make(chan chanData)
	u.server.addResponse(id, &response{
		fn: fn,
		ch: ch,
	}, u.isSys)
	// 配置一个额外的删除，防止内存泄漏
	go func() {
		<-time.After(u.server.config.Timeout)
		close(ch)
		u.server.deleteResponse(id, u.isSys)
	}()
	dataBytes, err := u.server.config.Codec.Marshal(data)
	if err != nil {
		return err
	}
	err = u.send(&message{
		ID:    id,
		Route: route,
		Type:  messageTypeRequest,
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
		mutex := &u.server.userData.responsesLock
		responses := u.server.userData.responses
		if u.isSys {
			mutex = &u.server.sysData.responsesLock
			responses = u.server.sysData.responses
		}
		mutex.Lock()
		delete(responses, id)
		mutex.Unlock()
		return ctx.Err()
	}
}

func (u *user) Push(route string, data any) error {
	dataBytes, err := u.server.config.Codec.Marshal(data)
	if err != nil {
		return err
	}
	err = u.send(&message{
		ID:    uuid.New().String(),
		Route: route,
		Type:  messageTypePush,
		Data:  string(dataBytes),
	})
	return err
}

func (u *user) GetID() string {
	return u.id
}

func (u *user) GetRoom() IRoom {
	return u.room
}

func (u *user) SetRoom(r IRoom) {
	u.room = r.(*room)
}

func (u *user) GetContext() IServerContext {
	return u.ctx
}

func (u *user) SetContext(ctx IServerContext) {
	u.ctx = ctx
}
