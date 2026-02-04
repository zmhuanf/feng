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
	JoinRoom(IRoom)
	CreateAndJoinRoom() error
	GetContext() IServerContext

	Push(string, any) error
	Request(context.Context, string, any, any) error
	RequestAsync(string, any, any) error
}

type user struct {
	// 用户ID
	id string
	// 服务
	server *server
	// 上下文
	ctx IServerContext
	// 房间
	room *room
	// 链接
	conn *websocket.Conn
	// 是否是系统用户
	isSys bool
}

func newUser(s *server, ctx IServerContext, room *room, conn *websocket.Conn, isSys bool) *user {
	u := &user{
		id:     uuid.New().String(),
		ctx:    ctx,
		server: s,
		room:   room,
		conn:   conn,
		isSys:  isSys,
	}
	s.addUser(u, isSys)
	return u
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
		timer := time.NewTimer(u.server.config.Timeout)
		defer timer.Stop()

		select {
		case <-timer.C:
			u.server.deleteResponse(id, u.isSys)
		case <-ch:
		}
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
	// 超时删除
	timer := time.NewTimer(u.server.config.Timeout)
	defer timer.Stop()
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
		u.server.deleteResponse(id, u.isSys)
		close(ch)
		return ctx.Err()
	case <-timer.C:
		u.server.deleteResponse(id, u.isSys)
		close(ch)
		return fmt.Errorf("request timeout")
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

func (u *user) JoinRoom(r IRoom) {
	r.AddUser(u)
	u.room = r.(*room)
}

func (u *user) GetContext() IServerContext {
	return u.ctx
}

func (u *user) CreateAndJoinRoom() error {
	newRoom := newRoom(u.server, u.isSys)
	err := newRoom.AddUser(u)
	if err != nil {
		return err
	}
	u.room = newRoom
	return nil
}
