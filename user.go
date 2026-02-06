package feng

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type IUser interface {
	GetID() string
	GetRoom() IRoom
	JoinRoom(room IRoom)
	CreateAndJoinRoom() error
	GetContext() IServerContext
	// 获取额外的数据
	GetExtraData(key string) (any, bool)
	// 设置额外的数据
	SetExtraData(key string, value any)
	// 获取当前页码
	GetPage() int
	// 设置当前页码
	SetPage(page int)

	Push(route string, data any) error
	Request(ctx context.Context, route string, data any, callback any) error
	RequestAsync(route string, data any, callback any) error
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
	// 页码
	page int
	// 链接
	conn *websocket.Conn
	// 是否是系统用户
	isSys bool
	// 额外的数据
	extraData sync.Map
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

func (u *user) GetExtraData(key string) (any, bool) {
	return u.extraData.Load(key)
}

func (u *user) SetExtraData(key string, value any) {
	u.extraData.Store(key, value)
}

func (u *user) GetPage() int {
	return u.page
}

func (u *user) SetPage(page int) {
	u.page = page
}
