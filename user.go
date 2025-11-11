package feng

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type IUser interface {
	Push(string, any) error
	Request(context.Context, string, any, any) error
	RequestAsync(string, any, any) error
}

type user struct {
	// 用户ID
	id string
	// 房间
	room *room
	// 服务
	server *server
	// 链接
	conn *websocket.Conn
}

func (u *user) send(res *request) error {
	resByte, err := u.server.config.Codec.Marshal(res)
	if err != nil {
		return err
	}
	return u.conn.WriteMessage(websocket.TextMessage, resByte)
}

func (u *user) RequestAsync(route string, data any, fn any) error {
	id := uuid.New().String()
	ch := make(chan chanData)
	u.server.addResponse(id, &response{
		fn: fn,
		ch: ch,
	})
	// 配置一个额外的删除，防止内存泄漏
	go func() {
		<-time.After(u.server.config.Timeout)
		close(ch)
		u.server.deleteResponse(id)
	}()
	dataBytes, err := u.server.config.Codec.Marshal(data)
	if err != nil {
		return err
	}
	err = u.send(&request{
		ID:    id,
		Route: route,
		Type:  requestTypeRequest,
		Data:  dataBytes,
	})
	if err != nil {
		return err
	}
	return nil
}

func (u *user) Request(ctx context.Context, route string, data any, fn any) error {
	id := uuid.New().String()
	ch := make(chan chanData)
	u.server.addResponse(id, &response{
		fn: fn,
		ch: ch,
	})
	// 配置一个额外的删除，防止内存泄漏
	go func() {
		<-time.After(u.server.config.Timeout)
		close(ch)
		u.server.deleteResponse(id)
	}()
	dataBytes, err := u.server.config.Codec.Marshal(data)
	if err != nil {
		return err
	}
	err = u.send(&request{
		ID:    id,
		Route: route,
		Type:  requestTypeRequest,
		Data:  dataBytes,
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
		u.server.responsesLock.Lock()
		delete(u.server.responses, id)
		u.server.responsesLock.Unlock()
		return ctx.Err()
	}
}

func (u *user) Push(route string, data any) error {
	dataBytes, err := u.server.config.Codec.Marshal(data)
	if err != nil {
		return err
	}
	err = u.send(&request{
		ID:    uuid.New().String(),
		Route: route,
		Type:  requestTypePush,
		Data:  dataBytes,
	})
	return err
}
