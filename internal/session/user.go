package session

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/google/uuid"
	"github.com/zmhuanf/feng/internal/core"
	"github.com/zmhuanf/feng/internal/pending"
	"github.com/zmhuanf/feng/internal/protocol"
	"github.com/zmhuanf/feng/internal/router"
)

type Sender interface {
	Send(*protocol.Message) error
}

type User struct {
	id      string
	ctx     core.ServerContext
	server  core.Server
	rooms   *RoomStoreImpl
	pending *pending.Store
	sender  Sender
	room    *Room
	page    int
	lock    sync.RWMutex
	extra   sync.Map
}

func NewUser(server core.Server, ctx core.ServerContext, rooms *RoomStoreImpl, pending *pending.Store, sender Sender) *User {
	return &User{
		id:      uuid.New().String(),
		ctx:     ctx,
		server:  server,
		rooms:   rooms,
		pending: pending,
		sender:  sender,
	}
}

func (u *User) ID() string { return u.id }

func (u *User) Room() core.Room {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.room
}

func (u *User) JoinRoom(room core.Room) error {
	r, ok := room.(*Room)
	if !ok {
		return fmt.Errorf("invalid room type")
	}
	if err := r.AddUser(u); err != nil {
		return err
	}
	u.setRoom(r)
	u.setPage(r.Page())
	return nil
}

func (u *User) CreateAndJoinRoom() error {
	return u.JoinRoom(u.rooms.CreateRoom())
}

func (u *User) LeaveRoom() error {
	u.lock.RLock()
	room := u.room
	u.lock.RUnlock()
	if room == nil {
		return fmt.Errorf("user %s not in any room", u.id)
	}
	return room.RemoveUser(u)
}

func (u *User) Context() core.ServerContext { return u.ctx }

func (u *User) ExtraData(key string) (any, bool) { return u.extra.Load(key) }

func (u *User) SetExtraData(key string, value any) { u.extra.Store(key, value) }

func (u *User) Page() int { return u.page }

func (u *User) Push(route string, data any) error {
	bytes, err := u.server.Config().Codec.Marshal(data)
	if err != nil {
		return err
	}
	return u.sender.Send(&protocol.Message{ID: uuid.New().String(), Route: route, Type: protocol.MessageTypePush, Data: string(bytes)})
}

func (u *User) RequestAsync(route string, data any, callback any) error {
	if err := router.CheckHandler(callback, reflect.TypeFor[core.ServerContext]()); err != nil {
		return err
	}
	req := u.pending.Add(callback)
	go u.pending.AutoDelete(req)
	return u.sendRequest(req.ID, route, data)
}

func (u *User) Request(ctx context.Context, route string, data any, callback any) error {
	if err := router.CheckHandler(callback, reflect.TypeFor[core.ServerContext]()); err != nil {
		return err
	}
	req := u.pending.Add(callback)
	if err := u.sendRequest(req.ID, route, data); err != nil {
		u.pending.Delete(req.ID)
		return err
	}
	return u.pending.Wait(ctx, req)
}

func (u *User) sendRequest(id, route string, data any) error {
	bytes, err := u.server.Config().Codec.Marshal(data)
	if err != nil {
		return err
	}
	return u.sender.Send(&protocol.Message{ID: id, Route: route, Type: protocol.MessageTypeRequest, Data: string(bytes)})
}

func (u *User) setRoom(room *Room) {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.room = room
}

func (u *User) setPage(page int) { u.page = page }
